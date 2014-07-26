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
	actions actions.ActionRegistrar
	rules   *watcher
}

func NewEngine(status *status.Status) (engine *Engine, e error) {
	engine = &Engine{status, actions.StandardActions(), nil}
	engine.rules = newWatcher(status, rules_watch_url, engine.newRule)

	return engine, nil
}

func (e *Engine) Stop() {
	e.rules.Stop()
}

func (e *Engine) newRule(url string, body *status.Status) (stoppable, error) {
	// Find it's name.
	// status://adapter_name/rules/<name>/
	url_parts := strings.Split(url, "/")
	name := url_parts[len(url_parts)-1]

	return rules.NewRule(e.status, e.actionHelper, name, body)
}

// This method implements the function signature needed by rules to fire
// actions. It understands how to fire them, and how to handle errors (rules
// don't).
func (e *Engine) actionHelper(action *status.Status) {
	err := actions.FireAction(e.status, e.actions, action)
	if err != nil {
		log.Println("Fire Error: ", err)
	}
}
