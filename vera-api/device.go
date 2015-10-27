package veraapi

import ()

type ValuesMap map[string]interface{}

type Device struct {
	// General device description.
	Id          int
	Name        string
	Category    string
	Subcategory string
	Room        string

	// Any device specific values.
	Values ValuesMap

	// Known values:
	//   Motion Sensors.
	//     armed        bool  // 0 or 1
	//     armedtripped bool  // 0 or 1
	//     lasttrip     string // Seconds since unix epoch.
	//     tripped      bool  // 0 or 1

	//   Environmental Variables.
	//     temperature float64 // (85.3), unit is "temperature" of top level of response.
	//     light       int     // How much light is seen. (0-100)
	//     humidity    int     // 0-100, is percentage.

	//   Switch or dimmer settings.
	//     status bool // (0 or 1) Represent switch. Ignore if tripped present.
	//     level  int  // Dimming setting (0 - 100), if Non-zero, then Status is 1.

	//   Lock State.
	//     locked bool // Lock is locked. (0 - 1)

	//   Any battery device.
	//     batterylevel int // % battery remaining, 0 - 100.

	//   Power Usage
	//     watts

}
