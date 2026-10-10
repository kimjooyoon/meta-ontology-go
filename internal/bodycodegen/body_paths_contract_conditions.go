package bodycodegen

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math"

	decision "github.com/kimjooyoon/gooo-decision-runtime"
	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
)

// Condition goals are a separate input channel from desired function outputs.
// Normal receipts retain its identity; explicit inspection also exports rows.
type ContractConditionContext struct {
	FeatureVersion string `json:"feature_version"`
	FeatureSHA     string `json:"input_sha256"`
	Count          int    `json:"count"`
	Bytes          int    `json:"bytes"`
}

type ExportedContractConditions struct {
	ContractConditionContext
	Features [][decision.DeclaredConditionFeatureDim]float32 `json:"features"`
}

func contractConditionContext(ctx context.Context, input *pathplan.ContractInput) (*ContractConditionContext, error) {
	r := &ContractConditionContext{FeatureVersion: decision.DeclaredConditionFeatureVersion,
		Count: input.ConditionCount(), Bytes: input.ConditionCount() * decision.DeclaredConditionFeatureDim * 4}
	h := sha256.New()
	h.Write([]byte(r.FeatureVersion + "\x00"))
	var raw [decision.DeclaredConditionFeatureDim * 4]byte
	binary.LittleEndian.PutUint32(raw[:4], uint32(r.Count))
	h.Write(raw[:4])
	for i := range r.Count {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		row, err := input.ConditionFeatures(i)
		if err != nil {
			return nil, err
		}
		for j, value := range row {
			binary.LittleEndian.PutUint32(raw[j*4:], math.Float32bits(value))
		}
		h.Write(raw[:])
	}
	r.FeatureSHA = fmt.Sprintf("sha256:%x", h.Sum(nil))
	return r, ctx.Err()
}

func exportContractConditions(ctx context.Context, document pathplan.Document,
	prepared *pathplan.PreparedPlan) (*ExportedContractConditions, error) {
	input, err := prepared.InitialContractInput(document.TestCases)
	if err != nil {
		return nil, err
	}
	identity, err := contractConditionContext(ctx, input)
	if err != nil {
		return nil, err
	}
	r := &ExportedContractConditions{ContractConditionContext: *identity,
		Features: make([][decision.DeclaredConditionFeatureDim]float32, input.ConditionCount())}
	for i := range r.Features {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if err := input.ConditionFeaturesInto(i, &r.Features[i]); err != nil {
			return nil, err
		}
	}
	return r, nil
}
