package validate

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"sync"
	"time"
	
	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrResetTokenExpired    = errors.New("password reset token has expired")
	ErrResetTokenInvalid    = errors.New("invalid password reset token")
	ErrResetAttemptExceeded = errors.New("too many password reset attempts")
)

// PasswordResetOptions configures password reset behavior
type PasswordResetOptions struct {
	// TokenExpiry is how long reset tokens are valid for
	TokenExpiry time.Duration
	
	// TokenSecret is the secret used to sign reset tokens
	TokenSecret []byte
	
	// MaxResetAttempts is the maximum number of reset attempts per time window
	MaxResetAttempts int
	
	// ResetWindowDuration is the time window for rate limiting
	ResetWindowDuration time.Duration
}

// DefaultPasswordResetOptions returns sensible defaults for password reset
func DefaultPasswordResetOptions() PasswordResetOptions {
	return PasswordResetOptions{
		TokenExpiry:         time.Hour,
		TokenSecret:         nil, // Must be set by the application
		MaxResetAttempts:    3,
		ResetWindowDuration: time.Hour * 24,
	}
}

// ResetAttempt tracks a password reset attempt
type ResetAttempt struct {
	UserID    string
	IP        string
	Timestamp time.Time
}

// ResetRateLimiter manages rate limiting for password resets
type ResetRateLimiter struct {
	attempts map[string][]ResetAttempt
	mutex    sync.Mutex
}

// NewResetRateLimiter creates a new rate limiter for password resets
func NewResetRateLimiter() *ResetRateLimiter {
	return &ResetRateLimiter{
		attempts: make(map[string][]ResetAttempt),
	}
}

// CheckRateLimit checks if a user has exceeded the rate limit
func (rl *ResetRateLimiter) CheckRateLimit(userID, ip string, opts PasswordResetOptions) error {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()
	
	// Clean up old attempts first
	rl.cleanupOldAttempts(userID, opts.ResetWindowDuration)
	
	// Get the key for this attempt
	key := fmt.Sprintf("%s:%s", userID, ip)
	
	// Check if the user has exceeded the rate limit
	if len(rl.attempts[key]) >= opts.MaxResetAttempts {
		return ErrResetAttemptExceeded
	}
	
	// Record this attempt
	rl.attempts[key] = append(rl.attempts[key], ResetAttempt{
		UserID:    userID,
		IP:        ip,
		Timestamp: time.Now(),
	})
	
	return nil
}

// cleanupOldAttempts removes attempts older than the window duration
func (rl *ResetRateLimiter) cleanupOldAttempts(key string, window time.Duration) {
	cutoff := time.Now().Add(-window)
	attempts := rl.attempts[key]
	
	if len(attempts) == 0 {
		return
	}
	
	// Find the first attempt that's not expired
	i := 0
	for ; i < len(attempts); i++ {
		if attempts[i].Timestamp.After(cutoff) {
			break
		}
	}
	
	// Remove all expired attempts
	if i > 0 {
		if i >= len(attempts) {
			// All attempts are expired
			delete(rl.attempts, key)
		} else {
			// Some attempts are still valid
			rl.attempts[key] = attempts[i:]
		}
	}
}

// GenerateResetToken creates a secure token for password reset
func GenerateResetToken(userID string, opts PasswordResetOptions) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(opts.TokenExpiry).Unix(),
		"jti": generateRandomString(16),
		"iat": time.Now().Unix(),
	}
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(opts.TokenSecret)
}

// ValidateResetToken validates a password reset token
func ValidateResetToken(tokenString string, opts PasswordResetOptions) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return opts.TokenSecret, nil
	})
	
	if err != nil {
		return "", ErrResetTokenInvalid
	}
	
	if !token.Valid {
		return "", ErrResetTokenInvalid
	}
	
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", ErrResetTokenInvalid
	}
	
	// Check expiration
	exp, ok := claims["exp"].(float64)
	if !ok || time.Unix(int64(exp), 0).Before(time.Now()) {
		return "", ErrResetTokenExpired
	}
	
	// Get user ID
	userID, ok := claims["sub"].(string)
	if !ok {
		return "", ErrResetTokenInvalid
	}
	
	return userID, nil
}

// generateRandomString creates a random string for the JTI claim
func generateRandomString(length int) string {
	bytes := make([]byte, length)
	rand.Read(bytes)
	return base64.URLEncoding.EncodeToString(bytes)[:length]
}

