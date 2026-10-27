package cyclecontext_test

import (
	"context"
	"crypto/sha256"
	"embed"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodylibrary"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/workspaceexecution"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/workspaceinput"
)

//go:embed policy.gooo gooo.workspace.json
var sourceFiles embed.FS

func TestGeneratedCyclePolicyMatchesAuthoredSource(t *testing.T) {
	raw, err := sourceFiles.ReadFile("gooo.workspace.json")
	if err != nil {
		t.Fatal(err)
	}
	definition, err := workspaceinput.Decode(raw)
	if err != nil {
		t.Fatal(err)
	}
	inputs, err := workspaceinput.Load(sourceFiles, "gooo.workspace.json", definition)
	if err != nil {
		t.Fatal(err)
	}
	program, err := workspaceexecution.Prepare(inputs)
	if err != nil {
		t.Fatal(err)
	}
	generated, err := bodylibrary.Generate(context.Background(), "workspace.gooo", []byte(program.Source), program.Entry.LoweredName, "cyclecontext")
	if err != nil {
		t.Fatal(err)
	}
	stored, err := os.ReadFile("semantic.gooo.go")
	if err != nil || !strings.HasSuffix(string(stored), generated) {
		t.Fatal("cycle policy source and generated library differ; regenerate through package generate", err)
	}
	for _, binding := range []string{
		fmt.Sprintf("// Workspace manifest: sha256:%x\n", sha256.Sum256(raw)),
		"// Workspace image: " + program.Workspace.Digest + "\n",
		fmt.Sprintf("source_json_string_sha256=%q", program.Workspace.Packages[0].Sources[0].SourceDigest),
	} {
		if !strings.Contains(string(stored), binding) {
			t.Fatal("generated cycle policy lost its original input binding", binding)
		}
	}
}
