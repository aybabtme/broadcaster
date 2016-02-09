package broadcaster

import (
	"sync"

	"golang.org/x/net/context"
)

type broadcaster struct {
	mu       sync.Mutex
	send     chan Event
	register chan *listener
}

// New broadcaster that sends Events in its internal goroutine. It does
// not preserve a backlog and will not attempt to update new listeners.
func New() Broadcaster {
	b := &broadcaster{
		send:     make(chan Event, 1),
		register: make(chan *listener),
	}

	go b.broadcast()

	return b
}

func (b *broadcaster) broadcast() {
	listeners := make(map[*listener]struct{}, 0)
	defer b.closeListeners(listeners)

	for {
		select {
		case newList := <-b.register:
			listeners[newList] = struct{}{}
			continue
		case ev, ok := <-b.send:
			if !ok {
				return
			}
			b.sendToListeners(listeners, ev)
		}
	}

}

func (b *broadcaster) sendToListeners(listeners map[*listener]struct{}, ev Event) {
	for listn := range listeners {
		sendc, ok := <-listn.listnC
		if !ok {
			delete(listeners, listn)
			continue
		}
		select {
		case sendc <- ev:
		default:
		}
	}
}

func (b *broadcaster) closeListeners(listeners map[*listener]struct{}) {
	for listn := range listeners {
		sendc, ok := <-listn.listnC
		if !ok {
			continue
		}
		close(sendc)
	}
}

// Close the broadcaster goroutine and stop broadcasting. Can only be
// be called once.
func (b *broadcaster) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()
	close(b.send)
}

// Send an Event to all the listeners. Sending on a closed broacaster
// will panic.
func (b *broadcaster) Send(ctx context.Context, e Event) {
	b.mu.Lock()
	defer b.mu.Unlock()
	select {
	case <-ctx.Done():
	case b.send <- e:
	}
}

// Listen returns a listener for this broadcaster. The listener will
// pickup the broadcast from where it is right now, without any history.
func (b *broadcaster) Listen() Listener {
	b.mu.Lock()
	defer b.mu.Unlock()
	l := &listener{
		messenger: make(chan Event, 1),
		listnC:    make(chan chan Event, 1),
	}
	b.register <- l
	return l
}

type listener struct {
	messenger chan Event
	listnC    chan chan Event
}

// Next blocks until the next Event from the broadcaster. It returns
// false if the broadcaster is closed.
//
// Calling Next after a call to Close is an error and will panic.
func (l *listener) Next(ctx context.Context) (Event, bool) {
	select {
	case <-ctx.Done():
		return nil, false
	case l.listnC <- l.messenger:
	}
	select {
	case <-ctx.Done():
		return nil, false
	case ev, hasMore := <-l.messenger:
		return ev, hasMore
	}
}

// Close removes this listener from the broadcaster's listeners.
func (l *listener) Close() {
	close(l.listnC)
}
