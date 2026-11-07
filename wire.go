package tangent_sdk

import (
	"bytes"
	"errors"
	"sync"

	"github.com/ethanblackburn/tangent-sdk-go/internal/tangent/logs/log"
	"github.com/ethanblackburn/tangent-sdk-go/internal/tangent/logs/mapper"
	"github.com/mailru/easyjson"
	"github.com/mailru/easyjson/jwriter"

	"go.bytecodealliance.org/cm"
)

var (
	bufPool = sync.Pool{New: func() any { return new(bytes.Buffer) }}
)

type ProcessLog[T any] func(Log) (T, error)
type ProcessLogs[T any] func([]Log) ([]T, error)

// Wire connects metadata, probe selectors, and a handler to Tangent's ABI.
func Wire[T any](meta Metadata, selectors []Selector, handler ProcessLog[T], batchHandler ProcessLogs[T]) {
	mapper.Exports.Metadata = func() mapper.Meta {
		return meta.ToMapper()
	}

	mapper.Exports.Probe = func() cm.List[mapper.Selector] {
		mapped := make([]mapper.Selector, len(selectors))
		for i := range selectors {
			mapped[i] = selectors[i].toMapper()
		}
		return cm.ToList(mapped)
	}

	// The schema tangentgen wrote for T from its fields and json tags. Without
	// it (tangentgen did not run) the text is not JSON, and the runtime says
	// so when it loads the plugin.
	mapper.Exports.OutputSchema = func() string {
		var zero T
		if s, ok := any(zero).(interface{ TangentOutputSchema() string }); ok {
			return s.TangentOutputSchema()
		}
		return "no output schema: T has no TangentOutputSchema method. Did you recompile?"
	}

	mapper.Exports.ProcessLogs = func(input cm.List[log.Logview]) (res cm.Result[cm.List[uint8], cm.List[uint8], string]) {
		buf := bufPool.Get().(*bytes.Buffer)
		buf.Reset()
		defer bufPool.Put(buf)

		var jw jwriter.Writer

		writeOut := func(out T) error {
			outMarshal, ok := any(out).(easyjson.Marshaler)
			if !ok {
				return errors.New("output does not implement easyjson.Marshaler. Did you recompile?")
			}
			outMarshal.MarshalEasyJSON(&jw)
			jw.RawByte('\n')
			return nil
		}

		items := append([]log.Logview(nil), input.Slice()...)
		if batchHandler != nil {
			var logviews []Log
			for _, lv := range items {
				logviews = append(logviews, Log{logview: lv})
			}
			outs, err := batchHandler(logviews)
			for _, lv := range logviews {
				lv.logview.ResourceDrop()
			}
			logviews = logviews[:0]
			if err != nil {
				res.SetErr(err.Error())
				return
			}
			if len(outs) != len(items) {
				res.SetErr("batchHandler returned wrong number of outputs")
				return
			}
			for _, out := range outs {
				if err := writeOut(out); err != nil {
					res.SetErr(err.Error())
					return
				}
			}
		} else {
			for _, lv := range items {
				out, err := handler(Log{logview: lv})
				lv.ResourceDrop()
				if err != nil {
					res.SetErr(err.Error())
					return
				}
				if err := writeOut(out); err != nil {
					res.SetErr(err.Error())
					return
				}
			}
		}

		if jw.Error != nil {
			res.SetErr(jw.Error.Error())
			return
		}

		jw.DumpTo(buf)
		res.SetOK(cm.ToList(buf.Bytes()))
		return
	}
}
