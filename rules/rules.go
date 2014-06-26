package rules

import (
	// "github.com/cpucycle/astrotime"
	"fmt"
	"github.com/DonGar/go-house/status"
	"time"
)

type rule interface {
	Stop() error
	Revision() int
}

type newRule func(base base) (r rule, e error)

type base struct {
	manager  *Manager
	name     string         // Name of this rule.
	revision int            // Status revision of the rule definition.
	body     *status.Status // Substatus of the rule definition.
}

func (b *base) Revision() int {
	return b.revision
}

func (b *base) Stop() error {
	return nil
}

// Called when the rule decides to fire it's action.
func (b *base) fire() error {
	fmt.Println("Fire: ", b.name, " ", time.Now())
	return nil
}

type conditionalRule struct {
	base
}

func newConditionalRule(base base) (r rule, e error) {
	return nil, nil
}

type statusRule struct {
	base
}

func newStatusRule(base base) (r rule, e error) {
	return nil, nil
}
