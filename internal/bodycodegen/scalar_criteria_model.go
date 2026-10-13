package bodycodegen

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/kimjooyoon/gooo-decision-runtime/criteriadecision"
	"github.com/kimjooyoon/gooo-decision-runtime/criteriainput"
	"github.com/kimjooyoon/meta-ontology-go/internal/fileopen"
)

func loadScalarCriteriaModel(name string, metadata []byte) (*criteriadecision.Model, error) {
	if len(metadata) > 8192 {
		return nil, fmt.Errorf("scalar criteria metadata exceeds 8192 bytes")
	}
	f, err := fileopen.ReadOnly(filepath.Join(filepath.Dir(name), "weights.bin"))
	if err != nil {
		return nil, err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil || !info.Mode().IsRegular() || info.Size() != criteriadecision.WeightBytes {
		return nil, fmt.Errorf("scalar criteria weights require a regular %d-byte file", criteriadecision.WeightBytes)
	}
	weights, err := io.ReadAll(io.LimitReader(f, criteriadecision.WeightBytes+1))
	if err != nil {
		return nil, err
	}
	return criteriadecision.Load(metadata, weights)
}

func describeScalarCriteriaModel(model *criteriadecision.Model, info *RetainedModelInfo) {
	info.MetadataSHA256, info.WeightsSHA256 = model.MetadataSHA256(), model.Metadata().WeightsSHA256
	info.ModelSchema, info.FeatureVersion = criteriadecision.Schema, criteriainput.FeatureSchema
	info.ResidentTensorBytes = criteriadecision.WeightBytes
	info.ModelFingerprint = info.MetadataSHA256
	info.Scope = "one immutable scalar source-criteria model; parameter bytes exclude metadata, source preparation and per-call scratch"
}

func (g *TypedPathGenerator) rankScalarCriteria(ctx context.Context, plan recordAssemblyPlan, r *RecordAssemblyReceipt) error {
	info := g.Info()
	r.Model = &info
	features, sourceContext, err := scalarCriteriaModelInput(ctx, plan)
	r.Context = sourceContext
	if err != nil || sourceContext.Status != "ENCODED" {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	var work criteriadecision.Workspace
	var prediction criteriadecision.Prediction
	started := time.Now()
	err = g.criteria.PredictInto(features, &work, &prediction)
	r.ModelCalls, r.PredictNS = 1, time.Since(started).Nanoseconds()
	if err != nil {
		return fmt.Errorf("scalar criteria prediction: %w", err)
	}
	r.ScalarPrediction = &prediction
	if err := validateScalarCriteriaPrediction(features, &prediction, &info); err != nil {
		return err
	}
	for i := range prediction.CandidateCount {
		r.Ranking[i] = prediction.Candidates[i].Mask
	}
	return ctx.Err()
}
