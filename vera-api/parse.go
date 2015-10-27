package veraapi

import (
	"encoding/json"
	"fmt"
)

type section struct {
	id   int
	name string
}

type room struct {
	id      int
	name    string
	section string
}

type scene struct {
	Id     int
	Name   string
	Room   string
	Active bool
}

type category struct {
	id   int
	name string
}

type parseResult struct {
	loadtime    int
	dataversion int
	full        bool

	sections   map[int]section
	rooms      map[int]room
	scenes     map[int]scene
	categories map[int]category
	devices    map[int]Device
}

func newParseResult() *parseResult {
	return &parseResult{
		0, 0, false,
		map[int]section{},
		map[int]room{},
		map[int]scene{},
		map[int]category{},
		map[int]Device{},
	}
}

func parseVeraData(bodyText []byte) (result *parseResult, err error) {

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
	var rawResponse struct {
		Full        interface{}
		LoadTime    interface{}
		DataVersion interface{}
		Sections    []rawSection
		Rooms       []rawRoom
		Scenes      []rawScene
		Categories  []rawCategory
		Devices     []rawDevice
	}

	result = newParseResult()

	err = json.Unmarshal(bodyText, &rawResponse)
	if err != nil {
		return nil, fmt.Errorf("Can't unmarshel devices: %s\n%s", err, string(bodyText))
	}

	result.full, err = parseBool(rawResponse.Full)

	if result.full, err = parseBool(rawResponse.Full); err != nil {
		return nil, err
	}

	if result.loadtime, err = parseInt(rawResponse.LoadTime); err != nil {
		return nil, err
	}

	if result.dataversion, err = parseInt(rawResponse.DataVersion); err != nil {
		return nil, err
	}

	// sections
	for _, s := range rawResponse.Sections {
		id, err := parseInt(s.Id)
		if err != nil {
			return nil, err
		}
		result.sections[id] = section{id, s.Name}
	}

	// rooms
	for _, r := range rawResponse.Rooms {
		id, err := parseInt(r.Id)
		if err != nil {
			return nil, err
		}
		sectionId, err := parseInt(r.Section)
		if err != nil {
			return nil, err
		}
		section, ok := result.sections[sectionId]
		if !ok {
			return nil, fmt.Errorf("Room (%d) has unknown sectionId %d in:\n%s", id, sectionId, string(bodyText))
		}

		result.rooms[id] = room{id, r.Name, section.name}
	}

	// scenes
	for _, s := range rawResponse.Scenes {
		id, err := parseInt(s.Id)
		if err != nil {
			return nil, err
		}
		roomId, err := parseInt(s.Room)
		if err != nil {
			return nil, err
		}
		room, ok := result.rooms[roomId]
		if !ok && roomId != 0 {
			return nil, fmt.Errorf("Scene (%d) has unknown roomId %d in:\n%s", id, roomId, string(bodyText))
		}
		active, err := parseBool(s.Active)
		if err != nil {
			return nil, err
		}

		result.scenes[id] = scene{id, s.Name, room.name, active}
	}

	// categories
	for _, c := range rawResponse.Categories {
		id, err := parseInt(c.Id)
		if err != nil {
			return nil, err
		}
		result.categories[id] = category{id, c.Name}
	}

	for _, d := range rawResponse.Devices {
		id, err := parseInt(d.Id)
		if err != nil {
			return nil, err
		}

		categoryId, err := parseInt(d.Category)
		if err != nil {
			return nil, err
		}
		category, ok := result.categories[categoryId]
		if !ok && categoryId != 0 {
			return nil, fmt.Errorf("Device (%d) has unknown categoryId %d in:\n%s", id, categoryId, string(bodyText))
		}

		subCategoryId, err := parseInt(d.Subcategory)
		if err != nil {
			return nil, err
		}
		subcategory, ok := result.categories[subCategoryId]
		// Ignore lookup values. We sometimes get bad subcategoryId's.
		roomId, err := parseInt(d.Room)
		if err != nil {
			return nil, err
		}
		room, ok := result.rooms[roomId]
		if !ok {
			return nil, fmt.Errorf("Device (%d) has unknown roomId %d in:\n%s", id, roomId, string(bodyText))
		}

		// Any device specific values.
		values := ValuesMap{}

		if err = insertRawBool(values, "armed", d.Armed); err != nil {
			return nil, err
		}
		if err = insertRawBool(values, "armedtripped", d.Armedtripped); err != nil {
			return nil, err
		}
		if err = insertRawString(values, "lasttrip", d.Lasttrip); err != nil {
			return nil, err
		}
		if err = insertRawBool(values, "tripped", d.Tripped); err != nil {
			return nil, err
		}
		if err = insertRawFloat(values, "temperature", d.Temperature); err != nil {
			return nil, err
		}
		if err = insertRawInt(values, "light", d.Light); err != nil {
			return nil, err
		}
		if err = insertRawInt(values, "humidity", d.Humidity); err != nil {
			return nil, err
		}
		if err = insertRawBool(values, "status", d.Status); err != nil {
			return nil, err
		}
		if err = insertRawInt(values, "level", d.Level); err != nil {
			return nil, err
		}
		if err = insertRawBool(values, "locked", d.Locked); err != nil {
			return nil, err
		}
		if err = insertRawInt(values, "batterylevel", d.Batterylevel); err != nil {
			return nil, err
		}
		if err = insertRawFloat(values, "watts", d.Watts); err != nil {
			return nil, err
		}

		result.devices[id] = Device{id, d.Name, category.name, subcategory.name, room.name, values}
	}

	return result, nil
}
