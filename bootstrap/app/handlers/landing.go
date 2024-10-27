package handlers

import (
	"AABBCCDD/app/views/landing"

	"github.com/khulnasoft/superkit/kit"
)

func HandleLandingIndex(kit *kit.Kit) error {
	return kit.Render(landing.Index())
}
