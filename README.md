Package broadcaster provides a way to send events to multiple listeners.

It offers a normal broadcaster and a broadcaster that keeps a short history
of events to update new listeners.

# Godocs?

[Godoc!](http://godoc.org/github.com/aybabtme/broadcaster)

# Presentation

I present to you `Broadcaster`:

```go
b := broadcaster.New() // create a broadcaster
b.Send(event)          // send an event to listeners
l := b.Listen()        // create a new listener
b.Close()              // close the broadcast, release the listeners
```

and `Listener`:

```go
l := b.Listen()       // create a listener
event, ok := l.Next() // block for the next event
l.Close()             // stop listening
```

# Usage

Here's how you can use them:

```go
b := broadcaster.New()
b.Send("Possibly nobody will get this")

for i := 0; i < listenCount; i++ {
  l := b.Listen()
  go func() {
    defer l.Close()
    for {
      e, ok := l.Next()
      if !ok {
        // broadcast stopped!
        return
      }
      fmt.Println(e)
    }
  }()
}

time.Sleep(time.Second * 1)
b.Close()
```

Listeners can be closed to announce that they do not wish to receive further events.

```go
l := b.Listen()

b.Send("hello!") // l receives this one

l.Close()

b.Send("hello?") // l is not listening anymore
b.Close()
```

And broadcasters can announce to their listeners that no further events will be sent (and therefore the listeners should be released).

```go
l := b.Listen()
b.Close()

go func() {
  defer l.Close()
  _, ok := l.Next()
  if !ok {
    return // the broadcaster closed
  }
}()
```

It's pretty simple!

# Two variants

There are two implementations. The first is a broadcaster that sends the events that it receives to all its listeners, as you would expect.

```go
b := broadcaster.New()
b.Send("Possibly nobody will get this")

l1 := b.Listen()
defer l1.Close()

b.Send("Hello l1")

l2 := b.Listen()
defer l2.Close()

b.Send("Hello l1 and l2")

b.Close()
```

The second does the same thing, but also keeps an history of the last events and brings up to date listeners that just subscribed.

```go
b := broadcaster.NewBacklog(3)
b.Send("Everybody will get this")

l1 := b.Listen()
defer l1.Close()

b.Send("Hello l1, hello future l2")

l2 := b.Listen()
defer l2.Close()

b.Send("Hello l1 and l2")

b.Close()
```


# License

MIT.
