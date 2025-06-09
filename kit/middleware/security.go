package middleware

import (
	"fmt"
	"net/http"
	"strings"
)

// SecurityHeadersConfig allows customization of security headers
type SecurityHeadersConfig struct {
	// Content Security Policy
	CSP map[string][]string
	
	// HSTS configuration
	HSTSMaxAge               int
	HSTSIncludeSubdomains    bool
	HSTSPreload              bool
	
	// X-Frame-Options
	FrameOptions string
	
	// X-Content-Type-Options
	ContentTypeOptions string
	
	// Referrer-Policy
	ReferrerPolicy string
	
	// Permissions-Policy
	PermissionsPolicy map[string]string
	
	// X-XSS-Protection
	XSSProtection string
}

// DefaultSecurityHeadersConfig provides sensible defaults for security headers
func DefaultSecurityHeadersConfig() SecurityHeadersConfig {
	return SecurityHeadersConfig{
		CSP: map[string][]string{
			"default-src": {"'self'"},
			"script-src":  {"'self'"},
			"style-src":   {"'self'"},
			"img-src":     {"'self'", "data:"},
			"font-src":    {"'self'"},
			"connect-src": {"'self'"},
		},
		HSTSMaxAge:            31536000, // 1 year in seconds
		HSTSIncludeSubdomains: true,
		HSTSPreload:           false,
		FrameOptions:          "DENY",
		ContentTypeOptions:    "nosniff",
		ReferrerPolicy:        "strict-origin-when-cross-origin",
		PermissionsPolicy: map[string]string{
			"geolocation":     "'self'",
			"camera":          "'none'",
			"microphone":      "'none'",
			"payment":         "'none'",
			"usb":             "'none'",
			"fullscreen":      "'self'",
			"display-capture": "'none'",
		},
		XSSProtection: "1; mode=block",
	}
}

// BuildCSPHeader constructs a Content-Security-Policy header value from a map of directives
func BuildCSPHeader(csp map[string][]string) string {
	policies := make([]string, 0, len(csp))
	for directive, sources := range csp {
		policies = append(policies, fmt.Sprintf("%s %s", directive, strings.Join(sources, " ")))
	}
	return strings.Join(policies, "; ")
}

// WithSecurityHeaders adds important security headers to HTTP responses
func WithSecurityHeaders(config SecurityHeadersConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Content-Security-Policy
			w.Header().Set("Content-Security-Policy", BuildCSPHeader(config.CSP))
			
			// HTTP Strict Transport Security
			var hstsValue string
			if config.HSTSIncludeSubdomains && config.HSTSPreload {
				hstsValue = fmt.Sprintf("max-age=%d; includeSubDomains; preload", config.HSTSMaxAge)
			} else if config.HSTSIncludeSubdomains {
				hstsValue = fmt.Sprintf("max-age=%d; includeSubDomains", config.HSTSMaxAge)
			} else {
				hstsValue = fmt.Sprintf("max-age=%d", config.HSTSMaxAge)
			}
			w.Header().Set("Strict-Transport-Security", hstsValue)
			
			// X-Frame-Options
			w.Header().Set("X-Frame-Options", config.FrameOptions)
			
			// X-Content-Type-Options
			w.Header().Set("X-Content-Type-Options", config.ContentTypeOptions)
			
			// Referrer-Policy
			w.Header().Set("Referrer-Policy", config.ReferrerPolicy)
			
			// Permissions-Policy
			if len(config.PermissionsPolicy) > 0 {
				policies := make([]string, 0, len(config.PermissionsPolicy))
				for feature, allowList := range config.PermissionsPolicy {
					policies = append(policies, fmt.Sprintf("%s=%s", feature, allowList))
				}
				w.Header().Set("Permissions-Policy", strings.Join(policies, ", "))
			}
			
			// X-XSS-Protection
			w.Header().Set("X-XSS-Protection", config.XSSProtection)
			
			next.ServeHTTP(w, r)
		})
	}
}

