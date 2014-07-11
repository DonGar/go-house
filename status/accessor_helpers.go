package status

import (
	"encoding/json"
	"fmt"
)

// This is just like Set, except it accepts the value to set in Json format.
func (s *Status) SetJson(url string, valueJson []byte, revision int) (e error) {
	var value interface{}
	e = json.Unmarshal(valueJson, &value)
	if e != nil {
		return e
	}

	return s.Set(url, value, revision)
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

func (s *Status) PrettyDump(url string) string {
	value, _, e := s.Get(url)
	if e != nil {
		panic(e)
	}

	valueJson, e := json.MarshalIndent(value, "", "    ")
	if e != nil {
		panic(e)
	}

	return string(valueJson)
}

// Returns a copy of a sub-tree as a new Status object.
// Useful if you need a sub-tree that's frozen with friendly accessors.
func (s *Status) GetSubStatus(url string) (contents *Status, revision int, e error) {
	value, revision, e := s.Get(url)
	if e != nil {
		return nil, 0, e
	}

	contents = &Status{}
	e = contents.Set("status://", value, 0)
	if e != nil {
		return nil, 0, e
	}

	return contents, revision, nil
}

// Get the names of the children of a given node.
func (s *Status) GetChildNames(url string) (names []string, e error) {
	value, _, e := s.Get(url)
	if e != nil {
		return nil, e
	}

	childMap, ok := value.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("Status: Node is %T not a map", value)
	}

	names = make([]string, 0, len(childMap))
	for childName := range childMap {
		names = append(names, childName)
	}

	return names, nil
}

// Extract a string value.
func (s *Status) GetString(url string) (value string, e error) {
	rawValue, _, e := s.Get(url)
	if e != nil {
		return "", e
	}

	value, ok := rawValue.(string)
	if !ok {
		return "", fmt.Errorf("Status: %s is %T not string.", url, rawValue)
	}

	return value, nil
}

func (s *Status) GetStringWithDefault(url string, defaultValue string) (value string) {
	v, e := s.GetString(url)
	if e == nil {
		return v
	} else {
		return defaultValue
	}
}

func (s *Status) GetInt(url string) (value int, e error) {
	rawValue, _, e := s.Get(url)
	if e != nil {
		return 0, e
	}

	switch t := rawValue.(type) {
	case int:
		return t, nil
	case float64:
		return int(t), nil
	default:
		return 0, fmt.Errorf("Status: %s is %T not int.", url, rawValue)
	}
}

func (s *Status) GetIntWithDefault(url string, defaultValue int) (value int) {
	v, e := s.GetInt(url)
	if e == nil {
		return v
	} else {
		return defaultValue
	}
}

func (s *Status) GetFloat(url string) (value float64, e error) {
	rawValue, _, e := s.Get(url)
	if e != nil {
		return 0, e
	}

	value, ok := rawValue.(float64)
	if !ok {
		return 0, fmt.Errorf("Status: %s is %T not float64.", url, rawValue)
	}

	return value, nil
}

func (s *Status) GetFloatWithDefault(url string, defaultValue float64) (value float64) {
	v, e := s.GetFloat(url)
	if e == nil {
		return v
	} else {
		return defaultValue
	}
}

func (s *Status) GetStrings(urls []string) (values []string, e error) {
	values = make([]string, len(urls))
	for i, url := range urls {
		values[i], e = s.GetString(url)
		if e != nil {
			return nil, e
		}
	}

	return values, nil
}
