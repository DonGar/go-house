package rules

import (
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
	actionHelper actionHelper,
	name string,
	revision int,
	action *status.Status,
	condition conditions.Condition) *rule {

	result := &rule{
		actionHelper,
		name,
		revision,
		condition,
		action,
		make(chan bool)}

	go result.watchConditionResults()

	return result
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
