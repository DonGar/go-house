package rules

import (
	"github.com/DonGar/go-house/actions"
	"github.com/DonGar/go-house/status"
	"log"
	"strings"
)

type Manager struct {
	status      *status.Status
	actions     map[string]actions.Action // Action name to fuction to perform action.
	ruleFactory map[string]newRule        // Rule type to rele factory method.
	rules       map[string]rule           // URL of Rule definition to rule instance.
	stop        chan bool
}

func NewManager(status *status.Status) (mgr *Manager, e error) {
	mgr = &Manager{
		status,
		map[string]actions.Action{},
		map[string]newRule{},
		map[string]rule{},
		make(chan bool),
	}

	// Register the builtin actions.
	mgr.RegisterAction("set", actions.ActionSet)
	mgr.RegisterAction("wol", actions.ActionWol)
	mgr.RegisterAction("ping", actions.ActionPing)
	mgr.RegisterAction("fetch", actions.ActionFetch)
	mgr.RegisterAction("email;", actions.ActionEmail)

	mgr.ruleFactory["base"] = newBaseRule
	mgr.ruleFactory["periodic"] = newPeriodicRule
	mgr.ruleFactory["daily"] = newDailyRule

	// Start watching the status for rules updates.
	go mgr.rulesWatchReader()

	return mgr, nil
}

func (m *Manager) Stop() (e error) {
	m.stop <- true
	<-m.stop
	return nil
}

// Register additional actions for rules to perform. This is normally done by
// adapters.
func (m *Manager) RegisterAction(name string, action actions.Action) {
	m.actions[name] = action
}

func (m *Manager) LookupAction(name string) (action actions.Action, ok bool) {
	a, ok := m.actions[name]
	return a, ok
}

// This is our back ground process for noticing rules updates.
func (m *Manager) rulesWatchReader() {
	rulesWatch, e := m.status.WatchForUpdate("status://*/rule/*/*")
	if e != nil {
		panic("Failure should not be possible.")
	}

	for {
		select {
		case ruleMatches := <-rulesWatch:
			// First remove rules that were removed or updated.
			m.updateRules(ruleMatches)
		case <-m.stop:
			// Stop watching for changes, remove all existing rules, and signal done.
			m.status.ReleaseWatch(rulesWatch)
			m.updateRules(status.UrlMatches{})
			m.stop <- true
			return
		}
	}
}

// Remove any rules that have been removed, or updated.
func (m *Manager) updateRules(ruleMatches status.UrlMatches) {
	// Remove all rules that no longer exist, or which have been updated.
	for url, rule := range m.rules {
		match, ok := ruleMatches[url]
		if !ok || match.Revision != rule.Revision() {
			// It's no longer valid, remove it.
			log.Printf("Stop rule: %s: %s", rule.Name(), url)
			rule.Stop()
			delete(m.rules, url)
		}
	}

	// Create all rules that don't exist in our manager.
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

		base := base{
			m.status,
			m.actionHelper,
			ruleName,
			match.Revision,
			ruleBody,
		}

		log.Printf("Start rule: %s: %s", ruleName, url)
		newRule, e := factory(base)
		if e != nil {
			// TODO: Log error and continue, not panic.
			panic(e)
		}

		m.rules[url] = newRule
	}
}

// This method implements the function signature needed by rules to fire
// actions. It understands how to fire them, and how to handle errors (rules
// don't).
func (m *Manager) actionHelper(action *status.Status) {
	e := actions.FireAction(m, m.status, action)
	if e != nil {
		log.Println("Fire Error: ", e)
	}
}
