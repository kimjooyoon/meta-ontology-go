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
	if !strings.Contains(string(observation.Generated), "type Order struct") || !strings.Contains(string(observation.Generated), "OrderNumber string") { return fmt.Errorf("Go struct projection is incomplete") }
	return nil
}

func verifyObservation(observation entityfieldsv1.Observation) error {
	if observation.Schema != entityfieldsv1.Schema || observation.Profile.ID != "gooo.entityfields.go-projection.v1" || observation.Profile.Version != 1 || observation.Profile.Digest == "" { return fmt.Errorf("EntityFields observation profile mismatch") }
	if len(observation.Semantic.Entities) != 1 || len(observation.Semantic.Entities[0].Fields) != 1 || len(observation.Semantic.Activities) != 12 { return fmt.Errorf("EntityFields semantic shape mismatch") }
	field := observation.Semantic.Entities[0].Fields[0]
	if field.ID != entityfieldsv1.FieldID || field.Name != "OrderNumber" || field.TypeRefID != "urn:gooo:type:string" || field.Presence != "required" || field.Cardinality != "one" { return fmt.Errorf("EntityFields field contract mismatch") }
	if len(observation.SourceMap.Mappings) == 0 || !hasSymbol(observation.Symbols, "OrderNumber", entityfieldsv1.FieldID) { return fmt.Errorf("EntityFields source map or LSP evidence missing") }
	if len(observation.DeclarationOrder) != 13 || observation.DeclarationOrder[0] != entityfieldsv1.SourceID { return fmt.Errorf("EntityFields declaration order mismatch") }
	if len(observation.StableIDs) != 14 || !uniqueStableIDs(observation.StableIDs) { return fmt.Errorf("EntityFields stable identity domain mismatch") }
	for _, span := range []generator.SourceSpan{field.Source, field.IDSpan, field.NameSpan, field.TypeRefSpan, field.PresenceSpan, field.CardinalitySpan} { if span.URI == "" || span.End.Offset <= span.Start.Offset { return fmt.Errorf("EntityFields field span is incomplete") } }
	return nil
}

func uniqueStableIDs(ids []string) bool {
	seen := make(map[string]struct{}, len(ids))
	for _, id := range ids { if id == "" { return false }; if _, ok := seen[id]; ok { return false }; seen[id] = struct{}{} }
	return true
}

func hasSymbol(symbols []entityfieldsv1.NavigationSymbol, name, id string) bool {
	for _, symbol := range symbols { if symbol.Name == name && symbol.ID == id { return true } }
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
