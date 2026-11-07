//go:build !wasm

package tangent_sdk

import (
	"testing"
	"unsafe"

	"go.bytecodealliance.org/cm"
)

// lens is what the fake host's logview.len returns, by path.
var lens = map[string]uint32{"tags": 3}

// wasmimport_LogviewLen stands in for the host's logview.len import, which
// exists only when the plugin runs inside tangent. Off wasm the generated
// declaration has no body, so this gives it one.
//
//go:linkname wasmimport_LogviewLen github.com/ethanblackburn/tangent-sdk-go/internal/tangent/logs/log.wasmimport_LogviewLen
func wasmimport_LogviewLen(self0 uint32, path0 *uint8, path1 uint32, result *cm.Option[uint32]) {
	if n, ok := lens[unsafe.String(path0, path1)]; ok {
		*result = cm.Some(n)
	} else {
		*result = cm.None[uint32]()
	}
}

func TestLogLen(t *testing.T) {
	var v Log

	got := v.Len("tags")
	if got == nil || *got != 3 {
		t.Fatalf("Len(%q) = %v, want 3", "tags", got)
	}

	if got := v.Len("missing"); got != nil {
		t.Fatalf("Len(%q) = %d, want nil", "missing", *got)
	}
}
