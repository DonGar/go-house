package status

import (
	"fmt"
	"strings"
	"sync"
)

// When calling Set, use this revision to avoid revision checking.
const UNCHECKED_REVISION = -1

// The status structure.
type Status struct {
	node
	lock     sync.RWMutex
	watchers []*watcher
}

// Structure used at every node in a Status tree.
type node struct {
	revision int
	value    statusValue
}

// Internal type used as the value of Status nodes with children.
type statusMap map[string]*node

// Internal type used for the value stored in a Status. May be any of:
//   bool, float64, int, int64, string, nil, for basic values.
//   []statusValue, for JSON arrays
//   statusMap, for JSON objects
type statusValue interface{}

// This represents a single URL match after wildcards are expanded.
type UrlMatch struct {
	Revision int
	Value    interface{}
}

// This is a map of status URLs to values, used when wildcard URLs are expanded.
type UrlMatches map[string]UrlMatch

const urlBase = "status://"

// Get a value from the status as described by the URL.
func (s *Status) Get(url string) (value interface{}, revision int, e error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	pathParts, e := parseUrl(url)
	if e != nil {
		return nil, 0, e
	}

	nodes, e := s.urlPathToNodes(pathParts, false)
	if e != nil {
		return nil, 0, e
	}

	node := nodes[len(nodes)-1]

	revision = node.revision
	value, e = statusValueToValue(node.value)
	if e != nil {
		return nil, 0, e
	}

	return
}

// Set a value from the status as described by the URL. Revision numbers are
// updated as needed.
func (s *Status) Set(url string, value interface{}, revision int) (e error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	pathParts, e := parseUrl(url)
	if e != nil {
		return e
	}

	if e := s.validRevision(pathParts, revision); e != nil {
		return e
	}

	// Look up the new revision to update.
	newRevision := s.revision + 1

	// Covert the new value into internal format.
	newValue, e := valueToStatusValue(value, newRevision)
	if e != nil {
		return e
	}

	// TODO: Verify the value is different from the old value.

	nodes, e := s.urlPathToNodes(pathParts, true)
	if e != nil {
		return e
	}

	// Set the new value to the last node found.
	nodes[len(nodes)-1].value = newValue

	// Update the revision for all affected nodes.
	for _, v := range nodes {
		v.revision = newRevision
	}

	s.checkWatchers()
	return nil
}

// Remove a named child from a node.
func (s *Status) Remove(url string, revision int) (e error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	pathParts, e := parseUrl(url)
	if e != nil {
		return e
	}

	if e := s.validRevision(pathParts, revision); e != nil {
		return e
	}

	// urlPathToNodes proves that the node we wish to remove exists.
	nodes, e := s.urlPathToNodes(pathParts, false)
	if e != nil {
		return e
	}

	// Look up the new revision to update.
	newRevision := nodes[0].revision + 1

	// The final node is no longer relevant, since we are about to remove it.
	nodes = nodes[:len(nodes)-1]

	// Discover the map in the parent node.
	parentMap := nodes[len(nodes)-1].value.(statusMap)

	delete(parentMap, pathParts[len(pathParts)-1])

	// Update the revision for all affected nodes.
	for _, v := range nodes {
		v.revision = newRevision
	}

	s.checkWatchers()
	return nil
}

// The exported version of GetMatchingUrls requires locking.
func (s *Status) GetMatchingUrls(url string) (matches UrlMatches, e error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.getMatchingUrls(url)
}

// Does the URL contain wildcards?
func CheckForWildcard(url string) (e error) {
	if strings.Contains(url, "*") {
		return fmt.Errorf("Status: Wildcards not allowed here: %s", url)
	}
	return nil
}

// Find all URLs that exist, which match a wildcard URL.
//
// A wildcard URL contains zero or more elements of '*' which match all
// keys on the relevant node.
func (s *Status) getMatchingUrls(url string) (matches UrlMatches, e error) {
	// Create a slice for fully expanded URLs.
	matches = UrlMatches{}

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
		current := &s.node
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
		}

		// If we finished without continuing to the next URL, it's a match.
		value, e := statusValueToValue(current.value)
		if e != nil {
			return nil, e
		}
		matches[joinUrl(urlPath)] = UrlMatch{Revision: current.revision, Value: value}
	}

	return matches, nil
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
func (s *Status) urlPathToNodes(urlPath []string, fillInMissing bool) (result []*node, e error) {

	// Check for wildcards.
	if e = CheckForWildcard(joinUrl(urlPath)); e != nil {
		return nil, e
	}

	result = make([]*node, len(urlPath)+1)

	current := &s.node
	result[0] = current

	for i, u := range urlPath {
		// If there is nothing at all, and we are creating the path..
		if fillInMissing && current.value == nil {
			current.value = statusMap{}
		}

		childMap, ok := current.value.(statusMap)
		if !ok {
			return nil, fmt.Errorf(
				"Status: Node %s of %s is not a map", joinUrl(urlPath[:i+1]), joinUrl(urlPath))
		}

		current, ok = childMap[u]
		if !ok {
			if fillInMissing {
				current = &node{value: statusMap{}, revision: s.revision}
				childMap[u] = current
			} else {
				return nil, fmt.Errorf(
					"Status: Node %s of %s does not exist.", joinUrl(urlPath[:i+1]), joinUrl(urlPath))
			}
		}

		result[i+1] = current
	}

	return result, nil
}

func (s *Status) validRevision(urlPath []string, revision int) (e error) {

	// UNCHECKED_REVISION is always a valid revision. It means don't test.
	if revision == UNCHECKED_REVISION {
		return nil
	}

	current := &s.node

	for _, u := range urlPath {
		if current.revision == revision {
			// If we found a revision match, there is no error.
			return nil
		}

		childMap, ok := current.value.(statusMap)
		if !ok {
			break
		}

		child, ok := childMap[u]
		if !ok {
			break
		}

		current = child
	}

	if current.revision == revision {
		// If we found a revision match, there is no error.
		return nil
	}

	return fmt.Errorf(
		"Status: Invalid Revision: %d - %s. Expected %d",
		revision, joinUrl(urlPath), current.revision)
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
			case bool, float64, int, int64, string, nil, []interface{}, map[string]interface{}:
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
			valueMap[k] = &node{revision: revision, value: subValue}
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
