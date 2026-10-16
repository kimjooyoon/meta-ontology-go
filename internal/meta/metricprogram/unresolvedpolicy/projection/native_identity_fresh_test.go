package policyprojection

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestDispositionStableIdentityNativeFresh1866(t *testing.T) {
	for _, name := range []string{"renamed-record-activity-field", "source-controls-all-Boolean-values"} {
		t.Run(name, func(t *testing.T) {
			source, err := os.ReadFile("../policy.gooo")
			if err != nil {
				t.Fatal(err)
			}
			target, err := os.ReadFile("../disposition.go")
			if err != nil {
				t.Fatal(err)
			}
			source = renamedDisposition1866(source, name != "renamed-record-activity-field")
			target = bytes.Replace(target, []byte("Reason "), []byte("NativeCause "), 1)
			result, err := Generate("renamed1866.gooo", source, target)
			if err != nil {
				t.Fatal(err)
			}
			if result.Report.Activity != "ExplainOrigin" || result.Report.ActivityID != ActivityID ||
				result.Report.RecordAssembly.ModelCalls != 0 || result.Report.RecordAssembly.FieldsPassed != 24 {
				t.Fatal("display rename changed stable activity or finite source contract")
			}
			harness := nativeDispositionHarness1866(name != "renamed-record-activity-field")
			runNativeDisposition1866(t, source, target, result.Source, []byte(harness))
		})
	}
}

func renamedDisposition1866(source []byte, flags bool) []byte {
	text := strings.NewReplacer("OperationDisposition", "OperationSummary", "DescribeOperation", "ExplainOrigin",
		"OperationReason", "PickCause", "field reason id", "field cause id", "    reason: ", "    cause: ",
		`\"reason\":`, `\"cause\":`).Replace(string(source))
	if flags {
		for _, field := range []string{"readiness_credit", "repository_workspace_writes", "promotion_authorized"} {
			text = strings.ReplaceAll(text, field+": false", field+": true")
			text = strings.ReplaceAll(text, `\"`+field+`\":false`, `\"`+field+`\":true`)
		}
	}
	return []byte(text)
}

func runNativeDisposition1866(t *testing.T, source, target, projection, harness []byte) {
	t.Helper()
	directory := t.TempDir()
	for path, raw := range map[string][]byte{"go.mod": []byte("module native1866\ngo 1.27.2\n"),
		"target.go": target, "projection.go": projection, "native_test.go": harness} {
		if err := os.WriteFile(filepath.Join(directory, path), raw, 0600); err != nil {
			t.Fatal(err)
		}
	}
	binary := filepath.Join(directory, "disposition1866.test")
	if runtime.GOOS == "windows" {
		binary += ".exe"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, "go", "test", "-p=1", "-c", "-o", binary, ".")
	command.Dir = directory
	command.Env = append(os.Environ(), "GOWORK=off", "GOTOOLCHAIN=local")
	if raw, err := command.CombinedOutput(); err != nil {
		t.Fatalf("native fixture compile: %v: %s", err, raw)
	}
	raw, err := exec.CommandContext(ctx, binary, "-test.v", "-test.run=^TestNativeRecord1866$").CombinedOutput()
	if err != nil || bytes.Count(raw, []byte("NATIVE_ROW1866 ")) != 4 {
		t.Fatalf("native fixture execution: %v: %s", err, raw)
	}
	executable, err := os.ReadFile(binary)
	if err != nil {
		t.Fatal(err)
	}
	identity := map[string]any{"source_digest": dispositionDigest1866(source),
		"target_digest": dispositionDigest1866(target), "projection_digest": dispositionDigest1866(projection),
		"harness_digest": dispositionDigest1866(harness), "binary_digest": dispositionDigest1866(executable),
		"transcript_digest": dispositionDigest1866(raw), "go_version": runtime.Version(), "goos": runtime.GOOS,
		"goarch": runtime.GOARCH, "native_processes": 1, "inputs": 4, "field_values": 24,
		"scope": "New display/value mutation fixture. All four inputs overlap its declared cases; no holdout or execution authority."}
	encoded, err := json.Marshal(identity)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("DISPOSITION_NATIVE_ORIGIN1866 %s", encoded)
}

func dispositionDigest1866(raw []byte) string { return fmt.Sprintf("sha256:%x", sha256.Sum256(raw)) }
