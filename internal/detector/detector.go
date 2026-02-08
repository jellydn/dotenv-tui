// Package detector provides functions to identify secrets in environment variables.
package detector

import (
	"encoding/base64"
	"regexp"
	"strings"
)

var hexPattern = regexp.MustCompile("^[0-9a-fA-F]+$")

var (
	secretPatterns = []string{
		"SECRET", "TOKEN", "PASSWORD", "PASS",
		"AUTH", "CREDENTIAL", "PRIVATE",
		"API_KEY", "ACCESS_KEY",
	}
	commonNonSecrets = []string{
		"PORT", "HOST", "NODE_ENV", "APP_NAME",
		"DEBUG", "LOG_LEVEL",
		"ENV", "ENVIRONMENT", "VERSION", "LANG",
		"TIMEZONE", "REGION",
		"ENDPOINT", "URL", "URI", "DOMAIN",
		"SERVER", "CLUSTER",
	}
)

// prefixPlaceholder defines a secret prefix and its placeholder format.
type prefixPlaceholder struct {
	prefix      string
	placeholder string
}

// knownSecretPrefixes maps detection prefixes to their display placeholders.
// Keep sorted alphabetically by prefix for consistency.
var knownSecretPrefixes = []prefixPlaceholder{
	{"age-secret-key-", "age-***"},
	{"akiai", "akia***"},
	{"akia", "akia***"},
	{"gho_", "gho_***"},
	{"ghp_", "ghp_***"},
	{"ghs_", "ghs_***"},
	{"ghu_", "ghu_***"},
	{"github_pat_", "github_pat_***"},
	{"pk_live_", "pk_***"},
	{"pk_test_", "pk_***"},
	{"rk_live_", "rk_***"},
	{"rk_test_", "rk_***"},
	{"sk_live_", "sk_***"},
	{"sk_test_", "sk_***"},
	{"ssh-ed25519", "ssh-ed25519-***"},
	{"ssh-rsa", "ssh-rsa-***"},
	{"whsec_", "whsec_***"},
	{"xoxa-", "xox***"},
	{"xoxb-", "xox***"},
	{"xoxp-", "xox***"},
	{"ya29.", "ya29.***"},
}

// IsSecret determines if a key-value pair appears to contain a secret
func IsSecret(key string, value string) bool {
	if isCommonNonSecret(key) {
		return false
	}

	if len(value) == 0 {
		return false
	}

	if isSecretKey(key) {
		return true
	}

	if isSecretValue(value) {
		return true
	}

	return false
}

// GeneratePlaceholder creates a format-hint placeholder for a secret.
// The key parameter is kept for API consistency but not currently used.
func GeneratePlaceholder(_ string, value string) string {
	// Early return for empty values
	if len(value) == 0 {
		return "***"
	}

	// JWT tokens
	if strings.HasPrefix(value, "eyJ") && len(value) > 50 {
		return "eyJ***"
	}

	// URL patterns
	if strings.Contains(value, "://") && strings.Contains(value, "@") {
		return generateUrlPlaceholder(value)
	}

	// Check known prefixes (case-insensitive)
	lowerValue := strings.ToLower(value)
	if placeholder := findPrefixPlaceholder(lowerValue); placeholder != "" {
		return placeholder
	}

	return "***"
}

// generateUrlPlaceholder creates a placeholder for URL-style values.
func generateUrlPlaceholder(value string) string {
	if strings.HasPrefix(value, "http://") {
		return "http://***"
	}
	if strings.HasPrefix(value, "https://") {
		return "https://***"
	}
	return "***"
}

// findPrefixPlaceholder checks known secret prefixes and returns the appropriate placeholder.
func findPrefixPlaceholder(value string) string {
	for _, pp := range knownSecretPrefixes {
		if strings.HasPrefix(value, pp.prefix) {
			return pp.placeholder
		}
	}
	return ""
}

func isSecretKey(key string) bool {
	keyUpper := strings.ToUpper(key)
	for _, pattern := range secretPatterns {
		if strings.Contains(keyUpper, pattern) {
			return true
		}
	}
	return false
}

func isSecretValue(value string) bool {
	if len(value) == 0 {
		return false
	}

	// URLs with user:pass@ pattern
	if strings.Contains(value, "://") && strings.Contains(value, "@") {
		return true
	}

	// JWT tokens (must be longer than 50 chars to be considered a real JWT)
	if strings.HasPrefix(value, "eyJ") && len(value) > 50 {
		return true
	}

	// Known secret prefixes
	lowerValue := strings.ToLower(value)
	for _, pp := range knownSecretPrefixes {
		if strings.HasPrefix(lowerValue, pp.prefix) {
			return true
		}
	}

	// Base64 strings longer than 20 chars (but not JWT tokens)
	if len(value) > 20 && isBase64(value) && !strings.HasPrefix(value, "eyJ") {
		return true
	}

	// Hex strings longer than 32 chars
	if len(value) > 32 && isHex(value) {
		return true
	}

	return false
}

func isCommonNonSecret(key string) bool {
	keyUpper := strings.ToUpper(key)
	for _, pattern := range commonNonSecrets {
		if keyUpper == pattern {
			return true
		}
	}
	return false
}

func isBase64(s string) bool {
	s = strings.Map(func(r rune) rune {
		if r == '\n' || r == ' ' || r == '\t' || r == '\r' {
			return -1
		}
		return r
	}, s)

	if len(s) == 0 || len(s)%4 != 0 {
		return false
	}

	_, err := base64.StdEncoding.DecodeString(s)
	return err == nil
}

func isHex(s string) bool {
	return hexPattern.MatchString(s)
}
