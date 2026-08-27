package languageproofartifactverifier

const CoreParserDependencyInventoryTotal = 2
const ProofTotal = 3

const (
	ProofPhasePreliminary = "PRELIMINARY"
	ProofPhaseFinal       = "FINAL"
	ProofStateObserved    = "OBSERVED"
	ProofStateOpen        = "OPEN"
	ProofStateDischarged  = "DISCHARGED"
)

func MetricIDs() []string {
	return []string{
		"gooo.metric.language.proof-carrying-artifact-cases.v3",
		"gooo.metric.language.proof-carrying-artifact-claim-templates.v3",
		"gooo.metric.language.proof-carrying-artifact-claim-instances.v3",
		"gooo.metric.language.proof-carrying-artifact-valid.v3",
		"gooo.metric.language.proof-carrying-artifact-evidence-kinds.v3",
		"gooo.metric.language.proof-carrying-artifact-evidence-links.v3",
		"gooo.metric.language.proof-carrying-artifact-recipes.v3",
		"gooo.metric.language.proof-carrying-artifact-accepted-transitions.v3",
		"gooo.metric.language.proof-carrying-artifact-preserved-transitions.v3",
		"gooo.metric.language.proof-carrying-artifact-case-discharged.v3",
		"gooo.metric.language.proof-carrying-artifact-case-open.v3",
		"gooo.metric.language.proof-carrying-artifact-case-refuted.v3",
		"gooo.metric.language.proof-carrying-artifact-final-ledger-open.v3",
		"gooo.metric.language.proof-carrying-artifact-final-ledger-discharged.v3",
		"gooo.metric.language.proof-carrying-artifact-tampered-rejections.v3",
		"gooo.metric.language.proof-carrying-artifact-coherent-tamper-rejections.v3",
		"gooo.metric.language.proof-carrying-artifact-coherent-claim-structure-rejections.v3",
		"gooo.metric.language.proof-carrying-artifact-missing-evidence.v3",
		"gooo.metric.language.proof-carrying-artifact-byte-only-denials.v3",
		"gooo.metric.language.proof-carrying-artifact-recipe-rejections.v3",
		"gooo.metric.language.proof-carrying-artifact-recipe-only-rejections.v3",
		"gooo.metric.language.proof-carrying-artifact-missing-attachment.v3",
		"gooo.metric.language.proof-carrying-artifact-wrong-attachment.v3",
		"gooo.metric.language.proof-carrying-artifact-unrelated-tamper.v3",
		"gooo.metric.language.proof-carrying-artifact-stale-head.v3",
		"gooo.metric.language.proof-carrying-artifact-unauthorized-consumer.v3",
		"gooo.metric.language.proof-carrying-artifact-semantic-interventions.v3",
		"gooo.metric.language.proof-carrying-artifact-nonsemantic-interventions.v3",
		"gooo.metric.language.proof-carrying-artifact-read-only-authority.v3",
		"gooo.metric.language.proof-carrying-artifact-bundle-only.v3",
		"gooo.metric.language.proof-carrying-artifact-consumer-recheck.v3",
		"gooo.metric.language.proof-carrying-artifact-net-repository-state.v3",
		"gooo.metric.language.proof-carrying-artifact-producer-dependencies.guardrail.v3",
		"gooo.metric.language.proof-carrying-artifact-generated-authority.guardrail.v3",
		"gooo.metric.language.proof-carrying-artifact-semantic-claims.guardrail.v3",
		"gooo.metric.language.proof-carrying-artifact-writes.guardrail.v3",
		"gooo.metric.language.proof-carrying-artifact-mutation-authority.guardrail.v3",
		"gooo.metric.language.proof-carrying-artifact-promotion-authority.guardrail.v3",
		"gooo.metric.language.proof-carrying-artifact-semantic-authority.guardrail.v3",
		"gooo.metric.language.proof-carrying-artifact-core-parser-dependencies.guardrail.v3",
	}
}

// indicators is validator-owned: its claim-state targets come from the
// phase-indexed fixed contract. The producer uses observedIndicators instead
// and therefore never reads either expectation table.
func indicators(summary Summary, phase string) []Indicator {
	return indicatorRows(summary, fixedClaimStateTotals(phase))
}

func observedIndicators(summary Summary) []Indicator {
	return indicatorRows(summary, claimStateTotals{Discharged: summary.CaseDischargedClaims, Open: summary.CaseOpenClaims, Refuted: summary.CaseRefutedClaims})
}

func indicatorRows(summary Summary, expectedStates claimStateTotals) []Indicator {
	values := []struct {
		class, proof, operation string
		value, target           int
	}{
		{"OUTCOME", "COHERENCE", "evaluate-proof-carrying-cases", summary.CasesSatisfied, CaseTotal},
		{"DRIVER", "FOUNDATION", "declare-unique-claim-templates", summary.ClaimTemplates, ClaimTemplateTotal},
		{"DRIVER", "COHERENCE", "evaluate-claim-instances", summary.ClaimInstances, CaseTotal * ClaimTemplateTotal},
		{"OUTCOME", "FOUNDATION", "verify-valid-artifact", summary.ValidArtifacts, 1},
		{"DRIVER", "FOUNDATION", "carry-source-operation-invariant-evidence", summary.EvidenceKindsCarried, EvidenceTotal},
		{"OUTCOME", "COHERENCE", "recheck-evidence-links", summary.ExactEvidenceLinks, EvidenceTotal},
		{"OUTCOME", "FOUNDATION", "match-independent-recipe", summary.RecipeMatches, 1},
		{"DRIVER", "COHERENCE", "accept-digested-claim-transitions", summary.AcceptedTransitions, TransitionTotal},
		{"DRIVER", "COHERENCE", "preserve-transport-transitions", summary.PreservedTransitions, EvidenceTotal + 1},
		{"DRIVER", "COHERENCE", "count-case-discharged-claims", summary.CaseDischargedClaims, expectedStates.Discharged},
		{"GUARDRAIL", "FOUNDATION", "count-case-open-claims", summary.CaseOpenClaims, expectedStates.Open},
		{"GUARDRAIL", "REGRESSION", "count-case-refuted-claims", summary.CaseRefutedClaims, expectedStates.Refuted},
		{"DRIVER", "FOUNDATION", "persist-final-ledger-open-claims", summary.FinalLedgerOpenClaims, ClaimTemplateTotal},
		{"DRIVER", "COHERENCE", "persist-final-ledger-discharged-claims", summary.FinalLedgerDischargedClaims, ClaimTemplateTotal},
		{"GUARDRAIL", "REGRESSION", "reject-tampered-evidence", summary.TamperedRejections, 1},
		{"GUARDRAIL", "REGRESSION", "reject-coherent-tamper", summary.CoherentTamperRejections, 1},
		{"GUARDRAIL", "REGRESSION", "reject-coherent-claim-structure-tamper", summary.CoherentClaimStructureRejections, 4},
		{"GUARDRAIL", "FOUNDATION", "lower-missing-evidence", summary.MissingEvidenceRejections, 1},
		{"GUARDRAIL", "REGRESSION", "deny-byte-only-authority", summary.ByteOnlyDenials, 1},
		{"GUARDRAIL", "REGRESSION", "reject-recipe-drift", summary.RecipeRejections, 1},
		{"GUARDRAIL", "COHERENCE", "reject-recipe-only-drift", summary.RecipeOnlyRejections, 1},
		{"GUARDRAIL", "FOUNDATION", "lower-missing-attachment", summary.MissingAttachmentRejections, 1},
		{"GUARDRAIL", "COHERENCE", "reject-wrong-attachment", summary.WrongAttachmentRejections, 1},
		{"GUARDRAIL", "REGRESSION", "reject-unrelated-evidence-tamper", summary.UnrelatedEvidenceRejections, 1},
		{"GUARDRAIL", "FOUNDATION", "reject-stale-head", summary.StaleHeadRejections, 1},
		{"GUARDRAIL", "REGRESSION", "deny-unauthorized-consumer", summary.UnauthorizedConsumerDenials, 1},
		{"DRIVER", "COHERENCE", "observe-semantic-intervention", summary.SemanticInterventions, 1},
		{"DRIVER", "COHERENCE", "observe-comment-only-intervention", summary.NonsemanticInterventions, 1},
		{"OUTCOME", "COHERENCE", "scope-consumer-authority", summary.ReadOnlyAuthorities, 1},
		{"DRIVER", "COHERENCE", "recheck-bundle-only", summary.BundleOnlyVerification, 1},
		{"DRIVER", "COHERENCE", "consume-attested-target", summary.ConsumerRechecks, 1},
		{"GUARDRAIL", "FOUNDATION", "observe-net-repository-state", summary.NetRepositoryStateUnchanged, 1},
		{"GUARDRAIL", "FOUNDATION", "separate-verifier-from-producer", summary.ProducerDependencies, 0},
		{"GUARDRAIL", "REGRESSION", "deny-generated-authority", summary.GeneratedAuthority, 0},
		{"GUARDRAIL", "FOUNDATION", "bound-semantic-claims", summary.SemanticClaims, 0},
		{"GUARDRAIL", "REGRESSION", "deny-verifier-writes", summary.NetChangedPaths, 0},
		{"GUARDRAIL", "REGRESSION", "deny-mutation-authority", summary.MutationAuthorities, 0},
		{"GUARDRAIL", "REGRESSION", "deny-promotion-authority", summary.PromotionAuthorities, 0},
		{"GUARDRAIL", "REGRESSION", "deny-semantic-authority", summary.SemanticAuthorities, 0},
		{"GUARDRAIL", "FOUNDATION", "record-core-parser-dependencies", summary.CoreParserDependencies, CoreParserDependencyInventoryTotal},
	}
	result := make([]Indicator, len(values))
	ids := MetricIDs()
	for index, value := range values {
		result[index] = Indicator{MetricID: ids[index], Class: value.class, ProofChoice: value.proof,
			MetaOperation: value.operation, Value: value.value, Target: value.target, Satisfied: value.value == value.target}
	}
	return result
}

func proofs(report Report, cases []CaseResult, phase string) []Proof {
	valid := validCase(cases)
	if valid == nil {
		return []Proof{{Phase: phase, State: ProofStateOpen, Choice: "FOUNDATION", MetaOperation: "bind-source-bytes-and-projection", Passed: false}, {Phase: phase, State: ProofStateOpen, Choice: "COHERENCE", MetaOperation: "bind-operation-and-claim-relations", Passed: false}, {Phase: phase, State: ProofStateOpen, Choice: "REGRESSION", MetaOperation: "execute-negative-and-intervention-suite", Passed: false}}
	}
	allClaimsDischarged := len(valid.Claims) == ClaimTemplateTotal
	for _, claim := range valid.Claims {
		allClaimsDischarged = allClaimsDischarged && claim.Status == "DISCHARGED" && claim.Resolution == "EXACT" && claim.StateDigest == claimStateDigest(claim)
	}
	exactBinding := valid.SourceDigest == report.Checkout.SourceDigest && valid.OperationAttachmentDigest == report.Checkout.OperationDigest &&
		valid.RecipeAttachmentDigest == report.Checkout.RecipeDigest && report.ContractDigest == report.Checkout.ContractDigest
	receiptFieldsOK := report.BundleDigest != "" && (report.ConsumerReceipt.TargetPath == "artifact.json" && report.ConsumerReceipt.Authority == "READ_ONLY_CONSUMPTION" && report.ConsumerReceipt.OutputExists &&
		validDigest(report.ConsumerReceipt.TargetDigest) && validDigest(report.ConsumerReceipt.OutputDigest) && validDigest(report.ConsumerReceipt.AttestationDigest) &&
		report.ConsumerReceipt.AttestationDigest == attestationDigest(report) && report.ConsumerReceipt.Digest == consumerReceiptDigest(report.ConsumerReceipt))
	coherenceObserved := valid.ObservedDecision == "PASS" && allClaimsDischarged && exactBinding && len(report.Transitions) == TransitionTotal
	regressionObserved := negativeCaseInventoryOK(cases) && report.Summary.CasesSatisfied == CaseTotal && report.Summary.CoherentClaimStructureRejections == 4 && report.Summary.SemanticInterventions == 1 && report.Summary.NonsemanticInterventions == 1 && report.Summary.UnauthorizedConsumerDenials == 1
	coherenceEvidenceValidated := coherenceObserved && (phase == ProofPhasePreliminary || receiptFieldsOK)
	regressionEvidenceValidated := regressionObserved && (phase == ProofPhasePreliminary || receiptFieldsOK)
	coherencePassed := coherenceEvidenceValidated && phase == ProofPhaseFinal
	regressionPassed := regressionEvidenceValidated && phase == ProofPhaseFinal
	consumerGateOpen := phase == ProofPhasePreliminary || !receiptFieldsOK
	foundationEvidenceValidated := valid.ObservedDecision == "PASS" && allClaimsDischarged && exactBinding && valid.SourceDigest != "" && valid.SemanticDigest != ""
	foundationPassed := foundationEvidenceValidated && phase == ProofPhaseFinal
	proofState := func(evidenceValidated, passed, gateOpen bool) string {
		if phase == ProofPhasePreliminary {
			if evidenceValidated && !gateOpen {
				return ProofStateObserved
			}
			return ProofStateOpen
		}
		if evidenceValidated && passed && !gateOpen {
			return ProofStateDischarged
		}
		return ProofStateOpen
	}
	evidence := []string{}
	for _, claim := range valid.Claims {
		evidence = append(evidence, claim.EvidenceDigests...)
	}
	return []Proof{
		{Phase: phase, State: proofState(foundationEvidenceValidated, foundationPassed, false), EvidenceValidated: foundationEvidenceValidated, Choice: "FOUNDATION", MetaOperation: "bind-source-bytes-and-projection", TargetDigest: valid.SourceDigest, Dependency: "checkout.source.gooo->source-bytes-bound", EvidenceDigests: append([]string(nil), valid.Claims[0].EvidenceDigests...), ReceiptDigest: valid.OperationDigest, Passed: foundationPassed},
		{Phase: phase, State: proofState(coherenceEvidenceValidated, coherencePassed, consumerGateOpen), EvidenceValidated: coherenceEvidenceValidated, Choice: "COHERENCE", MetaOperation: "bind-operation-and-claim-relations", TargetDigest: valid.OperationDigest, Dependency: "source-bytes-bound->operation-receipt-bound->recipe-match->consumer-authority", EvidenceDigests: evidence, ReceiptDigest: valid.OperationDigest, Passed: coherencePassed, ConsumerGateOpen: consumerGateOpen},
		{Phase: phase, State: proofState(regressionEvidenceValidated, regressionPassed, consumerGateOpen), EvidenceValidated: regressionEvidenceValidated, Choice: "REGRESSION", MetaOperation: "execute-negative-and-intervention-suite", TargetDigest: report.WriteSet.Digest, Dependency: "negative-case-inventory->claim-local-effects->authority-boundary", EvidenceDigests: regressionEvidence(cases, report.Interventions), ReceiptDigest: valid.OperationDigest, Passed: regressionPassed, ConsumerGateOpen: consumerGateOpen},
	}
}

func proofSummary(proofs []Proof, phase, authority string) ProofSummary {
	observedEvidence := 0
	openProofs := 0
	dischargedProofs := 0
	for _, proof := range proofs {
		if proof.EvidenceValidated {
			observedEvidence++
		}
		if proof.State == ProofStateOpen {
			openProofs++
		}
		if proof.State == ProofStateDischarged {
			dischargedProofs++
		}
	}
	authorityCount := 0
	if authority == "READ_ONLY_CONSUMPTION" {
		authorityCount = 1
	}
	return ProofSummary{Phase: phase, Proofs: ProofTotal, EvidenceValidated: observedEvidence, EvidenceValidatedTotal: ProofTotal, ObservedState: observedStateCount(proofs), ObservedStateTotal: ProofTotal, Open: openProofs, OpenTotal: ProofTotal, Discharged: dischargedProofs, DischargedTotal: ProofTotal, Authority: authorityCount, AuthorityTotal: 1}
}

func observedStateCount(proofs []Proof) int {
	count := 0
	for _, proof := range proofs {
		if proof.State == ProofStateObserved {
			count++
		}
	}
	return count
}

func regressionEvidence(cases []CaseResult, interventions []InterventionResult) []string {
	result := make([]string, 0, len(cases)+len(interventions))
	for _, item := range cases {
		if item.ID != "valid-proof-carrying-artifact" && item.ArtifactDigest != "" {
			result = append(result, item.ArtifactDigest)
		}
	}
	for _, item := range interventions {
		if item.OperationReceiptDigestAfter != "" {
			result = append(result, item.OperationReceiptDigestAfter)
		}
	}
	return result
}
