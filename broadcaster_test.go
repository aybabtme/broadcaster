package broadcaster_test

import (
	"github.com/aybabtme/broadcaster"
	"log"
	"math"
	"sync"
	"testing"
)

func TestCanBroadcast(t *testing.T) {
	b := broadcaster.New()

	wg := sync.WaitGroup{}

	l1 := b.Listen()

	wg.Add(3)
	go startListener(t, l1, "l1", &wg, 2)

	b.Send(1)

	l2 := b.Listen()
	go startListener(t, l2, "l2", &wg, 2)

	b.Send(2)

	l3 := b.Listen()
	go startListener(t, l3, "l3", &wg, 2)

	b.Send(3)

	b.Close()

	wg.Wait()
}

func TestCanBroadcastWithHistory(t *testing.T) {
	b := broadcaster.NewBacklog(4)

	b.Send(-4)
	b.Send(-3)
	b.Send(-2)
	b.Send(-1)

	wg := sync.WaitGroup{}

	l1 := b.Listen()

	wg.Add(3)
	go startListener(t, l1, "l1", &wg, 2)

	b.Send(1)

	l2 := b.Listen()
	go startListener(t, l2, "l2", &wg, 6)

	b.Send(2)

	l3 := b.Listen()
	go startListener(t, l3, "l3", &wg, 2)

	b.Send(3)

	b.Close()

	wg.Wait()
}

func startListener(t *testing.T, l broadcaster.Listener, name string, wg *sync.WaitGroup, times int) {
	defer wg.Done()
	defer l.Close()

	before := int(math.MinInt64)

	for i := 0; i < times; i++ {
		n, ok := l.Next()
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
