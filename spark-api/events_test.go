package sparkapi

import (
	"bufio"
	"gopkg.in/check.v1"
	"io"
	"strings"
)

func (suite *MySuite) TestReadLine(c *check.C) {
	baseReader := strings.NewReader("\nFoo\n\n  Bar  \n\n")
	reader := bufio.NewReader(baseReader)

	line, err := readLine(reader)
	c.Check(err, check.IsNil)
	c.Check(line, check.Equals, "")

	line, err = readLine(reader)
	c.Check(err, check.IsNil)
	c.Check(line, check.Equals, "Foo")

	line, err = readLine(reader)
	c.Check(err, check.IsNil)
	c.Check(line, check.Equals, "")

	line, err = readLine(reader)
	c.Check(err, check.IsNil)
	c.Check(line, check.Equals, "Bar")

	line, err = readLine(reader)
	c.Check(err, check.IsNil)
	c.Check(line, check.Equals, "")

	line, err = readLine(reader)
	c.Check(err, check.Equals, io.EOF)
	c.Check(line, check.Equals, "")

}

func (suite *MySuite) TestOpenEventConnection(c *check.C) {
	if !*network {
		c.Skip("-network tests not enabled.")
	}

	sa := NewSparkApi(TEST_USER, TEST_PASS)
	defer sa.Stop()

	response, reader, err := sa.openEventConnection()
	c.Check(response, check.NotNil)
	c.Check(reader, check.NotNil)
	c.Check(err, check.IsNil)

	err = response.Body.Close()
	c.Check(err, check.IsNil)
}

func (suite *MySuite) TestParseEvent(c *check.C) {
	baseReader := strings.NewReader(`
event: Punch
data: {"data":"Foo","ttl":"60","published_at":"2014-10-18T06:05:45.100Z","coreid":"55ff6f065075555320371887"}
`)
	reader := bufio.NewReader(baseReader)

	// First new line, is a nil Event with no error.
	event, err := parseEvent(reader)
	c.Check(err, check.IsNil)
	c.Check(event, check.IsNil)

	expectedEvent := &Event{
		Name:         "Punch",
		Data:         "Foo",
		Published_at: "2014-10-18T06:05:45.100Z",
		CoreId:       "55ff6f065075555320371887",
	}

	// Parse event without error.
	event, err = parseEvent(reader)
	c.Check(err, check.IsNil)
	c.Check(event, check.DeepEquals, expectedEvent)

	// Final Close.
	event, err = parseEvent(reader)
	c.Check(err, check.Equals, io.EOF)
	c.Check(event, check.IsNil)

}
