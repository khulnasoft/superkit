package validate

import (
	"sort"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// PasswordHistory represents a single historical password entry
type PasswordHistory struct {
	// ID is the unique identifier for this history entry
	ID uint
	
	// UserID is the user this password history belongs to
	UserID uint
	
	// Hash is the bcrypt hash of the historical password
	Hash string
	
	// CreatedAt is when this password was set
	CreatedAt time.Time
}

// PasswordHistoryOptions configures password history behavior
type PasswordHistoryOptions struct {
	// MaxHistory is the maximum number of password hashes to store
	MaxHistory int
	
	// MinAge is the minimum age a password must be before it can be changed
	MinAge time.Duration
	
	// MaxAge is the maximum age a password can be before requiring change
	MaxAge time.Duration
	
	// EnforceHistory determines whether to prevent password reuse
	EnforceHistory bool
}

// DefaultPasswordHistoryOptions returns sensible defaults for password history
func DefaultPasswordHistoryOptions() PasswordHistoryOptions {
	return PasswordHistoryOptions{
		MaxHistory:     5,
		MinAge:         time.Hour * 24,  // 1 day
		MaxAge:         time.Hour * 24 * 90, // 90 days
		EnforceHistory: true,
	}
}

// CanChangePassword checks if a password is old enough to be changed
func CanChangePassword(lastChanged time.Time, opts PasswordHistoryOptions) bool {
	if opts.MinAge == 0 {
		return true
	}
	
	return time.Since(lastChanged) >= opts.MinAge
}

// IsPasswordExpired checks if a password is too old and needs to be changed
func IsPasswordExpired(lastChanged time.Time, opts PasswordHistoryOptions) bool {
	if opts.MaxAge == 0 {
		return false
	}
	
	return time.Since(lastChanged) >= opts.MaxAge
}

// IsPasswordReused checks if a password is in the history
func IsPasswordReused(password string, history []PasswordHistory) (bool, error) {
	if len(history) == 0 {
		return false, nil
	}
	
	for _, h := range history {
		err := bcrypt.CompareHashAndPassword([]byte(h.Hash), []byte(password))
		if err == nil {
			// Password matches an old password
			return true, nil
		} else if err != bcrypt.ErrMismatchedHashAndPassword {
			// An error occurred that wasn't "password doesn't match"
			return false, err
		}
	}
	
	return false, nil
}

// AddPasswordToHistory adds a new password to the history
func AddPasswordToHistory(userID uint, hash string, history []PasswordHistory, opts PasswordHistoryOptions) []PasswordHistory {
	// Create new history entry
	entry := PasswordHistory{
		UserID:    userID,
		Hash:      hash,
		CreatedAt: time.Now(),
	}
	
	// Add to history
	history = append(history, entry)
	
	// Trim if necessary

