package tangent_sdk

import (
	"github.com/ethanblackburn/tangent-sdk-go/internal/tangent/logs/log"
	"github.com/ethanblackburn/tangent-sdk-go/internal/tangent/logs/mapper"

	"go.bytecodealliance.org/cm"
)

// Predicate represents a single probe predicate.
type Predicate struct {
	inner mapper.Pred
}

// Has returns a predicate that matches when path exists.
func Has(path string) Predicate {
	return Predicate{inner: mapper.PredHas(path)}
}

// EqString matches when the field at path equals value.
func EqString(path, value string) Predicate {
	return Predicate{inner: mapper.PredEq(cm.Tuple[string, mapper.Scalar]{
		F0: path,
		F1: log.ScalarStr(value),
	})}
}

// EqInt matches when the field at path equals value.
func EqInt(path string, value int64) Predicate {
	return Predicate{inner: mapper.PredEq(cm.Tuple[string, mapper.Scalar]{
		F0: path,
		F1: log.ScalarInt(value),
	})}
}

// EqFloat matches when the field at path equals value.
func EqFloat(path string, value float64) Predicate {
	return Predicate{inner: mapper.PredEq(cm.Tuple[string, mapper.Scalar]{
		F0: path,
		F1: log.ScalarFloat(value),
	})}
}

// EqBool matches when the field at path equals value.
func EqBool(path string, value bool) Predicate {
	return Predicate{inner: mapper.PredEq(cm.Tuple[string, mapper.Scalar]{
		F0: path,
		F1: log.ScalarBoolean(value),
	})}
}

// Prefix matches when the field at path has the given prefix.
func Prefix(path, prefix string) Predicate {
	return Predicate{inner: mapper.PredPrefix([2]string{path, prefix})}
}

// Regex matches when the field at path matches pattern.
func Regex(path, pattern string) Predicate {
	return Predicate{inner: mapper.PredRegex([2]string{path, pattern})}
}

// InStrings matches when the field at path is one of values.
func InStrings(path string, values ...string) Predicate {
	scalars := make([]mapper.Scalar, len(values))
	for i := range values {
		scalars[i] = log.ScalarStr(values[i])
	}
	return Predicate{inner: mapper.PredIn(cm.Tuple[string, cm.List[mapper.Scalar]]{
		F0: path,
		F1: cm.ToList(scalars),
	})}
}

// Selector groups predicates into AND/OR/NONE sets.
type Selector struct {
	Any  []Predicate
	All  []Predicate
	None []Predicate
}

func (s Selector) toMapper() mapper.Selector {
	return mapper.Selector{
		Any:  toPredList(s.Any),
		All:  toPredList(s.All),
		None: toPredList(s.None),
	}
}

func toPredList(preds []Predicate) cm.List[mapper.Pred] {
	if len(preds) == 0 {
		return cm.ToList([]mapper.Pred{})
	}
	out := make([]mapper.Pred, len(preds))
	for i := range preds {
		out[i] = preds[i].inner
	}
	return cm.ToList(out)
}
