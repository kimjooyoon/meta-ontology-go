package toolchainrelease

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"reflect"

	"github.com/kimjooyoon/gooo-decision-runtime/contractdecision"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func validateRelativeV3Preflight(raw, source, model []byte) (*bodycodegen.TypedPathContextExport, error) {
	var r bodycodegen.TypedPathContextExport
	if err := json.Unmarshal(raw, &r); err != nil {
		return nil, err
	}
	c := r.ModelCompatibility
	if !typedPreflightZeros(raw, true) || r.OriginalSourceSHA256 != scalarPreflightDigest(source) ||
		r.Context == nil || r.Context.Status != "ENCODED" || r.Context.Schema != "gooo/compiler-source-relative-contract-model-context/v3" ||
		r.Context.ActivityID != "output-relations-v3://choose" || r.Context.FeatureVersion != contractdecision.SourceRelativeV3FeatureVersion ||
		!r.SourceBinding.Equivalent || r.SourceBinding.Decision != "PASS" || r.SourceBinding.SourceSemanticDigest == "" ||
		r.SourceBinding.SourceSemanticDigest != r.SourceBinding.GeneratedSemanticDigest || len(r.Inputs) != 2 || r.ContractCases == nil || r.ContractConditions == nil ||
		c == nil || c.Status != "READY_FOR_RANKING" || c.Reason != "SOURCE_INPUT_ENCODED" || c.Model.ArtifactSHA256 != scalarPreflightDigest(model) ||
		c.Model.ModelSchema != contractdecision.SourceRelativeV3Schema || c.Model.FeatureVersion != contractdecision.SourceRelativeV3FeatureVersion ||
		c.Model.ResidentTensorBytes != 76136 || !c.Model.Loaded || r.Context.ArtifactSHA != c.Model.ArtifactSHA256 ||
		r.Context.ModelFingerprint != c.Model.ModelFingerprint || r.Context.SourceSemanticSHA != r.SourceBinding.SourceSemanticDigest {
		return nil, fmt.Errorf("V3 preflight source, model identity or zero counters differ")
	}
	d, err := bodycodegen.DecodeSourcePathDocument(context.Background(), "relative-v3.gooo", source, "Choose", nil)
	if err != nil {
		return nil, err
	}
	p, err := d.Prepare()
	if err != nil {
		return nil, err
	}
	input, err := p.InitialContractInput(d.TestCases)
	if err != nil {
		return nil, err
	}
	if r.Context.OriginalPlanSHA != p.PlanSHA256() || r.ContractCases.CaseSHA != input.CaseSHA256() ||
		r.ContractCases.Count != 6 || len(r.ContractCases.Features) != 6 || r.ContractConditions.Count != 3 || len(r.ContractConditions.Features) != 3 {
		return nil, fmt.Errorf("V3 declared requirement identity differs")
	}
	for i, choice := range d.Plan.Decisions {
		var row [528]float32
		if err = input.SourceRelativeV3FeaturesInto(choice.ID, &row); err != nil {
			return nil, err
		}
		if r.Inputs[i].DecisionID != choice.ID || r.Inputs[i].SourceRelativeFeatures == nil || *r.Inputs[i].SourceRelativeFeatures != row {
			return nil, fmt.Errorf("V3 source array differs")
		}
		outputs, err := input.SourceRelativeOutputs(choice.ID)
		if err != nil || !outputs.Available || r.Inputs[i].SourceRelativeOutputs == nil ||
			!reflect.DeepEqual(*r.Inputs[i].SourceRelativeOutputs, outputs) {
			return nil, fmt.Errorf("V3 declared output relationships differ")
		}
		var bits [528 * 4]byte
		for j, v := range row {
			binary.LittleEndian.PutUint32(bits[j*4:], math.Float32bits(v))
		}
		if r.Inputs[i].InputSHA != scalarPreflightDigest(bits[:]) || r.Inputs[i].Bytes != len(bits) ||
			r.Inputs[i].OriginalIntentSHA != scalarPreflightDigest([]byte(choice.Intent)) || r.Inputs[i].Text != choice.Intent {
			return nil, fmt.Errorf("V3 source feature identity differs")
		}
	}
	for i, row := range r.ContractCases.Features {
		actual, err := input.CaseFeatures(i)
		if err != nil {
			return nil, err
		}
		if actual != row {
			return nil, fmt.Errorf("V3 output array differs")
		}
	}
	for i, row := range r.ContractConditions.Features {
		actual, err := input.ConditionFeatures(i)
		if err != nil {
			return nil, err
		}
		if actual != row {
			return nil, fmt.Errorf("V3 condition array differs")
		}
	}
	return &r, nil
}
