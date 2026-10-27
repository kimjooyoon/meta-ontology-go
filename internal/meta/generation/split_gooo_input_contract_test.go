package generation

import (
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
	"github.com/kimjooyoon/meta-ontology-go/internal/sourcebundle"
)

func TestSplitGoooOperationRetainsSourceActivityIdentity(t *testing.T) {
	evidence, err := SplitGoooInputContractEvidence()
	if err != nil || evidence.Operation != sourcepolicy.OperationSplitGooo || evidence.ActivityID == "" ||
		!evidence.UsedInputFact || !evidence.GeneratedOutputFact || evidence.InputSubjectKind != sourcepolicy.SubjectKindFile {
		t.Fatal("split operation lacks its source-owned contract", err)
	}
	found := 0
	for _, declaration := range evidence.SourceOrigins.Declarations {
		if declaration.ID == evidence.ActivityID {
			found++
			if declaration.Span.File == "" || declaration.Span.End <= declaration.Span.Start {
				t.Fatal("split activity lost its physical source span")
			}
		}
	}
	if found != 1 {
		t.Fatal("split activity must resolve to one authored origin")
	}
	raw := string(operationContractFixture(t))
	raw = strings.Replace(raw, "activity SplitGoooSections(FileInput) -> OperationResult",
		"activity SplitGoooSections(FileInput) -> OperationResult id \"urn:gooo:test:authored-partition-operation\"", 1)
	contract, err := parseOperationInputContractSources([]sourcebundle.Source{{Filename: "contract.gooo", Content: raw}})
	if err != nil || contract.Bindings[sourcepolicy.OperationSplitGooo].ActivityID != "urn:gooo:test:authored-partition-operation" {
		t.Fatal("native binding invented an activity ID instead of retaining the lowered source identity", err)
	}
}
