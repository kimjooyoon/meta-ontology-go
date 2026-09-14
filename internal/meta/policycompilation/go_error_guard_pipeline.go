package policycompilation

import "fmt"

type GoGuardPipelineProposal struct {
	State              string                 `json:"state"`
	Reason             string                 `json:"reason"`
	ProgramDigest      string                 `json:"program_digest"`
	SemanticDigest     string                 `json:"semantic_digest"`
	SourceDigest       string                 `json:"source_digest"`
	Guard              *GoErrorGuardProposal  `json:"guard,omitempty"`
	Rendering          *GoGuardRendering      `json:"rendering,omitempty"`
	CandidateDigest    string                 `json:"candidate_digest,omitempty"`
	CandidateSource    string                 `json:"candidate_source,omitempty"`
	Pending            *PolicyRevisionPending `json:"pending,omitempty"`
	Admission          PolicyRevisionPending  `json:"admission"`
	Improvement        string                 `json:"improvement"`
	MutationAuthority  int                    `json:"mutation_authority"`
	PromotionAuthority int                    `json:"promotion_authority"`
}

// ProposeGoErrorGuardPipeline executes two explicitly connected Gooo computes
// activities. Rendering is not semantic admission or source application.
func ProposeGoErrorGuardPipeline(filename string, program, source []byte) (GoGuardPipelineProposal, error) {
	report := GoGuardPipelineProposal{
		State: "UNKNOWN", ProgramDigest: DigestBytes(program), SourceDigest: DigestBytes(source),
		Admission: revisionPending("INDEPENDENT_VALIDATION", "OBSERVE_CANONICAL_GO_GUARD",
			"CANONICAL_GUARD_NATIVE_EVIDENCE_MISSING", "RUN_FROZEN_NATIVE_ORACLE"),
		Improvement: "UNKNOWN",
	}
	plan, err := compileGoGuardPipeline(filename, program)
	if err != nil {
		report.State, report.Reason = "REFUTED", "GOOO_GUARD_PIPELINE_INVALID"
		return report, fmt.Errorf("%s: %w", report.Reason, err)
	}
	report.SemanticDigest = plan.semanticDigest
	raw := goGuardPipelineRawProposal(report, plan)
	raw, err = proposeGoErrorGuardSource(raw, plan.guard, source)
	report.Guard = &raw
	if err != nil {
		report.State, report.Reason, report.Pending = raw.State, raw.Reason, raw.Pending
		blocked := blockedGoGuardRendering(plan, raw)
		report.Rendering = &blocked
		return report, err
	}
	rendered, candidate, err := renderGoGuardPipeline(plan, source, raw)
	report.Rendering = &rendered
	if err != nil {
		report.State, report.Reason, report.Pending = rendered.State, rendered.Reason, rendered.Pending
		return report, err
	}
	report.CandidateSource, report.CandidateDigest = candidate, rendered.OutputDigest
	report.State, report.Reason = "PROPOSED", "GOOO_BOUND_CANONICAL_GO_GUARD"
	return report, nil
}

func goGuardPipelineRawProposal(report GoGuardPipelineProposal, plan goGuardPipelinePlan) GoErrorGuardProposal {
	return GoErrorGuardProposal{
		State: "UNKNOWN", ActivityID: plan.guardActivity,
		ProgramDigest: report.ProgramDigest, SemanticDigest: report.SemanticDigest,
		SourceDigest: report.SourceDigest,
		Admission: revisionPending("INDEPENDENT_VALIDATION", "OBSERVE_GO_GUARD_CANDIDATE",
			"NATIVE_GO_GUARD_EVIDENCE_MISSING", "RUN_FROZEN_NATIVE_ORACLE"),
		Improvement: "UNKNOWN",
	}
}
