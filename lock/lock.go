package lock

import (
	internallock "github.com/ethanblackburn/tangent-sdk-go/internal/tangent/logs/lock"
)

func Acquire(key string) bool {
	return internallock.Acquire(key)
}

func Release(key string) {
	internallock.Release(key)
}
