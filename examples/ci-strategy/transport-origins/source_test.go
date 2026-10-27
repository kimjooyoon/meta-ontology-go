package transportorigins

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodylibrary"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/workspaceexecution"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/workspaceinput"
)

type exampleSources string

func (directory exampleSources) ReadFile(name string) ([]byte, error) {
	return os.ReadFile(filepath.Join(string(directory), name))
}

func TestGeneratedReportViewMatchesItsSevenSourcesAndOriginalCompiler(t *testing.T) {
	const manifest = "transport-origins.workspace.json"
	sources := exampleSources("../..")
	raw, err := sources.ReadFile(manifest)
	if err != nil {
		t.Fatal(err)
	}
	definition, err := workspaceinput.Decode(raw)
	if err != nil {
		t.Fatal(err)
	}
	inputs, err := workspaceinput.Load(sources, manifest, definition)
	if err != nil {
		t.Fatal(err)
	}
	program, err := workspaceexecution.Prepare(inputs)
	if err != nil {
		t.Fatal(err)
	}
	generated, err := bodylibrary.GenerateWithStageExports(context.Background(), "workspace.gooo", []byte(program.Source),
		program.Entry.LoweredName, "transportorigins", []bodylibrary.StageExport{
			{Name: "ReportByteLimit", ActivityID: "gooo://transport-origins/report-byte-limit"},
		})
	if err != nil {
		t.Fatal(err)
	}
	stored, err := os.ReadFile("semantic.gooo.go")
	core, _, hasOrigin := strings.Cut(string(stored), "\n//gooo:generated:start id=\"urn:gooo:library:generation-origin:")
	if err != nil || !hasOrigin || !strings.HasSuffix(strings.TrimSpace(core), strings.TrimSpace(generated)) {
		t.Fatal("regenerate the report view from its authored sources", err)
	}
	var origin struct {
		Context bodylibrary.GenerationContext `json:"context"`
	}
	if err := json.Unmarshal([]byte(GoooGenerationOrigin), &origin); err != nil {
		t.Fatal(err)
	}
	if origin.Context.CompilerSourceSHA != "ba353f5af2e94389cae3af05470ae09e83109141" ||
		origin.Context.CompilerSourceStatus != "CLEAN_VCS" || len(origin.Context.Sources) != 7 ||
		origin.Context.WorkspaceManifestSHA != fmt.Sprintf("sha256:%x", sha256.Sum256(raw)) ||
		origin.Context.WorkspaceImageSHA != program.Workspace.Digest || strings.Count(generated, "type R_") != 20 {
		t.Fatal("original producer or source scope changed", origin.Context)
	}
	for _, pkg := range program.Workspace.Packages {
		for _, source := range pkg.Sources {
			if !strings.Contains(GoooGenerationOrigin, source.SourceDigest) {
				t.Fatal("lost source digest", source.Filename, source.SourceDigest)
			}
		}
	}
}
