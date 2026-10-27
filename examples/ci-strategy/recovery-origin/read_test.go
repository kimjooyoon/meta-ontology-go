package recoveryorigin

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestReaderUsesJobCompletionInsteadOfRunMetadataUpdate(t *testing.T) {
	root, scope := historyFixture(t)
	path := filepath.Join(root, "readback-run.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var run map[string]any
	if err := json.Unmarshal(raw, &run); err != nil {
		t.Fatal(err)
	}
	run["updated_at"] = time.Unix(scope.F_observed_at+3600, 0).UTC().Format(time.RFC3339)
	raw, err = json.Marshal(run)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
	p, _ := Read(root, "publication", scope.F_publish_job)
	r, _ := Read(root, "readback", scope.F_readback_job)
	if got := Decide(scope, p, r); got.F_state != "PROPOSED_RECOVERY" {
		t.Fatal("a later metadata update changed the original job completion", got)
	}
	if r.F_completed_at != 1792615620 { // 2026-10-21T20:47:00Z, original selected job.
		t.Fatal("timestamp is not the original selected job completion", r.F_completed_at)
	}
}

func TestReaderRejectsChangedOrMissingOriginals(t *testing.T) {
	for _, change := range []string{"missing archive", "changed archive", "wrong run", "wrong attempt", "false completion", "oversized record", "future completion", "missing completion"} {
		t.Run(change, func(t *testing.T) {
			root, scope := historyFixture(t)
			changeRetainedInput(t, root, change)
			p, _ := Read(root, "publication", scope.F_publish_job)
			r, _ := Read(root, "readback", scope.F_readback_job)
			got := Decide(scope, p, r)
			if got.F_state != "UNRESOLVED_ORIGIN" || len(got.F_completed_effect_ids) != 0 || got.F_execution_authorized {
				t.Fatalf("changed input was accepted: %+v", got)
			}
		})
	}
}

func changeRetainedInput(t *testing.T, root, change string) {
	t.Helper()
	name := "publication-run.json"
	if change == "missing archive" {
		if err := os.Remove(filepath.Join(root, "publication.zip")); err != nil {
			t.Fatal(err)
		}
		return
	}
	if change == "changed archive" {
		name = "publication.zip"
	}
	if change == "false completion" || change == "future completion" || change == "missing completion" {
		name = "publication-jobs.json"
	}
	path := filepath.Join(root, name)
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if change == "changed archive" {
		raw = append(raw, 'x') // Valid ZIP with changed bytes; the artifact digest must still match.
	} else if change == "oversized record" {
		raw = make([]byte, (1<<20)+1)
	} else {
		var record map[string]any
		if err := json.Unmarshal(raw, &record); err != nil {
			t.Fatal(err)
		}
		switch change {
		case "wrong run":
			record["id"] = float64(39672180711)
		case "wrong attempt":
			record["run_attempt"] = float64(2)
		case "false completion":
			job := record["jobs"].([]any)[1].(map[string]any)
			job["steps"].([]any)[4].(map[string]any)["conclusion"] = "failure"
		case "future completion":
			record["jobs"].([]any)[1].(map[string]any)["completed_at"] = "2026-10-21T22:00:00Z"
		case "missing completion":
			delete(record["jobs"].([]any)[1].(map[string]any), "completed_at")
		}
		raw, err = json.Marshal(record)
		if err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
}
