package generation

import (
    "strings"
    "testing"
)

func TestRewriteCandidateProvenancePart01PassAndTamper(t *testing.T) {
    digest := "sha256:" + strings.Repeat("a", 64)
    receipt, err := NewRewriteCandidateProvenancePart01(RewriteCandidateInputPart01{
        InputIRDigest: digest, RewriteSetDigest: digest, CandidateDigest: digest,
        RuleIdentity: "rewrite/constant-fold/v1", SynthesisPolicyIdentity: "gooo/self-improvement/v1",
        Saturation: RewriteCandidateSaturatedPart01, Status: RewriteCandidatePassPart01,
        ExtractionCost: 3, MissingStageIndex: -1,
    })
    if err != nil || !receipt.ValidPart01() {
        t.Fatalf("expected valid candidate receipt: %#v, %v", receipt, err)
    }
    receipt.CandidateDigest = "sha256:" + strings.Repeat("b", 64)
    if receipt.ValidPart01() {
        t.Fatal("tampered candidate digest must be rejected")
    }
}

func TestRewriteCandidateProvenancePart01UnknownPreservesCounterexampleFrontier(t *testing.T) {
    digest := "sha256:" + strings.Repeat("c", 64)
    receipt, err := NewRewriteCandidateProvenancePart01(RewriteCandidateInputPart01{
        InputIRDigest: digest, RewriteSetDigest: digest, CandidateDigest: digest,
        RuleIdentity: "rewrite/constant-fold/v1", SynthesisPolicyIdentity: "gooo/self-improvement/v1",
        Saturation: RewriteCandidateTimeoutPart01, Status: RewriteCandidateUnknownPart01,
        Reason: "saturation timed out before verifier evidence", ExtractionCost: 0, MissingStageIndex: 1,
    })
    if err != nil || !receipt.ValidPart01() || receipt.MissingStageIndex != 1 {
        t.Fatalf("expected UNKNOWN frontier: %#v, %v", receipt, err)
    }
}