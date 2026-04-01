package utils

import (
	"fmt"
	"regexp"
	"strings"
)

var nonSlugPattern = regexp.MustCompile(`[^a-z0-9]+`)

// Slugify converts a free-form title into a URL-safe ASCII slug.
func Slugify(value string, fallbackPrefix string) string {
	normalized := strings.ToLower(strings.TrimSpace(value))
	normalized = nonSlugPattern.ReplaceAllString(normalized, "-")
	normalized = strings.Trim(normalized, "-")

	if normalized == "" {
		return fallbackPrefix
	}

	return normalized
}

// SlugWithIDFallback returns a deterministic slug for legacy rows when a title
// does not produce an ASCII slug.
func SlugWithIDFallback(value string, fallbackPrefix string, id int64) string {
	base := Slugify(value, "")
	if base != "" {
		return base
	}

	return fmt.Sprintf("%s-%d", fallbackPrefix, id)
}
