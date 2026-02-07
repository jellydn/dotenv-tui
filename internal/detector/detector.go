// Package detector provides functions to identify secrets in environment variables.
package detector

import (
	"encoding/base64"
	"regexp"
	"strings"
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
	if strings.HasPrefix(value, "eyJ") && len(value) > 50 {
		return "eyJ***"
	}

	if strings.Contains(value, "://") && strings.Contains(value, "@") {
		if strings.HasPrefix(value, "http://") {
			return "http://***"
		}
		if strings.HasPrefix(value, "https://") {
			return "https://***"
		}
		return "***"
	}

	lowerValue := strings.ToLower(value)
	if strings.HasPrefix(lowerValue, "sk_live_") || strings.HasPrefix(lowerValue, "sk_test_") {
		return "sk_***"
	}
	if strings.HasPrefix(lowerValue, "ghp_") || strings.HasPrefix(lowerValue, "gho_") || strings.HasPrefix(lowerValue, "ghu_") {
		return "ghp_***"
	}
	if strings.HasPrefix(lowerValue, "pk_test_") || strings.HasPrefix(lowerValue, "pk_live_") {
		return "pk_***"
	}
	if strings.HasPrefix(lowerValue, "xoxb-") || strings.HasPrefix(lowerValue, "xoxp-") {
		return "xox***"
	}
	if strings.HasPrefix(lowerValue, "ya29.") {
		return "ya29.***"
	}
	if strings.HasPrefix(lowerValue, "ssh-rsa") || strings.HasPrefix(lowerValue, "ssh-ed25519") {
		return "ssh-***"
	}

	return "***"
}

func isSecretKey(key string) bool {
	keyUpper := strings.ToUpper(key)

	secretPatterns := []string{
		"SECRET", "TOKEN", "PASSWORD", "PASS", "AUTH",
		"CREDENTIAL", "PRIVATE", "API_KEY", "ACCESS_KEY",
	}

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
	knownSecretPrefixes := []string{
		"sk_live_", "sk_test_", "rk_live_", "rk_test_",
		"ghp_", "gho_", "ghu_", "ghs_", "github_pat_",
		"pk_live_", "pk_test_",
		"xoxb-", "xoxp-", "xoxa-",
		"ya29.",
		"whsec_",
		"akiai", "akia", // AWS access key
		"age-secret-key-",
	}
	for _, prefix := range knownSecretPrefixes {
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
	keyUpper := strings.ToUpper(key)

	commonNonSecrets := []string{
		"PORT", "HOST", "NODE_ENV", "APP_NAME", "DEBUG", "LOG_LEVEL",
		"ENV", "ENVIRONMENT", "VERSION", "LANG", "TIMEZONE", "REGION",
		"ENDPOINT", "URL", "URI", "DOMAIN", "SERVER", "CLUSTER",
	}

	for _, nonSecret := range commonNonSecrets {
		if keyUpper == nonSecret {
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
	// Check if string contains only hex characters
	matched, _ := regexp.MatchString("^[0-9a-fA-F]+$", s)
	return matched
}
