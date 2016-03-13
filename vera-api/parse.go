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
	result = sectionMap{}

	for _, s := range raw {
		id, err := parseInt(s.Id)
		if err != nil {
			return nil, err
		}
		result[id] = section{id, s.Name}
	}

	return result, nil
}

func parseRoomsMap(raw []rawRoom, sections sectionMap) (result roomMap, err error) {
	result = roomMap{}

	for _, r := range raw {
		id, err := parseInt(r.Id)
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

func parseScenesMap(raw []rawScene, rooms roomMap) (result sceneMap, err error) {
	result = sceneMap{}

	for _, s := range raw {
		id, err := parseInt(s.Id)
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

func parseCategoryMap(raw []rawCategory) (result categoryMap, err error) {
	result = categoryMap{}

	for _, c := range raw {
		id, err := parseInt(c.Id)
		if err != nil {
			return nil, err
		}
		result[id] = category{id, c.Name}
	}

	return result, nil
}

func parseDevicesMap(raw []rawDevice, categories categoryMap, rooms roomMap) (result deviceMap, err error) {
	result = deviceMap{}

	for _, d := range raw {
		id, err := parseInt(d.Id)
		if err != nil {
			return nil, err
		}

		categoryId, err := parseInt(d.Category)
		if err != nil {
			return nil, err
		}
		category, ok := categories[categoryId]
		if !ok && categoryId != 0 {
			return nil, fmt.Errorf("Device (%d) has unknown categoryId %d", id, categoryId)
		}

		subCategoryId, err := parseInt(d.Subcategory)
		if err != nil {
			return nil, err
		}
		subcategory, ok := categories[subCategoryId]
		// Ignore lookup values. We sometimes get bad subcategoryId's.
		roomId, err := parseInt(d.Room)
		if err != nil {
			return nil, err
		}
		room, ok := rooms[roomId]
		if !ok {
			return nil, fmt.Errorf("Device (%d) has unknown roomId %d", id, roomId)
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

		result[id] = Device{id, d.Name, category.name, subcategory.name, room.name, values}
	}

	return result, nil
}

func parseVeraData(bodyText []byte) (result *parseResult, err error) {
	// Parse the json into 'raw' format.
	raw, err := parseVeraDataJson(bodyText)
	if err != nil {
		return nil, err
	}

	result = newParseResult()

	// Parse top level values that should always be present.
	if result.full, err = parseBool(raw.Full); err != nil {
		return nil, err
	}
	if result.loadtime, err = parseInt(raw.LoadTime); err != nil {
		return nil, err
	}
	if result.dataversion, err = parseInt(raw.DataVersion); err != nil {
		return nil, err
	}

	if result.sections, err = parseSectionsMap(raw.Sections); err != nil {
		return nil, err
	}

	if result.rooms, err = parseRoomsMap(raw.Rooms, result.sections); err != nil {
		return nil, err
	}

	if result.scenes, err = parseScenesMap(raw.Scenes, result.rooms); err != nil {
		return nil, err
	}

	if result.categories, err = parseCategoryMap(raw.Categories); err != nil {
		return nil, err
	}

	if result.devices, err = parseDevicesMap(raw.Devices, result.categories, result.rooms); err != nil {
		return nil, err
	}

	return result, nil
}
