package bodycodegen

import (
	"context"
	"fmt"

	"github.com/kimjooyoon/gooo-decision-runtime/jointdecision"
)

type RecordModelCompatibility struct {
	Schema string            `json:"schema"`
	Status string            `json:"status"`
	Reason string            `json:"reason"`
	Model  RetainedModelInfo `json:"model"`
	Scope  string            `json:"scope"`
}

type RecordModelContextOptions struct {
	FeatureVersion         string
	IncludePlan, ValueFlow bool
}

// ExportRecordAssemblyModelContext validates one local model artifact and uses
// its feature contract to export source input. It never predicts or scores cases.
func ExportRecordAssemblyModelContext(ctx context.Context, filename string, source []byte,
	activity, modelPath string, options RecordModelContextOptions) (RecordAssemblyContextExport, error) {
	if ctx == nil || modelPath == "" {
		return RecordAssemblyContextExport{}, fmt.Errorf("context and explicit local model required")
	}
	if err := ctx.Err(); err != nil {
		return RecordAssemblyContextExport{}, err
	}
	g, err := NewTypedPathGenerator(modelPath)
	if err != nil {
		return RecordAssemblyContextExport{}, err
	}
	info := g.Info()
	if !recordContextModelFeature(info.FeatureVersion) || g.three == nil {
		return RecordAssemblyContextExport{}, fmt.Errorf("record context requires a record-field model feature; got %q", info.FeatureVersion)
	}
	if options.FeatureVersion != "" && options.FeatureVersion != info.FeatureVersion {
		return RecordAssemblyContextExport{}, fmt.Errorf("explicit feature version differs from loaded model feature %q", info.FeatureVersion)
	}
	r, err := exportRecordAssemblyContext(ctx, filename, source, activity, options.IncludePlan, info.FeatureVersion, options.ValueFlow)
	if err != nil {
		return RecordAssemblyContextExport{}, err
	}
	status, reason := r.Context.Status, r.Context.Reason
	if status == "ENCODED" {
		status, reason = "READY_FOR_RANKING", "SOURCE_INPUT_ENCODED"
	}
	r.ModelCompatibility = &RecordModelCompatibility{Schema: "gooo/record-model-compatibility/v1", Status: status,
		Reason: reason, Model: info, Scope: "verified local model and source input representation; no prediction, candidate outcomes or correctness guarantee"}
	return r, ctx.Err()
}

func recordContextModelFeature(feature string) bool {
	switch feature {
	case jointdecision.RecordFieldFeatureVersion, jointdecision.RecordSharedFeatureVersion,
		jointdecision.RecordOriginSharedFeatureVersion, jointdecision.RecordGraphSharedFeatureVersion:
		return true
	}
	return false
}
