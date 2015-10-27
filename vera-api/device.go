package veraapi

import ()

type Device struct {
	// General device description.
	Id          int
	Name        string
	Category    string
	Subcategory string
	Room        string

	// Connectivity State
	State int

	// Values produced by device.
	Armed        string
	Batterylevel string
	Status       string
	Temperature  string
	Light        string
	Humidity     string
	Tripped      string
	Armedtripped string
	Lasttrip     string
	Level        string
	Locked       string
}
