package recoveryorigin

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodylibrary"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/workspaceexecution"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/workspaceinput"
)

type sourceDirectory string

func (directory sourceDirectory) ReadFile(name string) ([]byte, error) {
	return os.ReadFile(filepath.Join(string(directory), name))
}

func TestGeneratedRecoveryMatchesAuthoredSource(t *testing.T) {
	const manifest = "recovery-origin.workspace.json"
	sources := sourceDirectory("..")
	raw, err := os.ReadFile("../" + manifest)
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
	generated, err := bodylibrary.GenerateWithStageExports(context.Background(), "workspace.gooo", []byte(program.Source), program.Entry.LoweredName, "recoveryorigin", []bodylibrary.StageExport{
		{Name: "DeclaredScope", ActivityID: "gooo://ci-strategy/historical-scope"},
		{Name: "Decide", ActivityID: "gooo://ci-strategy/bind-recovery"},
	})
	if err != nil {
		t.Fatal(err)
	}
	stored, err := os.ReadFile("semantic.gooo.go")
	if err != nil || !strings.HasSuffix(string(stored), generated) {
		t.Fatal("regenerate the source-owned library", err)
	}
	for _, binding := range []string{
		fmt.Sprintf("// Workspace manifest: sha256:%x\n", sha256.Sum256(raw)),
		"// Workspace image: " + program.Workspace.Digest + "\n",
	} {
		if !strings.Contains(string(stored), binding) {
			t.Fatal("lost source binding", binding)
		}
	}
	for _, source := range program.Workspace.Packages[0].Sources {
		if !strings.Contains(string(stored), fmt.Sprintf("source_json_string_sha256=%q", source.SourceDigest)) {
			t.Fatal("lost source digest", source.SourceDigest)
		}
	}
}
