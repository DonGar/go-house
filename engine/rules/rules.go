package rules

import (
	"fmt"
	"github.com/DonGar/go-house/engine/actions"
	"github.com/DonGar/go-house/engine/conditions"
	"github.com/DonGar/go-house/status"
	"log"
)

type Rule struct {
	status          *status.Status
	actionRegistrar actions.ActionRegistrar
	name            string // name of this rule.
	condition       conditions.Condition
	action          *status.Status // Substatus of the rule's action.
	stop            chan bool
}

func NewRule(
	status *status.Status,
	actionRegistrar actions.ActionRegistrar,
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
		status,
		actionRegistrar,
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
	r.stop <- true
	<-r.stop
}

func (r *Rule) watchConditionResults() {
	for {
		select {
		case condValue := <-r.condition.Result():
			if condValue {
				log.Println("Firing rule: ", r.name)
				e := actions.FireAction(r.status, r.actionRegistrar, r.action)
				if e != nil {
					log.Println("Fire Error: ", e)
				}
			}
		case <-r.stop:
			r.condition.Stop()
			log.Printf("Stop rule: %s", r.name) // url)
			r.stop <- true
			return
		}
	}
}
