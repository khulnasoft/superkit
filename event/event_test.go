package event

import (
	"context"
	"reflect"
	"sync/atomic"
	"testing"
	"time"
)

func TestEventSubscribeEmit(t *testing.T) {
	expect := 1
	ctx, cancel := context.WithCancel(context.Background())
	Subscribe("foo.a", func(_ context.Context, event any) {
		defer cancel()
		value, ok := event.(int)
		if !ok {
			t.Errorf("expected int got %v", reflect.TypeOf(event))
		}
		if value != 1 {
			t.Errorf("expected %d got %d", expect, value)
		}
	})
	Emit("foo.a", expect)
	<-ctx.Done()
}

func TestUnsubscribe(t *testing.T) {
	sub := Subscribe("foo.b", func(_ context.Context, _ any) {})
	Unsubscribe(sub)
	if _, ok := stream.subs["foo.b"]; ok {
		t.Errorf("expected topic foo.bar to be deleted")
	}
}

func TestEmitWithRetryRetriesOnPanic(t *testing.T) {
	topic := "foo.retry"
	var attempts atomic.Int32
	done := make(chan struct{}, 1)
	Subscribe(topic, func(_ context.Context, _ any) {
		if attempts.Add(1) == 1 {
			panic("transient failure")
		}
		done <- struct{}{}
	})

	EmitWithRetry(topic, 42, 3, 10*time.Millisecond)

	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("expected retry to succeed within timeout")
	}

	if got := attempts.Load(); got != 2 {
		t.Fatalf("expected 2 attempts, got %d", got)
	}
}
