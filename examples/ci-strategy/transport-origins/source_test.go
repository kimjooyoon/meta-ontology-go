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

	resolution "github.com/kimjooyoon/meta-ontology-go/examples/ci-strategy/transport-resolution"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodylibrary"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/workspaceexecution"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/workspaceinput"
)

type exampleSources string

func (directory exampleSources) ReadFile(name string) ([]byte, error) {
	return os.ReadFile(filepath.Join(string(directory), name))
}

func TestGeneratedReportViewMatchesItsSourcePartitionsAndOriginalCompiler(t *testing.T) {
	assertOriginalGeneration(t, generationFixture{
		"transport-origins.workspace.json", "transportorigins", "semantic.gooo.go", GoooGenerationOrigin, 9, 20,
		[]bodylibrary.StageExport{{Name: "ReportByteLimit", ActivityID: "gooo://transport-origins/report-byte-limit"}},
	})
}

func TestGeneratedReassessmentMatchesItsSourcesAndOriginalCompiler(t *testing.T) {
	assertOriginalGeneration(t, generationFixture{
		"transport-resolution.workspace.json", "transportresolution", "../transport-resolution/semantic.gooo.go",
		resolution.GoooGenerationOrigin, 10, 22, nil,
	})
}

type generationFixture struct {
	manifest, goPackage, stored, origin string
	sources, records                    int
	exports                             []bodylibrary.StageExport
}

func assertOriginalGeneration(t *testing.T, fixture generationFixture) {
	t.Helper()
	manifest := fixture.manifest
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
		program.Entry.LoweredName, fixture.goPackage, fixture.exports)
	if err != nil {
		t.Fatal(err)
	}
	stored, err := os.ReadFile(fixture.stored)
	core, _, hasOrigin := strings.Cut(string(stored), "\n//gooo:generated:start id=\"urn:gooo:library:generation-origin:")
	if err != nil || !hasOrigin || !strings.HasSuffix(strings.TrimSpace(core), strings.TrimSpace(generated)) {
		t.Fatal("regenerate the report view from its authored sources", err)
	}
	var origin struct {
		Context bodylibrary.GenerationContext `json:"context"`
	}
	if err := json.Unmarshal([]byte(fixture.origin), &origin); err != nil {
		t.Fatal(err)
	}
	if origin.Context.CompilerSourceSHA != "ba353f5af2e94389cae3af05470ae09e83109141" ||
		origin.Context.CompilerSourceStatus != "CLEAN_VCS" || len(origin.Context.Sources) != fixture.sources ||
		origin.Context.WorkspaceManifestSHA != fmt.Sprintf("sha256:%x", sha256.Sum256(raw)) ||
		origin.Context.WorkspaceImageSHA != program.Workspace.Digest || strings.Count(generated, "type R_") != fixture.records {
		t.Fatal("original producer or source scope changed", origin.Context)
	}
	for _, pkg := range program.Workspace.Packages {
		for _, source := range pkg.Sources {
			if !strings.Contains(fixture.origin, source.SourceDigest) {
				t.Fatal("lost source digest", source.Filename, source.SourceDigest)
			}
		}
	}
}
