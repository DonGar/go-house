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
const urlBase = "status://"

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
func (s *Status) GetJson(url string) (valueJson []byte, revision int, e error) {
	value, revision, e := s.Get(url)
	if e != nil {
		return nil, 0, e
	}

	valueJson, e = json.Marshal(value)
	if e != nil {
		return nil, 0, e
	}

	return valueJson, revision, e
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
		var revisionValid bool = false
		for _, v := range statuses {
			if v.revision == revision {
				revisionValid = true
				break
			}
		}
		if !revisionValid {
			return fmt.Errorf(
				"Status: Invalid Revision: %d - %s. Expected %d",
				revision, url, s.revision)
		}
	}

	// Look up the new revision to update.
	newRevision := statuses[0].revision + 1

	// Covert the new value into internal format.
	newValue, e := valueToStatusValue(value, newRevision)
	if e != nil {
		return e
	}

	// Set the new value to the last node found.
	statuses[len(statuses)-1].value = newValue

	// Update the revision for all affected nodes.
	for _, v := range statuses {
		v.revision = newRevision
	}

	return nil
}

// This is just like Set, except it accepts the value to set in Json format.
func (s *Status) SetJson(url string, valueJson []byte, revision int) (e error) {
	var value interface{}
	e = json.Unmarshal(valueJson, &value)
	if e != nil {
		return e
	}

	return s.Set(url, value, revision)
}

// Find all URLs that exist, which match a wildcard URL.
//
// A wildcard URL contains zero or more elements of '*' which match all
// keys on the relevant node.
func (s *Status) getMatchingUrls(url string) (urls []string, e error) {

	// We special case the root directory.
	if url == urlBase {
		return []string{url}, nil
	}

	// Create a slice for fully expanded URLs.
	matchedUrls := []string{}

	// Create a slice of URLs that haven't been expanded yet.
	unfinishedUrls := []string{url}

UnfinishedUrls:
	for len(unfinishedUrls) > 0 {
		// Pop off the last element for processing.
		testingUrl := unfinishedUrls[len(unfinishedUrls)-1]
		unfinishedUrls = unfinishedUrls[:len(unfinishedUrls)-1]

		// Parse the Url.
		urlPath, e := parseUrl(testingUrl)
		if e != nil {
			return nil, e
		}

		// Walk down the tree looking for matches.
		current := s
		for i, v := range urlPath {

			currentMap, ok := current.value.(statusMap)
			if !ok {
				continue UnfinishedUrls
			}

			if v == "*" {
				expandedPathParts := make([]string, len(urlPath))
				copy(expandedPathParts, urlPath)

				// Expand the star for each key in the current node.
				for k := range currentMap {
					expandedPathParts[i] = k
					unfinishedUrls = append(unfinishedUrls, joinUrl(expandedPathParts))
				}

				continue UnfinishedUrls
			}

			// Step to the next child, if it exists.
			if current, ok = currentMap[v]; !ok {
				continue UnfinishedUrls
			}

			// If we found the final element in the path, we fully expanded the URL
			// and found a match.
			if i == len(urlPath)-1 {
				matchedUrls = append(matchedUrls, joinUrl(urlPath))
			}
		}
	}

	return matchedUrls, nil
}

// Parse a status url, and return it as a slice of strings.
// One for each step in the URL. Error if it's not a legal
// status URL.
func parseUrl(url string) (pathParts []string, e error) {
	// Handle this special case quickly.
	if url == urlBase {
		return []string{}, nil
	}

	if !strings.HasPrefix(url, urlBase) {
		return nil, fmt.Errorf("Status: Invalid status url: %s", url)
	}

	// remove status:// from beginning, and / from end.
	preppedUrl := strings.TrimPrefix(url, urlBase)
	preppedUrl = strings.TrimRight(preppedUrl, "/")

	pathParts = strings.Split(preppedUrl, "/")

	// If we still have an empty string in the slice after the trimming above, the
	// URL contained a double slash like "foo//bar", which we consider invalid.
	for _, part := range pathParts {
		if part == "" {
			return nil, fmt.Errorf("Status: Invalid status url: %s", url)
		}
	}

	return pathParts, nil
}

// The inverse of parseUrl.
func joinUrl(pathParts []string) (url string) {
	return urlBase + strings.Join(pathParts, "/")
}

// Given a status URL, return a slice of *Status to each of the nodes referenced
// by the URL. If fillInMissing is true, create missing nodes as needed to do
// this.
func (s *Status) urlPathToStatuses(url string, fillInMissing bool) (result []*Status, e error) {

	urlPath, e := parseUrl(url)
	if e != nil {
		return
	}

	current := s
	result = make([]*Status, len(urlPath)+1)
	result[0] = current

	for i, u := range urlPath {
		// If there is nothing at all, and we are creating the path..
		if fillInMissing && current.value == nil {
			current.value = statusMap{}
		}

		childMap, ok := current.value.(statusMap)
		if !ok {
			return nil, fmt.Errorf("Status: Node is not a map")
		}

		current, ok = childMap[u]
		if !ok {
			if fillInMissing {
				current = &Status{value: statusMap{}, revision: s.revision}
				childMap[u] = current
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
		valueArray := make([]statusValue, len(t))
		for i, v := range t {
			valueArray[i] = v
		}
		result = valueArray
	case map[string]interface{}:
		// Convert each sub-value in a map.
		valueMap := statusMap{}
		for k, v := range t {
			subValue, e := valueToStatusValue(v, revision)
			if e != nil {
				return nil, e
			}
			valueMap[k] = &Status{revision: revision, value: subValue}
		}
		result = valueMap
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
		valueArray := make([]interface{}, len(t))
		for i, v := range t {
			valueArray[i] = v
		}
		result = valueArray
	case statusMap:
		// Convert each sub-value in a map.
		valueMap := map[string]interface{}{}
		for k, v := range t {
			if valueMap[k], e = statusValueToValue(v.value); e != nil {
				return nil, e
			}
		}
		result = valueMap
	default:
		return nil, fmt.Errorf("Status: Can't convert type: %T to Status value", t)
	}

	return result, nil
}
