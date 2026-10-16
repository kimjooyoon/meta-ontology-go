package metricstrategy

import (
	"reflect"
	"strings"
	"testing"

	metric "github.com/kimjooyoon/meta-ontology-go/internal/meta/metriccounterfactualverify/intervention"
)

func cohortIndicatorFresh1673(id, operation, family string) metric.Indicator {
	trilemma := map[string]string{"foundation": "AXIOM", "coherence": "COHERENCE", "regression": "REGRESS"}[family]
	return metric.Indicator{ID: id, MetaOperation: operation, Family: family,
		Trilemma: trilemma, EvidenceDigest: "sha256:" + strings.Repeat("a", 64)}
}

func TestUpstreamDeclaredCohortFresh1673(t *testing.T) {
	first := cohortIndicatorFresh1673("first", "bind-exact-source-metrics", "foundation")
	second := cohortIndicatorFresh1673("second", first.MetaOperation, first.Family)
	terminal := cohortIndicatorFresh1673("terminal", "terminate-at-fixed-point", "regression")
	for _, row := range []struct {
		name       string
		indicators []metric.Indicator
		want       []string
	}{
		{"declared-only", nil, []string{"preserve-non-promoting-terminal", "terminate-at-fixed-point"}},
		{"upstream-union", []metric.Indicator{first}, []string{"bind-exact-source-metrics", "preserve-non-promoting-terminal", "terminate-at-fixed-point"}},
		{"same-upstream-operation", []metric.Indicator{first, second}, []string{"bind-exact-source-metrics", "preserve-non-promoting-terminal", "terminate-at-fixed-point"}},
		{"declared-terminal-overlap", []metric.Indicator{terminal}, []string{"preserve-non-promoting-terminal", "terminate-at-fixed-point"}},
	} {
		t.Run(row.name, func(t *testing.T) {
			cohort, err := ConceptOperationCohort(row.indicators)
			if err != nil {
				t.Fatal(err)
			}
			var subjects []string
			for _, operation := range cohort {
				subjects = append(subjects, operation.Subject)
				if operation.Subject == "terminate-at-fixed-point" && operation.Carrier != "replay-counterfactual" {
					t.Fatal("source-declared terminal carrier was lost")
				}
			}
			if !reflect.DeepEqual(subjects, row.want) {
				t.Fatalf("expected cohort must precede observations: %v", subjects)
			}
		})
	}
}

func TestCohortProofConflictFresh1673(t *testing.T) {
	indicators := []metric.Indicator{
		cohortIndicatorFresh1673("a", "bind-exact-source-metrics", "foundation"),
		cohortIndicatorFresh1673("b", "bind-exact-source-metrics", "coherence"),
	}
	if _, err := ConceptOperationCohort(indicators); err == nil {
		t.Fatal("conflicting upstream proof families changed the declared set")
	}
}
