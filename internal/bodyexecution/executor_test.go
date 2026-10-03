package bodyexecution

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func nativeTool() string {
	name := "go"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	return filepath.Join(runtime.GOROOT(), "bin", name)
}

func TestExecutorAcceptsLiveTypedScopeParentWithoutChangingIt(t *testing.T) {
	source, doc, prior, _ := fixture(t)
	prior.Report.CompletenessReceipt.Scope["typed_live_record"] = struct {
		Z int `json:"z"`
		A int `json:"a"`
	}{Z: 2, A: 1}
	parent, err := json.Marshal(prior.Report.CompletenessReceipt)
	if err != nil {
		t.Fatal(err)
	}
	e := NewExecutor()
	defer e.Close()
	r, err := e.Execute(context.Background(), "fixture.gooo", source, doc, prior, parent, doc.TestCases, nativeTool())
	if err != nil || !bytes.Equal(parent, r.ParentReceipt) {
		t.Fatal("live typed parent differs", err)
	}
	verifyReceipt(t, r)
	wrong := bytes.Replace(parent, []byte(`"z":2`), []byte(`"z":3`), 1)
	if _, err := e.Execute(context.Background(), "fixture.gooo", source, doc, prior, wrong, doc.TestCases, nativeTool()); err == nil {
		t.Fatal("changed parent accepted")
	}
}

func TestExecutorReusesOnlyArtifactAndExecutesCurrentCases(t *testing.T) {
	source, doc, prior, parent := fixture(t)
	e := NewExecutor()
	defer e.Close()
	first, err := e.Execute(context.Background(), "fixture.gooo", source, doc, prior, parent, doc.TestCases, nativeTool())
	if err != nil {
		t.Fatal(err)
	}
	if first.Observation.Artifact.Reused || !first.Observation.Build.Completed {
		t.Fatal("first request did not build")
	}
	root := e.artifact.root
	// Caller mutation of a returned process record cannot rewrite owned history.
	*first.Observation.Artifact.SourceBuild.ExitCode = 99
	cases := []pathplan.TestCase{{Input: 9, Expected: 50}, {Input: -9, Expected: 999}}
	second, err := e.Execute(context.Background(), "fixture.gooo", source, doc, prior, parent, cases, nativeTool())
	if err != nil {
		t.Fatal(err)
	}
	verifyReceipt(t, second)
	o := second.Observation
	if !o.Artifact.Reused || o.Build.Started || *o.Artifact.SourceBuild.ExitCode != 0 || len(o.Runs) != 2 {
		t.Fatal("reused record invents a build or aliases prior history")
	}
	if o.Cases[0].Actual != 50 || o.Cases[1].Actual != -46 || o.Cases[1].Passed || o.RuntimeSuiteSHA256 == first.Observation.RuntimeSuiteSHA256 {
		t.Fatalf("old outcomes used: %+v", o.Cases)
	}
	if d := dimension(t, second, "runtime_finite_accuracy"); d.Numerator != 1 || d.Denominator != 2 {
		t.Fatal(d)
	}
	if d := dimension(t, second, "runtime_child_resources"); d.Denominator != 2 {
		t.Fatal("previous build resources counted as current", d)
	}
	if dimension(t, second, "runtime_build").Status != "PASS" || second.CompletenessReceipt.ProfileID != RuntimeProfileV3 {
		t.Fatal("missing source-bound actual build reference")
	}
	if !bytes.Equal(second.ParentReceipt, parent) {
		t.Fatal("parent changed")
	}
	if err := e.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(root); !os.IsNotExist(err) {
		t.Fatal("workspace retained after Close", err)
	}
	if _, err := e.Execute(context.Background(), "fixture.gooo", source, doc, prior, parent, cases, nativeTool()); !errors.Is(err, ErrExecutorClosed) {
		t.Fatal("closed executor accepted work", err)
	}
	if err := e.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestExecutorWarmCacheDoesNotBypassSourceOrParentReplay(t *testing.T) {
	source, doc, prior, parent := fixture(t)
	e := NewExecutor()
	defer e.Close()
	if _, err := e.Execute(context.Background(), "fixture.gooo", source, doc, prior, parent, doc.TestCases, nativeTool()); err != nil {
		t.Fatal(err)
	}
	for _, change := range []string{"source", "plan", "generation", "parent"} {
		t.Run(change, func(t *testing.T) {
			s, d, p, receipt := fixture(t)
			switch change {
			case "source":
				s = append(s, '\n')
			case "plan":
				d.Plan.Decisions[0].Intent += "changed"
			case "generation":
				p.Source += "\n"
			case "parent":
				receipt = []byte(`{}`)
			}
			r, err := e.Execute(context.Background(), "fixture.gooo", s, d, p, receipt, d.TestCases, nativeTool())
			if err == nil || r.Observation.Toolchain.Started || r.Observation.Build.Started || len(r.Observation.Runs) != 0 {
				t.Fatal("warm cache bypassed replay", err)
			}
		})
	}
	for _, change := range []string{"tamper", "remove"} {
		if change == "tamper" {
			if err := os.WriteFile(e.artifact.executable, []byte("changed bytes"), 0700); err != nil {
				t.Fatal(err)
			}
		} else if err := os.Remove(e.artifact.executable); err != nil {
			t.Fatal(err)
		}
		r, err := e.Execute(context.Background(), "fixture.gooo", source, doc, prior, parent, doc.TestCases, nativeTool())
		if err != nil || r.Observation.Artifact.Reused || !r.Observation.Build.Completed || r.Observation.Artifact.MissReason != "executable_changed_or_missing" {
			t.Fatal("changed executable not rebuilt", err)
		}
	}
}

func TestExecutorQueueCancellationAndConcurrentClose(t *testing.T) {
	source, doc, prior, parent := fixture(t)
	e := NewExecutor()
	<-e.gate // Simulate an operation already holding the one workspace.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	_, err := e.Execute(ctx, "fixture.gooo", source, doc, prior, parent, doc.TestCases, nativeTool())
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal("waiting request ignored cancellation", err)
	}
	done := make(chan error, 2)
	go func() { done <- e.Close() }()
	go func() { done <- e.Close() }()
	e.gate <- struct{}{}
	for range 2 {
		select {
		case err := <-done:
			if err != nil {
				t.Fatal(err)
			}
		case <-time.After(2 * time.Second):
			t.Fatal("Close deadlocked")
		}
	}
}

func TestExecutorChangedProjectionAndConcurrentInputs(t *testing.T) {
	source, doc, prior, parent := fixture(t)
	e := NewExecutor()
	defer e.Close()
	if _, err := e.Execute(context.Background(), "fixture.gooo", source, doc, prior, parent, doc.TestCases, nativeTool()); err != nil {
		t.Fatal(err)
	}
	oldRoot := e.artifact.root
	source = bytes.Replace(source, []byte("package sample"), []byte("package renamed"), 1)
	var err error
	prior, err = bodycodegen.GenerateWithTypedPaths(context.Background(), "fixture.gooo", source, "Combined", doc, "")
	if err != nil {
		t.Fatal(err)
	}
	parent, err = json.Marshal(prior.Report.CompletenessReceipt)
	if err != nil {
		t.Fatal(err)
	}
	r, err := e.Execute(context.Background(), "fixture.gooo", source, doc, prior, parent, doc.TestCases, nativeTool())
	if err != nil || r.Observation.Artifact.Reused || r.Observation.Artifact.MissReason != "key_changed" {
		t.Fatal("changed projection did not replace workspace", err)
	}
	if _, err := os.Stat(oldRoot); !os.IsNotExist(err) {
		t.Fatal("old workspace retained", err)
	}
	done := make(chan Result, 2)
	for _, input := range []int64{2, 3} {
		go func() {
			cases := []pathplan.TestCase{{Input: input, Expected: 5*input + 5}}
			r, err := e.Execute(context.Background(), "fixture.gooo", source, doc, prior, parent, cases, nativeTool())
			if err != nil {
				r.Observation.Failure = err.Error()
			}
			done <- r
		}()
	}
	for range 2 {
		r := <-done
		if r.Observation.Failure != "" || len(r.Observation.Cases) != 1 || !r.Observation.Cases[0].Passed || !r.Observation.Artifact.Reused {
			t.Fatal("concurrent input leaked", r.Observation.Failure)
		}
	}
}

func TestExecutorCanceledRunDropsArtifact(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX Go-tool fixture")
	}
	source, doc, prior, parent := fixture(t)
	tool := filepath.Join(t.TempDir(), "go-stub")
	marker := filepath.Join(t.TempDir(), "running")
	stub := "#!/bin/sh\nif [ \"$1\" = version ]; then echo 'go version go1.27.1 " + runtime.GOOS + "/" + runtime.GOARCH + "'; exit 0; fi\n" +
		"while [ \"$1\" != -o ]; do shift; done\nshift\nprintf '#!/bin/sh\\ntouch \"" + marker + "\"\\nsleep 30 &\\nwait\\n' > \"$1\"\nchmod 700 \"$1\"\n"
	if err := os.WriteFile(tool, []byte(stub), 0700); err != nil {
		t.Fatal(err)
	}
	e := NewExecutor()
	defer e.Close()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	type outcome struct {
		result Result
		err    error
	}
	done := make(chan outcome, 1)
	go func() {
		r, err := e.Execute(ctx, "fixture.gooo", source, doc, prior, parent, doc.TestCases, tool)
		done <- outcome{r, err}
	}()
	deadline := time.Now().Add(5 * time.Second)
	for {
		if _, err := os.Stat(marker); err == nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("runtime did not start")
		}
		time.Sleep(5 * time.Millisecond)
	}
	cancel()
	var observed outcome
	select {
	case observed = <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("runtime cancellation deadlocked")
	}
	r, err := observed.result, observed.err
	if !errors.Is(err, context.Canceled) || e.artifact != nil {
		t.Fatal("canceled execution retained work", err)
	}
	if len(r.Observation.Runs) != 1 || !r.Observation.Runs[0].Canceled || len(r.Observation.Cases) != 0 {
		t.Fatal("canceled runtime invented outputs", r.Observation)
	}
	verifyReceipt(t, r)
}

func TestExecutorGoToolBytesInvalidateAndCloseCancelsBuild(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX Go-tool fixture")
	}
	source, doc, prior, parent := fixture(t)
	tool := filepath.Join(t.TempDir(), "go-wrapper")
	quoted := "'" + strings.ReplaceAll(nativeTool(), "'", "'\"'\"'") + "'"
	stub := "#!/bin/sh\nexec " + quoted + " \"$@\"\n"
	if err := os.WriteFile(tool, []byte(stub), 0700); err != nil {
		t.Fatal(err)
	}
	e := NewExecutor()
	defer e.Close()
	first, err := e.Execute(context.Background(), "fixture.gooo", source, doc, prior, parent, doc.TestCases, tool)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(tool, []byte(stub+"# new tool bytes\n"), 0700); err != nil {
		t.Fatal(err)
	}
	second, err := e.Execute(context.Background(), "fixture.gooo", source, doc, prior, parent, doc.TestCases, tool)
	if err != nil || second.Observation.Artifact.Reused || second.Observation.Artifact.KeySHA256 == first.Observation.Artifact.KeySHA256 {
		t.Fatal("tool change did not invalidate", err)
	}
	marker := filepath.Join(t.TempDir(), "building")
	stub = "#!/bin/sh\nif [ \"$1\" = version ]; then echo 'go version go1.27.1 " + runtime.GOOS + "/" + runtime.GOARCH + "'; exit 0; fi\ntouch '" + marker + "'\nsleep 30 &\nwait\n"
	if err := os.WriteFile(tool, []byte(stub), 0700); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() {
		_, err := e.Execute(context.Background(), "fixture.gooo", source, doc, prior, parent, doc.TestCases, tool)
		done <- err
	}()
	deadline := time.Now().Add(2 * time.Second)
	for {
		if _, err := os.Stat(marker); err == nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("build did not start")
		}
		time.Sleep(5 * time.Millisecond)
	}
	if err := e.Close(); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("canceled build succeeded")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("build child did not terminate")
	}
	if e.artifact != nil {
		t.Fatal("canceled workspace retained")
	}
}
