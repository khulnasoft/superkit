package auth

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/khulnasoft/superkit/bootstrap/app/conf"
	"github.com/khulnasoft/superkit/bootstrap/app/db"
	"github.com/khulnasoft/superkit/kit"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) {
	t.Helper()

	os.Setenv("SUPERKIT_SECRET", "12345678901234567890123456789012")
	os.Setenv("DB_DRIVER", "sqlite3")
	os.Setenv("DB_NAME", ":memory:")
	os.Setenv("SUPERKIT_AUTH_SKIP_VERIFY", "true")

	conf.Load()
	kit.Setup()

	if err := db.Initialize(); err != nil {
		t.Fatalf("failed to initialize db: %v", err)
	}

	gormDB := db.Get()
	if err := gormDB.AutoMigrate(&User{}, &Session{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
	})
}

func createTestUser(t *testing.T, gormDB *gorm.DB, email, password string) User {
	t.Helper()

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	user := User{
		Email:        email,
		FirstName:    "Test",
		LastName:     "User",
		PasswordHash: string(hash),
	}

	if err := gormDB.Create(&user).Error; err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	return user
}

func TestHandleLoginIndex(t *testing.T) {
	setupTestDB(t)

	req := httptest.NewRequest(http.MethodGet, "/login", nil)
	w := httptest.NewRecorder()

	k := &kit.Kit{
		Response: w,
		Request:  req,
	}

	err := HandleLoginIndex(k)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandleLoginCreateInvalidCredentials(t *testing.T) {
	setupTestDB(t)

	form := url.Values{}
	form.Set("email", "nonexistent@example.com")
	form.Set("password", "wrongpassword")

	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	k := &kit.Kit{
		Response: w,
		Request:  req,
	}

	err := HandleLoginCreate(k)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "invalid credentials")
}

func TestHandleLoginCreateValidCredentials(t *testing.T) {
	setupTestDB(t)
	gormDB := db.Get()
	createTestUser(t, gormDB, "test@example.com", "password123")

	form := url.Values{}
	form.Set("email", "test@example.com")
	form.Set("password", "password123")

	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	k := &kit.Kit{
		Response: w,
		Request:  req,
	}

	err := HandleLoginCreate(k)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, w.Code)
}

func TestHandleLoginCreateWrongPassword(t *testing.T) {
	setupTestDB(t)
	gormDB := db.Get()
	createTestUser(t, gormDB, "test@example.com", "password123")

	form := url.Values{}
	form.Set("email", "test@example.com")
	form.Set("password", "wrongpassword")

	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	k := &kit.Kit{
		Response: w,
		Request:  req,
	}

	err := HandleLoginCreate(k)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "invalid credentials")
}

func TestHandleSignupCreate(t *testing.T) {
	setupTestDB(t)

	form := url.Values{}
	form.Set("email", "new@example.com")
	form.Set("firstName", "New")
	form.Set("lastName", "User")
	form.Set("password", "Password123!")
	form.Set("passwordConfirm", "Password123!")

	req := httptest.NewRequest(http.MethodPost, "/signup", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	k := &kit.Kit{
		Response: w,
		Request:  req,
	}

	err := HandleSignupCreate(k)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandleSignupCreatePasswordMismatch(t *testing.T) {
	setupTestDB(t)

	form := url.Values{}
	form.Set("email", "new@example.com")
	form.Set("firstName", "New")
	form.Set("lastName", "User")
	form.Set("password", "Password123!")
	form.Set("passwordConfirm", "DifferentPass123!")

	req := httptest.NewRequest(http.MethodPost, "/signup", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	k := &kit.Kit{
		Response: w,
		Request:  req,
	}

	err := HandleSignupCreate(k)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "passwords do not match")
}

func TestAuthenticateUserNoSession(t *testing.T) {
	setupTestDB(t)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	k := &kit.Kit{
		Response: w,
		Request:  req,
	}

	auth, err := AuthenticateUser(k)
	assert.NoError(t, err)
	assert.False(t, auth.Check())
}

func TestAuthenticateUserWithSession(t *testing.T) {
	setupTestDB(t)
	gormDB := db.Get()
	user := createTestUser(t, gormDB, "test@example.com", "password123")

	session := Session{
		UserID:    user.ID,
		Token:     "test-session-token",
		ExpiresAt: time.Now().Add(time.Hour * 24),
	}

	if err := gormDB.Create(&session).Error; err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	k := &kit.Kit{
		Response: w,
		Request:  req,
	}

	sess := k.GetSession(userSessionName)
	sess.Values["sessionToken"] = "test-session-token"
	sess.Save(req, w)

	for _, cookie := range w.Result().Cookies() {
		req.AddCookie(cookie)
	}

	k = &kit.Kit{
		Response: w,
		Request:  req,
	}

	authResult, err := AuthenticateUser(k)
	assert.NoError(t, err)
	assert.True(t, authResult.Check())

	auth, ok := authResult.(Auth)
	assert.True(t, ok)
	assert.Equal(t, user.ID, auth.UserID)
	assert.Equal(t, user.Email, auth.Email)
}

func TestHandleLoginCreateUnverifiedUser(t *testing.T) {
	setupTestDB(t)
	_ = db.Get()
	createTestUser(t, db.Get(), "test@example.com", "password123")

	t.Setenv("SUPERKIT_AUTH_SKIP_VERIFY", "false")

	form := url.Values{}
	form.Set("email", "test@example.com")
	form.Set("password", "password123")

	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	k := &kit.Kit{
		Response: w,
		Request:  req,
	}

	err := HandleLoginCreate(k)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "please verify your email")
}

func TestHandleEmailVerifyInvalidToken(t *testing.T) {
	setupTestDB(t)

	req := httptest.NewRequest(http.MethodGet, "/email/verify?token=invalid", nil)
	w := httptest.NewRecorder()

	k := &kit.Kit{
		Response: w,
		Request:  req,
	}

	err := HandleEmailVerify(k)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "invalid verification token")
}

func TestHandleEmailVerifyMissingToken(t *testing.T) {
	setupTestDB(t)

	req := httptest.NewRequest(http.MethodGet, "/email/verify", nil)
	w := httptest.NewRecorder()

	k := &kit.Kit{
		Response: w,
		Request:  req,
	}

	err := HandleEmailVerify(k)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "invalid verification token")
}

func TestHandleProfileShowUnauthorized(t *testing.T) {
	setupTestDB(t)

	req := httptest.NewRequest(http.MethodGet, "/profile", nil)
	w := httptest.NewRecorder()

	k := &kit.Kit{
		Response: w,
		Request:  req,
	}

	assert.Panics(t, func() {
		HandleProfileShow(k)
	})
}

func TestHandleProfileUpdateUnauthorized(t *testing.T) {
	setupTestDB(t)

	form := url.Values{}
	form.Set("id", "1")
	form.Set("firstName", "New")
	form.Set("lastName", "Name")

	req := httptest.NewRequest(http.MethodPut, "/profile", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	k := &kit.Kit{
		Response: w,
		Request:  req,
	}

	assert.Panics(t, func() {
		HandleProfileUpdate(k)
	})
}

func TestHandleLoginDelete(t *testing.T) {
	setupTestDB(t)
	gormDB := db.Get()
	user := createTestUser(t, gormDB, "test@example.com", "password123")

	session := Session{
		UserID:    user.ID,
		Token:     "test-session-token",
		ExpiresAt: time.Now().Add(time.Hour * 24),
	}

	if err := gormDB.Create(&session).Error; err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	req := httptest.NewRequest(http.MethodDelete, "/logout", nil)
	w := httptest.NewRecorder()

	k := &kit.Kit{
		Response: w,
		Request:  req,
	}

	sess := k.GetSession(userSessionName)
	sess.Values["sessionToken"] = "test-session-token"
	sess.Save(req, w)

	err := HandleLoginDelete(k)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, w.Code)

	var deleted Session
	err = gormDB.Where("token = ?", "test-session-token").First(&deleted).Error
	assert.Error(t, err)
	assert.Equal(t, gorm.ErrRecordNotFound, err)
}
