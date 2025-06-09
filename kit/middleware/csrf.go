package middleware

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/khulnasoft/superkit/kit"
)

const (
	csrfCookieName  = "csrf-token"
	csrfHeaderName  = "X-CSRF-Token"
	csrfFormField   = "_csrf"
	csrfTokenLength = 32 // 32 bytes = 256 bits
	defaultMaxAge   = 3600
)

var (
	ErrCSRFTokenMissing = errors.New("CSRF token missing")
	ErrCSRFTokenInvalid = errors.New("CSRF token invalid")
)

// CSRFConfig allows customization of CSRF protection
type CSRFConfig struct {
	// CookiePath sets the Path attribute of the CSRF cookie
	CookiePath string
	
	// CookieDomain sets the Domain attribute of the CSRF cookie
	CookieDomain string
	
	// CookieMaxAge sets the MaxAge attribute of the CSRF cookie
	CookieMaxAge int
	
	// CookieSecure sets the Secure attribute of the CSRF cookie
	CookieSecure bool
	
	// CookieHTTPOnly sets the HttpOnly attribute of the CSRF cookie
	CookieHTTPOnly bool
	
	// TrustedOrigins is a list of origins that are trusted to make cross-origin requests
	TrustedOrigins []string
	
	// ErrorHandler is called when CSRF validation fails
	ErrorHandler func(kit *kit.Kit, err error)
	
	// ExemptPaths is a list of path prefixes that are exempt from CSRF protection
	ExemptPaths []string
	
	// ExemptedMethods is a list of HTTP methods that are exempt from CSRF protection
	ExemptedMethods []string
}

// DefaultCSRFConfig provides sensible defaults for CSRF configuration
func DefaultCSRFConfig() CSRFConfig {
	return CSRFConfig{
		CookiePath:     "/",
		CookieMaxAge:   defaultMaxAge,
		CookieSecure:   true,
		CookieHTTPOnly: true,
		ExemptedMethods: []string{
			http.MethodGet,
			http.MethodHead,
			http.MethodOptions,
			http.MethodTrace,
		},
		ErrorHandler: defaultCSRFErrorHandler,
	}
}

// defaultCSRFErrorHandler returns a 403 Forbidden error with a message
func defaultCSRFErrorHandler(k *kit.Kit, err error) {
	k.Text(http.StatusForbidden, "Forbidden: "+err.Error())
}

// CSRFToken represents a CSRF token with its creation time
type CSRFToken struct {
	Token     string
	CreatedAt time.Time
}

// CSRFManager manages CSRF token creation and validation
type CSRFManager struct {
	config CSRFConfig
	mutex  sync.Mutex
	// Optional token store for server-side validation
	// tokenStore map[string]CSRFToken
}

// NewCSRFManager creates a new CSRF manager with the provided configuration
func NewCSRFManager(config CSRFConfig) *CSRFManager {
	return &CSRFManager{
		config: config,
		// tokenStore: make(map[string]CSRFToken),
	}
}

// GenerateToken creates a new CSRF token
func (cm *CSRFManager) GenerateToken() (string, error) {
	bytes := make([]byte, csrfTokenLength)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	token := base64.StdEncoding.EncodeToString(bytes)
	
	// Optionally store in token store for server-side validation
	// cm.mutex.Lock()
	// cm.tokenStore[token] = CSRFToken{
	// 	Token:     token,
	// 	CreatedAt: time.Now(),
	// }
	// cm.mutex.Unlock()
	
	return token, nil
}

// IsPathExempt checks if the given path is exempt from CSRF protection
func (cm *CSRFManager) IsPathExempt(path string) bool {
	for _, exemptPath := range cm.config.ExemptPaths {
		if strings.HasPrefix(path, exemptPath) {
			return true
		}
	}
	return false
}

// IsMethodExempt checks if the given HTTP method is exempt from CSRF protection
func (cm *CSRFManager) IsMethodExempt(method string) bool {
	for _, exemptMethod := range cm.config.ExemptedMethods {
		if method == exemptMethod {
			return true
		}
	}
	return false
}

// withCSRFCookie sets a CSRF token cookie
func (cm *CSRFManager) setCSRFCookie(w http.ResponseWriter, token string) {
	cookie := &http.Cookie{
		Name:     csrfCookieName,
		Value:    token,
		Path:     cm.config.CookiePath,
		Domain:   cm.config.CookieDomain,
		MaxAge:   cm.config.CookieMaxAge,
		Secure:   cm.config.CookieSecure,
		HttpOnly: cm.config.CookieHTTPOnly,
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(w, cookie)
}

// getCSRFToken extracts the CSRF token from the request
func (cm *CSRFManager) getCSRFToken(r *http.Request) string {
	// Check header first
	token := r.Header.Get(csrfHeaderName)
	if token != "" {
		return token
	}
	
	// Then check form value
	if r.PostForm == nil {
		r.ParseForm()
	}
	token = r.PostForm.Get(csrfFormField)
	if token != "" {
		return token
	}
	
	// Finally check cookie
	cookie, err := r.Cookie(csrfCookieName)
	if err == nil {
		return cookie.Value
	}
	
	return ""
}

// WithCSRF adds CSRF protection middleware
func WithCSRF(config CSRFConfig) func(http.Handler) http.Handler {
	csrfManager := NewCSRFManager(config)
	
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			kit := &kit.Kit{
				Response: w,
				Request:  r,
			}
			
			// Skip CSRF check for exempted paths
			if csrfManager.IsPathExempt(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}
			
			// Skip CSRF check for exempted methods
			if csrfManager.IsMethodExempt(r.Method) {
				// Generate and set CSRF token for subsequent requests
				token, err := csrfManager.GenerateToken()
				if err != nil {
					config.ErrorHandler(kit, fmt.Errorf("failed to generate CSRF token: %w", err))
					return
				}
				
				csrfManager.setCSRFCookie(w, token)
				
				// Store token in context for template rendering
				ctx := context.WithValue(r.Context(), csrfContextKey{}, token)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
			
			// Validate CSRF token for non-exempted methods
			requestToken := csrfManager.getCSRFToken(r)
			if requestToken == "" {
				config.ErrorHandler(kit, ErrCSRFTokenMissing)
				return
			}
			
			// For better security, validate against server-side store
			// But for simplicity, we're just checking if the token exists in the cookie
			cookie, err := r.Cookie(csrfCookieName)
			if err != nil || cookie.Value != requestToken {
				config.ErrorHandler(kit, ErrCSRFTokenInvalid)
				return
			}
			
			// Regenerate token for security (token rotation)
			newToken, err := csrfManager.GenerateToken()
			if err != nil {
				config.ErrorHandler(kit, fmt.Errorf("failed to generate CSRF token: %w", err))
				return
			}
			
			csrfManager.setCSRFCookie(w, newToken)
			
			// Store new token in context for template rendering
			ctx := context.WithValue(r.Context(), csrfContextKey{}, newToken)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// csrfContextKey is used to store the CSRF token in the request context
type csrfContextKey struct{}

// GetCSRFToken gets the CSRF token from the request context
func GetCSRFToken(r *http.Request) string {
	token, ok := r.Context().Value(csrfContextKey{}).(string)
	if !ok {
		return ""
	}
	return token
}

// CSRFField returns a hidden input field with the CSRF token
func CSRFField(r *http.Request) string {
	token := GetCSRFToken(r)
	if token == "" {
		return ""
	}
	return fmt.Sprintf(`<input type="hidden" name="%s" value="%s">`, csrfFormField, token)
}

