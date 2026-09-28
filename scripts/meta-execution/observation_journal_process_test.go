package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
)

func TestJournalRetainsPlannerBoundEntryAfterProcessKill(t *testing.T) {
	manifestPath := filepath.Join(t.TempDir(), "manifest.json")
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestJournalInterruptionChild$")
	command.Env = append(os.Environ(), "GOOO_JOURNAL_CHILD_OUTPUT="+manifestPath)
	output, err := command.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := command.Start(); err != nil {
		t.Fatal(err)
	}
	defer command.Process.Kill()
	line, err := bufio.NewReader(output).ReadString('\n')
	if err != nil || line != "boundary-written\n" {
		_ = command.Process.Kill()
		_ = command.Wait()
		t.Fatalf("child did not reach declared boundary: %q %v", line, err)
	}
	if err := command.Process.Kill(); err != nil {
		t.Fatal(err)
	}
	if err := command.Wait(); err == nil {
		t.Fatal("killed process was reported successful")
	}
	data, err := os.ReadFile(manifestPath + ".trace.ndjson")
	if err != nil {
		t.Fatal(err)
	}
	var event metaExecutionTraceEvent
	if err := json.Unmarshal(data, &event); err != nil {
		t.Fatal(err)
	}
	plan, action, _ := collapsePlannerFixture(t, "fixture.go:3:CollapseFixture", strings.Repeat("1", 40))
	manifest := generation.BuildExecutionManifest(plan)
	if event.Boundary != "ACTION_ENTERED" || event.Activity != action.Activity ||
		event.ActionIndicatorID != action.IndicatorID || event.PlanDigest != plan.PlanDigest ||
		event.ManifestDigest != manifest.ManifestDigest || event.InvocationID == "" ||
		event.InputContractSourceDigest != action.InputContractSourceDigest ||
		event.Permission != "UNOBSERVED" || !event.DiagnosticOnly || event.ReturnErrorObserved != nil {
		t.Fatalf("interruption changed planner identity or invented return: %+v", event)
	}
}

func TestJournalInterruptionChild(t *testing.T) {
	output := os.Getenv("GOOO_JOURNAL_CHILD_OUTPUT")
	if output == "" {
		t.Skip("helper runs only in the bounded interruption subprocess")
	}
	plan, action, _ := collapsePlannerFixture(t, "fixture.go:3:CollapseFixture", strings.Repeat("1", 40))
	manifest := generation.BuildExecutionManifest(plan)
	file, state, err := openObservationJournal(output)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	trace := newMetaExecutionTrace(plan, manifest, action, 1, state)
	trace.emitActionEntered()
	fmt.Println("boundary-written")
	time.Sleep(time.Minute)
	t.Fatal("parent did not interrupt the subprocess")
}
