package unresolvedpolicy

import (
	"runtime"
	"testing"
	"unsafe"
)

func TestDispositionLayoutObservationFresh1867(t *testing.T) {
	t.Logf("DISPOSITION_LAYOUT goos=%s goarch=%s disposition_inline_bytes=%d refusal_inline_bytes=%d scope=Go_sizeof_excludes_string_backing_and_allocator_RSS;no_baseline_improvement_claim",
		runtime.GOOS, runtime.GOARCH, unsafe.Sizeof(SourceDisposition{}), unsafe.Sizeof(OperationRefusal{}))
}
