package rules

import (
	"fmt"
	"github.com/DonGar/go-house/rules/conditions"
	"github.com/DonGar/go-house/status"
	"log"
)

// An interface for firing actions that should be provided to all rules.
type actionHelper func(action *status.Status)

// The base type all rules should compose with.
type rule struct {
	actionHelper actionHelper
	name         string // Name of this rule.
	revision     int    // Status revision of the rule definition.
	condition    conditions.Condition
	action       *status.Status // Substatus of the rule's action.
	stop         chan bool
}

func newRule(
	status *status.Status,
	actionHelper actionHelper,
	name string,
	revision int,
	ruleBody *status.Status) (*rule, error) {

	// Find the sub-expression contents.
	conditionBody, _, e := ruleBody.GetSubStatus("status://condition")
	if e != nil {
		return nil, fmt.Errorf("No 'condition' section.")
	}

	actionBody, _, e := ruleBody.GetSubStatus("status://action")
	if e != nil {
		return nil, fmt.Errorf("No 'action' section.")
	}

	// Create the condition (last, because it needs Stopping on failure).
	condition, e := conditions.NewCondition(status, conditionBody)
	if e != nil {
		return nil, e
	}

	result := &rule{
		actionHelper,
		name,
		revision,
		condition,
		actionBody,
		make(chan bool)}

	result.start()

	return result, nil
}

func (r *rule) start() {
	go r.watchConditionResults()
}

func (r *rule) Stop() {
	r.condition.Stop()
	r.stop <- true
	<-r.stop
}

func (r *rule) watchConditionResults() {
	for {
		select {
		case condValue := <-r.condition.Result():
			if condValue {
				log.Println("Firing rule: ", r.name)
				r.actionHelper(r.action)
			}
		case <-r.stop:
			r.stop <- true
			return
		}
	}
}
