package veraapi

import (
	"fmt"
)

type parseError struct {
	area  string
	id    int
	value string
	msg   string
}

func (e parseError) Error() string {
	result := ""
	if e.area != "" {
		result += fmt.Sprintf("%s: ", e.area)
	}
	if e.id != 0 {
		result += fmt.Sprintf("%d: ", e.id)
	}
	if e.value != "" {
		result += fmt.Sprintf("%s: ", e.value)
	}

	return result + e.msg
}

func insertErrorArea(area string, err error) error {
	if err == nil {
		return nil
	}

	pe, ok := err.(parseError)
	pe.area = area
	if !ok {
		pe.msg = err.Error()
	}
	return pe
}

func insertErrorId(id int, err error) error {
	if err == nil {
		return nil
	}

	pe, ok := err.(parseError)
	pe.id = id
	if !ok {
		pe.msg = err.Error()
	}
	return pe
}

func insertErrorValue(value string, err error) error {
	if err == nil {
		return nil
	}

	pe, ok := err.(parseError)
	pe.value = value
	if !ok {
		pe.msg = err.Error()
	}
	return pe
}
