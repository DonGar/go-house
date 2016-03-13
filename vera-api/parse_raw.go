package veraapi

// Json parser helper types not used outside of the parsing code.

type rawSection struct {
	Id   interface{}
	Name string
}
type rawRoom struct {
	Id      interface{}
	Name    string
	Section interface{}
}
type rawScene struct {
	Id     interface{}
	Name   string
	Room   interface{}
	Active interface{}
}
type rawCategory struct {
	Id   interface{}
	Name string
}
type rawDevice struct {
	Id          interface{}
	Name        string
	Category    interface{}
	Subcategory interface{}
	Room        interface{}

	// Optional device specific status variables.

	// Motion Sensor Values
	Armed        interface{} // 0 or 1
	Armedtripped interface{} // 0 or 1
	Lasttrip     interface{} // Seconds since unix epoch
	Tripped      interface{} // 0 or 1 (also can be on/off of a switch)

	// Environmental Sensors
	Temperature interface{} // (85.3), unit is "temperature" of top level response
	Light       interface{} // How much light is seen. (0-100)
	Humidity    interface{} // 0-100, percentage

	// Switches and Dimmers
	Status interface{} // 0 or 1. Represents switch state. Ignore if tripped present.
	Level  interface{} // Dimming setting (0-100). if Non-zero, then Status is 1.

	Locked interface{} // 0 or 1. Door lock is locks.

	Batterylevel interface{} // 0-100, percentage of battery remaining.

	Watts interface{} // (1.649) Number of watts currently being consumed.
}

// Parse the response.
type rawResponse struct {
	Full        interface{}
	LoadTime    interface{}
	DataVersion interface{}
	Sections    []rawSection
	Rooms       []rawRoom
	Scenes      []rawScene
	Categories  []rawCategory
	Devices     []rawDevice
}
