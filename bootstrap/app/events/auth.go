package events

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/khulnasoft/superkit/bootstrap/plugins/auth"
	"github.com/khulnasoft/superkit/event"
)

func init() {
	event.Subscribe(auth.UserSignupEvent, OnUserSignup)
	event.Subscribe(auth.ResendVerificationEvent, OnResendVerificationToken)
}

func OnUserSignup(ctx context.Context, event any) {
	userWithToken, ok := event.(auth.UserWithVerificationToken)
	if !ok {
		return
	}

	if err := auth.SendVerificationEmail(ctx, userWithToken.User.Email, userWithToken.Token); err != nil {
		fmt.Printf("failed to send verification email: %v\n", err)
	}

	b, _ := json.MarshalIndent(userWithToken, "   ", "    ")
	fmt.Println(string(b))
}

func OnResendVerificationToken(ctx context.Context, event any) {
	userWithToken, ok := event.(auth.UserWithVerificationToken)
	if !ok {
		return
	}

	if err := auth.SendVerificationEmail(ctx, userWithToken.User.Email, userWithToken.Token); err != nil {
		fmt.Printf("failed to send verification email: %v\n", err)
	}

	b, _ := json.MarshalIndent(userWithToken, "   ", "    ")
	fmt.Println(string(b))
}
