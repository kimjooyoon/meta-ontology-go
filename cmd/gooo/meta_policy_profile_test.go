package main

import (
	"bytes"
	"encoding/json"
	"fmt"
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
		name  string
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

type publicProfileDeliveryFixture struct {
	args       []string
	source     []byte
	policyPath string
	outputDir  string
	profile    string
	reportName string
	files      []string
}

type publicProfileDeliveryMode struct {
	name     string
	jsonMode bool
	rejected bool
}

type publicProfileDeliveryRejectedWriter struct {
	calls int
}

func (writer *publicProfileDeliveryRejectedWriter) Write([]byte) (int, error) {
	writer.calls++
	return 0, os.ErrClosed
}

func TestPublicMetaPolicyProfilesOutputDelivery(t *testing.T) {
	profiles := []struct {
		name     string
		revision bool
	}{
		{"generation", false},
		{"revision", true},
	}
	modes := []publicProfileDeliveryMode{
		{"human-denied", false, true},
		{"json-denied", true, true},
		{"human-success", false, false},
		{"json-success", true, false},
	}
	for _, profile := range profiles {
		t.Run(profile.name, func(t *testing.T) {
			for _, mode := range modes {
				t.Run(mode.name, func(t *testing.T) {
					runPublicProfileDeliveryCase(t, profile.revision, mode)
				})
			}
		})
	}
}

func newPublicProfileDeliveryFixture(t *testing.T, revision bool) publicProfileDeliveryFixture {
	t.Helper()
	projectRoot, policyPath := writeMetaPolicyProfileFixture(t)
	source, err := os.ReadFile(policyPath)
	if err != nil {
		t.Fatal(err)
	}
	outputDir := filepath.Join(t.TempDir(), "public")
	fixture := publicProfileDeliveryFixture{
		args:   metaPolicyProfileArgs(policyPath, projectRoot, outputDir, "metapolicycompilation", "metapolicycompilation"),
		source: source, policyPath: policyPath, outputDir: outputDir,
		profile: policycompilation.PublicProfileID, reportName: "generation-manifest.json",
		files: []string{"artifact.json", "generation-manifest.json", "judge.go", "policy.json"},
	}
	if revision {
		fixture.profile = policycompilation.PublicPolicyRevisionProfileID
		fixture.reportName = "proposal.json"
		fixture.files = []string{"candidate.gooo", "proposal.json"}
		fixture.args[2] = fixture.profile
		fixture.args = append(fixture.args,
			"--profile-source-digest", policycompilation.DigestBytes(source),
			"--profile-condition", "SEMANTIC_EQUIVALENCE",
			"--profile-from-decision", "PASS",
			"--profile-to-decision", "FAIL_CLOSED")
	}
	return fixture
}

func runPublicProfileDeliveryCase(t *testing.T, revision bool, mode publicProfileDeliveryMode) {
	t.Helper()
	fixture := newPublicProfileDeliveryFixture(t, revision)
	if mode.jsonMode {
		fixture.args = append(fixture.args, "--json")
	}
	var stdout, stderr bytes.Buffer
	var writer interfaceWriter = &stdout
	rejected := &publicProfileDeliveryRejectedWriter{}
	wantCode := exitOK
	if mode.rejected {
		writer = rejected
		wantCode = exitFailure
	}
	code := runGenerate(fixture.args, OSFileReader{}, SyntaxSourceParser{}, writer, &stderr)
	report := requirePublicProfileDeliveryArtifacts(t, fixture, revision)
	t.Logf("public profile delivery: profile=%s mode=%s rejected=%t writer_calls=%d exit=%d expected_exit=%d artifact_files=%d source=%s report=%s",
		fixture.profile, mode.name, mode.rejected, rejected.calls, code, wantCode,
		len(fixture.files), policycompilation.DigestBytes(fixture.source), policycompilation.DigestBytes(report))
	if mode.rejected && rejected.calls != 1 {
		t.Fatalf("rejected output was not attempted exactly once: calls=%d stderr=%s", rejected.calls, stderr.String())
	}
	if code != wantCode {
		t.Fatalf("profile delivery exit=%d want=%d stderr=%s", code, wantCode, stderr.String())
	}
	if !mode.rejected {
		requirePublicProfileDeliveryBytes(t, fixture, revision, mode.jsonMode, stdout.Bytes(), report)
	}
}

func requirePublicProfileDeliveryArtifacts(t *testing.T, fixture publicProfileDeliveryFixture, revision bool) []byte {
	t.Helper()
	entries, err := os.ReadDir(fixture.outputDir)
	if err != nil {
		t.Fatalf("generated files missing after output delivery: %v", err)
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		names = append(names, entry.Name())
		if !entry.Type().IsRegular() {
			t.Fatalf("generated artifact is not a regular file: %s", entry.Name())
		}
	}
	if !reflect.DeepEqual(names, fixture.files) {
		t.Fatalf("delivery changed generated files: got=%v want=%v", names, fixture.files)
	}
	after, err := os.ReadFile(fixture.policyPath)
	if err != nil || !bytes.Equal(after, fixture.source) {
		t.Fatalf("delivery changed the input source: %v", err)
	}
	report, err := os.ReadFile(filepath.Join(fixture.outputDir, fixture.reportName))
	if err != nil {
		t.Fatal(err)
	}
	requirePublicProfileDeliveryAuthority(t, report, revision)
	return report
}

func requirePublicProfileDeliveryAuthority(t *testing.T, report []byte, revision bool) {
	t.Helper()
	schema := policycompilation.PublicGenerationManifestSchema
	if revision {
		schema = policycompilation.PublicPolicyRevisionReportSchema
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(report, &fields); err != nil {
		t.Fatal(err)
	}
	expected := []struct {
		name  string
		value any
	}{
		{"schema", schema},
		{"repository_writes", 0},
		{"mutation_authority", 0},
		{"promotion_authority", 0},
		{"execution_observed", false},
		{"current_conformance", policycompilation.PublicGenerationConformanceUnknown},
	}
	for _, field := range expected {
		want, err := json.Marshal(field.value)
		if err != nil {
			t.Fatal(err)
		}
		got, exists := fields[field.name]
		if !exists || !bytes.Equal(bytes.TrimSpace(got), want) {
			t.Fatalf("delivery authority field %s: got=%s want=%s present=%t", field.name, got, want, exists)
		}
	}
}

func requirePublicProfileDeliveryBytes(t *testing.T, fixture publicProfileDeliveryFixture, revision, jsonMode bool, got, report []byte) {
	t.Helper()
	if jsonMode {
		if !bytes.Equal(got, report) {
			t.Fatalf("JSON delivery differs from the generated report: got=%q want=%q", got, report)
		}
		return
	}
	root, err := filepath.EvalSymlinks(fixture.outputDir)
	if err != nil {
		t.Fatal(err)
	}
	want := fmt.Sprintf("generated profile: %s\npolicy: %s\nartifact: %s\njudge: %s\nmanifest: %s\n",
		fixture.profile, filepath.Join(root, "policy.json"), filepath.Join(root, "artifact.json"),
		filepath.Join(root, "judge.go"), filepath.Join(root, fixture.reportName))
	if revision {
		want = fmt.Sprintf("generated profile: %s\ncandidate: %s\nproposal: %s\n",
			fixture.profile, filepath.Join(root, "candidate.gooo"), filepath.Join(root, fixture.reportName))
	}
	if string(got) != want {
		t.Fatalf("human delivery bytes differ: got=%q want=%q", got, want)
	}
}
