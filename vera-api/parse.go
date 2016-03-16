package veraapi

import (
	"encoding/json"
	"fmt"
)

func parseVeraDataJson(bodyText []byte) (result *rawResponse, err error) {
	result = &rawResponse{}

	err = json.Unmarshal(bodyText, result)
	if err != nil {
		return nil, fmt.Errorf("Can't unmarshel devices: %s\n%s", err, string(bodyText))
	}

	return result, nil
}

func parseSectionsMap(raw []rawSection) (result sectionMap, err error) {
	var id int
	defer func() {
		err = insertErrorArea("Sections", err)
		err = insertErrorId(id, err)
	}()

	result = sectionMap{}

	for _, s := range raw {
		id, err = parseInt(s.Id)
		if err != nil {
			return nil, err
		}
		result[id] = section{id, s.Name}
	}

	return result, nil
}

func parseRoomsMap(raw []rawRoom, sections sectionMap) (result roomMap, err error) {
	var id int
	defer func() {
		err = insertErrorArea("Rooms", err)
		err = insertErrorId(id, err)
	}()

	result = roomMap{}

	for _, r := range raw {
		id, err = parseInt(r.Id)
		if err != nil {
			return nil, err
		}
		sectionId, err := parseInt(r.Section)
		if err != nil {
			return nil, err
		}
		section, ok := sections[sectionId]
		if !ok {
			return nil, fmt.Errorf("Room (%d) has unknown sectionId %d", id, sectionId)
		}

		result[id] = room{id, r.Name, section.name}
	}

	return result, nil
}

func parseCategoryMap(raw []rawCategory) (result categoryMap, err error) {
	var id int
	defer func() {
		err = insertErrorArea("Categories", err)
		err = insertErrorId(id, err)
	}()

	result = categoryMap{}

	for _, c := range raw {
		id, err = parseInt(c.Id)
		if err != nil {
			return nil, err
		}
		result[id] = category{id, c.Name}
	}

	return result, nil
}

func parseScenesMap(raw []rawScene, rooms roomMap) (result sceneMap, err error) {
	var id int
	defer func() {
		err = insertErrorArea("Scenes", err)
		err = insertErrorId(id, err)
	}()

	result = sceneMap{}

	for _, s := range raw {
		id, err = parseInt(s.Id)
		if err != nil {
			return nil, err
		}
		roomId, err := parseInt(s.Room)
		if err != nil {
			return nil, err
		}
		room, ok := rooms[roomId]
		if !ok && roomId != 0 {
			return nil, fmt.Errorf("Scene (%d) has unknown roomId %d", id, roomId)
		}
		active, err := parseBool(s.Active)
		if err != nil {
			return nil, err
		}

		result[id] = scene{id, s.Name, room.name, active}
	}

	return result, nil
}

func updateScenesMap(raw []rawScene, scenes sceneMap) (err error) {
	var id int
	defer func() {
		err = insertErrorArea("Scenes", err)
		err = insertErrorId(id, err)
	}()

	for _, s := range raw {
		id, err = parseInt(s.Id)
		if err != nil {
			return err
		}

		scene, ok := scenes[id]
		if !ok {
			return fmt.Errorf("Received update for unknown scene: %d", id)
		}

		scene.Active, err = parseBool(s.Active)
		if err != nil {
			return err
		}

		scenes[id] = scene
	}
	return nil
}

func parseDevieValues(raw rawDevice, values ValuesMap) (err error) {
	if err = insertRawBool(values, "armed", raw.Armed); err != nil {
		return err
	}
	if err = insertRawBool(values, "armedtripped", raw.Armedtripped); err != nil {
		return err
	}
	if err = insertRawString(values, "lasttrip", raw.Lasttrip); err != nil {
		return err
	}
	if err = insertRawBool(values, "tripped", raw.Tripped); err != nil {
		return err
	}
	if err = insertRawFloat(values, "temperature", raw.Temperature); err != nil {
		return err
	}
	if err = insertRawInt(values, "light", raw.Light); err != nil {
		return err
	}
	if err = insertRawInt(values, "humidity", raw.Humidity); err != nil {
		return err
	}
	if err = insertRawBool(values, "status", raw.Status); err != nil {
		return err
	}
	if err = insertRawInt(values, "level", raw.Level); err != nil {
		return err
	}
	if err = insertRawBool(values, "locked", raw.Locked); err != nil {
		return err
	}
	if err = insertRawInt(values, "batterylevel", raw.Batterylevel); err != nil {
		return err
	}
	if err = insertRawFloat(values, "watts", raw.Watts); err != nil {
		return err
	}

	return nil
}

func parseDevicesMap(raw []rawDevice, categories categoryMap, rooms roomMap) (result deviceMap, err error) {
	var id int
	defer func() {
		err = insertErrorArea("Devices", err)
		err = insertErrorId(id, err)
	}()

	result = deviceMap{}

	for _, d := range raw {
		id, err = parseInt(d.Id)
		if err != nil {
			return nil, insertErrorValue("id", err)
		}

		categoryId, err := parseInt(d.Category)
		if err != nil {
			return nil, insertErrorValue("category", err)
		}
		category, ok := categories[categoryId]
		if !ok && categoryId != 0 {
			return nil, fmt.Errorf("Device (%d) has unknown categoryId %d", id, categoryId)
		}

		subCategoryId, err := parseInt(d.Subcategory)
		if err != nil {
			return nil, insertErrorValue("subcategory", err)
		}
		subcategory, ok := categories[subCategoryId]
		// Ignore lookup values. We sometimes get bad subcategoryId's.
		roomId, err := parseInt(d.Room)
		if err != nil {
			return nil, insertErrorValue("room", err)
		}
		room, ok := rooms[roomId]
		if !ok {
			return nil, fmt.Errorf("Device (%d) has unknown roomId %d", id, roomId)
		}

		// Any device specific values.
		values := ValuesMap{}
		if err = parseDevieValues(d, values); err != nil {
			return nil, err
		}

		result[id] = Device{id, d.Name, category.name, subcategory.name, room.name, values}
	}

	return result, nil
}

func updateDevicesMap(raw []rawDevice, devices deviceMap) (err error) {
	var id int
	defer func() {
		err = insertErrorArea("Devices", err)
		err = insertErrorId(id, err)
	}()

	for _, d := range raw {
		id, err = parseInt(d.Id)
		if err != nil {
			return insertErrorValue("id", err)
		}

		device, ok := devices[id]
		if !ok {
			return fmt.Errorf("Received update for unknown device: %d", id)
		}

		if err = parseDevieValues(d, device.Values); err != nil {
			return err
		}
	}

	return nil
}

func parsePartialData(raw *rawResponse, previous *parseResult) (result *parseResult, err error) {
	if previous == nil || !previous.full {
		// If there is no valid previous state, we can't process a partial.
		return nil, fmt.Errorf("Received partial result with no previous status.")
	}

	result = previous.copy()

	loadtime, err := parseInt(raw.LoadTime)
	if err != nil {
		return nil, insertErrorValue("loadtime", err)
	}
	if loadtime != result.loadtime {
		return nil, fmt.Errorf("Partial update to %d doesn't match known history %d", loadtime, result.loadtime)
	}
	if result.dataversion, err = parseInt(raw.DataVersion); err != nil {
		return nil, insertErrorValue("dateaversion", err)
	}

	if err = updateScenesMap(raw.Scenes, result.scenes); err != nil {
		return nil, err
	}

	if err = updateDevicesMap(raw.Devices, result.devices); err != nil {
		return nil, err
	}

	return result, nil
}

func parseVeraData(bodyText []byte, previous *parseResult) (result *parseResult, err error) {
	// Parse the json into 'raw' format.
	raw, err := parseVeraDataJson(bodyText)
	if err != nil {
		return nil, err
	}

	// Parse top level values that should always be present.
	full, err := parseBool(raw.Full)
	if err != nil {
		return nil, insertErrorValue("full", err)
	}

	if !full {
		return parsePartialData(raw, previous)
	}

	result = newParseResult()
	result.full = full

	if result.loadtime, err = parseInt(raw.LoadTime); err != nil {
		return nil, insertErrorValue("loadtime", err)
	}
	if result.dataversion, err = parseInt(raw.DataVersion); err != nil {
		return nil, insertErrorValue("dateaversion", err)
	}

	if result.sections, err = parseSectionsMap(raw.Sections); err != nil {
		return nil, err
	}

	if result.rooms, err = parseRoomsMap(raw.Rooms, result.sections); err != nil {
		return nil, err
	}

	if result.categories, err = parseCategoryMap(raw.Categories); err != nil {
		return nil, err
	}

	if result.scenes, err = parseScenesMap(raw.Scenes, result.rooms); err != nil {
		return nil, err
	}

	if result.devices, err = parseDevicesMap(raw.Devices, result.categories, result.rooms); err != nil {
		return nil, err
	}

	return result, nil
}
