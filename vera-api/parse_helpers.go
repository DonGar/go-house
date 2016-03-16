package veraapi

import (
	"fmt"
	"strconv"
)

type emptyError struct {
}

func (e emptyError) Error() string {
	return "Value \"\" is not valid."
}

var emptyString = &emptyError{}

func parseString(raw interface{}) (string, error) {
	switch t := raw.(type) {
	case string:
		if t == "" {
			return "", emptyString
		}
		return t, nil
	default:
		return "", fmt.Errorf("Expecting string not: %v(%T)", raw, raw)
	}
}

func parseBool(raw interface{}) (bool, error) {
	switch t := raw.(type) {
	case bool:
		return t, nil
	case string:
		if t == "" {
			return false, emptyString
		}
		return strconv.ParseBool(t)
	case float64:
		if t == 0 || t == 1 {
			return t == 1, nil
		}
		return false, fmt.Errorf("Expecting bool not: %v(%T)", raw, raw)
	default:
		return false, fmt.Errorf("Expecting bool not: %v(%T)", raw, raw)
	}
}

func parseInt(raw interface{}) (int, error) {
	switch t := raw.(type) {
	case float64:
		return int(t), nil
	case string:
		if t == "" {
			return 0, emptyString
		}
		parsed, err := strconv.ParseInt(t, 0, 0)
		return int(parsed), err
	default:
		return 0, fmt.Errorf("Expecting int not: %v(%T)", raw, raw)
	}
}

func parseFloat(raw interface{}) (float64, error) {
	switch t := raw.(type) {
	case float64:
		return t, nil
	case string:
		if t == "" {
			return 0, emptyString
		}
		return strconv.ParseFloat(t, 64)
	default:
		return 0, fmt.Errorf("Expecting float not: %v(%T)", raw, raw)
	}
}

func insertRawString(values ValuesMap, key string, raw interface{}) (err error) {
	defer func() { err = insertErrorValue(key, err) }()

	if raw == nil {
		return nil
	}

	value, err := parseString(raw)
	if err == emptyString {
		return nil
	}
	if err != nil {
		return err
	}

	if value == "" {
		return nil
	}

	values[key] = value
	return nil
}

func insertRawBool(values ValuesMap, key string, raw interface{}) (err error) {
	defer func() { err = insertErrorValue(key, err) }()

	if raw == nil {
		return nil
	}

	value, err := parseBool(raw)
	if err == emptyString {
		return nil
	}
	if err != nil {
		return err
	}

	values[key] = value
	return nil
}

func insertRawInt(values ValuesMap, key string, raw interface{}) (err error) {
	defer func() { err = insertErrorValue(key, err) }()

	if raw == nil {
		return nil
	}

	value, err := parseInt(raw)
	if err == emptyString {
		return nil
	}
	if err != nil {
		return err
	}

	values[key] = value
	return nil
}

func insertRawFloat(values ValuesMap, key string, raw interface{}) (err error) {
	defer func() { err = insertErrorValue(key, err) }()

	if raw == nil {
		return nil
	}

	value, err := parseFloat(raw)
	if err == emptyString {
		return nil
	}
	if err != nil {
		return err
	}

	values[key] = value
	return nil
}
