// Package cache is the store tangent called its cache before naming it state.
// Its calls go to package state.
//
// Deprecated: use package state.
package cache

import (
	"time"

	internallog "github.com/ethanblackburn/tangent-sdk-go/internal/tangent/logs/log"
	"github.com/ethanblackburn/tangent-sdk-go/state"
	"go.bytecodealliance.org/cm"
)

// Value wraps Tangent cache scalars behind a small, user-friendly API.
// Construct values with helpers like String/Int/Float/Bool/Bytes and use the
// As* accessors to convert a Value back into native Go types.
type Value struct {
	scalar internallog.Scalar
}

// String returns a Value that stores a string.
func String(v string) Value {
	return Value{scalar: internallog.ScalarStr(v)}
}

// Int returns a Value that stores an int64.
func Int(v int64) Value {
	return Value{scalar: internallog.ScalarInt(v)}
}

// Float returns a Value that stores a float64.
func Float(v float64) Value {
	return Value{scalar: internallog.ScalarFloat(v)}
}

// Bool returns a Value that stores a bool.
func Bool(v bool) Value {
	return Value{scalar: internallog.ScalarBoolean(v)}
}

// Bytes returns a Value that stores an arbitrary byte slice.
func Bytes(data []byte) Value {
	return Value{scalar: internallog.ScalarBytes(cm.ToList(data))}
}

// AsString attempts to read a Value that was created with String.
func (v Value) AsString() (string, bool) {
	if s := v.scalar.Str(); s != nil {
		return *s, true
	}
	return "", false
}

// AsInt attempts to read a Value that was created with Int.
func (v Value) AsInt() (int64, bool) {
	if s := v.scalar.Int(); s != nil {
		return *s, true
	}
	return 0, false
}

// AsFloat attempts to read a Value that was created with Float.
func (v Value) AsFloat() (float64, bool) {
	if s := v.scalar.Float(); s != nil {
		return *s, true
	}
	return 0, false
}

// AsBool attempts to read a Value that was created with Bool.
func (v Value) AsBool() (bool, bool) {
	if s := v.scalar.Boolean(); s != nil {
		return *s, true
	}
	return false, false
}

// AsBytes attempts to read a Value that was created with Bytes.
func (v Value) AsBytes() ([]byte, bool) {
	if s := v.scalar.Bytes(); s != nil {
		return append([]byte(nil), s.Slice()...), true
	}
	return nil, false
}

// Get returns the current Value stored at key.
// ok will be false when the key is missing.
//
// Deprecated: use state.Get.
func Get(key string) (interface{}, bool, error) {
	return state.Get(key)
}

// Set stores value at key. A nil ttl takes the pipeline's default and a ttl of
// 0 never expires; see package state.
//
// Deprecated: use state.Set.
func Set(key string, value interface{}, ttl *time.Duration) error {
	return state.Set(key, value, ttl)
}

// Delete removes key and reports whether the key previously existed.
//
// Deprecated: use state.Delete.
func Delete(key string) (bool, error) {
	return state.Delete(key)
}
