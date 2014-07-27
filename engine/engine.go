package engine

import (
	"github.com/DonGar/go-house/engine/actions"
	"github.com/DonGar/go-house/engine/properties"
	"github.com/DonGar/go-house/engine/rules"
	"github.com/DonGar/go-house/status"
	"strings"
)

const rules_watch_url = "status://*/rule/*"
const properties_watch_url = "status://*/property/*"

type Engine struct {
	status     *status.Status
	actions    actions.ActionRegistrar
	rules      *watcher
	properties *watcher
}

func NewEngine(status *status.Status) (engine *Engine, e error) {
	engine = &Engine{status, actions.StandardActions(), nil, nil}
	engine.rules = newWatcher(status, rules_watch_url, engine.newRule)
	engine.properties = newWatcher(status, properties_watch_url, engine.newProperty)

	return engine, nil
}

func (e *Engine) Stop() {
	e.rules.Stop()
	e.properties.Stop()
}

func nameFromUrl(url string) string {
	// Find it's name.
	// status://adapter_name/rules/<name>/
	url_parts := strings.Split(url, "/")
	return url_parts[len(url_parts)-1]
}

func (e *Engine) newRule(url string, body *status.Status) (stoppable, error) {
	return rules.NewRule(e.status, e.actions, nameFromUrl(url), body)
}

func (e *Engine) newProperty(url string, body *status.Status) (stoppable, error) {
	return properties.NewProperty(e.status, nameFromUrl(url), body)
}
