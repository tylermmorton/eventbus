# EventBus

This package provides an event bus implementation using a combination of Go generics and Go channels. Use an EventBus to create a simple and flexible pub/sub system in any project.

# Usage

```go
// Create a new EventBus[K, V] object. The key can be anything
// you want to use to identify the different channels in a bus.
// The value is the data type you expect to pass through the bus.
bus := eventbus.New[string, string]()

// To subscribe to an EventBus, we use built-in Go channels
ch := make(chan string)
cl := make(chan error)

// Subscribe to the bus using a key and a new channel
bus.Subscribe("channelID", ch)

// Use the closer channel to clean things up when done.
go func() {
    defer close(ch)

    <-cl
    bus.Unsubscribe(convo.ID, ch)
}()

// Use the main channel to receive events dispatched by the bus.
go func() {
    for {
        msg, ok := <-ch
        if !ok {
            // the channel was probably closed
            break
        }

        // do something with your message data
        err := doSomething(msg)
        if err != nil {
            // if something goes wrong, signal the closer channel
            cl <- errors.Wrap(err, "channel closed by server")
            break
        }
    }
}()

// Finally, send events through the bus using Dispatch.
// All subscribers under 'channelID' will receive the same message.
bus.Dispatch("channelID", "Hello world!")
```
