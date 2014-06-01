package status

import (
	"encoding/json"
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
//   bool, float64, int, int64, string, nil, for basic values.
//   []statusValue, for JSON arrays
//   statusMap, for JSON objects
type statusValue interface{}

// When calling Set, use this revision to avoid revision checking.
const UNCHECKED_REVISION = -1

// Get a value from the status as described by the URL.
func (s *Status) Get(url string) (value interface{}, revision int, e error) {
	statuses, e := s.urlPathToStatuses(url, false)
	if e != nil {
		return nil, 0, e
	}

	node := statuses[len(statuses)-1]

	revision = node.revision
	value, e = statusValueToValue(node.value)
	if e != nil {
		return nil, 0, e
	}

	return
}

// This is just like Get, except it returns the value in Json format.
func (s *Status) GetJson(url string) (value_json []byte, revision int, e error) {
	value, revision, e := s.Get(url)
	if e != nil {
		return nil, 0, e
	}

	value_json, e = json.Marshal(value)
	if e != nil {
		return nil, 0, e
	}

	return value_json, revision, e
}

// Set a value from the status as described by the URL. Revision numbers are
// updated as needed.
func (s *Status) Set(url string, value interface{}, revision int) (e error) {
	statuses, e := s.urlPathToStatuses(url, true)
	if e != nil {
		return e
	}

	// UNCHECKED_REVISION is always a valid revision. It means don't test.
	if revision != UNCHECKED_REVISION {
		// Find out if the passed in revision is an exact match for the selected node,
		// or any of it's parents.
		var revision_valid bool = false
		for _, v := range statuses {
			if v.revision == revision {
				revision_valid = true
				break
			}
		}
		if !revision_valid {
			return fmt.Errorf(
				"Status: Invalid Revision: %d - %s. Expected %d",
				revision, url, s.revision)
		}
	}

	// Look up the new revision to update.
	new_revision := statuses[0].revision + 1

	// Covert the new value into internal format.
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

// This is just like Set, except it accepts the value to set in Json format.
func (s *Status) SetJson(url string, value_json []byte, revision int) (e error) {
	var value interface{}
	e = json.Unmarshal(value_json, &value)
	if e != nil {
		return e
	}

	return s.Set(url, value, revision)
}

// Find all URLs that exist, which match a wildcard URL.
//
// A wildcard URL contains zero or more elements of '*' which match all
// keys on the relevant node.
func (s *Status) GetMatchingUrls(url string) (urls []string, revision int, e error) {

	// We special case the root directory.
	if url == "status://" {
		return []string{url}, s.revision, nil
	}

	// Create a slice for fully expanded URLs.
	matched_urls := []string{}

	// Create a slice of URLs that haven't been expanded yet.
	unfinished_urls := []string{url}

UnfinishedUrls:
	for len(unfinished_urls) > 0 {
		// Pop off the last element for processing.
		testing_url := unfinished_urls[len(unfinished_urls)-1]
		unfinished_urls = unfinished_urls[:len(unfinished_urls)-1]

		// Parse the Url.
		url_path, e := parseUrl(testing_url)
		if e != nil {
			return nil, 0, e
		}

		// Walk down the tree looking for matches.
		current := s
		for i, v := range url_path {

			current_map, ok := current.value.(statusMap)
			if !ok {
				continue UnfinishedUrls
			}

			if v == "*" {
				expanded_path_parts := make([]string, len(url_path))
				copy(expanded_path_parts, url_path)

				// Expand the star for each key in the current node.
				for k := range current_map {
					expanded_path_parts[i] = k
					unfinished_urls = append(unfinished_urls, joinUrl(expanded_path_parts))
				}

				continue UnfinishedUrls
			}

			// Step to the next child, if it exists.
			if current, ok = current_map[v]; !ok {
				continue UnfinishedUrls
			}

			// If we found the final element in the path, we fully expanded the URL
			// and found a match.
			if i == len(url_path)-1 {
				matched_urls = append(matched_urls, joinUrl(url_path))
			}
		}
	}

	return matched_urls, s.revision, nil
}

const url_base = "status://"

// Parse a status url, and return it as a slice of strings.
// One for each step in the URL. Error if it's not a legal
// status URL.
func parseUrl(url string) (path_parts []string, e error) {
	// Handle this special case quickly.
	if url == url_base {
		return []string{}, nil
	}

	if !strings.HasPrefix(url, url_base) {
		return nil, fmt.Errorf("Status: Invalid status url: %s", url)
	}

	// remove status:// from beginning, and / from end.
	prepped_url := strings.TrimPrefix(url, url_base)
	prepped_url = strings.TrimRight(prepped_url, "/")

	path_parts = strings.Split(prepped_url, "/")

	// If we still have an empty string in the slice after the trimming above, the
	// URL contained a double slash like "foo//bar", which we consider invalid.
	for _, part := range path_parts {
		if part == "" {
			return nil, fmt.Errorf("Status: Invalid status url: %s", url)
		}
	}

	return path_parts, nil
}

// The inverse of parseUrl.
func joinUrl(path_parts []string) (url string) {
	return url_base + strings.Join(path_parts, "/")
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
				current = &Status{value: statusMap{}, revision: s.revision}
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
	case bool, float64, int, int64, string, nil:
		// Immutable values are simply assigned.
		result = t
	case []interface{}:
		// Verify the array only contains supported values.
		for _, v := range t {
			switch element := v.(type) {
			case bool, float64, int, int64, string, nil:
			default:
				return nil, fmt.Errorf("Status: Illegal type: %T in Status array.", element)
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
		return nil, fmt.Errorf("Status: Can't convert type: %T to Status value", t)
	}

	return result, nil
}

// Convert an internal value to the external (JSON structure) equivalent.
func statusValueToValue(value statusValue) (result interface{}, e error) {
	switch t := value.(type) {
	case bool, float64, int, int64, string, nil:
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
		return nil, fmt.Errorf("Status: Can't convert type: %T to Status value", t)
	}

	return result, nil
}
