package cache

import (
	"errors"
	"fmt"
	"time"

	internalcache "github.com/ethanblackburn/tangent-sdk-go/internal/tangent/logs/cache"
	internallog "github.com/ethanblackburn/tangent-sdk-go/internal/tangent/logs/log"
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
func Get(key string) (interface{}, bool, error) {
	result := internalcache.Get(key)
	if result.IsErr() {
		return nil, false, errors.New(*result.Err())
	}

	opt := result.OK()
	if opt.None() {
		return nil, false, nil
	}

	scalar := opt.Value()
	if b := scalar.Boolean(); b != nil {
		return *b, true, nil
	}
	if bytes := scalar.Bytes(); bytes != nil {
		return append([]byte(nil), bytes.Slice()...), true, nil
	}
	if f := scalar.Float(); f != nil {
		return *f, true, nil
	}
	if i := scalar.Int(); i != nil {
		return *i, true, nil
	}
	if s := scalar.Str(); s != nil {
		return *s, true, nil
	}

	return nil, false, errors.New("unknown value type")
}

// Set stores value at key. If ttl is nil the value never expires. When ttl is
// provided it is rounded down to the nearest millisecond before being sent to
// the host.
func Set(key string, value interface{}, ttl *time.Duration) error {
	var ttlOpt cm.Option[uint64]

	if ttl == nil {
		ttlOpt = cm.None[uint64]()
	} else {
		if ttl.Milliseconds() < 0 {
			return fmt.Errorf("ttl must be >= 0")
		}
		ms := uint64(ttl.Milliseconds())
		ttlOpt = cm.Some(ms)
	}

	var scalarVal internalcache.Scalar
	switch value := value.(type) {
	case bool:
		scalarVal = internallog.ScalarBoolean(value)
	case int:
		scalarVal = internallog.ScalarInt(int64(value))
	case int16:
		scalarVal = internallog.ScalarInt(int64(value))
	case int32:
		scalarVal = internallog.ScalarInt(int64(value))
	case int64:
		scalarVal = internallog.ScalarInt(int64(value))
	case float32:
		scalarVal = internallog.ScalarFloat(float64(value))
	case float64:
		scalarVal = internallog.ScalarFloat(value)
	case string:
		scalarVal = internallog.ScalarStr(value)
	case []uint8:
		scalarVal = internallog.ScalarBytes(cm.ToList(value))
	}

	result := internalcache.Set(key, scalarVal, ttlOpt)
	if result.IsErr() {
		return errors.New(*result.Err())
	}

	return nil
}

// Delete removes key from the cache and reports whether the key previously
// existed.
func Delete(key string) (bool, error) {
	result := internalcache.Del(key)
	if result.IsErr() {
		return false, errors.New(*result.Err())
	}
	ok := result.OK()
	if ok == nil {
		return false, errors.New("cache: delete returned nil success flag")
	}
	return *ok, nil
}
