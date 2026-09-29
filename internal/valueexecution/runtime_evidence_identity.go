package valueexecution

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"runtime"
	"strings"
	"sync"
)

var runtimeEvaluatorIdentityOnce sync.Once
var runtimeEvaluatorIdentityDigest string

// RuntimeEvidenceIdentities reports the runtime/toolchain identity and the
// exact bytes of the evaluator executable. If the executable cannot be read,
// its identity stays empty so the consuming provenance boundary remains
// UNKNOWN.
func RuntimeEvidenceIdentities() (toolchainDigest, evaluatorDigest string) {
	toolchainDigest = digestBytes([]byte(strings.Join([]string{
		runtime.Version(), runtime.GOOS, runtime.GOARCH,
	}, "\x00")))
	runtimeEvaluatorIdentityOnce.Do(func() {
		runtimeEvaluatorIdentityDigest = runtimeEvaluatorDigest()
	})
	return toolchainDigest, runtimeEvaluatorIdentityDigest
}

func runtimeEvaluatorDigest() string {
	executable, err := os.Executable()
	if err != nil {
		return ""
	}
	file, err := os.Open(executable)
	if err != nil {
		return ""
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return ""
	}
	return "sha256:" + hex.EncodeToString(hash.Sum(nil))
}
