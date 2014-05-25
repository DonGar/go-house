package status

import (
	// "errors"
	// "encoding/json"
	"fmt"
	"strings"
)

//
type Status struct {
	revision int
	value    statusValue
}

// Internal type used as the value of Status nodes with children.
type statusMap map[string]*Status

// Internal type used for the value stored in a Status. May be any of:
//   bool, float64, int, string, nil, for basic values.
//   []statusValue, for JSON arrays
//   statusMap, for JSON objects
type statusValue interface{}

const url_base = "status://"

// Json methods.
//   return json.Marshal(s.value)
//   return json.Marshal(s.Get(url))

func parseUrl(url string) (path_parts []string, e error) {

	// Handle this special case quickly.
	if url == url_base {
		return []string{}, nil
	}

	if !strings.HasPrefix(url, url_base) {
		return nil, fmt.Errorf("Invalid status url: %s", url)
	}

	// remove status:// from beginning, and / from end.
	prepped_url := strings.TrimPrefix(url, url_base)
	prepped_url = strings.TrimRight(prepped_url, "/")

	path_parts = strings.Split(prepped_url, "/")

	// If we still have an empty string in the slice after the trimming above, the
	// URL contained a double slash like "foo//bar", which we consider invalid.
	for _, part := range path_parts {
		if part == "" {
			return nil, fmt.Errorf("Invalid status url: %s", url)
		}
	}

	return path_parts, nil
}

func valueToStatus(revision int, value statusValue) (result *Status, e error) {
	result = &Status{revision: revision}

	switch t := value.(type) {
	case bool, float64, int, string, nil:
		// Immutable values are simply assigned.
		result.value = t
	case []interface{}:
		// Verify the array only contains supported values.
		for _, v := range t {
			switch element := v.(type) {
			case bool, float64, int, string, nil:
			default:
				return nil, fmt.Errorf("Illegal type: %T in Status array.", element)
			}
		}
		// Duplicate the array.
		value_array := make([]statusValue, len(t))
		for i, v := range t {
			value_array[i] = v
		}
		result.value = value_array
	case map[string]interface{}:
		// Convert each sub-value in a map.
		value_map := statusMap{}
		for k, v := range t {
			if value_map[k], e = valueToStatus(revision, v); e != nil {
				return nil, e
			}
		}
		result.value = value_map
	default:
		return nil, fmt.Errorf("Can't convert type: %T to Status value", t)
	}

	return result, nil
}

func (s *Status) toValue() (result interface{}, e error) {
	switch t := s.value.(type) {
	case bool, float64, int, string, nil:
		// Immutable values are simply assigned.
		result = t
	case []statusValue:
		// Convert each sub-value in an array.
		value_array := make([]interface{}, len(t))
		for i, v := range t {
			value_array[i] = v
		}
		result = value_array
	case statusMap:
		// Convert each sub-value in a map.
		value_map := map[string]interface{}{}
		for k, v := range t {
			if value_map[k], e = v.toValue(); e != nil {
				return nil, e
			}
		}
		result = value_map
	default:
		return nil, fmt.Errorf("Can't convert type: %T to Status value", t)
	}

	return result, nil
}

func (s *Status) urlPathToStatuses(url string, fillInMissing bool) (result []*Status, e error) {

	url_path, e := parseUrl(url)
	if e != nil {
		return
	}

	current := s
	result = make([]*Status, len(url_path)+1)
	result[0] = current

	for i, u := range url_path {
		current, e := current.getChildStatus(u)
		if e != nil {
			return nil, e
		}

		result[i+1] = current
	}

	return result, nil
}

func (s *Status) getChildStatus(name string) (result *Status, e error) {
	child_map, ok := s.value.(statusMap)
	if !ok {
		return nil, fmt.Errorf("Status: Node is not a map")
	}

	child, ok := child_map[name]
	if !ok {
		return nil, fmt.Errorf("Status: Node does not have child: %s", name)
	}

	return child, nil
}

func (s *Status) setChildStatus(name string, child *Status) (e error) {
	child_map, ok := s.value.(statusMap)
	if !ok {
		return fmt.Errorf("Status: Node is not a map")
	}

	child_map[name] = child
	return nil
}

func (s *Status) Get(url string) (interface{}, error) {
	statuses, e := s.urlPathToStatuses(url, false)
	if e != nil {
		return nil, e
	}

	return statuses[len(statuses)-1].toValue()
}

func (s *Status) Set(url string, value interface{}) (e error) {
	statuses, e := s.urlPathToStatuses(url, true)
	if e != nil {
		return e
	}

	new_status, e := valueToStatus(22, value)
	if e != nil {
		return e
	}

	*statuses[len(statuses)-1] = *new_status

	return nil
}
