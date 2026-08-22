package directorykind

import "fmt"

func Build(source SourceMetrics) (Report, error) {
	if err := validateSource(source); err != nil {
		return Report{}, err
	}
	ontologyDigest, err := validateOntology()
	if err != nil {
		return Report{}, err
	}
	first, err := buildCore(source, ontologyDigest)
	if err != nil {
		return Report{}, err
	}
	second, err := buildCore(source, ontologyDigest)
	if err != nil {
		return Report{}, err
	}
	firstDigest, err := digest(first)
	if err != nil {
		return Report{}, err
	}
	secondDigest, err := digest(second)
	if err != nil || firstDigest != secondDigest {
		return Report{}, fmt.Errorf("directory kind replay diverged")
	}
	first.Summary.ReplayVerified = true
	first.Proofs = append(first.Proofs, Proof{Choice: "regression",
		MetaOperation: "replay-directory-kind-separation", Activity: "ReplayDirectoryKindSeparation",
		Satisfied: true, EvidenceDigest: firstDigest})
	sourceDigest, err := digest(source)
	if err != nil {
		return Report{}, err
	}
	decision, reason := "FIXED_POINT", "NO_KIND_SEPARATION_REQUIRED"
	if first.Summary.ViolatingIndicators > 0 {
		decision, reason = "PLAN_REVIEW", "MIXED_DIRECTORY_KINDS"
	}
	report := Report{Schema: ReportSchema, Repository: source.Repository, SubjectSHA: source.CommitSHA,
		SourceMetricsDigest: sourceDigest, OntologyDigest: ontologyDigest, Decision: decision, Reason: reason,
		RootPolicy: RootPolicy{CountsApplicability: "OBSERVED", TopologyApplicability: "NOT_APPLICABLE",
			TopologyReason: "ROOT_TOPOLOGY_EXEMPT", READMERequirement: "NOT_APPLICABLE"},
		Summary: first.Summary, Indicators: first.Indicators, Candidates: first.Candidates,
		Proofs: first.Proofs, PlanDigest: firstDigest}
	return sealReport(report)
}

func buildCore(source SourceMetrics, ontologyDigest string) (planCore, error) {
	targets, applicable, roots := kindTargets(source)
	candidates := make([]Candidate, 0, len(targets))
	for _, target := range targets {
		candidate, err := makeCandidate(source, target)
		if err != nil {
			return planCore{}, err
		}
		candidates = append(candidates, candidate)
	}
	summary := summarize(applicable, roots, targets, candidates)
	candidateDigest, err := digest(candidates)
	if err != nil {
		return planCore{}, err
	}
	proofs := []Proof{{Choice: "foundation", MetaOperation: "bind-directory-kind-ontology",
		Activity: "BindDirectoryKindFoundation", Satisfied: true, EvidenceDigest: ontologyDigest},
		{Choice: "coherence", MetaOperation: "resolve-directory-kind-separation",
			Activity: "ResolveMixedDirectoryKinds", Satisfied: true, EvidenceDigest: candidateDigest}}
	return planCore{Summary: summary, Indicators: buildIndicators(summary), Candidates: candidates, Proofs: proofs}, nil
}
