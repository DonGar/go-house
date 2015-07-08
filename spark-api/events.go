package sparkapi

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

var EVENTS_URL string = SPARK_IO_URL + "v1/devices/events"

type Event struct {
	Name         string
	Data         string
	Published_at string
	CoreId       string
}

func readLine(reader *bufio.Reader) (string, error) {
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	// Trim off trailing newline.
	return strings.TrimSpace(line), nil
}

func (a *SparkApi) openEventConnection() (*http.Response, *bufio.Reader, error) {
	response, err := a.urlToResponseWithTokenRefresh(EVENTS_URL)
	if err != nil {
		return nil, nil, err
	}

	reader := bufio.NewReader(response.Body)

	// the first line should always be ":ok"
	line, err := readLine(reader)
	if err != nil {
		response.Body.Close()
		return nil, nil, err
	}

	if line != ":ok" {
		response.Body.Close()
		return nil, nil, fmt.Errorf("Received unexpected ok response: %s", line)
	}

	return response, reader, nil
}

func parseEvent(reader *bufio.Reader) (*Event, error) {
	// Sample event data:
	//   event: Punch
	//   data: {"data":"..","ttl":"60","published_at":"2014-10-18T06:05:45.100Z","coreid":"55ff6f065075555320371887"}

	result := Event{}

	line, err := readLine(reader)
	if line == "" || err != nil {
		return nil, err
	}

	// If we have an event line, parse it.
	if !strings.HasPrefix(line, "event: ") {
		return nil, fmt.Errorf("Received event with unexpected content: %s", line)
	}
	result.Name = line[7:]

	// Look for a data line.
	line, err = readLine(reader)
	if err != nil {
		return nil, err
	}

	if !strings.HasPrefix(line, "data: ") {
		return nil, fmt.Errorf("Received event with unexpected content: %s", line)
	}
	data := line[6:]

	// Parse the JSON from the data line.
	err = json.Unmarshal([]byte(data), &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
