/*
Package broadcaster provides a way to send events to multiple listeners.
Listeners can be closed to announce that they do not wish to receive
further events, and broadcasters can announce to their listeners
that no further events will be sent (and therefore the listeners should
be released).

There are two implementations. The first is a broadcaster that sends
the events that it receives to all its listeners.

The second does the same thing, but also keeps an history of the last events
and brings up to date listeners that just subscribed.
*/
package broadcaster

import "golang.org/x/net/context"

// Broadcaster sends Events it receives to all its listeners.
//
// Failing to close a broadcaster that is not needed anymore is a
// memory leak.
type Broadcaster interface {
	// Close the broadcaster goroutine and stop broadcasting. Can only be
	// be called once.
	Close()
	// Send an Event to all the listeners. Sending on a closed broacaster
	// will panic.
	Send(ctx context.Context, e Event)
	// Listen returns a listener for this broadcaster.
	Listen() Listener
}

// Event that can be broadcasted.
type Event interface{}

// Listener can read Events and be closed when the owner wishes to
// stop listening.
//
// Failing to close a listener once you are done with it is a memory leak.
type Listener interface {
	// Next blocks until the next Event from the broadcaster. It returns
	// false if the broadcaster is closed.
	//
	// Calling Next after a call to Close is an error and will panic.
	Next(ctx context.Context) (Event, bool)
	// Close removes this listener from the broadcaster's listeners.
	Close()
}
