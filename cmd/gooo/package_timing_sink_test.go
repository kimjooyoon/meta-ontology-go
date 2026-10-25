package main

import (
	"bytes"
	"encoding/json"
	"io"
	"testing"
)

type unavailableTimingWriter struct {
	t      *testing.T
	output *bytes.Buffer
}

func (w unavailableTimingWriter) Write([]byte) (int, error) {
	if w.output.Len() == 0 {
		w.t.Error("diagnostic attempted before primary output publication")
	}
	return 0, io.ErrClosedPipe
}

func TestPackageTimingSinkDoesNotChangeInputValuesOrFailure(t *testing.T) {
	root := "../../examples/assembly-explainer/"
	base := []string{"package", "execute", "--inputs", root + "inputs.json", root + "gooo.workspace.json"}
	var original, observed, diagnostics bytes.Buffer
	if code := run(base, &original, &diagnostics); code != exitOK {
		t.Fatal(code, diagnostics.String())
	}
	if code := run(append(base, "--timing"), &observed, unavailableTimingWriter{t, &observed}); code != exitOK ||
		!bytes.Equal(original.Bytes(), observed.Bytes()) {
		t.Fatal("broken optional diagnostic changed actual values", code)
	}
	observed.Reset()
	if code := run(append(base, "--timing", "--json", "--go", "missing-go-timing-fixture"),
		&observed, unavailableTimingWriter{t, &observed}); code != exitFailure {
		t.Fatal("broken diagnostic hid original execution failure", code)
	}
	var receipt packageExecutionReceipt
	if err := json.Unmarshal(observed.Bytes(), &receipt); err != nil || receipt.Decision != "FAIL_CLOSED" || receipt.Error == "" {
		t.Fatal("original error receipt was lost", err)
	}
}
