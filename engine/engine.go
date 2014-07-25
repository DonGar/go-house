package engine

import (
	"github.com/DonGar/go-house/engine/actions"
	"github.com/DonGar/go-house/engine/rules"
	"github.com/DonGar/go-house/status"
	"log"
	"strings"
)

const rules_watch_url = "status://*/rule/*"

type Engine struct {
	status  *status.Status
	actions map[string]actions.Action // Action name to fuction to perform action.
	rules   map[string]*rules.Rule    // URL of Rule definition to rule instance.
	stop    chan bool
}

func NewEngine(status *status.Status) (mgr *Engine, e error) {
	mgr = &Engine{
		status,
		map[string]actions.Action{},
		map[string]*rules.Rule{},
		make(chan bool),
	}

	// Register the builtin actions.
	mgr.RegisterAction("set", actions.ActionSet)
	mgr.RegisterAction("wol", actions.ActionWol)
	mgr.RegisterAction("ping", actions.ActionPing)
	mgr.RegisterAction("fetch", actions.ActionFetch)
	mgr.RegisterAction("email", actions.ActionEmail)

	// Start watching the status for rules updates.
	go mgr.rulesWatchReader()

	return mgr, nil
}

func (m *Engine) Stop() (e error) {
	m.stop <- true
	<-m.stop
	return nil
}

// Register additional actions for rules to perform. This is normally done by
// adapters.
func (m *Engine) RegisterAction(name string, action actions.Action) {
	m.actions[name] = action
}

func (m *Engine) LookupAction(name string) (action actions.Action, ok bool) {
	a, ok := m.actions[name]
	return a, ok
}

// This is our back ground process for noticing rules updates.
func (m *Engine) rulesWatchReader() {
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
func (m *Engine) updateRules(ruleMatches status.UrlMatches) {
	// Remove all rules that no longer exist, or which have been updated.
	for url, rule := range m.rules {
		match, ok := ruleMatches[url]
		if !ok || match.Revision != rule.Revision {
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

func (m *Engine) RemoveRule(url string, rule *rules.Rule) {
	log.Printf("Stop rule: %s: %s", rule.Name, url)
	rule.Stop()
	delete(m.rules, url)
}

func (m *Engine) AddRule(url string, match status.UrlMatch) error {
	// status://adapter_name/rules/<name>/
	url_parts := strings.Split(url, "/")
	ruleName := url_parts[len(url_parts)-1]

	ruleBody := &status.Status{}
	e := ruleBody.Set("status://", match.Value, 0)
	if e != nil {
		log.Panic(e) // This is supposed to be impossible.
	}

	newRule, e := rules.NewRule(m.status, m.actionHelper, ruleName, match.Revision, ruleBody)
	if e != nil {
		return e
	}

	log.Printf("Start rule: %s: %s", ruleName, url)
	m.rules[url] = newRule
	return nil
}

// This method implements the function signature needed by rules to fire
// actions. It understands how to fire them, and how to handle errors (rules
// don't).
func (m *Engine) actionHelper(action *status.Status) {
	e := actions.FireAction(m, m.status, action)
	if e != nil {
		log.Println("Fire Error: ", e)
	}
}
