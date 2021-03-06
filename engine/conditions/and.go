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
	conditions []conditionValue
}

func newAndCondition(s *status.Status, body *status.Status) (*andCondition, error) {

	// look up the conditionValues.
	valuesRaw, _, e := body.Get("status://conditions")
	if e != nil {
		return nil, e
	}
	conditionValues, e := parseConditionValues(s, valuesRaw)
	if e != nil {
		return nil, e
	}

	// Create our condition.
	c := &andCondition{newBase(s), conditionValues}

	c.start()
	return c, nil
}

func parseConditionValues(s *status.Status, valuesRaw interface{}) ([]conditionValue, error) {

	valuesArray, ok := valuesRaw.([]interface{})
	if !ok {
		return nil, fmt.Errorf("'values' not an array ([]) on condition")
	}

	conditionValues := make([]conditionValue, len(valuesArray))
	for i, subConditionBody := range valuesArray {

		conditionBodyStatus := &status.Status{}
		conditionBodyStatus.Set("status://", subConditionBody, status.UNCHECKED_REVISION)

		condition, e := NewCondition(s, conditionBodyStatus)
		if e != nil {
			return nil, fmt.Errorf("And condition: %d (%#v): %s", i, subConditionBody, e.Error())
		}

		// We successfully parsed a value element. Remember it.
		conditionValues[i] = conditionValue{condition, false}
	}

	return conditionValues, nil
}

func (c *andCondition) start() {
	// Start it's goroutine.
	go c.Handler()
}

func (c *andCondition) Stop() {
	// Shut down inner conditions before stopping ourselves. This means we
	// can react to final result updates from them, and avoid race conditions.
	for _, condValue := range c.conditions {
		condValue.condition.Stop()
	}

	c.base.Stop()
}

func (c *andCondition) currentValue() bool {
	newResult := true

	for _, condValue := range c.conditions {
		newResult = newResult && condValue.result
	}

	return newResult
}

func (c *andCondition) Handler() {
	// Read initial state from all conditions, and send out our initial state.
	for i := range c.conditions {
		c.conditions[i].result = <-c.conditions[i].condition.Result()
	}
	c.sendResult(c.currentValue())

	// Build reflect list of all condition channels to listen for.
	channels := make([]reflect.SelectCase, len(c.conditions)+1)
	for i := range c.conditions {
		channels[i] = reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(c.conditions[i].condition.Result()),
		}
	}

	// We also listen to the stop channel.
	channels[len(channels)-1] = reflect.SelectCase{
		Dir: reflect.SelectRecv, Chan: reflect.ValueOf(c.StopChan)}

	for {
		receivedOnChannel, value, _ := reflect.Select(channels)

		if receivedOnChannel < len(c.conditions) {
			// One of our conditions updated it's result. Update ours accordingly.
			c.conditions[receivedOnChannel].result = value.Bool()
			c.sendResult(c.currentValue())
		} else {
			// If it's after the conditions, it's the stop channel.
			c.StopChan <- true
			return
		}
	}
}
