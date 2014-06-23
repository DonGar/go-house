package rules

import (
	"fmt"
	"github.com/DonGar/go-house/status"
	"strings"
	"time"
)

type Manager struct {
	status      *status.Status
	actions     map[string]Action  // Action name to fuction to perform action.
	ruleFactory map[string]NewRule // Rule type to rele factory method.
	rulesWatch  <-chan status.UrlMatches
	rules       map[string]Rule // URL of Rule definition to rule instance.
}

func NewManager(status *status.Status) (mgr *Manager, e error) {
	mgr = &Manager{
		status:      status,
		actions:     map[string]Action{},
		ruleFactory: map[string]NewRule{},
		rules:       map[string]Rule{},
	}

	// Register the builtin actions.
	mgr.RegisterAction("set", actionSet)
	mgr.RegisterAction("wol", actionWol)
	mgr.RegisterAction("ping", actionPing)
	mgr.RegisterAction("fetch", actionFetch)
	mgr.RegisterAction("email;", actionEmail)

	mgr.ruleFactory["periodic"] = newPeriodicRule
	mgr.ruleFactory["conditional"] = newConditionalRule
	mgr.ruleFactory["status"] = newStatusRule

	mgr.rulesWatch, e = mgr.status.WatchForUpdate("status://*/rules/*/*")
	if e != nil {
		return nil, e
	}

	// Start watching the status for rules updates.
	go mgr.rulesWatchReader()

	return mgr, nil
}

func (m *Manager) Stop() (e error) {
	m.status.ReleaseWatch(m.rulesWatch)
	time.Sleep(100 * time.Millisecond)
	return nil
}

// Register additional actions for rules to perform. This is normally done by
// adapters.
func (m *Manager) RegisterAction(name string, action Action) {
	m.actions[name] = action
}

// This is our back ground process for noticing rules updates.
func (m *Manager) rulesWatchReader() {
	for {
		ruleMatches := <-m.rulesWatch
		fmt.Println("Recieved.", ruleMatches)

		// First remove rules that were removed or updated.
		m.removeOutdatedRules(ruleMatches)

		// Add rules that aren't already present.
		m.createUpdatedRules(ruleMatches)

	}
}

// Remove any rules that have been removed, or updated.
func (m *Manager) removeOutdatedRules(ruleMatches status.UrlMatches) {
	for url, rule := range m.rules {
		match, ok := ruleMatches[url]
		if !ok || match.Revision != rule.Revision() {
			// It's no longer valid, remove it.
			rule.Stop()
			delete(m.rules, url)
		}
	}
}

// Remove any rules that have been removed, or updated.
func (m *Manager) createUpdatedRules(ruleMatches status.UrlMatches) {
	for url, match := range ruleMatches {

		// If the rule already exists, leave it alone.
		if _, ok := m.rules[url]; ok {
			continue
		}

		// status://adapter_name/rules/<type>/<name>/
		url_parts := strings.Split(url, "/")
		ruleName := url_parts[len(url_parts)-1]
		ruleType := url_parts[len(url_parts)-2]

		ruleBody := &status.Status{}
		e := ruleBody.Set("status://", match.Value, 0)
		if e != nil {
			panic(e) // This is supposed to be impossible.
		}

		factory, ok := m.ruleFactory[ruleType]
		if !ok {
			// TODO: Log error and continue, not panic.
			panic("Unknown rule type: " + ruleType)
		}

		newRule, e := factory(m.status, m, ruleName, ruleBody)
		if e != nil {
			// TODO: Log error and continue, not panic.
			panic(e)
		}

		m.rules[url] = newRule
	}
}
