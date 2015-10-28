package actions

import (
	"fmt"
	"github.com/DonGar/go-house/status"
	"log"
	"strings"
	"sync"
)

// This is the signature of an action implementation.
type Action func(s *status.Status, action *status.Status) (e error)

// Type for tracking the known actions in a thread safe manner.
type Manager struct {
	lock    sync.Mutex
	actions map[string]Action
}

func NewManager() *Manager {
	return &Manager{sync.Mutex{}, map[string]Action{}}
}

func (a *Manager) RegisterAction(name string, action Action) error {
	a.lock.Lock()
	defer a.lock.Unlock()

	_, ok := a.actions[name]
	if ok {
		return fmt.Errorf("Action: Already Exists: %s", name)
	}

	a.actions[name] = action
	return nil
}

func (a *Manager) UnRegisterAction(name string) error {
	a.lock.Lock()
	defer a.lock.Unlock()

	_, ok := a.actions[name]
	if !ok {
		return fmt.Errorf("Action: Not Registered: %s", name)
	}

	delete(a.actions, name)
	return nil
}

func (a *Manager) lookupAction(name string) (action Action, err error) {
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
func (am *Manager) FireAction(s *status.Status, action *status.Status) {

	var err error

	// If we exit with an error, we log it.
	defer func() {
		if err != nil {
			log.Println("Fire error: ", err)
		}
	}()

	actionValue, _, err := action.Get("status://")
	if err != nil {
		return
	}

	switch typedAction := actionValue.(type) {
	case string:
		// A string represents a redirection to another part of status.
		redirectAction, _, err := s.GetSubStatus(typedAction)

		if err != nil {

			// If the redirection URL isn't a status URL, it might be an HTTP
			// url. Retry as an HTTP fetch action.
			if strings.HasPrefix(err.Error(), "Status: Invalid status url:") {
				fetchStatus := &status.Status{}
				fetchStatus.Set("status://action", "fetch", 0)
				fetchStatus.Set("status://url", typedAction, 1)

				// Recurse. This let's us lookup and fire the fetch action normally.
				am.FireAction(s, fetchStatus)
			}

			// Some other error, probably that the status URL doesn't exist.
			return
		}

		// We found it, fire it off!
		am.FireAction(s, redirectAction)

	case []interface{}:
		// An array of actions means fire each one in order.
		// We do NOT return error results.
		for _, subActionValue := range typedAction {
			subActionStatus := &status.Status{}
			subActionStatus.Set("status://", subActionValue, 0)

			am.FireAction(s, subActionStatus)
		}

	case map[string]interface{}:
		// We received a dictionary, this is (hopefully) a registered action.
		actionName, _, err := action.GetString("status://action")
		if err != nil {
			err = fmt.Errorf("Action: No action specified: %s", actionName)
			return
		}

		actionMethod, err := am.lookupAction(actionName)
		if err != nil {
			return
		}

		// Fire the looked up action.
		log.Println("Firing action: ", actionName)
		err = actionMethod(s, action)

	default:
		err = fmt.Errorf("Action: Can't perform %#v", actionValue)
	}
}
