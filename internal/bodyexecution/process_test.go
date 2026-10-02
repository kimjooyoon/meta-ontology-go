package bodyexecution

import (
	"context"
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
	case "body-helper-exit":
		os.Exit(7)
	case "body-helper-diagnostics":
		fmt.Fprint(os.Stderr, "diagnostic")
		os.Exit(0)
	case "body-helper-overflow":
		fmt.Print(strings.Repeat("x", 128<<10))
		os.Exit(0)
	case "body-helper-sleep":
		time.Sleep(5 * time.Second)
		os.Exit(0)
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
			if mode == "overflow" && !observation.OutputTruncated {
				t.Fatal("output limit is unobserved")
			}
			if observation.WallNS > int64(4*time.Second) {
				t.Fatal("bounded child did not return")
			}
		})
	}
}
