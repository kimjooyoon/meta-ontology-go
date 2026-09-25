package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/compatibilitypolicy"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/compilercompatibility"
)

func runConformance(contractPath, predecessorPath, successorPath, authorizationPath, certificatePath, consumptionPath, publicationRoot, outputPath string) error {
	if contractPath == "" || predecessorPath == "" || successorPath == "" || authorizationPath == "" || certificatePath == "" || consumptionPath == "" || publicationRoot == "" || outputPath == "" {
		return errors.New("run requires contract, predecessor, successor, authorization, certificate, consumption-report, publication-root, and output")
	}
	started := time.Now()
	policy, err := loadPolicy(contractPath)
	if err != nil {
		return err
	}
	contractData, err := os.ReadFile(contractPath)
	if err != nil {
		return err
	}
	certificateData, certificate, err := readCertificate(certificatePath)
	if err != nil {
		return err
	}
	var predecessor, successor compilercompatibility.ExecutionReceipt
	predecessorData, err := readStrict(predecessorPath, &predecessor)
	if err != nil {
		return err
	}
	successorData, err := readStrict(successorPath, &successor)
	if err != nil {
		return err
	}
	var authorization compilercompatibility.Authorization
	authorizationData, err := readStrict(authorizationPath, &authorization)
	if err != nil {
		return err
	}
	if err := compilercompatibility.ValidateAuthorization(authorization); err != nil {
		return err
	}
	if !bytes.Equal(predecessorData, mustJSON(predecessor)) || !bytes.Equal(successorData, mustJSON(successor)) {
		return errors.New("execution receipt replay bytes are not canonical")
	}
	consumption, err := readConsumption(consumptionPath)
	if err != nil {
		return err
	}
	if consumption.Schema != compilercompatibility.ReportSchema || consumption.Decision != compilercompatibility.DecisionClosed ||
		consumption.Reason != compilercompatibility.ReasonBoundedSuccessorReplay || consumption.CompatibilityHits != 1 ||
		consumption.CompatibilityMisses != 0 || consumption.ImplementationDigestDifferent != true ||
		consumption.ImplementationDigestEqual || consumption.IdentityAxisCount != compatibilitypolicy.AxisCount ||
		consumption.SemanticEqual != true || consumption.GeneratedBytesEqual != true || consumption.GeneratedManifestEqual != true ||
		consumption.PolicyResultEqual != true || consumption.FullTestContractEqual != true ||
		consumption.IndependentReplayExecutions != 2 || consumption.TestContractReplays != 2 ||
		consumption.CompatibilityScopeSubjects != 1 || consumption.CertificateCount != 1 ||
		consumption.CertificateBytes != len(certificateData) ||
		consumption.CertificateDigest != cache.HashBytes(certificateData).String() ||
		consumption.EvidenceArtifactCount != compatibilitypolicy.EvidenceArtifactCount ||
		consumption.ContinuityEdgeCount != compatibilitypolicy.ContinuityEdgeCount ||
		consumption.UnsupportedFrontierDecision != compilercompatibility.UnsupportedFrontierDecision ||
		!sameStringSlice(consumption.UnsupportedFrontierClaims, compilercompatibility.UnsupportedFrontierClaims) ||
		consumption.RepositoryWrites != 0 || consumption.LocalTestExecutions != 0 {
		return errors.New("actual ordinary generate did not consume the bounded compatibility certificate")
	}
	baseRequest := compilercompatibility.Request{Mode: "OPT_IN", CandidateStableID: certificate.CandidateStableID,
		SubjectDigest: certificate.SubjectDigest, SourceDigest: certificate.SourceDigest, Current: successor, Certificate: &certificate}
	strictReplay := compilercompatibility.EvaluateStrict(predecessor, predecessor)
	boundedReplay := compilercompatibility.EvaluateOptIn(policy, baseRequest)
	missingCertificate := cloneCertificate(certificate)
	missingCertificate.Successor = compilercompatibility.ExecutionReceipt{}
	missingCertificate.SuccessorAxes = compilercompatibility.IdentityAxes{}
	missingCertificate.SuccessorReceiptDigest = ""
	missingCertificate.CertificateID = reseal(missingCertificate)
	missingReplay := compilercompatibility.EvaluateOptIn(policy, baseRequestWith(missingCertificate, baseRequest))
	unboundedScope := cloneCertificate(certificate)
	unboundedScope.ScopeBounded = false
	unboundedScope.Scope = nil
	unboundedScope.CertificateID = reseal(unboundedScope)
	unbounded := compilercompatibility.EvaluateOptIn(policy, baseRequestWith(unboundedScope, baseRequest))
	semanticMismatch := cloneCertificate(certificate)
	semanticMismatch.Successor.SemanticIRDigest = strings.Repeat("a", 64)
	semanticMismatch.SuccessorAxes = semanticMismatch.Successor.IdentityAxes()
	semanticMismatch.CertificateID = reseal(semanticMismatch)
	semantic := compilercompatibility.EvaluateOptIn(policy, baseRequestWith(semanticMismatch, baseRequest))
	tampered := cloneCertificate(certificate)
	tampered.CertificateID = strings.Repeat("b", 64)
	tamperedEval := compilercompatibility.EvaluateOptIn(policy, baseRequestWith(tampered, baseRequest))
	widened := cloneCertificate(certificate)
	widened.Scope = append(append([]compilercompatibility.ScopeSubject(nil), widened.Scope...), compilercompatibility.ScopeSubject{CandidateStableID: "widened", SubjectDigest: strings.Repeat("c", 64)})
	widened.CertificateID = reseal(widened)
	widenedEval := compilercompatibility.EvaluateOptIn(policy, baseRequestWith(widened, baseRequest))

	cases := []caseReport{
		caseResult(policy, compatibilitypolicy.CaseStrictExactReplay, strictReplay),
		caseResult(policy, compatibilitypolicy.CaseBoundedImplementationSuccessor, boundedReplay),
		caseResult(policy, compatibilitypolicy.CaseMissingSuccessorReplay, missingReplay),
		caseResult(policy, compatibilitypolicy.CaseUnboundedCompatibilityScope, unbounded),
		caseResult(policy, compatibilitypolicy.CaseSemanticPolicyOutputMismatch, semantic),
		caseResult(policy, compatibilitypolicy.CaseTamperedWidenedCertificate, tamperedEval),
	}
	for _, item := range cases {
		if item.ExpectedDecision != item.ObservedDecision {
			return fmt.Errorf("case %s = %s, want %s", item.ID, item.ObservedDecision, item.ExpectedDecision)
		}
		if item.ObservedDecision == compilercompatibility.DecisionUnknown && !completeUnknown(item.Unknown) {
			return fmt.Errorf("case %s has incomplete UNKNOWN causality", item.ID)
		}
		if item.ObservedDecision == compilercompatibility.DecisionRefuted && item.Unknown != nil {
			return fmt.Errorf("case %s carries UNKNOWN evidence on REFUTED", item.ID)
		}
	}
	if tamperedEval.Decision != compilercompatibility.DecisionRefuted || widenedEval.Decision != compilercompatibility.DecisionRefuted {
		return errors.New("tampered or widened certificates were not rejected")
	}
	axisProbes, err := runAxisMismatchProbes(policy, baseRequest, certificate)
	if err != nil {
		return err
	}
	fallback := baseRequest
	fallback.Current.SourceDigest = strings.Repeat("d", 64)
	fallbackEval := compilercompatibility.EvaluateOptIn(policy, fallback)
	if fallbackEval.Decision != compilercompatibility.DecisionRefuted || !fallbackEval.FallbackRejected {
		return errors.New("current-dev substitution was not rejected")
	}
	if !strictPredecessorMismatch(predecessor, successor) {
		return errors.New("strict predecessor consumption did not remain rejected for implementation mismatch")
	}
	counts := map[string]int{compilercompatibility.DecisionClosed: 0, compilercompatibility.DecisionUnknown: 0, compilercompatibility.DecisionRefuted: 0}
	for _, item := range cases {
		counts[item.ObservedDecision]++
	}
	if counts[compilercompatibility.DecisionClosed] != 2 || counts[compilercompatibility.DecisionUnknown] != 2 || counts[compilercompatibility.DecisionRefuted] != 2 {
		return errors.New("compatibility six-case denominator is not 2/2/2")
	}
	comparisons := compilercompatibility.CompareAxes(certificate.PredecessorAxes, certificate.SuccessorAxes)
	report := conformanceReport{
		Schema: "gooo/compiler-successor-compatibility-conformance/v1", Decision: compilercompatibility.DecisionClosed,
		Reason: "BOUNDED_IMPLEMENTATION_SUCCESSOR_REPLAY", PolicySourceDigest: policy.SourceDigest, PolicyEvaluatorDigest: policy.EvaluatorDigest,
		CaseDenominator: len(cases), CaseIDs: compatibilitypolicy.CaseIDs(), Cases: cases,
		ClosedCases: counts[compilercompatibility.DecisionClosed], UnknownCases: counts[compilercompatibility.DecisionUnknown], RefutedCases: counts[compilercompatibility.DecisionRefuted],
		IdentityAxisCount: len(comparisons), StrictPredecessorConsumption: certificate.StrictPredecessorConsumption,
		ImplementationDigestEqual: comparisons[1].Equal, ImplementationDigestDifferent: !comparisons[1].Equal,
		SemanticEqual: comparisons[0].Equal, GeneratedBytesEqual: certificate.GeneratedBytesEqual, GeneratedManifestEqual: certificate.GeneratedManifestEqual,
		PolicyResultEqual: certificate.PolicyResultEqual, FullTestContractEqual: certificate.FullTestContractEqual,
		IndependentReplayExecutions: certificate.IndependentReplayExecutions, TestContractReplays: 2,
		CompatibilityScopeSubjects: len(certificate.Scope), CertificateCount: 1, CertificateBytes: len(certificateData), CompatibilityHits: 1, CompatibilityMisses: 0,
		MismatchDetections: axisProbes, CertificateTamperDetections: 1, ScopeWideningDetections: 1,
		FallbackAttempts: 1, FallbackAccepted: 0, FallbackRejected: 1,
		EvidenceArtifactCount: policy.EvidenceArtifacts, ContinuityEdgeCount: policy.ContinuityEdges,
		Claim: "BOUNDED_IMPLEMENTATION_REUSE_ELIGIBILITY", PerformanceClaim: false, GeneralCompatibilityClaim: false,
		UnsupportedFrontierDecision: compilercompatibility.UnsupportedFrontierDecision, UnsupportedFrontierClaims: append([]string(nil), compilercompatibility.UnsupportedFrontierClaims...),
		RepositoryWrites: 0, LocalTestExecutions: 0, WallMS: compatibilityWallMS(started), PeakRSSKib: compatibilityPeakRSS(),
		EvidenceArtifactNames: append([]string(nil), publicationNames...),
	}
	if !comparisonsEqualExceptImplementation(comparisons) || !certificate.FullTestContractEqual || !certificate.PolicyResultEqual {
		return errors.New("bounded compatibility proof is broader than the implementation axis")
	}
	if err := writePublication(publicationRoot, contractData, policy, predecessor, successor, authorization, certificateData, certificate, consumption, cases, comparisons, report, predecessorData, successorData, authorizationData); err != nil {
		return err
	}
	return writeJSON(outputPath, report)
}

func caseResult(policy compatibilitypolicy.Policy, id string, evaluation compilercompatibility.Evaluation) caseReport {
	expected, _ := compilercompatibility.FixedPolicyCaseDecision(policy, id)
	return caseReport{ID: id, ExpectedDecision: expected, ObservedDecision: evaluation.Decision, Reason: evaluation.Reason,
		Unknown: evaluation.Unknown, AxisComparisons: evaluation.Axes, MismatchDetected: evaluation.MismatchDetected, FallbackRejected: evaluation.FallbackRejected}
}

func baseRequestWith(certificate *compilercompatibility.Certificate, base compilercompatibility.Request) compilercompatibility.Request {
	request := base
	request.Certificate = certificate
	return request
}

func cloneCertificate(certificate compilercompatibility.Certificate) *compilercompatibility.Certificate {
	data, _ := json.Marshal(certificate)
	var clone compilercompatibility.Certificate
	_ = json.Unmarshal(data, &clone)
	return &clone
}

func reseal(certificate *compilercompatibility.Certificate) string {
	certificate.CertificateID = ""
	digest, _ := certificate.ContentDigest()
	return digest
}

func strictPredecessorMismatch(predecessor, successor compilercompatibility.ExecutionReceipt) bool {
	return compilercompatibility.EvaluateStrict(predecessor, successor).Decision == compilercompatibility.DecisionRefuted
}

func runAxisMismatchProbes(policy compatibilitypolicy.Policy, request compilercompatibility.Request, certificate compilercompatibility.Certificate) (int, error) {
	mutators := []func(*compilercompatibility.Certificate){
		func(c *compilercompatibility.Certificate) {
			c.Successor.SemanticIRDigest = strings.Repeat("e", 64)
			c.SuccessorAxes = c.Successor.IdentityAxes()
		},
		func(c *compilercompatibility.Certificate) {
			c.Successor.GoToolchainDigest = strings.Repeat("f", 64)
			c.SuccessorAxes = c.Successor.IdentityAxes()
		},
		func(c *compilercompatibility.Certificate) {
			c.Successor.PolicyDigest = strings.Repeat("1", 64)
			c.SuccessorAxes = c.Successor.IdentityAxes()
		},
		func(c *compilercompatibility.Certificate) {
			c.Successor.GeneratedSource = append(c.Successor.GeneratedSource, 'x')
		},
		func(c *compilercompatibility.Certificate) {
			c.Successor.TestContractDigest = strings.Repeat("2", 64)
			c.SuccessorAxes = c.Successor.IdentityAxes()
		},
		func(c *compilercompatibility.Certificate) {
			c.Successor.AuthorizationDigest = strings.Repeat("3", 64)
			c.SuccessorAxes = c.Successor.IdentityAxes()
		},
	}
	detections := 0
	for _, mutate := range mutators {
		mutated := cloneCertificate(certificate)
		mutate(mutated)
		mutated.CertificateID = reseal(mutated)
		evaluation := compilercompatibility.EvaluateOptIn(policy, baseRequestWith(mutated, request))
		if evaluation.Decision != compilercompatibility.DecisionRefuted {
			return detections, errors.New("an identity-axis mismatch was inferred compatible")
		}
		detections++
	}
	return detections, nil
}

func completeUnknown(unknown *compilercompatibility.UnknownState) bool {
	return unknown != nil && unknown.Stage != "" && unknown.Step != "" && unknown.Reason != "" && unknown.UnknownClass != "" && unknown.NextOperation != "" && len(unknown.BlockedBy) > 0
}

func comparisonsEqualExceptImplementation(comparisons []compilercompatibility.AxisComparison) bool {
	if len(comparisons) != compatibilitypolicy.AxisCount {
		return false
	}
	for index, comparison := range comparisons {
		if index != 1 && !comparison.Equal {
			return false
		}
	}
	return !comparisons[1].Equal
}

func mustJSON(value any) []byte {
	data, _ := json.MarshalIndent(value, "", "  ")
	return append(data, '\n')
}

func compatibilityPeakRSS() int64 {
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)
	value := int64(stats.Sys / 1024)
	if value < 1 {
		return 1
	}
	return value
}

func compatibilityWallMS(started time.Time) int64 {
	return int64((time.Since(started).Nanoseconds() + int64(time.Millisecond) - 1) / int64(time.Millisecond))
}

func writePublication(root string, contractData []byte, policy compatibilitypolicy.Policy, predecessor, successor compilercompatibility.ExecutionReceipt, authorization compilercompatibility.Authorization, certificateData []byte, certificate compilercompatibility.Certificate, consumption compilercompatibility.ConsumptionReport, cases []caseReport, comparisons []compilercompatibility.AxisComparison, report conformanceReport, predecessorData, successorData, authorizationData []byte) error {
	if err := os.MkdirAll(root, 0o755); err != nil {
		return err
	}
	if entries, err := os.ReadDir(root); err != nil {
		return err
	} else if len(entries) != 0 {
		return errors.New("publication root must be empty")
	}
	policyData, _ := json.MarshalIndent(map[string]any{"source_digest": policy.SourceDigest, "evaluator_digest": policy.EvaluatorDigest, "mode": policy.Mode, "opt_in": policy.OptIn, "axes": policy.Axes, "transitions": policy.Transitions, "cases": policy.Cases, "test_contract": policy.TestContract, "continuity_edges": policy.ContinuityEdges, "evidence_artifacts": policy.EvidenceArtifacts}, "", "  ")
	transitionData, _ := json.MarshalIndent(policy.Transitions, "", "  ")
	caseData, _ := json.MarshalIndent(cases, "", "  ")
	axisData, _ := json.MarshalIndent(comparisons, "", "  ")
	replayData, _ := json.MarshalIndent(map[string]any{"predecessor": predecessor, "successor": successor, "independent_replay_executions": 2, "generated_bytes_equal": certificate.GeneratedBytesEqual, "generated_manifest_equal": certificate.GeneratedManifestEqual, "normalized_semantic_equal": certificate.NormalizedSemanticEqual, "policy_result_equal": certificate.PolicyResultEqual, "full_test_contract_equal": certificate.FullTestContractEqual}, "", "  ")
	scopeData, _ := json.MarshalIndent(certificate.Scope, "", "  ")
	edgesData, _ := json.MarshalIndent(map[string]any{"count": report.ContinuityEdgeCount, "edges": []string{"predecessor->successor", "predecessor->strict-consumption", "successor->certificate", "certificate->authorization", "certificate->consumer", "consumer->generated-bytes", "consumer->generated-manifest"}}, "", "  ")
	metricsData, _ := json.MarshalIndent(report, "", "  ")
	claimData, _ := json.MarshalIndent(map[string]any{"claim": report.Claim, "performance_claim": false, "general_compiler_compatibility_claim": false, "unsupported_frontier_decision": report.UnsupportedFrontierDecision, "unsupported_frontier_claims": report.UnsupportedFrontierClaims}, "", "  ")
	identityData, _ := json.MarshalIndent(map[string]any{"predecessor": predecessor.IdentityAxes(), "successor": successor.IdentityAxes(), "implementation_digest_equal": report.ImplementationDigestEqual, "implementation_digest_different": report.ImplementationDigestDifferent}, "", "  ")
	actualData, _ := json.MarshalIndent(consumption, "", "  ")
	strictData, _ := json.MarshalIndent(cases[0], "", "  ")
	boundedData, _ := json.MarshalIndent(cases[1], "", "  ")
	missingData, _ := json.MarshalIndent(cases[2], "", "  ")
	unboundedData, _ := json.MarshalIndent(cases[3], "", "  ")
	semanticData, _ := json.MarshalIndent(cases[4], "", "  ")
	tamperedData, _ := json.MarshalIndent(cases[5], "", "  ")
	values := map[string][]byte{
		"canonical-policy.gooo":                        contractData,
		"compatibility-policy.json":                    append(policyData, '\n'),
		"transition-table.json":                        append(transitionData, '\n'),
		"case-table.json":                              append(caseData, '\n'),
		"predecessor-execution-receipt.json":           predecessorData,
		"successor-execution-receipt.json":             successorData,
		"compatibility-authorization.json":             authorizationData,
		"compatibility-certificate.json":               certificateData,
		"compatibility-certificate-digest.txt":         []byte(cache.HashBytes(certificateData).String() + "\n"),
		"actual-compatibility-consumption.json":        append(actualData, '\n'),
		"strict-replay.json":                           append(strictData, '\n'),
		"bounded-implementation-successor-replay.json": append(boundedData, '\n'),
		"missing-successor-replay.json":                append(missingData, '\n'),
		"unbounded-compatibility-scope.json":           append(unboundedData, '\n'),
		"semantic-policy-output-mismatch.json":         append(semanticData, '\n'),
		"tampered-widened-certificate.json":            append(tamperedData, '\n'),
		"axis-comparisons.json":                        append(axisData, '\n'),
		"replay-evidence.json":                         append(replayData, '\n'),
		"compatibility-scope.json":                     append(scopeData, '\n'),
		"continuity-edges.json":                        append(edgesData, '\n'),
		"compatibility-metrics.json":                   append(metricsData, '\n'),
		"compatibility-claim.json":                     append(claimData, '\n'),
		"compiler-identity.json":                       append(identityData, '\n'),
		"compatibility-report.json":                    mustJSON(report),
	}
	_ = certificate
	for name, data := range values {
		if err := os.WriteFile(filepath.Join(root, name), data, 0o644); err != nil {
			return fmt.Errorf("write publication %s: %w", name, err)
		}
	}
	if len(values) != len(publicationNames) {
		return errors.New("publication artifact denominator changed")
	}
	return nil
}
