package transportoutcome

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodylibrary"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/workspaceexecution"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/workspaceinput"
)

func TestGeneratedOutcomeMatchesSourceAndOriginalCompiler(t *testing.T) {
	manifest, err := os.ReadFile("gooo.workspace.json")
	if err != nil {
		t.Fatal(err)
	}
	source, err := os.ReadFile("outcome.gooo")
	if err != nil {
		t.Fatal(err)
	}
	definition, err := workspaceinput.Decode(manifest)
	if err != nil {
		t.Fatal(err)
	}
	inputs, err := workspaceinput.Load(fstest.MapFS{"outcome.gooo": &fstest.MapFile{Data: source}}, "gooo.workspace.json", definition)
	if err != nil {
		t.Fatal(err)
	}
	program, err := workspaceexecution.Prepare(inputs)
	if err != nil {
		t.Fatal(err)
	}
	generated, err := bodylibrary.Generate(context.Background(), "workspace.gooo", []byte(program.Source), program.Entry.LoweredName, "transportoutcome")
	if err != nil {
		t.Fatal(err)
	}
	stored, err := os.ReadFile("semantic.gooo.go")
	core, _, found := strings.Cut(string(stored), "\n//gooo:generated:start id=\"urn:gooo:library:generation-origin:")
	if err != nil || !found || !strings.HasSuffix(strings.TrimSpace(core), strings.TrimSpace(generated)) {
		t.Fatal("regenerate the outcome library", err)
	}
	encoded, _ := json.Marshal(string(source))
	for _, value := range []string{
		fmt.Sprintf("sha256:%x", sha256.Sum256(manifest)), program.Workspace.Digest,
		fmt.Sprintf("sha256:%x", sha256.Sum256(encoded)), "gooo://transport-outcome/assess",
		`"compiler_source_sha":"b58dbc25d2616c7f32f11818be201a096229c5fa"`, `"compiler_source_status":"CLEAN_VCS"`,
	} {
		if !strings.Contains(GoooGenerationOrigin, value) {
			t.Fatal("origin changed", value)
		}
	}
}
