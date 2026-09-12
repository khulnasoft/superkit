package events

import (
	"context"
	"time"

	"github.com/khulnasoft/superkit/event"
)

func OnHealthCheck(ctx context.Context, _ any) {
	event.Emit("health.check", map[string]any{
		"timestamp": time.Now().Unix(),
		"status":    "ok",
	})
}
