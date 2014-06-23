package rules

import (
	// "github.com/cpucycle/astrotime"
	"fmt"
	"github.com/DonGar/go-house/status"
)

type Rule interface {
	Stop() error
	Revision() int
}

type NewRule func(
	s *status.Status,
	m *Manager,
	name string,
	body *status.Status) (r Rule, e error)

type base struct {
	revision int    // Status revision of the rule definition.
	target   string // Status url of target component (wildcards allowed)
	action   string
}

func (b *base) Revision() int {
	return b.revision
}

func (b *base) Stop() {
}

// Called when the rule decides to fire it's action.
func (b *base) fire() error {
	fmt.Println("Fire: ", b.target, " ", b.action)
	return nil
}

type periodicRule struct {
	base
	period string // how often is rule fired in seconds.
}

func newPeriodicRule(
	s *status.Status,
	m *Manager,
	name string,
	body *status.Status) (r Rule, e error) {
	return nil, nil
}

type conditionalRule struct {
	base
}

func newConditionalRule(
	s *status.Status,
	m *Manager,
	name string,
	body *status.Status) (r Rule, e error) {
	return nil, nil
}

type statusRule struct {
	base
}

func newStatusRule(
	s *status.Status,
	m *Manager,
	name string,
	body *status.Status) (r Rule, e error) {
	return nil, nil
}
