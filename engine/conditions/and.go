package conditions

import (
	"fmt"
	"github.com/DonGar/go-house/status"
	"reflect"
)

type conditionValue struct {
	condition Condition
	result    bool
}

type andCondition struct {
	base

	currentResult bool
	conditions    []conditionValue
}

func newAndCondition(s *status.Status, body *status.Status) (*andCondition, error) {

	// look up the conditionValues.
	valuesRaw, _, e := body.Get("status://conditions")
	if e != nil {
		return nil, fmt.Errorf("No 'values' on property: %s")
	}
	conditionValues, e := parseConditionValues(s, valuesRaw)
	if e != nil {
		return nil, e
	}

	// Create our condition.
	c := &andCondition{base{s, make(chan bool), make(chan bool)}, false, conditionValues}

	c.start()
	return c, nil
}

func parseConditionValues(s *status.Status, valuesRaw interface{}) ([]conditionValue, error) {
	conditionValues := []conditionValue{}

	valuesArray, ok := valuesRaw.([]interface{})
	if !ok {
		return nil, fmt.Errorf("'values' not an array ([]) on condition")
	}
	for i, valuesElementRaw := range valuesArray {
		valuesElement, ok := valuesElementRaw.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("Invalid value %d (%#v) on condition", i, valuesElementRaw)
		}

		conditionBody, ok := valuesElement["condition"]
		if !ok {
			return nil, fmt.Errorf(
				"No 'condition' on value %d (%#v) on condition", i, valuesElement)
		}

		conditionBodyStatus := &status.Status{}
		conditionBodyStatus.Set("status://", conditionBody, status.UNCHECKED_REVISION)

		condition, e := NewCondition(s, conditionBodyStatus)
		if e != nil {
			return nil, fmt.Errorf(
				"Invalid 'condition' on value %d (%#v) on condition: %s",
				i, valuesElement, e.Error())
		}

		// We successfully parsed a value element. Remember it.
		conditionValues = append(conditionValues, conditionValue{condition, false})
	}

	return conditionValues, nil
}

func (c *andCondition) start() {
	// Start it's goroutine.
	go c.handle()
}

func (c *andCondition) Stop() {
	// Shut down inner conditions before stopping ourselves. This means we
	// can react to final result updates from them, and avoid race conditions.
	for _, condValue := range c.conditions {
		condValue.condition.Stop()
	}

	c.base.Stop()
}

func (c *andCondition) updateTarget() {
	newResult := true

	for _, condValue := range c.conditions {
		newResult = newResult && condValue.result
	}

	if newResult != c.currentResult {
		c.currentResult = newResult
		c.result <- newResult
	}
}

func (c *andCondition) handle() {
	// We listen on all condition result channels.
	channels := make([]reflect.SelectCase, len(c.conditions)+1)
	for i, c := range c.conditions {
		channels[i] = reflect.SelectCase{
			Dir: reflect.SelectRecv, Chan: reflect.ValueOf(c.condition.Result())}
	}

	// We also listen to the stop channel.
	channels[len(channels)-1] = reflect.SelectCase{
		Dir: reflect.SelectRecv, Chan: reflect.ValueOf(c.stop)}

	for {
		receivedOnChannel, value, _ := reflect.Select(channels)

		if receivedOnChannel < len(c.conditions) {
			// One of our conditions updated it's result. Update ours accordingly.
			c.conditions[receivedOnChannel].result = value.Bool()
			c.updateTarget()
		} else {
			// If it's after the conditions, it's the stop channel.
			c.stop <- true
			return
		}
	}
}
