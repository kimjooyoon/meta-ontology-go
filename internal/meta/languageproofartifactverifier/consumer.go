package languageproofartifactverifier

import (
	"encoding/base64"
	"fmt"
	"reflect"
)

const ConsumerReceiptSchema = "gooo/read-only-consumption-receipt/v1"

type consumerAttestation struct {
	Decision        string
	Resolution      string
	Reason          string
	Authority       string
	SubjectDecision string
	BundleDigest    string
}

func attestationDigest(report Report) string {
	return digestValue(consumerAttestation{Decision: report.ConformanceDecision, Resolution: report.ConformanceResolution, Reason: report.ConformanceReason,
		Authority: report.ArtifactUseAuthority, SubjectDecision: report.SubjectArtifactDecision, BundleDigest: report.BundleDigest})
}

func expectedConsumerReceipt(report Report, targetPath string, target []byte) ConsumerReceipt {
	targetDigest := digestBytes(target)
	outputDigest := digestValue(struct {
		TargetPath   string `json:"target_path"`
		TargetDigest string `json:"target_digest"`
		Authority    string `json:"authority"`
	}{targetPath, targetDigest, "READ_ONLY_CONSUMPTION"})
	receipt := ConsumerReceipt{Schema: ConsumerReceiptSchema, Version: 1, TargetPath: targetPath, TargetDigest: targetDigest,
		OutputDigest: outputDigest, Authority: "READ_ONLY_CONSUMPTION", AttestationDigest: attestationDigest(report)}
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
		return ConsumerReceipt{}, err
	}
	if err := Validate(report); err != nil || report.ConformanceDecision != "PASS" || report.ArtifactUseAuthority != "READ_ONLY_CONSUMPTION" || report.BundleDigest != bundle.Digest {
		return ConsumerReceipt{}, fmt.Errorf("consumer attestation is not independently verified")
	}
	if targetPath == "" || targetPath == "artifact.json" {
		targetPath = "artifact.json"
	} else {
		return ConsumerReceipt{}, fmt.Errorf("consumer target is outside read-only policy: %s", targetPath)
	}
	for _, item := range bundle.Files {
		if item.Path != targetPath {
			continue
		}
		raw, err := base64.StdEncoding.DecodeString(item.Content)
		if err != nil {
			return ConsumerReceipt{}, err
		}
		receipt := expectedConsumerReceipt(report, targetPath, raw)
		if !reflect.DeepEqual(receipt, report.ConsumerReceipt) {
			return ConsumerReceipt{}, fmt.Errorf("consumer receipt does not match report")
		}
		return receipt, nil
	}
	return ConsumerReceipt{}, fmt.Errorf("consumer target is absent from bundle: %s", targetPath)
}
