package event

import (
	"context"
	"slices"
	"sync"
	"time"
)

// HandlerFunc is the function being called when receiving an event.
type HandlerFunc func(context.Context, any)

const (
	eventChannelSize = 128
	workerPoolSize   = 8
)

// Emit an event to the given topic. If the event channel is full, the
// event is dropped to avoid blocking request handlers.
func Emit(topic string, event any) {
	stream.emit(topic, event)
}

// EmitWithRetry dispatches the event and retries any panic-ing handler up to the
// requested number of attempts with a bounded delay between retries.
func EmitWithRetry(topic string, event any, retries int, delay time.Duration) {
	stream.emitWithRetry(topic, event, retries, delay)
}

// Subscribe a HandlerFunc to the given topic.
// A Subscription is being returned that can be used
// to unsubscribe from the topic.
func Subscribe(topic string, h HandlerFunc) Subscription {
	return stream.subscribe(topic, h)
}

// Unsubscribe unsubscribes the given Subscription from its topic.
func Unsubscribe(sub Subscription) {
	stream.unsubscribe(sub)
}

// Stop stops the event stream, cleaning up its resources.
func Stop() {
	stream.stop()
}

var stream *eventStream

type event struct {
	topic   string
	message any
	retries int
	delay   time.Duration
}

// Subscription represents a handler subscribed to a specific topic.
type Subscription struct {
	Topic     string
	CreatedAt int64
	Fn        HandlerFunc
}

type eventStream struct {
	mu           sync.RWMutex
	subs         map[string][]Subscription
	eventch      chan event
	quitch       chan struct{}
	workerQuitch chan struct{}
	dropped      int64
}

func newStream() *eventStream {
	e := &eventStream{
		subs:         make(map[string][]Subscription),
		eventch:      make(chan event, eventChannelSize),
		quitch:       make(chan struct{}),
		workerQuitch: make(chan struct{}),
	}
	go e.startWorkers()
	return e
}

func (e *eventStream) startWorkers() {
	ctx := context.Background()
	for i := 0; i < workerPoolSize; i++ {
		go e.worker(ctx, i)
	}
	<-e.quitch
	close(e.workerQuitch)
}

func (e *eventStream) worker(ctx context.Context, id int) {
	_ = id
	for {
		select {
		case <-e.workerQuitch:
			return
		case evt, ok := <-e.eventch:
			if !ok {
				return
			}
			if handlers, ok := e.subs[evt.topic]; ok {
				for _, sub := range handlers {
					select {
					case <-e.workerQuitch:
						return
					default:
						runWithRetry(ctx, sub.Fn, evt.message, evt.retries, evt.delay)
					}
				}
			}
		}
	}
}

func runWithRetry(ctx context.Context, fn HandlerFunc, value any, retries int, delay time.Duration) {
	if retries < 0 {
		retries = 0
	}
	for attempt := 0; attempt <= retries; attempt++ {
		panicRecovered := false
		func() {
			defer func() {
				if recovered := recover(); recovered != nil {
					panicRecovered = true
					if attempt < retries && delay > 0 {
						time.Sleep(delay)
					}
				}
			}()
			fn(ctx, value)
		}()
		if !panicRecovered {
			return
		}
	}
}

func (e *eventStream) stop() {
	e.quitch <- struct{}{}
}

func (e *eventStream) emit(topic string, v any) {
	e.emitWithRetry(topic, v, 0, 0)
}

func (e *eventStream) emitWithRetry(topic string, v any, retries int, delay time.Duration) {
	select {
	case e.eventch <- event{
		topic:   topic,
		message: v,
		retries: retries,
		delay:   delay,
	}:
	default:
		e.mu.Lock()
		e.dropped++
		e.mu.Unlock()
	}
}

func (e *eventStream) subscribe(topic string, h HandlerFunc) Subscription {
	e.mu.Lock()
	defer e.mu.Unlock()

	sub := Subscription{
		CreatedAt: time.Now().UnixNano(),
		Topic:     topic,
		Fn:        h,
	}

	if _, ok := e.subs[topic]; !ok {
		e.subs[topic] = []Subscription{}
	}

	e.subs[topic] = append(e.subs[topic], sub)

	return sub
}

func (e *eventStream) unsubscribe(sub Subscription) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if _, ok := e.subs[sub.Topic]; ok {
		e.subs[sub.Topic] = slices.DeleteFunc(e.subs[sub.Topic], func(e Subscription) bool {
			return sub.CreatedAt == e.CreatedAt
		})
	}
	if len(e.subs[sub.Topic]) == 0 {
		delete(e.subs, sub.Topic)
	}
}

// Stats holds runtime metrics for the event system.
type Stats struct {
	Dropped  int64
	QueueLen int
	Workers  int
}

// Stats returns current event system metrics.
func GetStats() Stats {
	stream.mu.RLock()
	defer stream.mu.RUnlock()
	return Stats{
		Dropped:  stream.dropped,
		QueueLen: len(stream.eventch),
		Workers:  workerPoolSize,
	}
}

func init() {
	stream = newStream()
}
