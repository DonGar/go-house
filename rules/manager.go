package rules

import (
	// "github.com/cpucycle/astrotime"
	"github.com/DonGar/go-house/status"
)

type Manager struct {
	status  *status.Status
	actions map[string]Action
}

func NewManager(status *status.Status) (mgr *Manager, e error) {
	mgr = &Manager{status: status, actions: map[string]Action{}}

	// Register the builtin actions.
	mgr.RegisterAction("set", actionSet)
	mgr.RegisterAction("wol", actionWol)
	mgr.RegisterAction("ping", actionPing)
	mgr.RegisterAction("fetch", actionFetch)
	mgr.RegisterAction("email;", actionEmail)

	return mgr, nil
}

func (m *Manager) Stop() (e error) {
	return nil
}

// Register additional actions for rules to perform. This is normally done by
// adapters.
func (m *Manager) RegisterAction(name string, action Action) {
	m.actions[name] = action
}
