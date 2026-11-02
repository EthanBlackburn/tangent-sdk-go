package config

import (
	internalconfig "github.com/ethanblackburn/tangent-sdk-go/internal/tangent/logs/config"
)

// Get returns the current Value stored at key.
// ok will be false when the key is missing.
func Get(key string) (string, bool) {
	result := internalconfig.Get(key)
	if result.Some() != nil {
		return result.Value(), true
	}

	return "", false
}
