package status

type watcher struct {
	watchUrl      string            // Wildcard URL to watch.
	lastSeen      map[string]int    // Map expanded URL to revision of last value.
	updateChannel chan<- UrlMatches // Channel to notify clients.
}

// Public method to create a watcher.
func (s *Status) WatchForUpdate(url string) (<-chan UrlMatches, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	// Discover the current state of our matches.
	matches, e := s.getMatchingUrls(url)
	if e != nil {
		return nil, e
	}

	// Convert to our storage format.
	currentSeen := map[string]int{}
	for url, match := range matches {
		currentSeen[url] = match.revision
	}

	// Create the channel
	notifyChannel := make(chan UrlMatches, 100)

	w := watcher{
		watchUrl:      url,
		lastSeen:      currentSeen,
		updateChannel: notifyChannel,
	}

	// Add new watcher to Status.
	s.watchers = append(s.watchers, &w)

	return notifyChannel, nil
}

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
func (w *watcher) checkForUpdate(status *Status) (e error) {
	matches, e := status.getMatchingUrls(w.watchUrl)
	if e != nil {
		return e
	}

	currentSeen := map[string]int{}

	for url, match := range matches {
		currentSeen[url] = match.revision
	}

	// If the list of seen values has changed,
	if !lastSeenEqual(w.lastSeen, currentSeen) {
		w.lastSeen = currentSeen
		w.updateChannel <- matches
	}

	return nil
}

// Method to check all watchers to see if notifications should be sent.
func (s *Status) checkWatchers() {
	for _, w := range s.watchers {
		e := w.checkForUpdate(s)
		if e != nil {
			panic(e) // This is supposed to be impossible.
		}
	}
}
