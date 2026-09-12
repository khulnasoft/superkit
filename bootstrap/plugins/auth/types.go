package auth

import (
	"github.com/khulnasoft/superkit/bootstrap/app/db"
	"database/sql"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Event name constants
const (
	UserSignupEvent         = "auth.signup"
	ResendVerificationEvent = "auth.resend.verification"
)

// UserWithVerificationToken is a struct that will be sent over the
// auth.signup event. It holds the User struct and the Verification token string.
type UserWithVerificationToken struct {
	User  User
	Token string
}

type Role string

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
)

var rolePermissions = map[Role][]string{
	RoleUser:  {"profile:read", "profile:write"},
	RoleAdmin: {"profile:read", "profile:write", "admin:access"},
}

type Auth struct {
	UserID    uint
	Email     string
	FirstName string
	LastName  string
	LoggedIn  bool
	Role      Role
}

func (auth Auth) Check() bool {
	return auth.LoggedIn
}

func (auth Auth) HasRole(role Role) bool {
	return auth.LoggedIn && auth.Role == role
}

func (auth Auth) Can(permission string) bool {
	if !auth.LoggedIn {
		return false
	}
	for _, allowed := range rolePermissions[auth.Role] {
		if allowed == permission {
			return true
		}
	}
	return false
}

type User struct {
	gorm.Model

	Email           string
	FirstName       string
	LastName        string
	Role            Role
	PasswordHash    string
	EmailVerifiedAt sql.NullTime
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func createUserFromFormValues(values SignupFormValues) (User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(values.Password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, err
	}
	user := User{
		Email:        values.Email,
		FirstName:    values.FirstName,
		LastName:     values.LastName,
		Role:         RoleUser,
		PasswordHash: string(hash),
	}
	result := db.Get().Create(&user)
	return user, result.Error
}

type Session struct {
	gorm.Model

	UserID    uint
	Token     string
	IPAddress string
	UserAgent string
	ExpiresAt time.Time
	CreatedAt time.Time
	User      User
}
