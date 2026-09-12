package handlers

import (
	"github.com/khulnasoft/superkit/bootstrap/app/views/landing"

	"github.com/khulnasoft/superkit/kit"
)

func HandleLandingIndex(kit *kit.Kit) error {
	return kit.Render(landing.Index())
}
