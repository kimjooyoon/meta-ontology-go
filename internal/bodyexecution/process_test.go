package bodyexecution

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
)

func TestRuntimeProcessHelper(t *testing.T) {
	if len(os.Args) < 2 {
		return
	}
	switch os.Args[len(os.Args)-1] {
	case "body-helper-ok":
		fmt.Print("ready")
		os.Exit(0)
	case "body-helper-exit":
		fmt.Fprint(os.Stderr, "native helper failed")
		os.Exit(7)
	case "body-helper-diagnostics":
		fmt.Fprint(os.Stderr, "diagnostic")
		os.Exit(0)
	case "body-helper-raw-diagnostics":
		_, _ = os.Stderr.Write([]byte{'x', 0xff, 0})
		os.Exit(0)
	case "body-helper-overflow":
		fmt.Print(strings.Repeat("x", 128<<10))
		os.Exit(0)
	case "body-helper-sleep":
		time.Sleep(5 * time.Second)
		os.Exit(0)
	}
}

func TestRuntimeChildKeepsFailureCauseAndDiagnostics(t *testing.T) {
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	_, observation, err := process(context.Background(), t.TempDir(), binary, nil,
		"-test.run=^TestRuntimeProcessHelper$", "--", "body-helper-exit")
	raw, _ := json.Marshal(observation)
	var fields struct {
		Failure     string
		Diagnostics []byte
	}
	if decodeErr := json.Unmarshal(raw, &fields); decodeErr != nil {
		t.Fatal(decodeErr)
	}
	if err == nil || fields.Failure != "exit status 7" || string(fields.Diagnostics) != "native helper failed" ||
		observation.DiagnosticsBytes != len("native helper failed") || observation.StderrSHA256 != digest([]byte("native helper failed")) {
		t.Fatalf("failed child lost the underlying cause or original diagnostics: %s %v", raw, err)
	}
	_, observation, err = process(context.Background(), t.TempDir(), binary, nil,
		"-test.run=^TestRuntimeProcessHelper$", "--", "body-helper-raw-diagnostics")
	raw, _ = json.Marshal(observation)
	if decodeErr := json.Unmarshal(raw, &fields); decodeErr != nil {
		t.Fatal(decodeErr)
	}
	if err == nil || string(fields.Diagnostics) != string([]byte{'x', 0xff, 0}) ||
		observation.StderrSHA256 != digest(fields.Diagnostics) || observation.DiagnosticsBytes != len(fields.Diagnostics) {
		t.Fatal("diagnostics changed during JSON round trip", fields)
	}
}

func TestRuntimeProcessCopyKeepsIndependentDiagnostics(t *testing.T) {
	budget := int64(123)
	original := ProcessObservation{Diagnostics: []byte("original"), Timing: &ProcessTiming{StartNS: 45, DeadlineRemainingNS: &budget}}
	copied := copyProcess(original)
	copied.Diagnostics[0] = 'x'
	copied.Timing.StartNS = 9
	*copied.Timing.DeadlineRemainingNS = 7
	if string(original.Diagnostics) != "original" || original.Timing.StartNS != 45 || *original.Timing.DeadlineRemainingNS != 123 {
		t.Fatal("copy changed the original diagnostic bytes")
	}
}

func TestRuntimeChildBoundsAndCancellation(t *testing.T) {
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for _, mode := range []string{"exit", "diagnostics", "overflow", "sleep"} {
		t.Run(mode, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			if mode == "sleep" {
				cancel()
				ctx, cancel = context.WithTimeout(context.Background(), 100*time.Millisecond)
			}
			defer cancel()
			_, observation, err := process(ctx, t.TempDir(), binary, nil, "-test.run=^TestRuntimeProcessHelper$", "--", "body-helper-"+mode)
			if err == nil || !observation.Started || observation.Completed || observation.ExitCode == nil {
				t.Fatalf("child failure was lost: %+v %v", observation, err)
			}
			if mode == "sleep" && !observation.TimedOut {
				t.Fatal("timeout is unobserved")
			}
			if observation.Timing == nil || observation.Timing.StartNS+observation.Timing.WaitNS != observation.WallNS {
				t.Fatal("start/wait accounting differs from total", observation)
			}
			if mode == "overflow" && !observation.OutputTruncated {
				t.Fatal("output limit is unobserved")
			}
			if observation.WallNS > int64(4*time.Second) {
				t.Fatal("bounded child did not return")
			}
		})
	}
}
