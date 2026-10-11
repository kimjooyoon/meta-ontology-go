package toolchainrelease

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/workspaceexecution"
)

func validateSingleConditionFill(r packageProfileReceipt, program workspaceexecution.Program, profile packageProfile) error {
	if profile.Name != "single-condition-byte" {
		if len(r.Result.Fills) != 0 || r.Result.SourceSHA != computationDigest([]byte(program.Source)) {
			return fmt.Errorf("fixed computation package unexpectedly filled a body")
		}
		return nil
	}
	if len(r.Result.Fills) != 1 {
		return fmt.Errorf("single-condition native fill is missing")
	}
	x := r.Result.Fills[0]
	g, f := x.Generation, x.Generation.Report.Fill
	if x.Activity.Package != profile.Package || x.Activity.Activity != profile.Activity ||
		x.Activity.Lowered != program.Entry.LoweredName || x.InputSHA != computationDigest([]byte(program.Source)) ||
		g.Report.Activity != program.Entry.LoweredName || g.Report.ID != profile.ID || g.Report.Decision != "PASS" ||
		g.Source == "" || g.GoooSource == "" || r.Result.SourceSHA != computationDigest([]byte(g.GoooSource)) ||
		g.Report.Generated != computationDigest([]byte(g.Source)) ||
		g.Report.Replay != g.Report.Generated || !g.Report.Typechecked || !g.Report.Replayed || f == nil ||
		f.Schema != "gooo/body-codegen-ir-fill-plan/v3-record" || f.Original != x.InputSHA || !boundedTextDigest(f.Plan) ||
		f.Decision == nil || f.Decision.Schema != "gooo/typed-decision-receipt/v1" ||
		f.Decision.Mode != "deterministic_fallback" || f.Decision.Selected != "lower" ||
		f.Decision.Provider != "deterministic" || f.Decision.Fallback != "NOT_CONFIGURED" ||
		!boundedTextDigest(f.Decision.Request) || !jointSmokeBool(f.ExternalKnown, true) ||
		f.Selected != "equal" || len(f.Holes) != 1 || f.Holes[0].ID != "match" ||
		f.Holes[0].Expression != "int64(input.text[offset]) == input.needle" ||
		!jointSmokeInt(f.Passed, 6) || !jointSmokeInt(f.Total, 6) || !jointSmokeInt(f.HoldoutPassed, 4) ||
		!jointSmokeInt(f.HoldoutTotal, 4) || !jointSmokeInt(f.LocalCalls, 0) || !jointSmokeInt(f.ExternalCalls, 0) {
		return fmt.Errorf("single-condition source fill or its finite results differ")
	}
	return nil
}
