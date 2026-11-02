package tangent_sdk

import "github.com/ethanblackburn/tangent-sdk-go/internal/tangent/logs/log"

type Log struct {
	logview log.Logview
}

func (v Log) Log() string {
	return v.logview.Log()
}

func (v Log) Has(path string) bool {
	return v.logview.Has(path)
}

func (v Log) Keys(path string) []string {
	keys := v.logview.Keys(path)

	return append([]string(nil), keys.Slice()...)
}

func (v Log) GetBool(path string) *bool {
	opt := v.logview.Get(path)
	if opt.None() {
		return nil
	}
	s := opt.Value()
	return s.Boolean()
}

func (v Log) GetInt64(path string) *int64 {
	opt := v.logview.Get(path)
	if opt.None() {
		return nil
	}
	s := opt.Value()
	return s.Int()
}

func (v Log) GetFloat64(path string) *float64 {
	opt := v.logview.Get(path)
	if opt.None() {
		return nil
	}
	s := opt.Value()
	return s.Float()
}

func (v Log) GetString(path string) *string {
	opt := v.logview.Get(path)
	if opt.None() {
		return nil
	}
	s := opt.Value()
	return s.Str()
}

func (v Log) Len(path string) *uint32 {
	return v.Len(path)
}

func (v Log) GetStringList(path string) ([]string, bool) {
	opt := v.logview.GetList(path)
	if opt.None() {
		return nil, false
	}
	lst := opt.Value()
	out := make([]string, 0, lst.Len())
	data := lst.Slice()
	for i := 0; i < int(lst.Len()); i++ {
		if p := data[i].Str(); p != nil {
			out = append(out, *p)
		}
	}
	return out, true
}

func (v Log) GetFloat64List(path string) ([]float64, bool) {
	opt := v.logview.GetList(path)
	if opt.None() {
		return nil, false
	}
	lst := opt.Value()
	out := make([]float64, 0, lst.Len())
	data := lst.Slice()
	for i := 0; i < int(lst.Len()); i++ {
		if p := data[i].Float(); p != nil {
			out = append(out, *p)
		}
	}
	return out, true
}

func (v Log) GetInt64List(path string) ([]int64, bool) {
	opt := v.logview.GetList(path)
	if opt.None() {
		return nil, false
	}
	lst := opt.Value()
	out := make([]int64, 0, lst.Len())
	data := lst.Slice()
	for i := 0; i < int(lst.Len()); i++ {
		if p := data[i].Int(); p != nil {
			out = append(out, *p)
		}
	}
	return out, true
}
