package rules

import ()

// Actions do something. They normally do something to a specific type
//   of component.
//
//
// * set - Set a status URI with a value.
//   * dest - Value to update on the component. ("foo")
//   * value - Value to write.
// * wol - Issue a Wake On Lan request to a given component.
// * ping - Ping a host, and store result.
//   * host - Status URI of the host component to ping. Result stored in <host>/up as a boolean. The result is NOT immediately available. Host can contain wildcards in it's path.
// * fetch_url - Fetch the specified URL.
//   * url - Url to fetch.
//   * download_name - Optional field. Name of file inside system downloads directy in which to store the downloaded value. '{time}' in the name will be filled in with a unique time based number.
// * email - Send email.
//   * to - Address to send email too.
//   * subject - Optional subject string.
//   * body - Optional body string.
//   * attachments - List of attachments.
//     * url - URL to fetch and attach to email.
//     * download_name - Name to download and attach as. Follows same rules as fetch_url:download_name.
//     * preserve - optional flag to keep in downloads directory.

// This is the signature of an action method.
type Action func(actionUrl, componentUrl string) (e error)

func actionSet(actionUrl, componentUrl string) (e error) {

	// dest string
	// Value interface{}

	return nil
}

// Send a Wake On Lan request to a component. The component must have a "mac"
// value defined with is the components network mac address.
func actionWol(actionUrl, componentUrl string) (e error) {
	// mac string

	return nil
}

// Ping a component, and set the "up" value on component to true or false. The
// name of the component is the name to ping. The "up" value is updated in the
// background after an arbitrary delay, not right away.
func actionPing(actionUrl, componentUrl string) (e error) {
	// Component Name

	// Component Up

	return nil
}

// Fetch a URL. If DownloadName is not present, the result is thrown away.
// otherwise,
// Happens ansynronously.
// Does not require a component to fire.
func actionFetch(actionUrl, componentUrl string) (e error) {
	// Url          string
	// DownloadName string

	return nil
}

func actionEmail(actionUrl, componentUrl string) (e error) {
	//   * to - Address to send email too.
	//   * subject - Optional subject string.
	//   * body - Optional body string.
	//   * attachments - List of attachments.
	//     * url - URL to fetch and attach to email.
	//     * download_name - Name to download and attach as. Follows same rules as fetch_url:download_name.
	//     * preserve - optional flag to keep in downloads directory.

	return nil
}
