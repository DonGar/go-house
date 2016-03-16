package veraapi

import ()

type section struct {
	id   int
	name string
}
type sectionMap map[int]section

type room struct {
	id      int
	name    string
	section string
}
type roomMap map[int]room

type scene struct {
	Id     int
	Name   string
	Room   string
	Active bool
}
type sceneMap map[int]scene

type category struct {
	id   int
	name string
}
type categoryMap map[int]category

type deviceMap map[int]Device

type parseResult struct {
	loadtime    int
	dataversion int
	full        bool

	sections   sectionMap
	rooms      roomMap
	scenes     sceneMap
	categories categoryMap
	devices    deviceMap
}

func newParseResult() *parseResult {
	return &parseResult{
		0, 0, false,
		sectionMap{},
		roomMap{},
		sceneMap{},
		categoryMap{},
		deviceMap{},
	}
}

func (r *parseResult) copy() *parseResult {
	sections := sectionMap{}
	for k, v := range r.sections {
		sections[k] = v
	}

	rooms := roomMap{}
	for k, v := range r.rooms {
		rooms[k] = v
	}

	scenes := sceneMap{}
	for k, v := range r.scenes {
		scenes[k] = v
	}

	categories := categoryMap{}
	for k, v := range r.categories {
		categories[k] = v
	}

	devices := deviceMap{}
	for k, v := range r.devices {
		// The values copy is copied by reference, so deep copy it.
		values := ValuesMap{}
		for k, v := range v.Values {
			values[k] = v
		}

		v.Values = values
		devices[k] = v
	}

	return &parseResult{
		r.loadtime,
		r.dataversion,
		r.full,
		sections,
		rooms,
		scenes,
		categories,
		devices,
	}
}
