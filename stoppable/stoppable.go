package stoppable

type Stoppable interface {
	// Stop the event handler goroutine associated with this object.
	Stop()
}

// Compose this type.
type Base struct {
	StopChan chan bool
}

// Use this helper to construct the composed value.
func NewBase() Base {
	return Base{make(chan bool)}
}

// Override only if you need additional work here.
func (b *Base) Stop() {
	b.StopChan <- true
	<-b.StopChan
}

// This is a default (and uninteresting) background process. Override as needed.
// call go x.Handler() in your constructor.
func (b *Base) Handler() {
	for {
		select {
		case <-b.StopChan:
			b.StopChan <- true
			return
		}
	}
}
