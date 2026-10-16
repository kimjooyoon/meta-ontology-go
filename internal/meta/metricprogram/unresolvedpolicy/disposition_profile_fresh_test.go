package unresolvedpolicy_test

import (
	"os"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const dispositionRecord1751 = `
entity OperationDisposition id "gooo://metric-operation-origin/disposition-record" fields {
    field resolution id "gooo://metric-operation-origin/disposition-record/resolution" type string required one
    field reason id "gooo://metric-operation-origin/disposition-record/reason" type string required one
    field next_operation id "gooo://metric-operation-origin/disposition-record/next-operation" type string required one
    field readiness_credit id "gooo://metric-operation-origin/disposition-record/readiness-credit" type boolean required one
    field repository_workspace_writes id "gooo://metric-operation-origin/disposition-record/repository-workspace-writes" type boolean required one
    field promotion_authorized id "gooo://metric-operation-origin/disposition-record/promotion-authorized" type boolean required one
}

activity DescribeOperation(registered Boolean, originRecorded Boolean) -> OperationDisposition
    id "gooo://metric-operation-origin/describe-operation" computes ` + "`" + `
return OperationDisposition{
    resolution: OperationResolution(registered, originRecorded),
    reason: OperationReason(registered, originRecorded),
    next_operation: OperationNextStep(registered, originRecorded),
    readiness_credit: false,
    repository_workspace_writes: false,
    promotion_authorized: false,
}
` + "`" + "\n"

func TestSourceDispositionProfileFresh1751(t *testing.T) {
	base, err := os.ReadFile("testdata/policy-labels-v1.txt")
	if err != nil {
		t.Fatal(err)
	}
	source := append(append([]byte(nil), base...), []byte("\n"+dispositionRecord1751)...)
	t.Run("ordinary-parser-retains-deferred-contract", func(t *testing.T) {
		_, diagnostics := syntax.ParseFile("disposition-profile-1751.gooo", string(source))
		if !diagnostics.HasErrors() || !strings.Contains(diagnostics.Error().Error(), "parse.entity-fields-deferred") {
			t.Fatalf("ordinary parser contract changed: %v", diagnostics)
		}
	})
	t.Run("explicit-record-profile", func(t *testing.T) {
		_, diagnostics := syntax.ParseFileWithEntityFieldsSupport("disposition-profile-1751.gooo", string(source),
			syntax.EntityFieldsV3Support())
		if diagnostics.HasErrors() {
			t.Fatal(diagnostics.Error())
		}
	})
	t.Run("body-profile-source-record", func(t *testing.T) {
		result, err := bodycodegen.Generate("disposition-profile-1751.gooo", source, "DescribeOperation")
		if err != nil {
			t.Fatal(err)
		}
		if !result.Report.TypecheckPassed || !result.Report.DeterministicReplay ||
			result.Report.RouteDecision.Provider != "deterministic" || len(result.Report.RecordTypes) != 1 ||
			len(result.Report.RecordTypes[0].Fields) != 6 || result.Report.CallClosure == nil ||
			len(result.Report.CallClosure.Activities) != 3 {
			t.Fatalf("source-owned record or fixed pure callees lost: %+v", result.Report)
		}
	})
}
