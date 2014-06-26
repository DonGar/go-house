package rules

import (
	// "github.com/cpucycle/astrotime"
	"time"
)

type periodicRule struct {
	base
	period time.Duration
	stop   chan bool
}

func newPeriodicRule(base base) (rule, error) {

	interval, e := base.body.GetString("status://interval")
	if e != nil {
		return nil, e
	}

	period, e := time.ParseDuration(interval)
	if e != nil {
		return nil, e
	}

	// Create the rule.
	r := &periodicRule{base: base, period: period, stop: make(chan bool)}

	// Start it processing.
	go r.handleTicks()

	return r, nil
}

func (r *periodicRule) handleTicks() {
	ticker := time.NewTicker(r.period)

	for {
		select {
		case <-ticker.C:
			r.fire()
		case <-r.stop:
			ticker.Stop()
			return
		}
	}

}

func (r *periodicRule) Stop() error {
	r.stop <- true
	return r.base.Stop()
}
