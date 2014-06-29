package rules

import (
	"fmt"
)

// This is the signature of an action implementation.
type Action func(actionUrl, componentUrl string) (e error)

func (m *Manager) fireAction(actionUrl string) error {
	actionValue, _, e := m.status.Get(actionUrl)
	if e != nil {
		return e
	}

	switch v := actionValue.(type) {
	case string:
		_ = v
		// If status URL.
		// If http/https/etc URL.
		return nil
	case []interface{}:
		return nil
	case map[string]interface{}:
		return nil
	default:
		return fmt.Errorf("Action: Can't perform action %s", actionValue)
	}
}
