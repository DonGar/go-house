package rules

import (
	"fmt"
	"github.com/DonGar/go-house/engine/conditions"
	"github.com/DonGar/go-house/status"
	"log"
)

// An interface for firing actions that should be provided to all rules.
type actionHelper func(action *status.Status)

// The base type all rules should compose with.
type Rule struct {
	actionHelper actionHelper
	name         string // name of this rule.
	condition    conditions.Condition
	action       *status.Status // Substatus of the rule's action.
	stop         chan bool
}

func NewRule(
	status *status.Status,
	actionHelper actionHelper,
	name string,
	ruleBody *status.Status) (*Rule, error) {

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

	result := &Rule{
		actionHelper,
		name,
		condition,
		actionBody,
		make(chan bool)}

	result.start()

	return result, nil
}

func (r *Rule) start() {
	log.Printf("Start rule: %s", r.name) // url)
	go r.watchConditionResults()
}

func (r *Rule) Stop() {
	r.condition.Stop()
	r.stop <- true
	<-r.stop
	log.Printf("Stop rule: %s", r.name) // url)
}

func (r *Rule) watchConditionResults() {
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
