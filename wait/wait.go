package wait

import (
	"time"
)

func Wait(timeout time.Duration, ready func() bool) bool {
	end_wait := time.Now().Add(timeout)
	for time.Now().Before(end_wait) {
		if ready() {
			return true
		}

		// If we didn't match, wait a little and try again.
		time.Sleep(time.Millisecond)
	}

	return false
}
