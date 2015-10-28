package rules

import (
	"fmt"
	"github.com/DonGar/go-house/engine/actions"
	"github.com/DonGar/go-house/engine/conditions"
	"github.com/DonGar/go-house/status"
	"github.com/DonGar/go-house/stoppable"
	"log"
)

type Rule struct {
	status        *status.Status
	actionManager *actions.Manager
	name          string // name of this rule.
	condition     conditions.Condition
	actionOn      *status.Status // Substatus of the rule's action.
	actionOff     *status.Status // Substatus of the rule's action.
	stoppable.Base
}

func NewRule(
	status *status.Status,
	actionManager *actions.Manager,
	name string,
	ruleBody *status.Status) (*Rule, error) {

	// Find the sub-expression contents.
	conditionBody, _, e := ruleBody.GetSubStatus("status://condition")
	if e != nil {
		return nil, fmt.Errorf("No 'condition' section.")
	}

	actionOn, _, _ := ruleBody.GetSubStatus("status://on")
	actionOff, _, _ := ruleBody.GetSubStatus("status://off")

	if actionOn == nil && actionOff == nil {
		return nil, fmt.Errorf("No On or Off action.")
	}

	// Create the condition (last, because it needs Stopping on failure).
	condition, e := conditions.NewCondition(status, conditionBody)
	if e != nil {
		return nil, e
	}

	result := &Rule{
		status,
		actionManager,
		name,
		condition,
		actionOn,
		actionOff,
		stoppable.NewBase()}

	result.start()
	return result, nil
}

func (r *Rule) start() {
	log.Printf("Start rule: %s", r.name) // url)
	go r.Handler()

}

func (r *Rule) Handler() {
	for {
		select {
		case condValue := <-r.condition.Result():

			if condValue {
				if r.actionOn != nil {
					log.Println("Firing rule On: ", r.name)
					r.actionManager.FireAction(r.status, r.actionOn)
				}
			} else {
				if r.actionOff != nil {
					log.Println("Firing rule Off: ", r.name)
					r.actionManager.FireAction(r.status, r.actionOff)
				}
			}

		case <-r.StopChan:
			r.condition.Stop()
			log.Printf("Stop rule: %s", r.name) // url)
			r.StopChan <- true
			return
		}
	}
}
