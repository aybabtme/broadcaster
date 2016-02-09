package broadcaster

import (
	"sync"
	"time"

	"golang.org/x/net/context"
)

type backlogCaster struct {
	mu       sync.Mutex
	timeout  time.Duration
	backlog  int
	send     chan Event
	register chan *listener
}

// NewBacklog broadcaster that sends Events in its internal goroutine.
// When a new listener is given out, it will have an history of recent
// events.
func NewBacklog(backlog int) Broadcaster {
	b := &backlogCaster{
		backlog:  backlog,
		send:     make(chan Event, 1),
		register: make(chan *listener),
	}

	go b.broadcast()

	return b
}

func (b *backlogCaster) broadcast() {
	listeners := make(map[*listener]struct{}, 0)
	backlog := &leakingQueue{max: b.backlog}
	defer b.closeListeners(listeners)

	for {
		select {
		case newList := <-b.register:
			if b.updateListener(newList, backlog) {
				listeners[newList] = struct{}{}
			}
			continue
		case ev, ok := <-b.send:
			if !ok {
				return
			}
			b.sendToListeners(listeners, ev)
			backlog.enqueue(ev)
		}
	}

}

func (b *backlogCaster) updateListener(l *listener, backlog *leakingQueue) bool {
	timeOut := time.NewTimer(b.timeout)
	for _, e := range backlog.data {
		send, ok := <-l.listnC
		if !ok {
			return false
		}
		select {
		case send <- e:
		default:
		}

	}
	timeOut.Stop()
	return true
}

func (b *backlogCaster) sendToListeners(listeners map[*listener]struct{}, ev Event) {
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

func (b *backlogCaster) closeListeners(listeners map[*listener]struct{}) {
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
func (b *backlogCaster) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()
	close(b.send)
}

// Send an Event to all the listeners. Sending on a closed broacaster
// will panic.
func (b *backlogCaster) Send(ctx context.Context, e Event) {
	b.mu.Lock()
	defer b.mu.Unlock()
	select {
	case <-ctx.Done():
	case b.send <- e:
	}
}

// Listen returns a listener for this broadcaster. The listener will
// updated with the backlog of the broadcaster.
func (b *backlogCaster) Listen() Listener {
	b.mu.Lock()
	defer b.mu.Unlock()
	// big enough to receive the full history without blocking
	// the broadcaster
	bufferedMessenger := make(chan Event, b.backlog)

	l := &listener{
		messenger: bufferedMessenger,
		listnC:    make(chan chan Event, 1),
	}
	b.register <- l
	return l
}

type leakingQueue struct {
	max  int
	data []Event
}

func (l *leakingQueue) enqueue(e Event) {
	if len(l.data) == l.max-1 {
		l.data = l.data[1:]
	}
	l.data = append(l.data, e)
}
