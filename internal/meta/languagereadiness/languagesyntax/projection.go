package languagesyntax

import (
	"fmt"
	"io/fs"
	"reflect"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesyntax/replay"
	"github.com/kimjooyoon/meta-ontology-go/internal/receiptprojection"
)

const FixedProjectionTotal = 2

type ProjectionDefinition struct {
	ID             string `json:"id"`
	Path           string `json:"path"`
	Profile        string `json:"profile"`
	Root           string `json:"root"`
	GoPath         string `json:"go_path"`
	JSONSchemaPath string `json:"json_schema_path"`
}

type ProjectionResult struct {
	Definition     ProjectionDefinition    `json:"definition"`
	Evidence       replay.ProjectionResult `json:"evidence"`
	Status         string                  `json:"status"`
	EvidenceDigest string                  `json:"evidence_digest"`
}

func expectedProjectionUnits() []ProjectionDefinition {
	return []ProjectionDefinition{{
		ID: "declared-completeness-receipt", Path: "internal/completeness/receipt.gooo",
		Profile: receiptprojection.Profile, Root: "CompletenessReceipt",
		GoPath: "internal/completeness/receipt.generated.go", JSONSchemaPath: "internal/completeness/receipt.schema.json",
	}, {
		ID: "declared-completeness-delta", Path: "internal/completenessdelta/delta.gooo",
		Profile: receiptprojection.Profile, Root: "CompletenessDelta",
		GoPath: "internal/completenessdelta/delta.generated.go", JSONSchemaPath: "internal/completenessdelta/delta.schema.json",
	}}
}

func evaluateProjections(repository fs.FS, source Source, execute bool) []ProjectionResult {
	results := make([]ProjectionResult, 0, FixedProjectionTotal)
	for _, def := range expectedProjectionUnits() {
		evidence := replay.ProjectionResult{ObservedDecision: replay.DecisionUnknown, Profile: def.Profile,
			Diagnostics: []string{"projection.source-or-registry-unresolved"}}
		if execute {
			evidence = replay.ExecuteProjection(repository, def.Path, def.Profile, def.Root, def.GoPath, def.JSONSchemaPath)
		}
		item := ProjectionResult{Definition: def, Evidence: evidence}
		item.Status = projectionStatus(item)
		item.EvidenceDigest = projectionDigest(item, source)
		results = append(results, item)
	}
	return results
}

func projectionStatus(item ProjectionResult) string {
	e := item.Evidence
	if e.ObservedDecision == replay.DecisionUnknown {
		return "UNRESOLVED"
	}
	if e.ObservedDecision == replay.DecisionPass && e.Profile == item.Definition.Profile &&
		e.ASTReplayed && e.ByteReplayed && e.StructureReplayed && e.ArtifactsMatched && len(e.Diagnostics) == 0 &&
		validDigest(e.SourceDigest) && validDigest(e.CanonicalDigest) && validDigest(e.StructureDigest) &&
		validDigest(e.GoArtifactDigest) && validDigest(e.JSONArtifactDigest) {
		return "SATISFIED"
	}
	return "NOT_SATISFIED"
}

func projectionDigest(item ProjectionResult, source Source) string {
	item.EvidenceDigest = ""
	return digestJSON(struct {
		Projection ProjectionResult `json:"projection"`
		Source     Source           `json:"source"`
	}{item, source})
}

func validateProjections(report Report) error {
	expected := expectedProjectionUnits()
	if len(report.ProjectionUnits) != FixedProjectionTotal {
		return fmt.Errorf("language syntax projection denominator mismatch")
	}
	for i, item := range report.ProjectionUnits {
		if !reflect.DeepEqual(item.Definition, expected[i]) || item.Status != projectionStatus(item) ||
			item.EvidenceDigest != projectionDigest(item, report.Source) {
			return fmt.Errorf("language syntax projection evidence mismatch")
		}
		if item.Status != "SATISFIED" {
			continue
		}
		bound := false
		for _, file := range report.Source.GoooFiles {
			if file.Path == item.Definition.Path && file.SourceDigest == item.Evidence.SourceDigest {
				bound = true
			}
		}
		if !bound {
			return fmt.Errorf("language syntax projection source mismatch")
		}
	}
	return nil
}
