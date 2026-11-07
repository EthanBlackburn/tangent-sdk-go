// Package state reads and writes the pipeline's state store: keyed values
// shared by every worker and scoped by the runtime to the partition of the
// records being processed.
//
// Values are a bool, an integer, a float, a string or a []byte. Get returns
// them as bool, int64, float64, string or []byte.
//
// A write's ttl decides when its entry expires: nil takes the pipeline's
// default (runtime.state.default_ttl_ms, 10 minutes unless configured), 0
// never expires, and anything else is rounded down to the millisecond. The
// runtime caps a default or explicit ttl at runtime.state.max_ttl_ms.
//
// Each call stands alone: two workers running the plugin can interleave one's
// read with the other's write. To read a key and write back something that
// depends on it, do both inside Transaction, or between Begin and Commit.
package state

import (
	"errors"
	"fmt"
	"time"

	internallog "github.com/ethanblackburn/tangent-sdk-go/internal/tangent/logs/log"
	internalstate "github.com/ethanblackburn/tangent-sdk-go/internal/tangent/logs/state"
	"go.bytecodealliance.org/cm"
)

// Entry is a key and its value, as Scan returns them.
type Entry struct {
	Key   string
	Value interface{}
}

// Get returns the value stored at key. ok is false when the key is missing or
// has expired.
func Get(key string) (interface{}, bool, error) {
	result := internalstate.Get(key)
	if result.IsErr() {
		return nil, false, errors.New(*result.Err())
	}

	opt := result.OK()
	if opt.None() {
		return nil, false, nil
	}

	value, err := fromScalar(opt.Value())
	if err != nil {
		return nil, false, err
	}
	return value, true, nil
}

// Set stores value at key.
func Set(key string, value interface{}, ttl *time.Duration) error {
	ttlOpt, err := ttlOption(ttl)
	if err != nil {
		return err
	}
	scalar, err := toScalar(value)
	if err != nil {
		return err
	}

	result := internalstate.Set(key, scalar, ttlOpt)
	if result.IsErr() {
		return errors.New(*result.Err())
	}
	return nil
}

// Delete removes key and reports whether it was there.
func Delete(key string) (bool, error) {
	result := internalstate.Del(key)
	if result.IsErr() {
		return false, errors.New(*result.Err())
	}
	ok := result.OK()
	if ok == nil {
		return false, errors.New("state: delete returned nil success flag")
	}
	return *ok, nil
}

// Incr adds by to the integer stored at key, a missing key counting as 0, and
// returns the new value. ttl applies to the entry as in Set.
func Incr(key string, by int64, ttl *time.Duration) (int64, error) {
	ttlOpt, err := ttlOption(ttl)
	if err != nil {
		return 0, err
	}

	result := internalstate.Incr(key, by, ttlOpt)
	if result.IsErr() {
		return 0, errors.New(*result.Err())
	}
	n := result.OK()
	if n == nil {
		return 0, errors.New("state: incr returned nil value")
	}
	return *n, nil
}

// Scan returns the entries with start <= key < end, in key order, at most
// limit of them. Keys that start with their time make a history a key range.
func Scan(start, end string, limit uint32) ([]Entry, error) {
	result := internalstate.Scan(start, end, limit)
	if result.IsErr() {
		return nil, errors.New(*result.Err())
	}
	list := result.OK()
	if list == nil {
		return nil, errors.New("state: scan returned nil list")
	}

	entries := make([]Entry, 0, list.Len())
	for _, kv := range list.Slice() {
		value, err := fromScalar(kv.F1)
		if err != nil {
			return nil, err
		}
		entries = append(entries, Entry{Key: kv.F0, Value: value})
	}
	return entries, nil
}

// Begin starts a transaction: the state calls after it see each other's
// writes, and land together on Commit or not at all on Rollback. No other
// state call in the pipeline runs while it is open, so begin late and commit
// early. A second Begin, or a remote call, inside one fails, and a call to the
// plugin that returns with one open has it rolled back.
func Begin() error {
	result := internalstate.Begin()
	if result.IsErr() {
		return errors.New(*result.Err())
	}
	return nil
}

// Commit keeps the open transaction's writes.
func Commit() error {
	result := internalstate.Commit()
	if result.IsErr() {
		return errors.New(*result.Err())
	}
	return nil
}

// Rollback undoes the open transaction's writes.
func Rollback() {
	internalstate.Rollback()
}

// Transaction runs fn in a transaction: committed when fn returns nil, rolled
// back when it returns an error, which Transaction returns. If fn panics, the
// plugin's call fails and the runtime rolls the transaction back.
func Transaction(fn func() error) error {
	if err := Begin(); err != nil {
		return err
	}
	if err := fn(); err != nil {
		Rollback()
		return err
	}
	return Commit()
}

func ttlOption(ttl *time.Duration) (cm.Option[uint64], error) {
	if ttl == nil {
		return cm.None[uint64](), nil
	}
	if ttl.Milliseconds() < 0 {
		return cm.None[uint64](), fmt.Errorf("ttl must be >= 0")
	}
	return cm.Some(uint64(ttl.Milliseconds())), nil
}

func toScalar(value interface{}) (internallog.Scalar, error) {
	switch value := value.(type) {
	case bool:
		return internallog.ScalarBoolean(value), nil
	case int:
		return internallog.ScalarInt(int64(value)), nil
	case int8:
		return internallog.ScalarInt(int64(value)), nil
	case int16:
		return internallog.ScalarInt(int64(value)), nil
	case int32:
		return internallog.ScalarInt(int64(value)), nil
	case int64:
		return internallog.ScalarInt(value), nil
	case float32:
		return internallog.ScalarFloat(float64(value)), nil
	case float64:
		return internallog.ScalarFloat(value), nil
	case string:
		return internallog.ScalarStr(value), nil
	case []byte:
		return internallog.ScalarBytes(cm.ToList(value)), nil
	}
	return internallog.Scalar{}, fmt.Errorf("state: unsupported value type %T", value)
}

func fromScalar(scalar internallog.Scalar) (interface{}, error) {
	if b := scalar.Boolean(); b != nil {
		return *b, nil
	}
	if bytes := scalar.Bytes(); bytes != nil {
		return append([]byte(nil), bytes.Slice()...), nil
	}
	if f := scalar.Float(); f != nil {
		return *f, nil
	}
	if i := scalar.Int(); i != nil {
		return *i, nil
	}
	if s := scalar.Str(); s != nil {
		return *s, nil
	}
	return nil, errors.New("state: unknown value type")
}
