package validate

import (
	"errors"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
	"time"
	"unicode"
)

// Common password-related errors
var (
	ErrPasswordTooShort         = errors.New("password is too short")
	ErrPasswordTooLong          = errors.New("password is too long")
	ErrPasswordNoUpper          = errors.New("password must contain at least one uppercase letter")
	ErrPasswordNoLower          = errors.New("password must contain at least one lowercase letter")
	ErrPasswordNoNumber         = errors.New("password must contain at least one number")
	ErrPasswordNoSpecial        = errors.New("password must contain at least one special character")
	ErrPasswordCommon           = errors.New("password is too common")
	ErrPasswordContainsUsername = errors.New("password contains username")
	ErrPasswordContainsEmail    = errors.New("password contains email address")
	ErrPasswordPreviouslyUsed   = errors.New("password has been used previously")
	ErrPasswordInvalid          = errors.New("password doesn't meet the required criteria")
)

// PasswordStrengthOptions configures password strength requirements
type PasswordStrengthOptions struct {
	// MinLength is the minimum password length
	MinLength int
	
	// MaxLength is the maximum password length
	MaxLength int
	
	// RequireUppercase requires the password to contain at least one uppercase letter
	RequireUppercase bool
	
	// RequireLowercase requires the password to contain at least one lowercase letter
	RequireLowercase bool
	
	// RequireNumbers requires the password to contain at least one number
	RequireNumbers bool
	
	// RequireSpecial requires the password to contain at least one special character
	RequireSpecial bool
	
	// DisallowCommonPasswords checks against a list of common passwords
	DisallowCommonPasswords bool
	
	// CommonPasswordsFile is the path to a file containing common passwords
	CommonPasswordsFile string
	
	// DisallowUsernameSimilarity prevents passwords that contain the username
	DisallowUsernameSimilarity bool
	
	// DisallowEmailSimilarity prevents passwords that contain the email address
	DisallowEmailSimilarity bool
	
	// CheckPasswordHistory determines whether to check against previously used passwords
	CheckPasswordHistory bool
	
	// PasswordHistoryCount is the number of previous passwords to check against
	PasswordHistoryCount int
}

// DefaultPasswordStrengthOptions returns the default password strength options
func DefaultPasswordStrengthOptions() PasswordStrengthOptions {
	return PasswordStrengthOptions{
		MinLength:                8,
		MaxLength:                72, // bcrypt limit is 72 bytes
		RequireUppercase:         true,
		RequireLowercase:         true,
		RequireNumbers:           true,
		RequireSpecial:           true,
		DisallowCommonPasswords:  true,
		CommonPasswordsFile:      "",  // No default file, can be set by the user
		DisallowUsernameSimilarity: true,
		DisallowEmailSimilarity:  true,
		CheckPasswordHistory:     true,
		PasswordHistoryCount:     5,
	}
}

// commonPasswords is a small set of extremely common passwords
// This is used as a fallback if no common passwords file is provided
var commonPasswords = []string{
	"123456", "password", "12345678", "qwerty", "123456789", "12345", "1234",
	"111111", "1234567", "dragon", "123123", "baseball", "abc123", "football",
	"monkey", "letmein", "696969", "shadow", "master", "666666", "qwertyuiop",
	"123321", "mustang", "1234567890", "michael", "654321", "superman", "1qaz2wsx",
	"7777777", "fuckyou", "121212", "000000", "qazwsx", "123qwe", "killer",
	"trustno1", "jordan", "jennifer", "zxcvbnm", "asdfgh", "hunter", "buster",
	"soccer", "harley", "batman", "andrew", "tigger", "sunshine", "iloveyou",
	"fuckme", "2000", "charlie", "robert", "thomas", "hockey", "ranger", "daniel",
	"starwars", "klaster", "112233", "george", "asshole", "computer", "michelle",
	"jessica", "pepper", "1111", "zxcvbn", "555555", "11111111", "131313", "freedom",
	"777777", "pass", "maggie", "159753", "aaaaaa", "ginger", "princess", "joshua",
	"cheese", "amanda", "summer", "love", "ashley", "6969", "nicole", "chelsea",
	"biteme", "matthew", "access", "yankees", "987654321", "dallas", "austin",
	"thunder", "taylor", "matrix", "william", "corvette", "hello", "martin", "heather",
	"secret", "merlin", "diamond", "1234qwer", "gfhjkm", "hammer", "silver", "222222",
	"88888888", "anthony", "justin", "test", "bailey", "q1w2e3r4t5", "patrick",
	"internet", "scooter", "orange", "golfer", "cookie", "richard", "samantha",
	"bigdog", "guitar", "jackson", "whatever", "mickey", "chicken", "sparky", "snoopy",
	"maverick", "phoenix", "camaro", "sexy", "peanut", "morgan", "welcome", "falcon",
	"cowboy", "ferrari", "samsung", "andrea", "smokey", "steelers", "joseph",
	"mercedes", "dakota", "arsenal", "eagles", "melissa", "boomer", "booboo",
	"spider", "nascar", "monster", "tigers", "yellow", "xxxxxx", "123123123", "gateway",
	"marina", "diablo", "bulldog", "qwer1234", "compaq", "purple", "hardcore", "banana",
	"junior", "hannah", "123654", "porsche", "lakers", "iceman", "money", "cowboys",
	"987654", "london", "tennis", "999999", "ncc1701", "coffee", "scooby", "0000",
	"miller", "boston", "q1w2e3r4", "brandon", "yamaha", "chester", "mother", "forever",
	"johnny", "edward", "333333", "oliver", "redsox", "player", "nikita", "knight",
	"fender", "barney", "midnight", "please", "brandy", "chicago", "badboy", "iwantu",
	"slayer", "rangers", "charles", "angel", "flower", "bigdaddy", "rabbit", "wizard",
	"bigdick", "jasper", "enter", "rachel", "chris", "steven", "winner", "adidas",
	"victoria", "natasha", "1q2w3e4r", "jasmine", "winter", "prince", "panties",
	"marine", "ghbdtn", "fishing", "cocacola", "casper", "james", "232323", "raiders",
	"888888", "marlboro", "gandalf", "asdfasdf", "crystal", "87654321", "dennis",
	"ncc1701d", "blue", "john", "rebecca", "physics", "happyday",
}

// loadCommonPasswords loads common passwords from a file, one per line
func loadCommonPasswords(filename string) ([]string, error) {
	if filename == "" {
		return commonPasswords, nil
	}
	
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	
	lines := strings.Split(string(data), "\n")
	passwords := make([]string, 0, len(lines))
	
	for _, line := range lines {
		password := strings.TrimSpace(line)
		if password != "" {
			passwords = append(passwords, password)
		}
	}
	
	return passwords, nil
}

// isCommonPassword checks if a password is in the list of common passwords
func isCommonPassword(password string, commonPasswords []string) bool {
	lowercasePassword := strings.ToLower(password)
	for _, commonPassword := range commonPasswords {
		if lowercasePassword == commonPassword {
			return true
		}
	}
	return false
}

// containsUsername checks if a password contains the username
func containsUsername(password, username string) bool {
	if username == "" {
		return false
	}
	
	lowercasePassword := strings.ToLower(password)
	lowercaseUsername := strings.ToLower(username)
	
	return strings.Contains(lowercasePassword, lowercaseUsername)
}

// containsEmail checks if a password contains the email address or username part of the email
func containsEmail(password, email string) bool {
	if email == "" {
		return false
	}
	
	lowercasePassword := strings.ToLower(password)
	lowercaseEmail := strings.ToLower(email)
	
	// Check if password contains the full email
	if strings.Contains(lowercasePassword, lowercaseEmail) {
		return true
	}
	
	// Check if password contains the username part of the email
	parts := strings.Split(lowercaseEmail, "@")
	if len(parts) == 2 && parts[0] != "" {
		if strings.Contains(lowercasePassword, parts[0]) {
			return true
		}
	}
	
	return false
}

// ValidatePassword checks if a password meets the requirements
func ValidatePassword(password string, opts PasswordStrengthOptions) error {
	// Check length
	if len(password) < opts.MinLength {
		return ErrPasswordTooShort
	}
	
	if opts.MaxLength > 0 && len(password) > opts.MaxLength {
		return ErrPasswordTooLong
	}
	
	// Check for uppercase letters
	if opts.RequireUppercase {
		hasUpper := false
		for _, char := range password {
			if unicode.IsUpper(char) {
				hasUpper = true
				break
			}
		}
		if !hasUpper {
			return ErrPasswordNoUpper
		}
	}
	
	// Check for lowercase letters
	if opts.RequireLowercase {
		hasLower := false
		for _, char := range password {
			if unicode.IsLower(char) {
				hasLower = true
				break
			}
		}
		if !hasLower {
			return ErrPasswordNoLower
		}
	}
	
	// Check for numbers
	if opts.RequireNumbers {
		hasNumber := false
		for _, char := range password {
			if unicode.IsDigit(char) {
				hasNumber = true
				break
			}
		}
		if !hasNumber {
			return ErrPasswordNoNumber
		}
	}
	
	// Check for special characters
	if opts.RequireSpecial {
		hasSpecial := false
		for _, char := range password {
			if !unicode.IsLetter(char) && !unicode.IsDigit(char) {
				hasSpecial = true
				break
			}
		}
		if !hasSpecial {
			return ErrPasswordNoSpecial
		}
	}
	
	return nil
}

// ValidatePasswordWithUsername validates a password against the requirements and checks if it contains the username
func ValidatePasswordWithUsername(password, username string, opts PasswordStrengthOptions) error {
	if err := ValidatePassword(password, opts); err != nil {
		return err
	}
	
	if opts.DisallowUsernameSimilarity && containsUsername(password, username) {
		return ErrPasswordContainsUsername
	}
	
	return nil
}

// ValidatePasswordWithEmail validates a password against the requirements and checks if it contains the email
func ValidatePasswordWithEmail(password, email string, opts PasswordStrengthOptions) error {
	if err := ValidatePassword(password, opts); err != nil {
		return err
	}
	
	if opts.DisallowEmailSimilarity && containsEmail(password, email) {
		return ErrPasswordContainsEmail
	}
	
	return nil
}

// ValidatePasswordComplete performs all password validation checks
func ValidatePasswordComplete(password, username, email string, previousPasswords []string, opts PasswordStrengthOptions) error {
	// Basic validation
	if err := ValidatePassword(password, opts); err != nil {
		return err
	}
	
	// Check for username similarity
	if opts.DisallowUsernameSimilarity && containsUsername(password, username) {
		return ErrPasswordContainsUsername
	}
	
	// Check for email similarity
	if opts.DisallowEmailSimilarity && containsEmail(password, email) {
		return ErrPasswordContainsEmail
	}
	
	// Check if password is common
	if opts.DisallowCommonPasswords {
		commonPwds, err := loadCommonPasswords(opts.CommonPasswordsFile)
		if err == nil && isCommonPassword(password, commonPwds) {
			return ErrPasswordCommon
		}
	}
	
	// Check password history
	if opts.CheckPasswordHistory && len(previousPasswords) > 0 {
		for _, prevPass := range previousPasswords {
			if password == prevPass {
				return ErrPasswordPreviouslyUsed
			}
		}
	}
	
	return nil
}

// PasswordStrength represents a calculated password strength score
type PasswordStrength struct {
	// Score from 0-100
	Score int
	
	// Feedback provides user-friendly suggestions for improving the password
	Feedback string
	
	// Issues is a list of specific issues with the password
	Issues []string
}

// CalculatePasswordStrength calculates a password strength score
func CalculatePasswordStrength(password string) PasswordStrength {
	strength := PasswordStrength{
		Score:    0,
		Feedback: "",
		Issues:   []string{},
	}
	
	// Base score is the length of the password (capped at 25 points)
	length := len(password)
	if length > 25 {
		length = 25
	}
	strength.Score += length * 2
	
	// Check for character variety
	hasUpper := false
	hasLower := false
	hasNumber := false
	hasSpecial := false
	
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasNumber = true
		case !unicode.IsLetter(char) && !unicode.IsDigit(char):
			hasSpecial = true
		}
	}
	
	// Add points for character variety
	if hasUpper {
		strength.Score += 10
	} else {
		strength.Issues = append(strength.Issues, "Add uppercase letters")
	}
	
	if hasLower {
		strength.Score += 10
	} else {
		strength.Issues = append(strength.Issues, "Add lowercase letters")
	}
	
	if hasNumber {
		strength.Score += 10
	} else {
		strength.Issues = append(strength.Issues, "Add numbers")
	}
	
	if hasSpecial {
		strength.Score += 15
	} else {
		strength.Issues = append(strength.Issues, "Add special characters")
	}
	
	// Check for repeating patterns
	if regexp.MustCompile(`(.)\1{2,}`).MatchString(password) {
		strength.Score -= 10
		strength.Issues = append(strength.Issues, "Avoid repeating characters")
	}
	
	// Check for sequences
	if regexp.MustCompile(`(?i)(abc|bcd|cde|def|efg|fgh|ghi|hij|ijk|jkl|klm|lmn|mno|nop|opq|pqr|qrs|rst|stu|tuv|uvw|vwx|wxy|xyz|012|123|234|345|456|567|678|789|890)`).MatchString(password) {
		strength.Score -= 5
		strength.Issues = append(strength.Issues, "Avoid sequential characters")
	}
	
	// Check for keyboard patterns
	if regexp.MustCompile(`(?i)(qwert|werty|ertyu|rtyui|tyuio|yuiop|asdfg|sdfgh|dfghj|fghjk|ghjkl|zxcvb|xcvbn|cvbnm)`).MatchString(password) {
		strength.Score -= 5
		strength.Issues = append(strength.Issues, "Avoid keyboard patterns")
	}
	
	// Bonus for mixed character types
	charTypesCount := 0
	if hasUpper {
		charTypesCount++
	}
	if hasLower {
		charTypesCount++
	}
	if hasNumber {
		charTypesCount++
	}
	if hasSpecial {
		charTypesCount++
	}
	
	if charTypesCount >= 3 {
		strength.Score += 10
	}
	
	// Cap the score at 100
	if strength.Score > 100 {
		strength.Score = 100
	} else if strength.Score < 0 {
		strength.Score = 0
	}
	
	// Generate feedback based on score
	switch {
	case strength.Score < 30:
		strength.Feedback = "Very weak password"
	case strength.Score < 50:
		strength.Feedback = "Weak password"
	case strength.Score < 70:
		strength.Feedback = "Moderate password"
	case strength.Score < 90:
		strength.Feedback = "Strong password"
	default:
		strength.Feedback = "Very strong password"
	}
	
	return strength
}
