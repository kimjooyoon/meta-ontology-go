package recoveryorigin

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

const historySHA256 = "a5142fa585b183dfc39b94ef497a2706aeebfb16c1df41da6320fb6ff7beacdf"

func historyFixture(t *testing.T) (string, R_RecoveryScope) {
	t.Helper()
	archive, err := os.ReadFile("testdata/history.zip")
	if err != nil || fmt.Sprintf("%x", sha256.Sum256(archive)) != historySHA256 {
		t.Fatal("retained public history changed", err)
	}
	root := t.TempDir()
	r, err := zip.OpenReader("testdata/history.zip")
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	for _, f := range r.File {
		if filepath.Base(f.Name) != f.Name {
			t.Fatal("non-flat fixture", f.Name)
		}
		in, err := f.Open()
		if err != nil {
			t.Fatal(err)
		}
		raw, err := io.ReadAll(in)
		in.Close()
		if err != nil || os.WriteFile(filepath.Join(root, f.Name), raw, 0600) != nil {
			t.Fatal("extract fixture", f.Name, err)
		}
	}
	var scope R_RecoveryScope
	raw, err := os.ReadFile("testdata/scope.json")
	if err != nil || json.Unmarshal(raw, &scope) != nil {
		t.Fatal("declared historical scope", err)
	}
	return root, scope
}

func TestRetainedHistory(t *testing.T) {
	root, scope := historyFixture(t)
	published, publicationDigests := Read(root, "publication", scope.F_publish_job)
	readback, readbackDigests := Read(root, "readback", scope.F_readback_job)
	if published.F_read_error != "" || readback.F_read_error != "" {
		t.Fatal(published.F_read_error, readback.F_read_error)
	}
	got := Evaluate(scope, published, readback)
	if got.F_state != "PROPOSED_RECOVERY" || got.F_binding_state != "BOUND_HISTORICAL_OBSERVATION" {
		t.Fatalf("actual retained inputs: %+v", got)
	}
	if got.F_revision_id != scope.F_revision_id || got.F_goal_id != scope.F_goal_id ||
		got.F_prior_plan_id != scope.F_prior_plan_id || got.F_recovery_plan_id != scope.F_recovery_plan_id ||
		got.F_failed_activity_id != scope.F_failed_activity_id || got.F_execution_authorized {
		t.Fatal("identity or authority changed", got)
	}
	if got.F_failure_observation_id != "https://github.com/kimjooyoon/gooo-go-ports/actions/runs/39672180710/job/119472846036" ||
		got.F_readback_observation_id != "https://github.com/kimjooyoon/gooo-go-ports/actions/runs/39673212888/job/119476221244" {
		t.Fatal("original observations lost", got)
	}
	if !slices.Equal(got.F_completed_effect_ids, []string{"urn:gooo:effect:oci-publish:" + scope.F_subject}) ||
		!slices.Equal(got.F_recovery_effect_ids, []string{"urn:gooo:effect:oci-original-build-proof:" + scope.F_subject + ":" + scope.F_source}) {
		t.Fatal("operation identity lost", got)
	}
	if len(publicationDigests) != 5 || len(readbackDigests) != 5 {
		t.Fatal("raw observation digests missing")
	}
	t.Log("retained history cases=1/1; preserved goal/plan/activity/observation identities=7/7; execution_authorized=false")
}
