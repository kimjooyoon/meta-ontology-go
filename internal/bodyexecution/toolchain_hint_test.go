package bodyexecution

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestWrongNativeGoRetainsObservedVersionAndNextAction(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX Go-version fixture")
	}
	tool := filepath.Join(t.TempDir(), "observed-go")
	if err := os.WriteFile(tool, []byte("#!/bin/sh\necho 'go version go1.26.5 "+runtime.GOOS+"/"+runtime.GOARCH+"'\n"), 0700); err != nil {
		t.Fatal(err)
	}
	var o Observation
	err := observeToolchain(context.Background(), tool, &o, nil)
	if err == nil || !strings.Contains(err.Error(), "go1.26.5") || !strings.Contains(err.Error(), "--go-bin") {
		t.Fatalf("missing observed version/next action: %v", err)
	}
	if o.GoVersion != "go version go1.26.5 "+runtime.GOOS+"/"+runtime.GOARCH || !o.Toolchain.Completed || len(o.Runs) != 0 {
		t.Fatalf("original tool observation = %+v", o)
	}
}
