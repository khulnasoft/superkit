package handlers

import (
	"AABBCCDD/app/types"

	"github.com/khulnasoft/superkit/kit"
)

func HandleAuthentication(kit *kit.Kit) (kit.Auth, error) {
	return types.AuthUser{}, nil
}
