package conditions

import (
	"github.com/DonGar/go-house/status"
)

type constCondition struct {
	base

	result bool
}

func newConstCondition(s *status.Status, body *status.Status, result bool) (*constCondition, error) {
	// Create our condition.
	c := &constCondition{newBase(s), result}

	c.start()
	return c, nil
}

func (c *constCondition) start() {
	// Start it's goroutine.
	go c.Handler()
}

func (c *constCondition) Handler() {
	c.sendResult(c.result)
	c.base.Handler()
}
