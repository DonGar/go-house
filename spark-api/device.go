package sparkapi

import ()

type Device struct {
	Id        string
	Name      string
	LastHeard string `json:"last_heard"`
	Connected bool
	Variables map[string]interface{} // Unknown values are nil.
	Functions []string
}

func (d Device) Copy() (result Device) {
	// Copy with shared pointers to maps/slices.
	result = d

	// Copy Variables map.
	result.Variables = map[string]interface{}{}
	for k, v := range d.Variables {
		result.Variables[k] = v
	}

	// Copy Functions slice.
	result.Functions = make([]string, len(d.Functions))
	for j := range d.Functions {
		result.Functions[j] = d.Functions[j]
	}

	return
}
