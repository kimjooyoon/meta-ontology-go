package verify

import (
	"os"
	"strings"
	"testing"
)

var readinessPushDependencies = []string{
	".github/workflows/transformation-effect.yml",
	"cmd/guarded-promotion-witness/**",
	"cmd/language-readiness-witness/**",
	"internal/meta/languageconcept/**",
	"internal/meta/languagereadiness/**",
}

func TestReadinessDependenciesTriggerMetricStrategy(t *testing.T) {
	raw, err := os.ReadFile("../../.github/workflows/metric-counterfactual.yml")
	if err != nil {
		t.Fatal(err)
	}
	text := string(raw)
	pullStart := strings.Index(text, "\n  pull_request:")
	pushStart := strings.Index(text, "\n  push:")
	permissionsStart := strings.Index(text, "\npermissions:")
	if pullStart < 0 || pushStart <= pullStart || permissionsStart <= pushStart {
		t.Fatal("metric strategy trigger sections malformed")
	}
	completed := countReadinessTriggerCoordinates(
		t, text[pullStart:pushStart], "pull_request")
	completed += countReadinessTriggerCoordinates(
		t, text[pushStart:permissionsStart], "push")
	total := len(readinessPushDependencies) * 2
	if completed != total {
		t.Fatalf("readiness trigger coordinates: %d/%d", completed, total)
	}
}

func countReadinessTriggerCoordinates(t *testing.T, section, event string) int {
	t.Helper()
	completed := 0
	for _, path := range readinessPushDependencies {
		coordinate := "'" + path + "'"
		if count := strings.Count(section, coordinate); count != 1 {
			t.Errorf("%s trigger %s count = %d, want 1", event, path, count)
			continue
		}
		completed++
	}
	return completed
}
