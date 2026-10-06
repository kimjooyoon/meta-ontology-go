package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPublicGenerateProjectsOptionalScalarFields(t *testing.T) {
	root := t.TempDir()
	output := filepath.Join(root, "generated")
	source := filepath.Join("..", "..", "examples", "entity-fields-v4", "main.gooo.fixture")
	var stdout, stderr bytes.Buffer
	if code := runCheck([]string{"--semantic", source}, OSFileReader{}, EntityFieldsCLIParser{}, &stdout, &stderr); code != exitOK {
		t.Fatalf("gooo check --semantic = %d, stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
	stdout.Reset()
	stderr.Reset()
	if code := runGenerate([]string{source, "--out", output}, OSFileReader{}, EntityFieldsCLIParser{}, &stdout, &stderr); code != exitOK {
		t.Fatalf("gooo generate = %d, stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
	generated, err := os.ReadFile(filepath.Join(output, generatedFileName))
	if err != nil {
		t.Fatal(err)
	}
	compact := strings.NewReplacer(" ", "", "\t", "", "\n", "").Replace(string(generated))
	for _, want := range []string{"DisplayNamestring", "Nickname*string", "Active*bool", "Score*int64"} {
		if !strings.Contains(compact, want) {
			t.Fatalf("generated source is missing %q:\n%s", want, generated)
		}
	}
	if _, err := os.Stat(filepath.Join(output, "runtime-plan.json")); err != nil {
		t.Fatalf("V4 source binding did not produce its runtime plan: %v", err)
	}
	if strings.Contains(stderr.String(), "unsupported shape") {
		t.Fatalf("public generator rejected the V4 profile: %s", stderr.String())
	}
	stdout.Reset()
	stderr.Reset()
	if code := runDiscover([]string{"--query", "What profile fields can this Gooo source represent?", source}, OSFileReader{}, &stdout, &stderr); code != exitOK {
		t.Fatalf("gooo discover = %d, stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
}
