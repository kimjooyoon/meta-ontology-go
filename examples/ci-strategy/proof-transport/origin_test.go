package prooftransport

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/workspaceexecution"
)

func checkTransportGenerationOrigin(t *testing.T, manifest []byte, program workspaceexecution.Program, stored string) {
	t.Helper()
	var origin struct {
		Schema    string `json:"schema"`
		Entry     string `json:"entry_activity_id"`
		Lowered   string `json:"lowered_source_sha256"`
		TypedPlan string `json:"typed_plan_sha256"`
		Context   struct {
			Compiler string `json:"compiler_source_sha"`
			Status   string `json:"compiler_source_status"`
			Manifest string `json:"workspace_manifest_sha256"`
			Image    string `json:"workspace_image_sha256"`
			Sources  []struct {
				Package string `json:"package_path"`
				File    string `json:"filename"`
				SHA     string `json:"source_json_string_sha256"`
			} `json:"sources"`
		} `json:"context"`
		Stages []struct {
			Name string `json:"name"`
			ID   string `json:"activity_id"`
		} `json:"stage_exports"`
	}
	if err := json.Unmarshal([]byte(GoooGenerationOrigin), &origin); err != nil {
		t.Fatal(err)
	}
	if origin.Schema != "gooo/library-generation-origin/v1" || origin.Entry != "gooo://ci-strategy/transport-next" || origin.Lowered != fmt.Sprintf("sha256:%x", sha256.Sum256([]byte(program.Source))) || !strings.Contains(stored, "// Typed plan: "+origin.TypedPlan+"\n") {
		t.Fatal("compiled source or entry identity lost", origin)
	}
	if origin.Context.Manifest != fmt.Sprintf("sha256:%x", sha256.Sum256(manifest)) || origin.Context.Image != program.Workspace.Digest || len(origin.Context.Sources) != 4 {
		t.Fatal("workspace binding lost", origin.Context)
	}
	if origin.Context.Compiler != "d5cd2eeec2b47cd8ca51408ecf7665079c486eec" || origin.Context.Status != "CLEAN_VCS" {
		t.Fatal("original compiler producer relabeled", origin.Context)
	}
	if len(origin.Stages) != 1 || origin.Stages[0].Name != "Plan" || origin.Stages[0].ID != "gooo://ci-strategy/transport-plan" {
		t.Fatal("stage identity lost", origin.Stages)
	}
	expected := map[string]bool{"strategy.gooo": true, "recovery-origin/origin.gooo": true, "recovery-origin/plan.gooo": true, "proof-transport/transport.gooo": true}
	for _, source := range origin.Context.Sources {
		if !expected[source.File] {
			t.Fatal("unexpected or repeated source", source.File)
		}
		delete(expected, source.File)
		raw, err := os.ReadFile(filepath.Join("..", source.File))
		if err != nil {
			t.Fatal(err)
		}
		encoded, err := json.Marshal(string(raw))
		if err != nil || source.Package != "examples/cistrategy" || source.SHA != fmt.Sprintf("sha256:%x", sha256.Sum256(encoded)) {
			t.Fatal("original source bytes changed", source, err)
		}
	}
}

func TestTransportReportRetainsCompiledOriginWithoutSourceReads(t *testing.T) {
	f := fixture(t, true)
	t.Chdir(t.TempDir())
	got := run(context.Background(), "transport-original-build-proof", f.material, "", false, f)
	if got.GenerationOrigin != GoooGenerationOrigin || got.Decision.F_state != "ALREADY_PRESENT" || got.Write.Attempted {
		t.Fatal("report lost compiled origin or source policy", got)
	}
}
