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
	repeat    time.Duration
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

	var repeat time.Duration
	repeatStr, _, e := body.GetString("status://repeat")
	if e == nil {
		repeat, e = time.ParseDuration(repeatStr)
		if e != nil {
			return nil, e
		}
	}

	// Create our condition.
	c := &afterCondition{newBase(s), condition, delay, repeat}

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

	c.sendResult(false)

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
			// If we are repeating, we force a switch to false so true registers.
			c.sendResult(false)
			c.sendResult(true)

			// If we repeat after being true,
			if c.repeat != 0 {
				timer.Reset(c.repeat)
			}

		case <-c.StopChan:
			c.StopChan <- true
			return
		}
	}
}
