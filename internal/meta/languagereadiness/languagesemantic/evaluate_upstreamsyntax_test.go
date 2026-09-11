package languagesemantic

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languageconcept"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesyntax"
)

const testSyntaxHead = "0000000000000000000000000000000000000000"

func TestSemanticConsumerAcceptsActualSyntaxProducerReceipt(t *testing.T) {
	root := testRepositoryRoot(t)
	report := actualSyntaxProducerReport(t, root)
	semantic, err := consumeSyntaxProducerReport(t, root, report)
	if err != nil {
		t.Fatal(err)
	}
	if semantic.Decision != DecisionPass || semantic.Resolution != ResolutionExact ||
		semantic.Summary.Total != FixedTotal || semantic.Summary.Unresolved != 0 {
		t.Fatalf("semantic report from actual syntax producer: decision=%s resolution=%s total=%d unresolved=%d", semantic.Decision, semantic.Resolution, semantic.Summary.Total, semantic.Summary.Unresolved)
	}
}

func TestSemanticConsumerRejectsRecomputedUpstreamTampering(t *testing.T) {
	root := testRepositoryRoot(t)
	canonical := actualSyntaxProducerReport(t, root)
	tests := []struct {
		name   string
		mutate func(*languagesyntax.Report)
	}{
		{
			name: "head",
			mutate: func(report *languagesyntax.Report) {
				report.HeadSHA = strings.Repeat("1", 40)
			},
		},
		{
			name: "source head",
			mutate: func(report *languagesyntax.Report) {
				report.Source.ExpectedHeadSHA = strings.Repeat("1", 40)
			},
		},
		{
			name: "registry digest",
			mutate: func(report *languagesyntax.Report) {
				report.Source.RegistryDigest = testDigest([]byte("tampered registry"))
			},
		},
		{
			name: "source count",
			mutate: func(report *languagesyntax.Report) {
				report.Source.GoooFiles = report.Source.GoooFiles[1:]
				report.Summary.GoooLines = 0
				for _, file := range report.Source.GoooFiles {
					report.Summary.GoooLines += file.GoooLines
				}
			},
		},
		{
			name: "source digest recomputed",
			mutate: func(report *languagesyntax.Report) {
				report.Source.GoooFiles[0].SourceDigest = testDigest([]byte("tampered source"))
			},
		},
		{
			name: "case identity",
			mutate: func(report *languagesyntax.Report) {
				report.Cases[0].Definition.ID = "tampered-case"
			},
		},
		{
			name: "package binding",
			mutate: func(report *languagesyntax.Report) {
				report.Source.PackageUnits[0].ID = "tampered-package"
			},
		},
		{
			name: "zero effects",
			mutate: func(report *languagesyntax.Report) {
				report.MutationAuthorized = true
			},
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			report := cloneSyntaxReport(t, canonical)
			testCase.mutate(&report)
			resignSyntaxReport(&report)
			semantic, err := consumeSyntaxProducerReport(t, root, report)
			if err != nil {
				t.Fatal(err)
			}
			if semantic.Decision != DecisionFailClosed || semantic.Resolution != ResolutionLower {
				t.Fatalf("tampered syntax report was accepted: decision=%s resolution=%s reason=%s", semantic.Decision, semantic.Resolution, semantic.ReasonCode)
			}
		})
	}
}

func actualSyntaxProducerReport(t *testing.T, root string) languagesyntax.Report {
	t.Helper()
	repository := os.DirFS(root)
	registry, err := os.ReadFile(filepath.Join(root, "examples/language-syntax-roundtrip/corpus.json"))
	if err != nil {
		t.Fatal(err)
	}
	report := languagesyntax.Evaluate(repository, testSyntaxHead, registry, languageconcept.BuildArtifact(repository))
	if err := languagesyntax.Validate(report, testSyntaxHead); err != nil {
		t.Fatalf("actual syntax producer report is invalid: %v", err)
	}
	return report
}

func consumeSyntaxProducerReport(t *testing.T, root string, syntaxReport languagesyntax.Report) (Report, error) {
	t.Helper()
	raw, err := json.MarshalIndent(syntaxReport, "", "  ")
	if err != nil {
		return Report{}, err
	}
	path := filepath.Join(t.TempDir(), "syntax-artifact.json")
	if err := os.WriteFile(path, append(raw, '\n'), 0o600); err != nil {
		return Report{}, err
	}
	return Evaluate(Input{
		Root:               root,
		ExpectedHeadSHA:    testSyntaxHead,
		RegistryPath:       filepath.Join(root, "examples/language-semantic-model/corpus.json"),
		SyntaxArtifactPath: path,
	})
}

func cloneSyntaxReport(t *testing.T, report languagesyntax.Report) languagesyntax.Report {
	t.Helper()
	raw, err := json.Marshal(report)
	if err != nil {
		t.Fatal(err)
	}
	var clone languagesyntax.Report
	if err := json.Unmarshal(raw, &clone); err != nil {
		t.Fatal(err)
	}
	return clone
}

func resignSyntaxReport(report *languagesyntax.Report) {
	report.Source.CorpusDigest = testDigestJSON(report.Source.GoooFiles)
	for index := range report.Cases {
		item := report.Cases[index]
		item.EvidenceDigest = ""
		item.EvidenceDigest = testDigestJSON(struct {
			Case   languagesyntax.CaseResult `json:"case"`
			Source languagesyntax.Source     `json:"source"`
		}{item, report.Source})
		report.Cases[index] = item
	}
	report.ReportDigest = ""
	report.ReportDigest = testDigestJSON(report)
}

func testDigestJSON(value any) string {
	raw, _ := json.Marshal(value)
	return testDigest(raw)
}

func testDigest(raw []byte) string {
	sum := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func testRepositoryRoot(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("test source path is unavailable")
	}
	root, err := filepath.Abs(filepath.Join(filepath.Dir(filename), "../../../../"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
		t.Fatalf("repository root is unavailable: %v", err)
	}
	return root
}
