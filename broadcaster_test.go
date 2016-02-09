package broadcaster

import (
	"log"
	"math"
	"sync"
	"testing"
	"time"

	"golang.org/x/net/context"
)

func TestCanBroadcast(t *testing.T) {
	ctx := context.Background()
	b := New()

	wg := sync.WaitGroup{}

	l1 := b.Listen()

	wg.Add(3)
	go startListener(t, ctx, l1, "l1", &wg, 2)

	b.Send(ctx, 1)

	l2 := b.Listen()
	go startListener(t, ctx, l2, "l2", &wg, 2)

	b.Send(ctx, 2)

	l3 := b.Listen()
	go startListener(t, ctx, l3, "l3", &wg, 2)

	b.Send(ctx, 3)

	b.Close()

	wg.Wait()
}

func TestCanBroadcastWithHistory(t *testing.T) {
	ctx := context.Background()
	b := NewBacklog(4)

	b.Send(ctx, -4)
	b.Send(ctx, -3)
	b.Send(ctx, -2)
	b.Send(ctx, -1)

	wg := sync.WaitGroup{}

	l1 := b.Listen()

	wg.Add(3)
	go startListener(t, ctx, l1, "l1", &wg, 2)

	b.Send(ctx, 1)

	l2 := b.Listen()
	go startListener(t, ctx, l2, "l2", &wg, 6)

	b.Send(ctx, 2)

	l3 := b.Listen()
	go startListener(t, ctx, l3, "l3", &wg, 2)

	b.Send(ctx, 3)

	b.Close()

	wg.Wait()
}

func startListener(t *testing.T, ctx context.Context, l Listener, name string, wg *sync.WaitGroup, times int) {
	defer wg.Done()
	defer l.Close()

	before := int(math.MinInt64)

	for i := 0; i < times; i++ {
		n, ok := l.Next(ctx)
		if !ok {
			log.Printf("%s: broadcaster closed!", name)
			return
		}

		now := n.(int)
		if now <= before {
			t.Errorf("%s received %d after %d", name, now, before)
		}
		before = now

		log.Printf("%s: msg %d/%d: got %d", name, i+1, times, n)
	}
	log.Printf("%s: done!", name)
}

func TestCanAbortSend(t *testing.T)            { testCanAbortSend(t, New()) }
func TestCanAbortSendWithHistory(t *testing.T) { testCanAbortSend(t, NewBacklog(100)) }

func testCanAbortSend(t *testing.T, b Broadcaster) {
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() { defer close(done); b.Send(ctx, 1) }()

	cancel()

	select {
	case <-time.After(time.Second):
		panic("too long")
	case <-done:
	}
}

func TestCanAbortListen(t *testing.T)            { testCanAbortListen(t, New()) }
func TestCanAbortListenWithHistory(t *testing.T) { testCanAbortListen(t, NewBacklog(100)) }

func testCanAbortListen(t *testing.T, b Broadcaster) {
	ctx, cancel := context.WithCancel(context.Background())
	l := b.Listen()

	done := make(chan struct{})
	go func() {
		defer close(done)
		ev, more := l.Next(ctx)
		if more {
			t.Errorf("should have no more (context cancelled): got %#v", ev)
		}
	}()

	// never send anything
	cancel()

	select {
	case <-time.After(time.Second):
		panic("too long")
	case <-done:
	}
}
