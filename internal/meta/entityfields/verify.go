package entityfields

import (
	"fmt"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/entityfieldsv1"
	"github.com/kimjooyoon/meta-ontology-go/internal/generator"
)

func Verify(observation entityfieldsv1.Observation, report Report) error {
	if err := verifyObservation(observation); err != nil { return err }
	if err := validateProfile(report); err != nil { return err }
	if report.Schema != ReportSchema || report.Decision != DecisionPass || report.Resolution != ResolutionExact { return fmt.Errorf("EntityFields report is not exact PASS") }
	if report.ActivityCount != len(cellSpecs) || report.BindingCount != len(cellSpecs) || len(report.Activities) != len(cellSpecs) { return fmt.Errorf("EntityFields activity binding denominator mismatch") }
	if err := verifyActivities(report.Activities); err != nil { return err }
	if len(report.Cells) != len(cellSpecs) { return fmt.Errorf("cell denominator mismatch") }
	if err := verifySummary(report.Summary); err != nil { return err }
	if err := verifyCounterexamples(report.Counterexamples); err != nil { return err }
	if report.Authority.RepositoryWrites != 0 || report.Authority.BranchSettingWrites != 0 || report.Authority.LocalTestExecutions != 0 || report.Authority.CrossProjectGates != 0 || report.Authority.PromotionAuthorized { return fmt.Errorf("EntityFields authority boundary changed") }
	if report.Improvement != "UNKNOWN" { return fmt.Errorf("EntityFields improvement is not an exact pair") }
	seenEvidence := map[string]bool{}
	for index, cell := range report.Cells {
		if cell.ID != cellSpecs[index].ID || cell.Activity != cellSpecs[index].Activity || cell.MetaOperation != cellSpecs[index].MetaOperation || cell.ProofChoice != cellSpecs[index].ProofChoice || cell.IndicatorClass != cellSpecs[index].IndicatorClass || cell.Producer != cellSpecs[index].Producer || cell.Consumer != cellSpecs[index].Consumer || cell.EvidenceKey != cellSpecs[index].EvidenceKey { return fmt.Errorf("cell %s is not canonical", cell.ID) }
		if err := cellEvidenceContract(cell, observation); err != nil { return err }
		if seenEvidence[cell.EvidenceDigest] { return fmt.Errorf("cell evidence digest is not distinct") }
		seenEvidence[cell.EvidenceDigest] = true
		if cell.Decision != DecisionPass || cell.Resolution != ResolutionExact { return fmt.Errorf("cell %s is not closed", cell.ID) }
	}
	if len(observation.Source) == 0 || len(observation.Formatted) == 0 || len(observation.Generated) == 0 || len(observation.SourceMap.Mappings) == 0 || !observation.GetPutRoundTrip { return fmt.Errorf("EntityFields output evidence incomplete") }
	generated := string(observation.Generated)
	if !strings.Contains(generated, "type Order struct") || !strings.Contains(generated, "OrderNumber string") { return fmt.Errorf("Go struct projection is incomplete") }
	for _, spec := range cellSpecs {
		if strings.Contains(generated, spec.Activity) { return fmt.Errorf("user Go projection contains meta activity %s", spec.Activity) }
	}
	return nil
}

func verifyObservation(observation entityfieldsv1.Observation) error {
	if observation.Schema != entityfieldsv1.Schema || observation.Profile.ID != "gooo.entityfields.go-projection.v1" || observation.Profile.Version != 1 || observation.Profile.Digest == "" { return fmt.Errorf("EntityFields observation profile mismatch") }
	if len(observation.Semantic.Entities) != 1 || len(observation.Semantic.Entities[0].Fields) != 2 || len(observation.Semantic.Activities) != 1 { return fmt.Errorf("EntityFields semantic shape mismatch") }
	fields := observation.Semantic.Entities[0].Fields
	if err := verifyField(fields[0], entityfieldsv1.FieldID, "OrderNumber"); err != nil { return err }
	if err := verifyField(fields[1], entityfieldsv1.SecondFieldID, "CustomerName"); err != nil { return err }
	if fields[0].Source.Start.Offset >= fields[1].Source.Start.Offset { return fmt.Errorf("EntityFields field declaration order mismatch") }
	if len(observation.SourceMap.Mappings) < 2 || !hasSymbol(observation.Symbols, "OrderNumber", entityfieldsv1.FieldID) || !hasSymbol(observation.Symbols, "CustomerName", entityfieldsv1.SecondFieldID) || !hasReference(observation.References, entityfieldsv1.SecondFieldID) { return fmt.Errorf("EntityFields source map or LSP evidence missing") }
	if len(observation.DeclarationOrder) != 2 || observation.DeclarationOrder[0] != entityfieldsv1.SourceID || observation.Semantic.Activities[0].Name != "LoadOrder" { return fmt.Errorf("EntityFields declaration order mismatch") }
	if len(observation.StableIDs) != 4 || !uniqueStableIDs(observation.StableIDs) { return fmt.Errorf("EntityFields stable identity domain mismatch") }
	if !observation.GetPutRoundTrip || !observation.PutGetRoundTrip || observation.GetPutOriginalDigest == "" || observation.GetPutWrittenDigest == "" || observation.PutGetInputDigest == "" || observation.PutGetObservedDigest == "" { return fmt.Errorf("EntityFields BX evidence is incomplete") }
	if len(observation.Counterexamples) != 6 { return fmt.Errorf("EntityFields counterexample denominator mismatch") }
	refuted, missing := 0, 0
	for _, counterexample := range observation.Counterexamples {
		if counterexample.PartialOutput || counterexample.InputDigest == "" || counterexample.OutputDigest == "" || counterexample.EvidenceDigest == "" {
			return fmt.Errorf("EntityFields counterexample is not executed evidence")
		}
		switch counterexample.Decision {
		case "REFUTED":
			if counterexample.Resolution != ResolutionExact || counterexample.Unknown != nil {
				return fmt.Errorf("EntityFields refuted counterexample is malformed")
			}
			refuted++
		case "UNKNOWN":
			if counterexample.ID != "missing-source" || counterexample.Resolution != ResolutionLower || counterexample.Unknown == nil || counterexample.Unknown.Stage == "" || counterexample.Unknown.Step == "" || counterexample.Unknown.Reason != "ENTITY_FIELDS_SOURCE_MISSING" || counterexample.Unknown.UnknownClass != "DIRECT_MISSING" || counterexample.Unknown.NextOperation != "RESTORE_ENTITY_FIELDS_SOURCE" || counterexample.Unknown.BlockedBy == nil || len(counterexample.Unknown.BlockedBy) != 0 {
				return fmt.Errorf("EntityFields missing-source counterexample is malformed")
			}
			missing++
		default:
			return fmt.Errorf("EntityFields counterexample has unknown decision %q", counterexample.Decision)
		}
	}
	if refuted != 5 || missing != 1 { return fmt.Errorf("EntityFields counterexample terminal mismatch") }
	return nil
}

func verifyCounterexamples(counterexamples []Counterexample) error {
	if len(counterexamples) != 6 { return fmt.Errorf("EntityFields report counterexample denominator mismatch") }
	seen := map[string]bool{}
	refuted, missing := 0, 0
	for _, counterexample := range counterexamples {
		if counterexample.ID == "" || seen[counterexample.ID] || counterexample.InputDigest == "" || counterexample.OutputDigest == "" || counterexample.EvidenceDigest == "" || counterexample.PartialOutput {
			return fmt.Errorf("EntityFields report counterexample evidence is incomplete")
		}
		seen[counterexample.ID] = true
		switch counterexample.Decision {
		case DecisionRefuted:
			if counterexample.Resolution != ResolutionExact || counterexample.Unknown != nil { return fmt.Errorf("EntityFields report refuted counterexample is malformed") }
			refuted++
		case "UNKNOWN":
			if counterexample.ID != "missing-source" || counterexample.Resolution != ResolutionLower || counterexample.Unknown == nil || counterexample.Unknown.Stage == "" || counterexample.Unknown.Step == "" || counterexample.Unknown.Reason != "ENTITY_FIELDS_SOURCE_MISSING" || counterexample.Unknown.UnknownClass != "DIRECT_MISSING" || counterexample.Unknown.NextOperation != "RESTORE_ENTITY_FIELDS_SOURCE" || counterexample.Unknown.BlockedBy == nil || len(counterexample.Unknown.BlockedBy) != 0 { return fmt.Errorf("EntityFields report unknown counterexample is malformed") }
			missing++
		default:
			return fmt.Errorf("EntityFields report counterexample decision is unknown")
		}
	}
	if refuted != 5 || missing != 1 { return fmt.Errorf("EntityFields report counterexample terminal mismatch") }
	return nil
}

func verifyField(field generator.Field, id, name string) error {
	if field.ID != id || field.Name != name || field.TypeRefID != "urn:gooo:type:string" || field.Presence != "required" || field.Cardinality != "one" { return fmt.Errorf("EntityFields field contract mismatch for %s", id) }
	for _, span := range []generator.SourceSpan{field.Source, field.IDSpan, field.NameSpan, field.TypeRefSpan, field.PresenceSpan, field.CardinalitySpan} { if span.URI == "" || span.End.Offset <= span.Start.Offset { return fmt.Errorf("EntityFields field span is incomplete for %s", id) } }
	return nil
}

func uniqueStableIDs(ids []string) bool {
	seen := make(map[string]struct{}, len(ids))
	for _, id := range ids { if id == "" { return false }; if _, ok := seen[id]; ok { return false }; seen[id] = struct{}{} }
	return true
}

func hasSymbol(symbols []entityfieldsv1.NavigationSymbol, name, id string) bool {
	for _, symbol := range symbols { if symbol.Name == name && symbol.ID == id && symbol.HasIdentity && symbol.Range.End.Line >= symbol.Range.Start.Line && symbol.SelectionRange.End.Line >= symbol.SelectionRange.Start.Line && symbol.IdentityRange.End.Line >= symbol.IdentityRange.Start.Line { return true } }
	return false
}

func hasReference(references []entityfieldsv1.NavigationReference, id string) bool {
	for _, reference := range references { if reference.ID == id && reference.Range.End.Line >= reference.Range.Start.Line { return true } }
	return false
}

func verifyActivities(activities []string) error {
	seen := map[string]bool{}
	for index, spec := range cellSpecs { if activities[index] != spec.Activity || seen[activities[index]] { return fmt.Errorf("activity binding is not one-to-one") }; seen[activities[index]] = true }
	return nil
}

func verifySummary(summary Summary) error {
	if summary.CellsTotal != 12 || summary.ClosedCells != 12 || summary.UnknownCells != 0 || summary.RefutedCells != 0 { return fmt.Errorf("cell terminal summary mismatch") }
	if summary.FoundationCells != 4 || summary.CoherenceCells != 4 || summary.RegressionCells != 4 { return fmt.Errorf("proof denominator mismatch") }
	if summary.DriverCells != 4 || summary.OutcomeCells != 4 || summary.GuardrailCells != 4 { return fmt.Errorf("indicator denominator mismatch") }
	return nil
}
