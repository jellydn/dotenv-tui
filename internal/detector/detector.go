// Package detector provides functions to identify secrets in environment variables.
package detector

import (
	"encoding/base64"
	"regexp"
	"strings"
)

var hexPattern = regexp.MustCompile("^[0-9a-fA-F]+$")

var (
	secretPatternsMap = map[string]bool{
		"SECRET": true, "TOKEN": true, "PASSWORD": true, "PASS": true,
		"AUTH": true, "CREDENTIAL": true, "PRIVATE": true,
		"API_KEY": true, "ACCESS_KEY": true,
	}
	knownSecretPrefixesMap = map[string]bool{
		"sk_live_": true, "sk_test_": true, "rk_live_": true, "rk_test_": true,
		"ghp_": true, "gho_": true, "ghu_": true, "ghs_": true, "github_pat_": true,
		"pk_live_": true, "pk_test_": true,
		"xoxb-": true, "xoxp-": true, "xoxa-": true,
		"ya29.":  true,
		"whsec_": true,
		"akiai":  true, "akia": true,
		"age-secret-key-": true,
	}
	commonNonSecretsMap = map[string]bool{
		"PORT": true, "HOST": true, "NODE_ENV": true, "APP_NAME": true,
		"DEBUG": true, "LOG_LEVEL": true,
		"ENV": true, "ENVIRONMENT": true, "VERSION": true, "LANG": true,
		"TIMEZONE": true, "REGION": true,
		"ENDPOINT": true, "URL": true, "URI": true, "DOMAIN": true,
		"SERVER": true, "CLUSTER": true,
	}
)

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
	prefixes := []struct {
		prefix      string
		placeholder string
	}{
		{"sk_live_", "sk_***"},
		{"sk_test_", "sk_***"},
		{"ghp_", "ghp_***"},
		{"gho_", "ghp_***"},
		{"ghu_", "ghp_***"},
		{"pk_test_", "pk_***"},
		{"pk_live_", "pk_***"},
		{"xoxb-", "xox***"},
		{"xoxp-", "xox***"},
		{"ya29.", "ya29.***"},
		{"ssh-rsa", "ssh-***"},
		{"ssh-ed25519", "ssh-***"},
	}

	for _, p := range prefixes {
		if strings.HasPrefix(value, p.prefix) {
			return p.placeholder
		}
	}
	return ""
}

func isSecretKey(key string) bool {
	keyUpper := strings.ToUpper(key)
	for pattern := range secretPatternsMap {
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
	for prefix := range knownSecretPrefixesMap {
		if strings.HasPrefix(lowerValue, prefix) {
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
	return commonNonSecretsMap[strings.ToUpper(key)]
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
