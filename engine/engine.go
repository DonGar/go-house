package engine

import (
	"github.com/DonGar/go-house/engine/actions"
	"github.com/DonGar/go-house/engine/rules"
	"github.com/DonGar/go-house/status"
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

	return rules.NewRule(e.status, e.actions, name, body)
}
