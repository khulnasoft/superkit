package app

import (
	"github.com/khulnasoft/superkit/bootstrap/app/events"
	"github.com/khulnasoft/superkit/event"
)

// RegisterEvents registers event subscriptions.
// Events are functions that are handled in separate goroutines.
// They are the perfect fit for offloading work in your handlers
// that otherwise would take up response time.
// - sending email
// - sending notifications (Slack, Telegram, Discord)
// - analytics..
func RegisterEvents() {
	event.Subscribe("health.check", events.OnHealthCheck)
}
