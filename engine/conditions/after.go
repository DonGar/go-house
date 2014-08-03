package conditions

import (
	"fmt"
	"github.com/DonGar/go-house/status"
	"time"
)

type afterCondition struct {
	base

	currentResult   bool
	conditionResult bool
	condition       Condition
	delay           time.Duration
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

	delayStr, e := body.GetString("status://delay")
	if e != nil {
		return nil, e
	}

	delay, e := time.ParseDuration(delayStr)
	if e != nil {
		return nil, e
	}

	// Create our condition.
	c := &afterCondition{base{s, make(chan bool), make(chan bool)}, false, false, condition, delay}

	c.start()
	return c, nil
}

func (c *afterCondition) start() {
	// Start it's goroutine.
	go c.handle()
}

func (c *afterCondition) Stop() {
	// Shut down inner condition before stopping ourselves. This means we
	// can react to final result updates from them, and avoid deadlocks.
	c.condition.Stop()
	c.base.Stop()
}

func (c *afterCondition) updateTarget(newResult bool) {
	if c.currentResult != newResult {
		c.currentResult = newResult
		c.result <- newResult
	}
}

func (c *afterCondition) handle() {
	// Create the timer with a really long timeout, then stop it.
	// We'll reset, when we're ready to realy start it.
	timer := time.NewTimer(time.Hour)
	timer.Stop()

	for {
		select {
		case condValue := <-c.condition.Result():
			if condValue != c.conditionResult {
				c.conditionResult = condValue

				if condValue {
					timer.Reset(c.delay)
				} else {
					timer.Stop()
					c.updateTarget(false)
				}
			}

		case <-timer.C:
			c.updateTarget(true)

		case <-c.stop:
			c.stop <- true
			return
		}
	}
}
