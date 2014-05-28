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

// Json methods.
//   return json.Marshal(s.value)
//   return json.Marshal(s.Get(url))

// Get a value from the status as described by the URL.
func (s *Status) Get(url string) (interface{}, error) {
	statuses, e := s.urlPathToStatuses(url, false)
	if e != nil {
		return nil, e
	}

	return statusValueToValue(statuses[len(statuses)-1].value)
}

// Set a value from the status as described by the URL. Revision numbers are
// updated as needed.
func (s *Status) Set(url string, value interface{}) (e error) {
	statuses, e := s.urlPathToStatuses(url, true)
	if e != nil {
		return e
	}

	new_revision := statuses[0].revision + 1

	new_value, e := valueToStatusValue(value, new_revision)
	if e != nil {
		return e
	}

	// Set the new value to the last node found.
	statuses[len(statuses)-1].value = new_value

	// Update the revision for all affected nodes.
	for _, v := range statuses {
		v.revision = new_revision
	}

	return nil
}

func (s *Status) Revision(url string) (result int, e error) {
	statuses, e := s.urlPathToStatuses(url, false)
	if e != nil {
		return 0, e
	}

	return statuses[len(statuses)-1].revision, nil
}

// Parse a status url, and return it as a slice of strings.
// One for each step in the URL. Error if it's not a legal
// status URL.
func parseUrl(url string) (path_parts []string, e error) {

	const url_base = "status://"

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

// Given a status URL, return a slice of *Status to each of the nodes referenced
// by the URL. If fillInMissing is true, create missing nodes as needed to do
// this.
func (s *Status) urlPathToStatuses(url string, fillInMissing bool) (result []*Status, e error) {

	url_path, e := parseUrl(url)
	if e != nil {
		return
	}

	current := s
	result = make([]*Status, len(url_path)+1)
	result[0] = current

	for i, u := range url_path {
		// If there is nothing at all, and we are creating the path..
		if fillInMissing && current.value == nil {
			current.value = statusMap{}
		}

		child_map, ok := current.value.(statusMap)
		if !ok {
			return nil, fmt.Errorf("Status: Node is not a map")
		}

		current, ok = child_map[u]
		if !ok {
			if fillInMissing {
				current = &Status{value: statusMap{}}
				child_map[u] = current
			} else {
				return nil, fmt.Errorf("Status: Node does not have child: %s", u)
			}
		}

		result[i+1] = current
	}

	return result, nil
}

// Convert an external value (matching a JSON structure), to the internal
// statusValue. The main difference is that all maps in the struct are
// converted to Status values.
func valueToStatusValue(value interface{}, revision int) (result statusValue, e error) {
	switch t := value.(type) {
	case bool, float64, int, string, nil:
		// Immutable values are simply assigned.
		result = t
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
		result = value_array
	case map[string]interface{}:
		// Convert each sub-value in a map.
		value_map := statusMap{}
		for k, v := range t {
			sub_value, e := valueToStatusValue(v, revision)
			if e != nil {
				return nil, e
			}
			value_map[k] = &Status{revision: revision, value: sub_value}
		}
		result = value_map
	default:
		return nil, fmt.Errorf("Can't convert type: %T to Status value", t)
	}

	return result, nil
}

// Convert an internal value to the external (JSON structure) equivalent.
func statusValueToValue(value statusValue) (result interface{}, e error) {
	switch t := value.(type) {
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
			if value_map[k], e = statusValueToValue(v.value); e != nil {
				return nil, e
			}
		}
		result = value_map
	default:
		return nil, fmt.Errorf("Can't convert type: %T to Status value", t)
	}

	return result, nil
}
