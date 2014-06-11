package status

type watcher struct {
	watchUrl      string          // Wildcard URL to watch.
	lastSeen      map[string]int  // Map expanded URL to revision of last value.
	updateChannel chan UrlMatches // Channel to notify clients.
}

// Public method to create a watcher. The channel received will have a
// notification available immediately, and will receive another after every
// change affecting the specified URL. If multiple updates happen and the
// channel is not read from, then intermediate updates may be lost.
func (s *Status) WatchForUpdate(url string) (Channel <-chan UrlMatches, e error) {
	if _, e = parseUrl(url); e != nil {
		return nil, e
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	// Create the channel
	notifyChannel := make(chan UrlMatches, 1)

	w := &watcher{
		watchUrl:      url,
		lastSeen:      map[string]int{"bad_url_to_force_no_match": 0},
		updateChannel: notifyChannel,
	}

	// Add new watcher to Status.
	s.watchers = append(s.watchers, w)

	// Do an initial look for updates to populate lastSeen, and send initial
	// update notifcation.
	w.checkForUpdate(s)

	return notifyChannel, nil
}

// Stop receiving updates on the given channel.
func (s *Status) ReleaseWatch(wc <-chan UrlMatches) {
	s.lock.Lock()
	defer s.lock.Unlock()

	trimmedWatchers := make([]*watcher, 0, len(s.watchers))

	for _, w := range s.watchers {
		if wc != w.updateChannel {
			trimmedWatchers = append(trimmedWatchers, w)
		}
	}

	s.watchers = trimmedWatchers
}

// Compare two map[string]int structures to see if they contain identical
// values. This is used to see if what a watch has been updated or not.
func lastSeenEqual(left, right map[string]int) bool {
	if len(left) != len(right) {
		return false
	}

	for k, v := range left {
		if w, ok := right[k]; !ok || v != w {
			return false
		}
	}

	return true
}

// Method to check one watch to see if a notification should be sent.
func (w *watcher) checkForUpdate(status *Status) {
	matches, e := status.getMatchingUrls(w.watchUrl)
	if e != nil {
		panic(e) // This is supposed to be impossible.
	}

	currentSeen := map[string]int{}

	for url, match := range matches {
		currentSeen[url] = match.revision
	}

	// If the list of seen values has changed,
	if !lastSeenEqual(w.lastSeen, currentSeen) {
		w.lastSeen = currentSeen

		select {
		case <-w.updateChannel: // Clear the channel if it hasn't been read from.
		default: // Read nothing if it's empty.
		}

		w.updateChannel <- matches
	}
}

// Method to check all watchers to see if notifications should be sent.
func (s *Status) checkWatchers() {
	for _, w := range s.watchers {
		w.checkForUpdate(s)
	}
}
