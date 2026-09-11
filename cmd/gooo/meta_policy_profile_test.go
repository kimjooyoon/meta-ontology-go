package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/policycompilation"
)

func TestMetaPolicyProfileRejectsIncompatibleFlagsBeforeOutput(t *testing.T) {
	projectRoot, policyPath := writeMetaPolicyProfileFixture(t)
	outputDir := filepath.Join(t.TempDir(), "public")
	args := []string{policyPath, "--profile", policycompilation.PublicProfileID,
		"--profile-package", "metapolicycompilation", "--profile-namespace", "metapolicycompilation",
		"--profile-project-root", projectRoot, "--previous-go", filepath.Join(projectRoot, "previous.go"),
		"--out", outputDir}
	if code := runGenerate(args, OSFileReader{}, SyntaxSourceParser{}, &bytes.Buffer{}, &bytes.Buffer{}); code != exitUsage {
		t.Fatalf("incompatible profile flags exit code = %d, want %d", code, exitUsage)
	}
	if _, err := os.Stat(outputDir); !os.IsNotExist(err) {
		t.Fatalf("incompatible profile flags created output: err=%v", err)
	}
}

func TestMetaPolicyProfileRejectsUnsupportedProfile(t *testing.T) {
	projectRoot, policyPath := writeMetaPolicyProfileFixture(t)
	outputDir := filepath.Join(t.TempDir(), "public")
	args := metaPolicyProfileArgs(policyPath, projectRoot, outputDir, "metapolicycompilation", "metapolicycompilation")
	args[2] = "unsupported-profile"
	if code := runGenerate(args, OSFileReader{}, SyntaxSourceParser{}, &bytes.Buffer{}, &bytes.Buffer{}); code != exitUsage {
		t.Fatalf("unsupported profile exit code = %d, want %d", code, exitUsage)
	}
	if _, err := os.Stat(outputDir); !os.IsNotExist(err) {
		t.Fatalf("unsupported profile created output: err=%v", err)
	}
}

func TestMetaPolicyProfileRejectsWrongIdentity(t *testing.T) {
	tests := []struct {
		name      string
		pkg       string
		namespace string
	}{
		{name: "package", pkg: "wrongpackage", namespace: "metapolicycompilation"},
		{name: "namespace", pkg: "metapolicycompilation", namespace: "wrongnamespace"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			projectRoot, policyPath := writeMetaPolicyProfileFixture(t)
			outputDir := filepath.Join(t.TempDir(), "public")
			var stderr bytes.Buffer
			code := runGenerate(metaPolicyProfileArgs(policyPath, projectRoot, outputDir, test.pkg, test.namespace), OSFileReader{}, SyntaxSourceParser{}, &bytes.Buffer{}, &stderr)
			if code != exitFailure {
				t.Fatalf("wrong identity exit code = %d, want %d; stderr=%q", code, exitFailure, stderr.String())
			}
			if _, err := os.Stat(outputDir); !os.IsNotExist(err) {
				t.Fatalf("wrong identity created output: err=%v", err)
			}
		})
	}
}

func TestMetaPolicyProfileRejectsBoundaryViolations(t *testing.T) {
	tests := []struct {
		name string
		setup func(t *testing.T, projectRoot, policyPath string) string
	}{
		{
			name: "output nested in project",
			setup: func(_ *testing.T, projectRoot, policyPath string) string {
				_ = policyPath
				return filepath.Join(projectRoot, "nested-output")
			},
		},
		{
			name: "output symlink",
			setup: func(t *testing.T, _, _ string) string {
				actual := t.TempDir()
				parent := t.TempDir()
				link := filepath.Join(parent, "public")
				if err := os.Symlink(actual, link); err != nil {
					t.Skipf("symlinks unavailable: %v", err)
				}
				return link
			},
		},
		{
			name: "output nonempty",
			setup: func(t *testing.T, _, _ string) string {
				output := t.TempDir()
				if err := os.WriteFile(filepath.Join(output, "existing"), []byte("occupied\n"), 0o640); err != nil {
					t.Fatal(err)
				}
				return output
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			projectRoot, policyPath := writeMetaPolicyProfileFixture(t)
			outputDir := test.setup(t, projectRoot, policyPath)
			var stderr bytes.Buffer
			code := runGenerate(metaPolicyProfileArgs(policyPath, projectRoot, outputDir, "metapolicycompilation", "metapolicycompilation"), OSFileReader{}, SyntaxSourceParser{}, &bytes.Buffer{}, &stderr)
			if code != exitFailure {
				t.Fatalf("boundary violation exit code = %d, want %d; stderr=%q", code, exitFailure, stderr.String())
			}
		})
	}
}

func TestMetaPolicyProfileRejectsInputOutsideProject(t *testing.T) {
	projectRoot, _ := writeMetaPolicyProfileFixture(t)
	sourceRoot := t.TempDir()
	_, policyPath := writeMetaPolicyProfileFixtureAt(t, sourceRoot)
	outputDir := filepath.Join(t.TempDir(), "public")
	var stderr bytes.Buffer
	code := runGenerate(metaPolicyProfileArgs(policyPath, projectRoot, outputDir, "metapolicycompilation", "metapolicycompilation"), OSFileReader{}, SyntaxSourceParser{}, &bytes.Buffer{}, &stderr)
	if code != exitFailure {
		t.Fatalf("outside input exit code = %d, want %d; stderr=%q", code, exitFailure, stderr.String())
	}
}

func TestMetaPolicyProfileWritesExactExternalFourFileBoundary(t *testing.T) {
	workingDirectory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	repositoryRoot, err := filepath.Abs(filepath.Join(workingDirectory, "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	projectRoot, policyPath := writeMetaPolicyProfileFixture(t)
	t.Chdir(repositoryRoot)
	outputDir := filepath.Join(t.TempDir(), "public")
	var stdout, stderr bytes.Buffer
	code := runGenerate(metaPolicyProfileArgs(policyPath, projectRoot, outputDir, "metapolicycompilation", "metapolicycompilation"), OSFileReader{}, SyntaxSourceParser{}, &stdout, &stderr)
	if code != exitOK || stderr.Len() != 0 {
		t.Fatalf("valid profile exit = %d, stdout=%q, stderr=%q", code, stdout.String(), stderr.String())
	}
	entries, err := os.ReadDir(outputDir)
	if err != nil {
		t.Fatal(err)
	}
	gotFiles := make([]string, 0, len(entries))
	for _, entry := range entries {
		gotFiles = append(gotFiles, entry.Name())
	}
	wantFiles := []string{"artifact.json", "generation-manifest.json", "judge.go", "policy.json"}
	if !reflect.DeepEqual(gotFiles, wantFiles) {
		t.Fatalf("profile files = %v, want %v", gotFiles, wantFiles)
	}
	var manifest policycompilation.PublicGenerationManifest
	manifestBytes, err := os.ReadFile(filepath.Join(outputDir, "generation-manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		t.Fatal(err)
	}
	if manifest.OutputRootClass != policycompilation.PublicGenerationOutputRootClass || manifest.ExecutionObserved || manifest.CurrentConformance != policycompilation.PublicGenerationConformanceUnknown || manifest.RepositoryWrites != 0 || manifest.MutationAuthority != 0 || manifest.PromotionAuthority != 0 {
		t.Fatalf("profile manifest state = %#v", manifest)
	}
	manifestFiles := []string{"policy.json", "artifact.json", "judge.go", "generation-manifest.json"}
	if !reflect.DeepEqual(manifest.GeneratedFiles, manifestFiles) {
		t.Fatalf("manifest files = %v, want %v", manifest.GeneratedFiles, manifestFiles)
	}
}

func metaPolicyProfileArgs(policyPath, projectRoot, outputDir, pkg, namespace string) []string {
	return []string{policyPath, "--profile", policycompilation.PublicProfileID,
		"--profile-package", pkg, "--profile-namespace", namespace,
		"--profile-project-root", projectRoot, "--out", outputDir}
}

func writeMetaPolicyProfileFixture(t *testing.T) (string, string) {
	return writeMetaPolicyProfileFixtureAt(t, t.TempDir())
}

func writeMetaPolicyProfileFixtureAt(t *testing.T, projectRoot string) (string, string) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", "examples", "meta-policy-compilation", "policy.gooo"))
	if err != nil {
		t.Fatal(err)
	}
	policyPath := filepath.Join(projectRoot, "policy.gooo")
	if err := os.WriteFile(policyPath, data, 0o640); err != nil {
		t.Fatal(err)
	}
	return projectRoot, policyPath
}
