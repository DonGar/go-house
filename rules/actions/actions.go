package actions

import (
	"fmt"
	"github.com/DonGar/go-house/status"
	"strings"
)

// This is the signature of an action implementation.
type Action func(r ActionRegistrar, s *status.Status, action *status.Status) (e error)

// How do we look up named types? The manager does this in production, but
// Mocks do it in tests.
type ActionRegistrar interface {
	LookupAction(name string) (action Action, ok bool)
}

// This method should always be used to fire any action.
func FireAction(r ActionRegistrar, s *status.Status, action *status.Status) error {
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
				return FireAction(r, s, fetchStatus)
			}

			// Some other error, probably that the status URL doesn't exist.
			return e
		}

		// We found it, fire it off!
		return FireAction(r, s, redirectAction)

	case []interface{}:
		// An array of actions means fire each one in order.
		for _, subActionValue := range typedAction {
			subActionStatus := &status.Status{}
			subActionStatus.Set("status://", subActionValue, 0)

			e = FireAction(r, s, subActionStatus)
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

		actionMethod, ok := r.LookupAction(actionName)
		if !ok {
			return fmt.Errorf("Action: No registered action: %s", actionName)
		}

		// Fire the looked up action.
		return actionMethod(r, s, action)

	default:
		return fmt.Errorf("Action: Can't perform %s", action)
	}
}
