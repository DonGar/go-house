package properties

import (
	"fmt"
	"github.com/DonGar/go-house/engine/conditions"
	"github.com/DonGar/go-house/status"
	"log"
	"reflect"
)

type conditionValue struct {
	condition conditions.Condition
	result    bool
	value     interface{}
}

// The base type all rules should compose with.
type Property struct {
	status      *status.Status
	name        string
	targetUrl   string
	conditions  []conditionValue
	hasDefault  bool
	defautValue interface{}
	stop        chan bool
}

func NewProperty(
	status *status.Status,
	name string,
	body *status.Status) (*Property, error) {

	// Look up the target URL.
	targetUrl, e := body.GetString("status://target")
	if e != nil {
		return nil, fmt.Errorf("No 'target' on property: %s", name)
	}

	// look up the conditionValues.
	valuesRaw, _, e := body.Get("status://values")
	if e != nil {
		return nil, fmt.Errorf("No 'values' on property: %s", name)
	}
	conditionValues, e := parseConditionValues(status, name, valuesRaw)
	if e != nil {
		return nil, e
	}

	// Look up the optional default value.
	defaultValue, e := body.GetString("status://default")
	hasDefault := e == nil

	result := &Property{
		status,
		name,
		targetUrl,
		conditionValues,
		hasDefault,
		defaultValue,
		make(chan bool)}

	result.start()
	return result, nil
}

func parseConditionValues(s *status.Status, name string, valuesRaw interface{}) ([]conditionValue, error) {
	conditionValues := []conditionValue{}

	valuesArray, ok := valuesRaw.([]interface{})
	if !ok {
		return nil, fmt.Errorf("'values' not an array ([]) on property: %s", name)
	}
	for i, valuesElementRaw := range valuesArray {
		valuesElement, ok := valuesElementRaw.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf(
				"Invalid value %d (%#v) on property: %s",
				i, valuesElementRaw, name)
		}

		conditionBody, ok := valuesElement["condition"]
		if !ok {
			return nil, fmt.Errorf(
				"No 'condition' on value %d (%#v) on property: %s",
				i, valuesElement, name)
		}

		value, ok := valuesElement["value"]
		if !ok {
			return nil, fmt.Errorf(
				"No 'value' on value %d (%#v) on property: %s",
				i, valuesElement, name)
		}

		conditionBodyStatus := &status.Status{}
		conditionBodyStatus.Set("status://", conditionBody, status.UNCHECKED_REVISION)

		condition, e := conditions.NewCondition(s, conditionBodyStatus)
		if e != nil {
			return nil, fmt.Errorf(
				"Invalid 'condition' on value %d (%#v) on property: %s: %s",
				i, valuesElement, name, e.Error())
		}

		// We successfully parsed a value element. Remember it.
		conditionValues = append(conditionValues, conditionValue{condition, false, value})
	}

	return conditionValues, nil
}

func (p *Property) start() {
	log.Printf("Start property: %s", p.name) // url)
	go p.watchConditionResults()
}

func (p *Property) Stop() {
	// Stop the conditions BEFORE stopping our listener, or we will deadlock
	// if they send a shutdown result.
	for _, c := range p.conditions {
		c.condition.Stop()
	}

	p.stop <- true
	<-p.stop
}

func (p *Property) updateTarget() {

	setTarget := func(value interface{}) {
		e := p.status.Set(p.targetUrl, value, status.UNCHECKED_REVISION)
		if e == nil {
			log.Printf("Property (%s: %s) updated: %#v", p.name, p.targetUrl, value)
		} else {
			log.Printf("Property (%s) update failed: %s", p.name, e.Error())
		}
	}

	for _, c := range p.conditions {
		if c.result {
			setTarget(c.value)
			return
		}
	}

	if p.hasDefault {
		setTarget(p.defautValue)
	}
}

func (p *Property) watchConditionResults() {
	// Set with default, if present.
	p.updateTarget()

	// We listen on all condition result channels.
	channels := make([]reflect.SelectCase, len(p.conditions)+1)
	for i, c := range p.conditions {
		channels[i] = reflect.SelectCase{
			Dir: reflect.SelectRecv, Chan: reflect.ValueOf(c.condition.Result())}
	}

	// We also listen to the stop channel.
	channels[len(channels)-1] = reflect.SelectCase{
		Dir: reflect.SelectRecv, Chan: reflect.ValueOf(p.stop)}

	for {
		receivedOnChannel, value, _ := reflect.Select(channels)

		if receivedOnChannel < len(p.conditions) {
			// One of our conditions updated it's result. Update ours accordingly.
			p.conditions[receivedOnChannel].result = value.Bool()
			p.updateTarget()
		} else {
			// If it's after the conditions, it's the stop channel.
			log.Printf("Stop property: %s", p.name) // url)
			p.stop <- true
			return
		}
	}
}
