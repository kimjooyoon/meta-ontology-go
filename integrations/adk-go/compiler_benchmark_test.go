package adkcheck

import (
	"context"
	"encoding/json"
	"errors"
	"os/exec"
	"testing"
	"time"
)

// Both routes start the same real compiler. Allocation counts belong to this
// parent Go process; ns/op includes the child, but B/op does not describe it.
func BenchmarkCompiler(b *testing.B) {
	binary, root := compilerParityFixture(b)
	for _, workload := range []struct {
		name string
		path string
		exit int
	}{
		{"valid_check", "valid.gooo", 0},
		{"semantic_error", "invalid.gooo", 1},
		{"discovery", "valid.gooo", 0},
	} {
		args := []string{"check", "--semantic", "--json", workload.path}
		bridge := func() (CheckResult, error) {
			return check(context.Background(), binary, root, 15*time.Second, CheckArgs{Path: workload.path})
		}
		if workload.name == "discovery" {
			args = []string{"discover", "--json", "--query", "billing order", workload.path}
			bridge = func() (CheckResult, error) {
				return discover(context.Background(), binary, root, 15*time.Second,
					DiscoverArgs{Path: workload.path, Query: "billing order"})
			}
		}
		direct := func() (CheckResult, error) { return benchmarkDirectCompiler(binary, root, args) }
		want, err := direct()
		if err != nil || want.ExitCode != workload.exit {
			b.Fatalf("%s fixture: exit=%d err=%v", workload.name, want.ExitCode, err)
		}
		for _, route := range []struct {
			name string
			call func() (CheckResult, error)
		}{{"direct", direct}, {"bridge", bridge}} {
			b.Run(workload.name+"/"+route.name, func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					got, err := route.call()
					if err != nil || got != want {
						b.Fatalf("compiler result changed: got=%+v want=%+v err=%v", got, want, err)
					}
				}
			})
		}
	}
}

func benchmarkDirectCompiler(binary, root string, args []string) (CheckResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, binary, args...)
	cmd.Dir = root
	output, err := cmd.Output()
	if ctx.Err() != nil {
		return CheckResult{}, ctx.Err()
	}
	code := 0
	if err != nil {
		exit, ok := errors.AsType[*exec.ExitError](err)
		if !ok || exit.ExitCode() < 0 {
			return CheckResult{}, err
		}
		code = exit.ExitCode()
	}
	if !json.Valid(output) {
		return CheckResult{}, errors.New("compiler did not return JSON")
	}
	return CheckResult{ReportJSON: string(output), ExitCode: code}, nil
}
