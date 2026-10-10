package bodycodegen

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math"

	decision "github.com/kimjooyoon/gooo-decision-runtime"
	"github.com/kimjooyoon/gooo-decision-runtime/contractdecision"
	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
)

// Goals and observed failures have separate identities. Only explicit context
// export discloses the case arrays; ordinary receipts retain their digests.
type ContractCaseContext struct {
	FeatureVersion string `json:"feature_version"`
	CaseSHA        string `json:"finite_cases_sha256"`
	FeatureSHA     string `json:"input_sha256"`
	Count          int    `json:"count"`
	Bytes          int    `json:"bytes"`
}

type ExportedContractCases struct {
	ContractCaseContext
	Features [][decision.DeclaredCaseFeatureDim]float32 `json:"features"`
}

func contractCaseContext(input *pathplan.ContractInput) (ContractCaseContext, error) {
	r := ContractCaseContext{FeatureVersion: decision.DeclaredCaseFeatureVersion, CaseSHA: input.CaseSHA256(),
		Count: input.CaseCount(), Bytes: input.CaseCount() * decision.DeclaredCaseFeatureDim * 4}
	h := sha256.New()
	for i := range input.CaseCount() {
		row, err := input.CaseFeatures(i)
		if err != nil {
			return r, err
		}
		var raw [decision.DeclaredCaseFeatureDim * 4]byte
		for j, v := range row {
			binary.LittleEndian.PutUint32(raw[j*4:], math.Float32bits(v))
		}
		h.Write(raw[:])
	}
	r.FeatureSHA = fmt.Sprintf("sha256:%x", h.Sum(nil))
	return r, nil
}

func exportContractCases(document pathplan.Document, prepared *pathplan.PreparedPlan) (*ExportedContractCases, error) {
	input, err := prepared.InitialContractInput(document.TestCases)
	if err != nil {
		return nil, err
	}
	identity, err := contractCaseContext(input)
	if err != nil {
		return nil, err
	}
	r := &ExportedContractCases{ContractCaseContext: identity,
		Features: make([][decision.DeclaredCaseFeatureDim]float32, input.CaseCount())}
	for i := range r.Features {
		if err := input.CaseFeaturesInto(i, &r.Features[i]); err != nil {
			return nil, err
		}
	}
	return r, nil
}

func exportedContractInput(input *pathplan.ContractInput, choice pathplan.Choice, ordered bool) (ExportedPathInput, error) {
	intent := digest([]byte(choice.Intent))
	r := ExportedPathInput{DecisionID: choice.ID,
		OriginalIntentSHA: intent, NaturalIntentSHA: intent, Text: choice.Intent}
	if ordered {
		var features [contractdecision.OrderedFeatureDim]float32
		if err := input.OrderedSourceFeaturesInto(choice.ID, &features); err != nil {
			return r, err
		}
		r.OrderedFeatures, r.Bytes, r.InputSHA = &features, len(features)*4, candidateFeatureDigest(features[:])
		return r, nil
	}
	var features [decision.ExecutionFlowFeatureDim]float32
	if err := input.RelationalSourceFeaturesInto(choice.ID, &features); err != nil {
		return r, err
	}
	r.FlowFeatures, r.Bytes, r.InputSHA = &features, len(features)*4, candidateFeatureDigest(features[:])
	return r, nil
}

func exportedContractInputs(document pathplan.Document, prepared *pathplan.PreparedPlan, ordered bool) ([]ExportedPathInput, error) {
	input, err := prepared.InitialContractInput(document.TestCases)
	if err != nil {
		return nil, err
	}
	r := make([]ExportedPathInput, 0, len(document.Plan.Decisions))
	for _, choice := range document.Plan.Decisions {
		exported, err := exportedContractInput(input, choice, ordered)
		if err != nil {
			return nil, err
		}
		if ordered {
			view, err := prepared.OrderedBranchContext(choice.ID)
			if err != nil {
				return nil, err
			}
			if !view.Available {
				return nil, fmt.Errorf("ordered explanation unavailable: %s", view.Reason)
			}
			exported.OrderedBranch = &view
		}
		r = append(r, exported)
	}
	return r, nil
}

func prepareContractModelContext(ctx context.Context, document pathplan.Document, prepared *pathplan.PreparedPlan,
	model *conditionPathModel, activityID, semanticSHA string) (*pathplan.PreparedPlan, *PathModelContextReceipt, bool, error) {
	r := &PathModelContextReceipt{Schema: "gooo/compiler-contract-model-context/v1", Status: "ENCODED", ActivityID: activityID,
		SourceSemanticSHA: semanticSHA, OriginalPlanSHA: prepared.PlanSHA256(), RankedPlanSHA: prepared.PlanSHA256(),
		ArtifactSHA: model.artifactSHA, ModelFingerprint: "sha256:" + model.fingerprint(), FeatureVersion: model.featureVersion(),
		Inputs: []PathContextInput{}, Scope: "source arrays and all declared input/expected-output cases; no candidate execution or prediction; FP32 feature digests use little-endian bytes"}
	input, err := prepared.InitialContractInput(document.TestCases)
	if err != nil {
		return prepared, r, false, err
	}
	cases, err := contractCaseContext(input)
	if err != nil {
		return prepared, r, false, err
	}
	r.ContractCases = &cases
	if model.orderedContract() {
		r.Schema = "gooo/compiler-interaction-contract-model-context/v1"
		if model.contract.ArtifactSchema() == contractdecision.OrderedRequirementSchema {
			r.Schema = "gooo/compiler-ordered-contract-model-context/v1"
		}
		r.Scope = "ordered source arrays, all declared outputs and Boolean conditions; zero predictions or candidate executions; FP32 digests use little-endian bytes"
		r.ContractConditions, err = contractConditionContext(ctx, input)
		if err != nil {
			return prepared, r, false, err
		}
	}
	return appendContractSourceContext(ctx, document, prepared, input, r)
}

func appendContractSourceContext(ctx context.Context, document pathplan.Document, prepared *pathplan.PreparedPlan,
	input *pathplan.ContractInput, r *PathModelContextReceipt) (*pathplan.PreparedPlan, *PathModelContextReceipt, bool, error) {
	for _, choice := range document.Plan.Decisions {
		if err := ctx.Err(); err != nil {
			return prepared, r, false, err
		}
		exported, err := exportedContractInput(input, choice, r.FeatureVersion == contractdecision.OrderedSourceFeatureVersion)
		if err != nil {
			r.Status, r.Reason, r.DeclinedDecision = "DECLINED_TO_DETERMINISTIC", err.Error(), choice.ID
			return prepared, r, true, nil
		}
		r.Inputs = append(r.Inputs, exported.PathContextInput)
	}
	return prepared, r, false, ctx.Err()
}
