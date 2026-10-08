package toolchainrelease

import (
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestLanguageSmokeEntryIsSelectedOnlyDuringConstruction(t *testing.T) {
	example := languageSmokeCase{source: "source.gooo", cases: "cases.json", entry: "Classify"}
	directory := filepath.Join("saved", "filename")
	want := []string{"body-compose", "--source", example.source, "--cases", example.cases,
		"--entry", "Classify", "--out", directory}
	if got := languageSmokeArgs(example, directory, false); !slices.Equal(got, want) {
		t.Fatalf("construction arguments: %v; want %v", got, want)
	}
	want = []string{"body-compose", "--source", filepath.Join(directory, "original.gooo"),
		"--cases", example.cases, "--composition", filepath.Join(directory, "composition.json")}
	if got := languageSmokeArgs(example, directory, true); !slices.Equal(got, want) {
		t.Fatalf("saved entry must come from composition: %v; want %v", got, want)
	}
	example.entry = ""
	want = []string{"body-compose", "--source", example.source, "--cases", example.cases, "--out", directory}
	if got := languageSmokeArgs(example, directory, false); !slices.Equal(got, want) {
		t.Fatalf("default entry arguments: %v; want %v", got, want)
	}
}

func TestLanguageSmokeRequiresNativeCasesAndSameSavedProgram(t *testing.T) {
	valid := `{"generated_now":true,"runtime":{"finite_passed":12,"finite_total":12,"model_calls":0,"projection_replayed":true,"runtime_replayed":true},"composition":{"generated_sha256":"sha256:program"}}`
	selected, err := validateLanguageSmoke([]byte(valid), 12, false, "")
	if err != nil {
		t.Fatal(err)
	}
	replay := strings.Replace(valid, `"generated_now":true`, `"generated_now":false`, 1)
	if _, err := validateLanguageSmoke([]byte(replay), 12, true, selected); err != nil {
		t.Fatal(err)
	}
	for _, raw := range []string{
		`{}`, valid + `{}`, strings.Replace(valid, `"finite_passed":12`, `"finite_passed":11`, 1),
		strings.Replace(valid, `"finite_total":12`, `"finite_total":11`, 1),
		strings.Replace(valid, `"model_calls":0`, `"model_calls":1`, 1),
		strings.Replace(valid, `"model_calls":0,`, ``, 1),
		strings.Replace(valid, `"projection_replayed":true`, `"projection_replayed":false`, 1),
		strings.Replace(valid, `"runtime_replayed":true`, `"runtime_replayed":false`, 1),
		strings.Replace(valid, `"generated_now":true,`, ``, 1),
	} {
		if _, err := validateLanguageSmoke([]byte(raw), 12, false, ""); err == nil {
			t.Fatal("incomplete native observation accepted", raw)
		}
	}
	for _, original := range []string{"", "sha256:other"} {
		if _, err := validateLanguageSmoke([]byte(replay), 12, true, original); err == nil {
			t.Fatal("unbound replay accepted")
		}
	}
}

func TestReleaseProofDescriptionUsesCurrentCorpus(t *testing.T) {
	corpus := Corpus{Version: 1, Targets: Targets(), Cases: expectedCases()}
	proof := buildProofs(corpus, "corpus", "concept", nil)[0]
	if proof.Claim != "the v1 corpus declares 4 native runner targets and 26 release cases" {
		t.Fatal("release proof description differs from current denominator", proof.Claim)
	}
}
