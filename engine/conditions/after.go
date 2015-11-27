package conditions

import (
	"fmt"
	"github.com/DonGar/go-house/status"
	"time"
)

type afterCondition struct {
	base

	condition Condition
	delay     time.Duration
}

func newAfterCondition(s *status.Status, body *status.Status) (*afterCondition, error) {

	// look up the conditionValues.
	subConditionBody, _, e := body.Get("status://condition")
	if e != nil {
		return nil, e
	}

	conditionBodyStatus := &status.Status{}
	conditionBodyStatus.Set("status://", subConditionBody, status.UNCHECKED_REVISION)

	condition, e := NewCondition(s, conditionBodyStatus)
	if e != nil {
		return nil, fmt.Errorf("After condition: (%#v): %s", subConditionBody, e.Error())
	}

	delayStr, _, e := body.GetString("status://delay")
	if e != nil {
		return nil, e
	}

	delay, e := time.ParseDuration(delayStr)
	if e != nil {
		return nil, e
	}

	// Create our condition.
	c := &afterCondition{newBase(s), condition, delay}

	c.start()
	return c, nil
}

func (c *afterCondition) start() {
	// Start it's goroutine.
	go c.Handler()
}

func (c *afterCondition) Stop() {
	// Shut down inner condition before stopping ourselves. This means we
	// can react to final result updates from them, and avoid deadlocks.
	c.condition.Stop()
	c.base.Stop()
}

func (c *afterCondition) Handler() {
	// Create the timer with a long timeout, then stop it.
	// We'll reset, when we're ready to really start it.
	timer := time.NewTimer(time.Hour)
	timer.Stop()

	for {
		select {
		case condValue := <-c.condition.Result():
			if condValue {
				timer.Reset(c.delay)
			} else {
				timer.Stop()
				c.sendResult(false)
			}

		case <-timer.C:
			c.sendResult(true)

		case <-c.StopChan:
			c.StopChan <- true
			return
		}
	}
}
