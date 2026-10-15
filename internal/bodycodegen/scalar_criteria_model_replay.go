package bodycodegen

import (
	"context"
	"fmt"
	"math"
	"reflect"
	"slices"

	"github.com/kimjooyoon/gooo-decision-runtime/criteriadecision"
	"github.com/kimjooyoon/gooo-decision-runtime/criteriainput"
)

// Saved scores are observations. Replay checks their source bindings and finite
// numerical consistency, then independently re-evaluates constructed candidates.
// It loads no historical model and makes no new prediction.
func verifyScalarCriteriaRanking(ctx context.Context, plan recordAssemblyPlan, r, expected *RecordAssemblyReceipt) error {
	info := r.Model
	if info.Schema != "gooo/retained-path-model/v1" || !info.Loaded || info.FeatureVersion != criteriainput.FeatureSchema ||
		!validDigest(info.MetadataSHA256) || !validDigest(info.WeightsSHA256) ||
		math.IsNaN(info.SetupMS) || math.IsInf(info.SetupMS, 0) || info.SetupMS < 0 ||
		info.ResidentTensorBytes != criteriadecision.WeightBytes || info.ModelFingerprint != info.MetadataSHA256 ||
		r.Context == nil || r.Prediction != nil || r.ModelCalls < 0 || r.ModelCalls > 1 || r.PredictNS < 0 {
		return fmt.Errorf("complete scalar criteria model observation required")
	}
	features, sourceContext, err := scalarCriteriaModelInput(ctx, plan)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(sourceContext, r.Context) {
		return fmt.Errorf("scalar criteria source context differs")
	}
	if sourceContext.Status == "DECLINED_TO_DETERMINISTIC" {
		if r.ModelCalls != 0 || r.PredictNS != 0 || r.ScalarPrediction != nil || !slices.Equal(r.Ranking, expected.Ranking) {
			return fmt.Errorf("declined scalar context must retain deterministic ranking")
		}
		return nil
	}
	if r.ModelCalls != 1 || r.ScalarPrediction == nil {
		return fmt.Errorf("one scalar criteria model prediction required")
	}
	if err := validateScalarCriteriaPrediction(features, r.ScalarPrediction, info); err != nil {
		return err
	}
	for i, mask := range r.Ranking {
		if mask != r.ScalarPrediction.Candidates[i].Mask {
			return fmt.Errorf("scalar ranking differs from recorded score order")
		}
	}
	return nil
}

func validateScalarCriteriaPrediction(features *criteriainput.Features, prediction *criteriadecision.Prediction, info *RetainedModelInfo) error {
	featureSHA, err := criteriadecision.FeatureSHA256(features)
	if err != nil {
		return err
	}
	if prediction.Schema != criteriadecision.RankingSchema || prediction.FeatureSchema != features.Schema ||
		prediction.InputSHA256 != features.InputSHA256 || prediction.CandidatePlanSHA256 != features.CandidatePlanSHA256 ||
		prediction.FeatureSHA256 != featureSHA || prediction.ChoiceCount != features.ChoiceCount ||
		prediction.PredicateCount != features.PredicateCount || prediction.CandidateCount != 1<<features.ChoiceCount ||
		prediction.ModelMetadataSHA256 != info.MetadataSHA256 || prediction.ModelWeightsSHA256 != info.WeightsSHA256 ||
		prediction.ProposedMask != prediction.Candidates[0].Mask {
		return fmt.Errorf("scalar prediction bindings differ from source, features or model")
	}
	return validateScalarCriteriaDistribution(prediction)
}

func validateScalarCriteriaDistribution(p *criteriadecision.Prediction) error {
	var seen [criteriadecision.MaxCandidates]bool
	var marginals [criteriainput.MaxChoices][2]float64
	var total float64
	for i, candidate := range p.Candidates[:p.CandidateCount] {
		if int(candidate.Mask) >= int(p.CandidateCount) || seen[candidate.Mask] ||
			math.IsNaN(candidate.Score) || math.IsInf(candidate.Score, 0) ||
			math.IsNaN(candidate.Probability) || candidate.Probability < 0 || candidate.Probability > 1 {
			return fmt.Errorf("finite complete scalar candidate distribution required")
		}
		seen[candidate.Mask] = true
		if i > 0 {
			previous := p.Candidates[i-1]
			if previous.Score < candidate.Score || previous.Score == candidate.Score && previous.Mask > candidate.Mask {
				return fmt.Errorf("scalar candidate score order differs")
			}
		}
		total += math.Exp(candidate.Score - p.Candidates[0].Score)
		for choice := range p.ChoiceCount {
			marginals[choice][candidate.Mask>>choice&1] += candidate.Probability
		}
	}
	return validateScalarCriteriaProbabilities(p, marginals, total)
}

func validateScalarCriteriaProbabilities(p *criteriadecision.Prediction, marginals [criteriainput.MaxChoices][2]float64, total float64) error {
	for _, candidate := range p.Candidates[:p.CandidateCount] {
		expected := math.Exp(candidate.Score-p.Candidates[0].Score) / total
		if math.Abs(candidate.Probability-expected) > 1e-12 {
			return fmt.Errorf("scalar probability differs from the recorded scores")
		}
	}
	for choice := range marginals {
		for option := range 2 {
			value := p.Marginals[choice][option]
			if math.IsNaN(value) || math.IsInf(value, 0) || math.Abs(value-marginals[choice][option]) > 1e-12 {
				return fmt.Errorf("scalar marginals differ from the joint distribution")
			}
		}
	}
	for _, candidate := range p.Candidates[p.CandidateCount:] {
		if candidate != (criteriadecision.Candidate{}) {
			return fmt.Errorf("unused scalar candidate slots must be empty")
		}
	}
	return nil
}
