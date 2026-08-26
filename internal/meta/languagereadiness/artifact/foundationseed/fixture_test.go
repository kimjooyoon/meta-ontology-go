package foundationseed

import (
	"fmt"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/artifact/predecessorresolution"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/artifact/predecessorselection"
)

func exactResolution(t *testing.T) predecessorresolution.Report {
	t.Helper()
	current := testSHA(100)
	ancestors := make([]string, predecessorresolution.SearchLimit)
	for index := range ancestors {
		ancestors[index] = testSHA(index + 1)
	}
	attempts := make([]predecessorresolution.Attempt, 0, len(ancestors))
	for index, ancestor := range ancestors {
		selected, err := predecessorselection.Select(predecessorselection.Input{
			Repository: "owner/repository", CurrentHeadSHA: current,
			PredecessorSHA: ancestor, Branch: canonicalBranch,
			Workflow: canonicalWorkflow,
		})
		if err != nil {
			t.Fatal(err)
		}
		attempt := predecessorresolution.Attempt{
			Depth: index, AncestorSHA: ancestor, Selection: selected,
		}
		if index+1 < len(ancestors) {
			attempt.ParentSHA = ancestors[index+1]
		}
		attempts = append(attempts, attempt)
	}
	report, err := predecessorresolution.Build(predecessorresolution.Input{
		Repository: "owner/repository", CurrentHeadSHA: current,
		ImmediatePredecessorSHA: ancestors[0],
		SearchLimit: predecessorresolution.SearchLimit, Attempts: attempts,
	})
	if err != nil {
		t.Fatal(err)
	}
	return report
}

func testSHA(value int) string {
	return fmt.Sprintf("%040x", value)
}
