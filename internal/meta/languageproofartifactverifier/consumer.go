package languageproofartifactverifier

import (
	"encoding/base64"
	"errors"
	"fmt"
	"reflect"
)

const ConsumerReceiptSchema = "gooo/read-only-consumption-receipt/v2"

func DecodeConsumerReceipt(raw []byte) (ConsumerReceipt, error) {
	return decodeStrict[ConsumerReceipt](raw)
}

// ConsumerReceiptEmpty is the zero check used at the CLI and verifier phase
// boundary. ConsumerReceipt carries an observed issue slice, so comparing the
// struct directly is not a valid Go operation.
func ConsumerReceiptEmpty(receipt ConsumerReceipt) bool {
	return reflect.DeepEqual(receipt, ConsumerReceipt{})
}

type ConsumerErrorClass string

const (
	ConsumerErrorBundleInvalid       ConsumerErrorClass = "BUNDLE_INVALID"
	ConsumerErrorAttestationMismatch ConsumerErrorClass = "ATTESTATION_MISMATCH"
	ConsumerErrorTargetMissing       ConsumerErrorClass = "TARGET_MISSING"
	ConsumerErrorTargetPolicy        ConsumerErrorClass = "TARGET_POLICY"
	ConsumerErrorReceiptMismatch     ConsumerErrorClass = "RECEIPT_MISMATCH"
)

// ConsumerError preserves the failed consumer boundary. In particular, a
// malformed bundle or absent target is not an authorization failure.
type ConsumerError struct {
	Class  ConsumerErrorClass
	Detail string
}

func (e *ConsumerError) Error() string { return string(e.Class) + ": " + e.Detail }

func (e *ConsumerError) Digest() string {
	return digestValue(struct {
		Class  ConsumerErrorClass `json:"class"`
		Detail string             `json:"detail"`
	}{e.Class, e.Detail})
}

func consumerError(class ConsumerErrorClass, detail string) *ConsumerError {
	return &ConsumerError{Class: class, Detail: detail}
}

type consumerAttestation struct {
	Decision                    string   `json:"decision"`
	Resolution                  string   `json:"resolution"`
	Reason                      string   `json:"reason"`
	Authority                   string   `json:"authority"`
	SubjectDecision             string   `json:"subject_decision"`
	BundleDigest                string   `json:"bundle_digest"`
	PreliminaryDigest           string   `json:"preliminary_digest"`
	Producer                    string   `json:"producer"`
	Consumer                    string   `json:"consumer"`
	TargetPath                  string   `json:"target_path"`
	TargetDigest                string   `json:"target_digest"`
	PolicyRawSourceDigest       string   `json:"policy_raw_source_digest"`
	PolicySemanticDigest        string   `json:"policy_semantic_digest"`
	PolicyUniqueIssueRows       int      `json:"policy_unique_issue_rows"`
	PolicyUniqueRankRows        int      `json:"policy_unique_rank_rows"`
	PolicyRowTotal              int      `json:"policy_row_total"`
	PolicySelectionOperation    string   `json:"policy_selection_operation"`
	PolicyObservedIssueSet      []string `json:"policy_observed_issue_set"`
	PolicySelectedIssue         string   `json:"policy_selected_issue"`
	PolicySelectedRank          int      `json:"policy_selected_rank"`
	PolicyMembershipDigest      string   `json:"policy_membership_digest"`
	PolicyObservedIssueCount    int      `json:"policy_observed_issue_count"`
	PolicyClaimTransitionDigest string   `json:"policy_claim_transition_digest"`
	CaseEnvelopeDigest          string   `json:"case_envelope_digest"`
}

func attestationDigest(report Report) string {
	return attestationDigestFor(report, report.ConsumerReceipt)
}

func attestationDigestFor(report Report, receipt ConsumerReceipt) string {
	return digestValue(consumerAttestation{Decision: report.ConformanceDecision, Resolution: report.ConformanceResolution, Reason: report.ConformanceReason,
		Authority: report.ArtifactUseAuthority, SubjectDecision: report.SubjectArtifactDecision, BundleDigest: report.BundleDigest,
		PreliminaryDigest: receipt.PreliminaryDigest, Producer: report.Producer, Consumer: report.Consumer, TargetPath: receipt.TargetPath, TargetDigest: receipt.TargetDigest,
		PolicyRawSourceDigest: receipt.PolicyRawSourceDigest, PolicySemanticDigest: receipt.PolicySemanticDigest, PolicyUniqueIssueRows: receipt.PolicyUniqueIssueRows,
		PolicyUniqueRankRows: receipt.PolicyUniqueRankRows, PolicyRowTotal: receipt.PolicyRowTotal, PolicySelectionOperation: receipt.PolicySelectionOperation, PolicyObservedIssueSet: receipt.PolicyObservedIssueSet,
		PolicySelectedIssue: receipt.PolicySelectedIssue, PolicySelectedRank: receipt.PolicySelectedRank, PolicyMembershipDigest: receipt.PolicyMembershipDigest,
		PolicyObservedIssueCount: receipt.PolicyObservedIssueCount, PolicyClaimTransitionDigest: receipt.PolicyClaimTransitionDigest, CaseEnvelopeDigest: receipt.CaseEnvelopeDigest})
}

func consumerAttestedReport(report Report) Report {
	report.ConformanceDecision = "PASS"
	report.ConformanceResolution = "EXACT"
	report.ConformanceReason = "PROOF_CARRYING_ARTIFACT_CONTRACT_SATISFIED"
	report.ConformanceCoordinate = Coordinate{"CONSUME_AUTHORITY", "grant-read-only-consumption", report.ConformanceReason}
	report.ArtifactUseAuthority = "READ_ONLY_CONSUMPTION"
	return report
}

func expectedConsumerReceipt(report Report, preliminaryDigest, targetPath string, target []byte) ConsumerReceipt {
	targetDigest := digestBytes(target)
	outputDigest := digestValue(struct {
		TargetPath   string `json:"target_path"`
		TargetDigest string `json:"target_digest"`
		Authority    string `json:"authority"`
	}{targetPath, targetDigest, "READ_ONLY_CONSUMPTION"})
	policy := CaseEnvelopePolicyObservation{}
	caseDigest := ""
	if valid := validCase(report.Cases); valid != nil {
		policy = valid.Policy
		caseDigest = valid.EnvelopeDigest
	}
	receipt := ConsumerReceipt{Schema: ConsumerReceiptSchema, Version: 2, PreliminaryDigest: preliminaryDigest, Producer: report.Producer, Consumer: report.Consumer,
		TargetPath: targetPath, TargetDigest: targetDigest, OutputDigest: outputDigest, OutputExists: true, Authority: "READ_ONLY_CONSUMPTION",
		PolicyRawSourceDigest: policy.RawSourceDigest, PolicySemanticDigest: policy.SemanticDigest, PolicyUniqueIssueRows: policy.UniqueIssueRows, PolicyUniqueRankRows: policy.UniqueRankRows, PolicyRowTotal: CaseEnvelopePolicyRowTotal,
		PolicySelectionOperation: policy.SelectionOperation, PolicyObservedIssueSet: append([]string(nil), policy.ObservedIssueSet...), PolicySelectedIssue: policy.SelectedIssue,
		PolicySelectedRank: policy.SelectedRank, PolicyMembershipDigest: policy.ObservedIssueMembershipDigest, PolicyObservedIssueCount: policy.ObservedIssueCount}
	if valid := validCase(report.Cases); valid != nil {
		receipt.PolicyClaimTransitionDigest = policyClaimTransitionDigest(valid.Claims)
	}
	receipt.AttestationDigest = attestationDigestFor(report, receipt)
	receipt.Digest = consumerReceiptDigest(receipt)
	return receipt
}

func consumerReceiptDigest(receipt ConsumerReceipt) string {
	receipt.Digest = ""
	return digestValue(receipt)
}

// consumerReceiptOK is an independent consumer-boundary decision used while
// Evaluate is assembling a report. It reconstructs the target bytes from the
// validated bundle, derives the canonical preliminary projection, and compares
// the complete receipt reconstructed by this verifier. It does not use the
// producer's conformance decision as an authority input.
func consumerReceiptOK(report Report, bundle Bundle) bool {
	if ValidateBundle(bundle) != nil || report.BundleDigest == "" || report.BundleDigest != bundle.Digest {
		return false
	}
	receipt := report.ConsumerReceipt
	if receipt.Schema != ConsumerReceiptSchema || receipt.Version != 2 || receipt.Producer != report.Producer || receipt.Consumer != report.Consumer ||
		receipt.TargetPath != "artifact.json" || !receipt.OutputExists || receipt.Authority != "READ_ONLY_CONSUMPTION" ||
		!validDigest(receipt.PreliminaryDigest) || !validDigest(receipt.TargetDigest) || !validDigest(receipt.OutputDigest) ||
		!validDigest(receipt.AttestationDigest) || !validDigest(receipt.Digest) {
		return false
	}
	preliminary := canonicalPreliminaryProjection(report)
	if receipt.PreliminaryDigest != preliminary.Digest {
		return false
	}
	target, err := bundleTargetBytes(bundle, receipt.TargetPath)
	if err != nil || digestBytes(target) != receipt.TargetDigest {
		return false
	}
	attested := consumerAttestedReport(report)
	expected := expectedConsumerReceipt(attested, preliminary.Digest, receipt.TargetPath, target)
	return reflect.DeepEqual(expected, receipt)
}

// ConsumeBundle is the only downstream consumption API. It accepts a portable
// bundle plus a verifier report; it does not accept raw artifact bytes. The
// report is checked first, then the target is reconstructed from its bundle
// content-addressed entry and the actual consumed digest is recorded.
func ConsumeBundle(bundle Bundle, report Report, targetPath string) (ConsumerReceipt, error) {
	if err := ValidateBundle(bundle); err != nil {
		return ConsumerReceipt{}, consumerError(ConsumerErrorBundleInvalid, err.Error())
	}
	if report.ConformanceDecision == "PASS" && report.ArtifactUseAuthority == "READ_ONLY_CONSUMPTION" && !report.ConsumerReceipt.OutputExists {
		return ConsumerReceipt{}, consumerError(ConsumerErrorReceiptMismatch, "consumer receipt declares output_exists=false")
	}
	if targetPath == "" || targetPath == "artifact.json" {
		targetPath = "artifact.json"
	} else {
		if _, err := bundleTargetBytes(bundle, targetPath); err != nil {
			return ConsumerReceipt{}, consumerError(ConsumerErrorTargetMissing, err.Error())
		}
		return ConsumerReceipt{}, consumerError(ConsumerErrorTargetPolicy, fmt.Sprintf("consumer target is outside read-only policy: %s", targetPath))
	}
	raw, err := bundleTargetBytes(bundle, targetPath)
	if err != nil {
		return ConsumerReceipt{}, consumerError(ConsumerErrorTargetMissing, err.Error())
	}
	finalErr := Validate(report)
	finalReport := finalErr == nil
	preliminaryErr := error(nil)
	if !finalReport {
		preliminaryErr = ValidatePreliminary(report)
	}
	if !finalReport && preliminaryErr != nil {
		// A report that declares final authority or carries a receipt was
		// intended as FINAL. Preserve its final validation coordinate instead
		// of hiding a semantic final-state failure behind the preliminary
		// authority-present error.
		if !ConsumerReceiptEmpty(report.ConsumerReceipt) || report.ConformanceDecision == "PASS" || report.ArtifactUseAuthority != "" {
			return ConsumerReceipt{}, consumerError(ConsumerErrorAttestationMismatch, finalErr.Error())
		}
		var typed *ValidationError
		if errors.As(preliminaryErr, &typed) {
			return ConsumerReceipt{}, consumerError(ConsumerErrorAttestationMismatch, typed.Error())
		}
		return ConsumerReceipt{}, consumerError(ConsumerErrorAttestationMismatch, "consumer attestation is not independently verified: "+preliminaryErr.Error())
	}
	if report.BundleDigest != bundle.Digest {
		return ConsumerReceipt{}, consumerError(ConsumerErrorAttestationMismatch, "consumer attestation is not independently verified")
	}
	attested := consumerAttestedReport(report)
	if !finalReport {
		// Re-run the case oracle in FINAL mode from the bundle's content. This
		// is what records the declared evidence-time transition; copying the
		// preliminary cases would erase OPEN -> REFUTED semantics.
		replayInput, replayErr := InputFromBundle(bundle)
		if replayErr != nil {
			return ConsumerReceipt{}, consumerError(ConsumerErrorBundleInvalid, replayErr.Error())
		}
		replayInput.ConsumerReceiptProvided = true
		replayInput.ConsumerReceipt = ConsumerReceipt{}
		replayed := Evaluate(replayInput)
		attested = consumerAttestedReport(replayed)
		// A preliminary report is accepted only when it can be lifted to the
		// exact final report by this kernel, including all cases, indicators,
		// proofs, bindings, and the receipt for the reconstructed target.
		attested.Summary.BundleOnlyVerification, attested.Summary.ConsumerRechecks = verificationMode(bundle.Digest, true)
		// ValidatePreliminary above proves that report.Digest is the exact
		// producer/consumer-independent preliminary subject. Do not inherit a
		// digest from a receipt field while creating the first receipt.
		attested.ConsumerReceipt = expectedConsumerReceipt(attested, report.Digest, targetPath, raw)
		attested.Indicators = indicators(attested.Summary, ProofPhaseFinal)
		attested.Proofs = proofs(attested, attested.Cases, ProofPhaseFinal)
		attested.ConformanceDecision, attested.ConformanceResolution, attested.ConformanceReason = "PASS", "EXACT", "PROOF_CARRYING_ARTIFACT_CONTRACT_SATISFIED"
		attested.ConformanceCoordinate = Coordinate{"CONSUME_AUTHORITY", "grant-read-only-consumption", attested.ConformanceReason}
		attested.ArtifactUseAuthority = "READ_ONLY_CONSUMPTION"
		attested.ProofSummary = proofSummary(attested.Proofs, ProofPhaseFinal, attested.ArtifactUseAuthority)
		attested.Digest = reportDigest(attested)
		if Validate(attested) != nil {
			return ConsumerReceipt{}, consumerError(ConsumerErrorAttestationMismatch, "consumer attestation is not independently verified")
		}
	}
	if finalReport && report.ConsumerReceipt.PreliminaryDigest != canonicalPreliminaryProjection(report).Digest {
		return ConsumerReceipt{}, consumerError(ConsumerErrorAttestationMismatch, "consumer receipt preliminary binding does not match canonical projection")
	}
	preliminaryDigest := report.ConsumerReceipt.PreliminaryDigest
	if !finalReport {
		preliminaryDigest = report.Digest
	}
	receipt := expectedConsumerReceipt(attested, preliminaryDigest, targetPath, raw)
	if finalReport && !reflect.DeepEqual(receipt, report.ConsumerReceipt) {
		return ConsumerReceipt{}, consumerError(ConsumerErrorReceiptMismatch, "consumer receipt does not match report")
	}
	return receipt, nil
}

func bundleTargetBytes(bundle Bundle, targetPath string) ([]byte, error) {
	for _, item := range bundle.Files {
		if item.Path == targetPath {
			return base64.StdEncoding.DecodeString(item.Content)
		}
	}
	return nil, fmt.Errorf("consumer target is absent from bundle: %s", targetPath)
}

func bundleTargetDigests(bundle Bundle, targetPath string) (string, string, bool) {
	raw, err := bundleTargetBytes(bundle, targetPath)
	if err != nil {
		return "", "", false
	}
	return digestBytes(raw), "", false
}
