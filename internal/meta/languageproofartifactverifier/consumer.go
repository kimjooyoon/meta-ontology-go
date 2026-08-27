package languageproofartifactverifier

import (
	"encoding/base64"
	"fmt"
	"reflect"
)

const ConsumerReceiptSchema = "gooo/read-only-consumption-receipt/v1"

func DecodeConsumerReceipt(raw []byte) (ConsumerReceipt, error) {
	return decodeStrict[ConsumerReceipt](raw)
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
	Decision          string `json:"decision"`
	Resolution        string `json:"resolution"`
	Reason            string `json:"reason"`
	Authority         string `json:"authority"`
	SubjectDecision   string `json:"subject_decision"`
	BundleDigest      string `json:"bundle_digest"`
	PreliminaryDigest string `json:"preliminary_digest"`
	Producer          string `json:"producer"`
	Consumer          string `json:"consumer"`
	TargetPath        string `json:"target_path"`
	TargetDigest      string `json:"target_digest"`
}

func attestationDigest(report Report) string {
	return attestationDigestFor(report, report.ConsumerReceipt)
}

func attestationDigestFor(report Report, receipt ConsumerReceipt) string {
	preliminaryDigest := receipt.PreliminaryDigest
	if preliminaryDigest == "" {
		preliminaryDigest = report.Digest
	}
	return digestValue(consumerAttestation{Decision: report.ConformanceDecision, Resolution: report.ConformanceResolution, Reason: report.ConformanceReason,
		Authority: report.ArtifactUseAuthority, SubjectDecision: report.SubjectArtifactDecision, BundleDigest: report.BundleDigest,
		PreliminaryDigest: preliminaryDigest, Producer: report.Producer, Consumer: report.Consumer, TargetPath: receipt.TargetPath, TargetDigest: receipt.TargetDigest})
}

func consumerAttestedReport(report Report) Report {
	report.ConformanceDecision = "PASS"
	report.ConformanceResolution = "EXACT"
	report.ConformanceReason = "PROOF_CARRYING_ARTIFACT_CONTRACT_SATISFIED"
	report.ConformanceCoordinate = Coordinate{"CONSUME_AUTHORITY", "grant-read-only-consumption", report.ConformanceReason}
	report.ArtifactUseAuthority = "READ_ONLY_CONSUMPTION"
	return report
}

func expectedConsumerReceipt(report Report, targetPath string, target []byte) ConsumerReceipt {
	targetDigest := digestBytes(target)
	outputDigest := digestValue(struct {
		TargetPath   string `json:"target_path"`
		TargetDigest string `json:"target_digest"`
		Authority    string `json:"authority"`
	}{targetPath, targetDigest, "READ_ONLY_CONSUMPTION"})
	preliminaryDigest := report.ConsumerReceipt.PreliminaryDigest
	if preliminaryDigest == "" {
		preliminaryDigest = report.Digest
	}
	receipt := ConsumerReceipt{Schema: ConsumerReceiptSchema, Version: 1, PreliminaryDigest: preliminaryDigest, Producer: report.Producer, Consumer: report.Consumer,
		TargetPath: targetPath, TargetDigest: targetDigest, OutputDigest: outputDigest, OutputExists: true, Authority: "READ_ONLY_CONSUMPTION"}
	receipt.AttestationDigest = attestationDigestFor(report, receipt)
	receipt.Digest = consumerReceiptDigest(receipt)
	return receipt
}

func consumerReceiptDigest(receipt ConsumerReceipt) string {
	receipt.Digest = ""
	return digestValue(receipt)
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
	finalReport := Validate(report) == nil
	if !finalReport && ValidatePreliminary(report) != nil {
		return ConsumerReceipt{}, consumerError(ConsumerErrorAttestationMismatch, "consumer attestation is not independently verified")
	}
	if report.BundleDigest != bundle.Digest {
		return ConsumerReceipt{}, consumerError(ConsumerErrorAttestationMismatch, "consumer attestation is not independently verified")
	}
	attested := consumerAttestedReport(report)
	if !finalReport {
		// A preliminary report is accepted only when it can be lifted to the
		// exact final report by this kernel, including all cases, indicators,
		// proofs, bindings, and the receipt for the reconstructed target.
		attested.Summary.ConsumerRechecks = 1
		attested.ConsumerReceipt = expectedConsumerReceipt(attested, targetPath, raw)
		attested.Indicators = indicators(attested.Summary)
		attested.Proofs = proofs(attested, attested.Cases)
		attested.ConformanceDecision, attested.ConformanceResolution, attested.ConformanceReason = "PASS", "EXACT", "PROOF_CARRYING_ARTIFACT_CONTRACT_SATISFIED"
		attested.ConformanceCoordinate = Coordinate{"CONSUME_AUTHORITY", "grant-read-only-consumption", attested.ConformanceReason}
		attested.ArtifactUseAuthority = "READ_ONLY_CONSUMPTION"
		attested.Digest = reportDigest(attested)
		if Validate(attested) != nil {
			return ConsumerReceipt{}, consumerError(ConsumerErrorAttestationMismatch, "consumer attestation is not independently verified")
		}
	}
	receipt := expectedConsumerReceipt(attested, targetPath, raw)
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
