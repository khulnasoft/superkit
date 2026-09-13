package handlers

import (
	"github.com/khulnasoft/superkit/bootstrap/app/types"

	"github.com/khulnasoft/superkit/kit"
)

func HandleAuthentication(kit *kit.Kit) (kit.Auth, error) {
	return types.AuthUser{}, nil
}
