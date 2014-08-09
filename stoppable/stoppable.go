package stoppable

type Stoppable interface {
	// Stop the event handler goroutine associated with this object.
	Stop()
}

type Base struct {
	StopChan chan bool
}

func NewBase() Base {
	return Base{make(chan bool)}
}

func (b *Base) Stop() {
	b.StopChan <- true
	<-b.StopChan
}

// This is a default (and uninteresting) background process.
func (b *Base) Handler() {
	for {
		select {
		case <-b.StopChan:
			b.StopChan <- true
			return
		}
	}
}
