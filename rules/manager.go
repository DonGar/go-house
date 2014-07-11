package rules

import (
	"fmt"
	"github.com/DonGar/go-house/rules/actions"
	"github.com/DonGar/go-house/rules/conditions"
	"github.com/DonGar/go-house/status"
	"log"
	"strings"
)

const rules_watch_url = "status://*/rule/*"

type Manager struct {
	status  *status.Status
	actions map[string]actions.Action // Action name to fuction to perform action.
	rules   map[string]*rule          // URL of Rule definition to rule instance.
	stop    chan bool
}

func NewManager(status *status.Status) (mgr *Manager, e error) {
	mgr = &Manager{
		status,
		map[string]actions.Action{},
		map[string]*rule{},
		make(chan bool),
	}

	// Register the builtin actions.
	mgr.RegisterAction("set", actions.ActionSet)
	mgr.RegisterAction("wol", actions.ActionWol)
	mgr.RegisterAction("ping", actions.ActionPing)
	mgr.RegisterAction("fetch", actions.ActionFetch)
	mgr.RegisterAction("email;", actions.ActionEmail)

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
	rulesWatch, e := m.status.WatchForUpdate(rules_watch_url)
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
		if !ok || match.Revision != rule.revision {
			// It's no longer valid, remove it.
			m.RemoveRule(url, rule)
		}
	}

	// Create all rules that don't exist in our manager.
	for url, match := range ruleMatches {

		// If the rule already exists, leave it alone.
		if _, ok := m.rules[url]; ok {
			continue
		}

		// Add the new/updated rule.
		e := m.AddRule(url, match)
		if e != nil {
			log.Printf("INVALID RULE: %s: %s", url, e.Error())
		}
	}
}

func (m *Manager) RemoveRule(url string, rule *rule) {
	log.Printf("Stop rule: %s: %s", rule.name, url)
	rule.Stop()
	delete(m.rules, url)
}

func (m *Manager) AddRule(url string, match status.UrlMatch) error {
	// status://adapter_name/rules/<name>/
	url_parts := strings.Split(url, "/")
	ruleName := url_parts[len(url_parts)-1]

	ruleBody := &status.Status{}
	e := ruleBody.Set("status://", match.Value, 0)
	if e != nil {
		log.Panic(e) // This is supposed to be impossible.
	}

	// Find the sub-expression contents.
	conditionBody, _, e := ruleBody.GetSubStatus("status://condition")
	if e != nil {
		return fmt.Errorf("No 'condition' section.")
	}

	actionBody, _, e := ruleBody.GetSubStatus("status://action")
	if e != nil {
		return fmt.Errorf("No 'action' section.")
	}

	// Create the condition (last, because it needs Stopping on failure).
	condition, e := conditions.NewCondition(m.status, conditionBody)
	if e != nil {
		return e
	}

	newRule := newRule(m.actionHelper, ruleName, match.Revision, actionBody, condition)

	log.Printf("Start rule: %s: %s", ruleName, url)
	m.rules[url] = newRule
	return nil
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
