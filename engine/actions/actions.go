package actions

import (
	"fmt"
	"github.com/DonGar/go-house/status"
	"strings"
	"sync"
)

// This is the signature of an action implementation.
type Action func(s *status.Status, action *status.Status) (e error)

// Type for tracking the known actions in a thread safe manner.
type ActionManager struct {
	lock    sync.Mutex
	actions map[string]Action
}

func NewActionManager() *ActionManager {
	return &ActionManager{sync.Mutex{}, map[string]Action{}}
}

func (a *ActionManager) RegisterAction(name string, action Action) error {
	a.lock.Lock()
	defer a.lock.Unlock()

	_, ok := a.actions[name]
	if ok {
		return fmt.Errorf("Action: Already Exists: %s", name)
	}

	a.actions[name] = action
	return nil
}

func (a *ActionManager) UnRegisterAction(name string) error {
	a.lock.Lock()
	defer a.lock.Unlock()

	_, ok := a.actions[name]
	if !ok {
		return fmt.Errorf("Action: Not Registered: %s", name)
	}

	delete(a.actions, name)
	return nil
}

func (a *ActionManager) lookupAction(name string) (action Action, err error) {
	a.lock.Lock()
	defer a.lock.Unlock()

	action, ok := a.actions[name]

	if ok {
		return action, nil
	} else {
		return nil, fmt.Errorf("Action: Not Registered: %s", name)
	}
}

// This method should always be used to fire any action.
func (am *ActionManager) FireAction(s *status.Status, action *status.Status) error {
	actionValue, _, e := action.Get("status://")
	if e != nil {
		return e
	}

	switch typedAction := actionValue.(type) {
	case string:
		// A string represents a redirection to another part of status.
		redirectAction, _, e := s.GetSubStatus(typedAction)

		if e != nil {

			// If the redirection URL isn't a status URL, it might be an HTTP
			// url. Retry as an HTTP fetch action.
			if strings.HasPrefix(e.Error(), "Status: Invalid status url:") {
				fetchStatus := &status.Status{}
				fetchStatus.Set("status://action", "fetch", 0)
				fetchStatus.Set("status://url", typedAction, 1)

				// Recurse. This let's us lookup and fire the fetch action normally.
				return am.FireAction(s, fetchStatus)
			}

			// Some other error, probably that the status URL doesn't exist.
			return e
		}

		// We found it, fire it off!
		return am.FireAction(s, redirectAction)

	case []interface{}:
		// An array of actions means fire each one in order.
		for _, subActionValue := range typedAction {
			subActionStatus := &status.Status{}
			subActionStatus.Set("status://", subActionValue, 0)

			e = am.FireAction(s, subActionStatus)
			if e != nil {
				return e
			}
		}
		return nil

	case map[string]interface{}:
		// We received a dictionary, this is (hopefully) a registered action.
		actionName, e := action.GetString("status://action")
		if e != nil {
			return fmt.Errorf("Action: No action specified: %s", actionName)
		}

		actionMethod, e := am.lookupAction(actionName)
		if e != nil {
			return e
		}

		// Fire the looked up action.
		return actionMethod(s, action)

	default:
		return fmt.Errorf("Action: Can't perform %#v", actionValue)
	}
}
