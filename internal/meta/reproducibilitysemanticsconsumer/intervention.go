package reproducibilitysemanticsconsumer

import (
	"fmt"
	"reflect"
)

func BuildInterventionArtifact(base, semantic, presentation Judgment) (InterventionArtifact, error) {
	if err := validInterventionJudgment(base); err != nil {
		return InterventionArtifact{}, err
	}
	if err := validInterventionJudgment(semantic); err != nil {
		return InterventionArtifact{}, err
	}
	if err := validInterventionJudgment(presentation); err != nil {
		return InterventionArtifact{}, err
	}
	semanticCase, err := buildInterventionCase("semantic-source-change", "SEMANTIC_SOURCE_CHANGE", base, semantic)
	if err != nil {
		return InterventionArtifact{}, err
	}
	if semantic.SourceDigest == base.SourceDigest || semantic.SemanticDigest == base.SemanticDigest ||
		semantic.Summary.MeaningClaim == base.Summary.MeaningClaim || semantic.Summary.JointClaim == base.Summary.JointClaim ||
		!transitionChannelChanged(base.Cases, semantic.Cases, "meaning") || !transitionChannelChanged(base.Cases, semantic.Cases, "joint") {
		return InterventionArtifact{}, fmt.Errorf("semantic intervention did not change meaning and joint transitions")
	}
	presentationCase, err := buildInterventionCase("presentation-only-source-change", "PRESENTATION_ONLY_SOURCE_CHANGE", base, presentation)
	if err != nil {
		return InterventionArtifact{}, err
	}
	if presentation.SourceDigest == base.SourceDigest || presentation.SemanticDigest != base.SemanticDigest ||
		!reflect.DeepEqual(presentation.Summary, base.Summary) || !reflect.DeepEqual(presentation.Cases, base.Cases) {
		return InterventionArtifact{}, fmt.Errorf("presentation intervention changed semantic evidence")
	}
	artifact := InterventionArtifact{Schema: InterventionSchema, Version: 1, ContractID: ContractID,
		SourcePath: base.SourcePath, Denominator: InterventionCaseCount, Cases: []InterventionCase{semanticCase, presentationCase},
		Decision: StatusDischarged, Resolution: "EXACT", Reason: "TWO_SEPARATE_INTERVENTIONS_PERSISTED", Authority: Authority{}}
	return sealIntervention(artifact), nil
}

func validInterventionJudgment(judgment Judgment) error {
	if judgment.ConformanceDecision != StatusDischarged || judgment.ConformanceResolution != "EXACT" {
		return fmt.Errorf("intervention input is not exact conformance")
	}
	if len(judgment.Cases) != CaseCount {
		return fmt.Errorf("intervention input case count is %d", len(judgment.Cases))
	}
	return nil
}

func buildInterventionCase(id, kind string, before, after Judgment) (InterventionCase, error) {
	item := InterventionCase{ID: id, Kind: kind, Stage: "intervention", Step: "fixed-two-case-contract",
		Reason: "PERSISTED_CHANNEL_TRANSITIONS", SourceDigestBefore: before.SourceDigest, SourceDigestAfter: after.SourceDigest,
		SemanticDigestBefore: before.SemanticDigest, SemanticDigestAfter: after.SemanticDigest,
		MeaningBefore: before.Summary.MeaningClaim, MeaningAfter: after.Summary.MeaningClaim,
		JointBefore: before.Summary.JointClaim, JointAfter: after.Summary.JointClaim,
		TransitionsBefore: append([]JudgmentCase(nil), before.Cases...), TransitionsAfter: append([]JudgmentCase(nil), after.Cases...)}
	item.EvidenceDigest = digestValue(item)
	return item, nil
}

func transitionChannelChanged(before, after []JudgmentCase, channel string) bool {
	for index := range before {
		var left, right Transition
		switch channel {
		case "meaning":
			left, right = before[index].MeaningTransition, after[index].MeaningTransition
		case "joint":
			left, right = before[index].JointTransition, after[index].JointTransition
		default:
			return false
		}
		if left != right {
			return true
		}
	}
	return false
}

func ValidateIntervention(artifact InterventionArtifact) error {
	if artifact.Schema != InterventionSchema || artifact.Version != 1 || artifact.ContractID != ContractID || artifact.Denominator != InterventionCaseCount ||
		len(artifact.Cases) != InterventionCaseCount || artifact.Decision != StatusDischarged || artifact.Resolution != "EXACT" || artifact.Reason != "TWO_SEPARATE_INTERVENTIONS_PERSISTED" || artifact.Authority != (Authority{}) {
		return fmt.Errorf("intervention artifact contract invalid")
	}
	if artifact.Cases[0].ID != "semantic-source-change" || artifact.Cases[0].Kind != "SEMANTIC_SOURCE_CHANGE" || artifact.Cases[1].ID != "presentation-only-source-change" || artifact.Cases[1].Kind != "PRESENTATION_ONLY_SOURCE_CHANGE" {
		return fmt.Errorf("intervention artifact case order invalid")
	}
	for index := range artifact.Cases {
		item := artifact.Cases[index]
		want := item.EvidenceDigest
		item.EvidenceDigest = ""
		if want == "" || digestValue(item) != want {
			return fmt.Errorf("intervention evidence digest invalid at %d", index)
		}
	}
	unsigned := artifact.ArtifactDigest
	artifact.ArtifactDigest = ""
	if unsigned == "" || digestValue(artifact) != unsigned {
		return fmt.Errorf("intervention artifact digest invalid")
	}
	return nil
}

func sealIntervention(artifact InterventionArtifact) InterventionArtifact {
	artifact.ArtifactDigest = ""
	artifact.ArtifactDigest = digestValue(artifact)
	return artifact
}
