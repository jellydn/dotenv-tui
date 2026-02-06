package detector

import (
	"encoding/base64"
	"regexp"
	"strings"
)

// IsSecret determines if a key-value pair appears to contain a secret
func IsSecret(key string, value string) bool {
	// Check for common non-secrets first
	if isCommonNonSecret(key) {
		return false
	}

	// Empty values cannot be secrets
	if len(value) == 0 {
		return false
	}

	// Check key patterns
	if isSecretKey(key) {
		return true
	}

	// Check value patterns
	if isSecretValue(value) {
		return true
	}

	return false
}

// GeneratePlaceholder creates a format-hint placeholder for a secret
func GeneratePlaceholder(key string, value string) string {
	// JWT token pattern
	if strings.HasPrefix(value, "eyJ") && len(value) > 50 {
		return "eyJ***"
	}

	// URL with credentials pattern
	if strings.Contains(value, "://") && strings.Contains(value, "@") {
		if strings.HasPrefix(value, "http://") {
			return "http://***"
		}
		if strings.HasPrefix(value, "https://") {
			return "https://***"
		}
		return "***://***"
	}

	// API key prefixes
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

	// Default placeholder
	return "***"
}

func isSecretKey(key string) bool {
	keyUpper := strings.ToUpper(key)

	secretPatterns := []string{
		"SECRET", "KEY", "TOKEN", "PASSWORD", "PASS", "AUTH",
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
			return true // This is a common non-secret
		}
	}

	return false // This is not in the list of common non-secrets
}

func isBase64(s string) bool {
	// Remove any whitespace
	s = strings.ReplaceAll(s, "\n", "")
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "\t", "")
	s = strings.ReplaceAll(s, "\r", "")

	// Empty string cannot be base64
	if len(s) == 0 {
		return false
	}

	// Check if the string length is a multiple of 4
	if len(s)%4 != 0 {
		return false
	}

	// Try to decode as base64
	_, err := base64.StdEncoding.DecodeString(s)
	return err == nil
}

func isHex(s string) bool {
	// Check if string contains only hex characters
	matched, _ := regexp.MatchString("^[0-9a-fA-F]+$", s)
	return matched
}
