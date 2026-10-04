package bodycodegen

import (
	"context"
	"fmt"

	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

type Realization struct {
	Schema                  string `json:"schema"`
	Source                  string `json:"gooo_source"`
	OriginalSourceSHA256    string `json:"original_source_sha256"`
	GenerationSourceSHA256  string `json:"generation_source_sha256"`
	RealizedSourceSHA256    string `json:"realized_source_sha256"`
	DocumentSHA256          string `json:"document_sha256"`
	TestSuiteSHA256         string `json:"test_suite_sha256"`
	ActivityID              string `json:"activity_id"`
	FinitePassed            int    `json:"finite_passed"`
	FiniteTotal             int    `json:"finite_total"`
	ModelCalls              int    `json:"model_calls"`
	LegacyCheckpointUpgrade bool   `json:"legacy_checkpoint_upgrade"`
}

// RealizeSourceAssembly replays observations and returns a reusable Gooo
// checkpoint. Partial finite scores remain visible; no search or inference runs.
func RealizeSourceAssembly(ctx context.Context, filename string, source []byte, prior Result) (Realization, error) {
	document, err := DecodeSourcePathDocument(ctx, filename, source, prior.Report.Activity, nil)
	if err != nil {
		return Realization{}, err
	}
	var completed []byte
	if err := replayTypedPathProjection(ctx, filename, source, document, prior, &completed); err != nil {
		return Realization{}, err
	}
	file, diagnostics := syntax.ParseFile(filename, string(source))
	if diagnostics.HasErrors() {
		return Realization{}, fmt.Errorf("realization source: %v", diagnostics)
	}
	for _, declaration := range file.Declarations {
		activity, ok := declaration.(*syntax.ActivityDecl)
		if !ok || activity.Name != prior.Report.Activity {
			continue
		}
		if activity.Assembly == nil {
			return Realization{}, fmt.Errorf("body-realize requires a source assembling contract")
		}
		if prior.Report.BodyPaths.SourceFormat == "" {
			completed, err = upgradeLegacyCheckpoint(source, activity, document, prior)
			if err != nil {
				return Realization{}, err
			}
		}
		passed, total := bodyPathSelectedScore(prior.Report.BodyPaths)
		return Realization{Schema: "gooo/body-realization/v1", Source: string(completed),
			OriginalSourceSHA256: digest(source), GenerationSourceSHA256: prior.Report.SourceDigest,
			RealizedSourceSHA256: digest(completed), DocumentSHA256: prior.Report.BodyPaths.DocumentSHA256,
			TestSuiteSHA256: prior.Report.BodyPaths.TestSuiteSHA256, ActivityID: prior.Report.ActivityID,
			FinitePassed: passed, FiniteTotal: total,
			LegacyCheckpointUpgrade: prior.Report.BodyPaths.SourceFormat == ""}, nil
	}
	return Realization{}, fmt.Errorf("realization activity missing")
}

func upgradeLegacyCheckpoint(source []byte, activity *syntax.ActivityDecl,
	document pathplan.Document, prior Result) ([]byte, error) {
	prepared, err := document.Prepare()
	if err != nil {
		return nil, err
	}
	choices := bodyPathSelection(prior.Report.BodyPaths).Choices
	selected, err := prepared.Compile(choices)
	if err != nil {
		return nil, err
	}
	return selectedTypedPathSource(source, activity, selected.GoooBody(), choices, true)
}
