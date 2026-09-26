package generation

import (
    "crypto/sha256"
    "encoding/hex"
    "encoding/json"
    "errors"
    "fmt"
)

const RewriteCandidateProvenanceSchemaPart01 = "gooo/generation/rewrite-candidate-provenance/v1"

type RewriteCandidateStatusPart01 string

const (
    RewriteCandidatePassPart01    RewriteCandidateStatusPart01 = "PASS"
    RewriteCandidateUnknownPart01 RewriteCandidateStatusPart01 = "UNKNOWN"
)

type RewriteCandidateSaturationPart01 string

const (
    RewriteCandidateSaturatedPart01   RewriteCandidateSaturationPart01 = "SATURATED"
    RewriteCandidateTimeoutPart01     RewriteCandidateSaturationPart01 = "TIMEOUT"
    RewriteCandidateResourceLimitPart01 RewriteCandidateSaturationPart01 = "RESOURCE_LIMIT"
    RewriteCandidateSaturationUnknownPart01 RewriteCandidateSaturationPart01 = "UNKNOWN"
)

type RewriteCandidateInputPart01 struct {
    InputIRDigest              string                           `json:"input_ir_digest"`
    RewriteSetDigest           string                           `json:"rewrite_set_digest"`
    CandidateDigest            string                           `json:"candidate_digest"`
    RuleIdentity               string                           `json:"rule_identity"`
    SynthesisPolicyIdentity    string                           `json:"synthesis_policy_identity"`
    Saturation                 RewriteCandidateSaturationPart01 `json:"saturation"`
    Status                     RewriteCandidateStatusPart01    `json:"status"`
    Reason                     string                           `json:"reason,omitempty"`
    ExtractionCost             float64                          `json:"extraction_cost"`
    CounterexampleDigest       string                           `json:"counterexample_digest,omitempty"`
    ReverseObservationDigest  string                           `json:"reverse_observation_digest,omitempty"`
    MissingStageIndex          int                              `json:"missing_stage_index"`
}

type RewriteCandidateProvenancePart01 struct {
    Schema                     string                           `json:"schema"`
    InputIRDigest              string                           `json:"input_ir_digest"`
    RewriteSetDigest           string                           `json:"rewrite_set_digest"`
    CandidateDigest            string                           `json:"candidate_digest"`
    RuleIdentity               string                           `json:"rule_identity"`
    SynthesisPolicyIdentity    string                           `json:"synthesis_policy_identity"`
    Saturation                 RewriteCandidateSaturationPart01 `json:"saturation"`
    Status                     RewriteCandidateStatusPart01    `json:"status"`
    Reason                     string                           `json:"reason,omitempty"`
    ExtractionCost             float64                          `json:"extraction_cost"`
    CounterexampleDigest       string                           `json:"counterexample_digest,omitempty"`
    ReverseObservationDigest  string                           `json:"reverse_observation_digest,omitempty"`
    MissingStageIndex          int                              `json:"missing_stage_index"`
    EvidencePrefixDigest      string                           `json:"evidence_prefix_digest"`
    CandidateReceiptDigest    string                           `json:"candidate_receipt_digest"`
}

func NewRewriteCandidateProvenancePart01(input RewriteCandidateInputPart01) (RewriteCandidateProvenancePart01, error) {
    receipt := RewriteCandidateProvenancePart01{
        Schema: RewriteCandidateProvenanceSchemaPart01,
        InputIRDigest: input.InputIRDigest,
        RewriteSetDigest: input.RewriteSetDigest,
        CandidateDigest: input.CandidateDigest,
        RuleIdentity: input.RuleIdentity,
        SynthesisPolicyIdentity: input.SynthesisPolicyIdentity,
        Saturation: input.Saturation,
        Status: input.Status,
        Reason: input.Reason,
        ExtractionCost: input.ExtractionCost,
        CounterexampleDigest: input.CounterexampleDigest,
        ReverseObservationDigest: input.ReverseObservationDigest,
        MissingStageIndex: input.MissingStageIndex,
    }
    if err := receipt.validateShapePart01(); err != nil {
        return RewriteCandidateProvenancePart01{}, err
    }
    receipt.EvidencePrefixDigest = digestRewriteCandidatePart01(receipt.evidencePrefixPart01())
    receipt.CandidateReceiptDigest = digestRewriteCandidatePart01(receipt.receiptViewPart01())
    return receipt, nil
}

func (r RewriteCandidateProvenancePart01) ValidPart01() bool {
    if r.validateShapePart01() != nil || !validRewriteCandidateDigestPart01(r.EvidencePrefixDigest) || !validRewriteCandidateDigestPart01(r.CandidateReceiptDigest) {
        return false
    }
    if r.EvidencePrefixDigest != digestRewriteCandidatePart01(r.evidencePrefixPart01()) {
        return false
    }
    return r.CandidateReceiptDigest == digestRewriteCandidatePart01(r.receiptViewPart01())
}

func (r RewriteCandidateProvenancePart01) validateShapePart01() error {
    if r.Schema != RewriteCandidateProvenanceSchemaPart01 || r.RuleIdentity == "" || r.SynthesisPolicyIdentity == "" {
        return errors.New("invalid rewrite candidate identity")
    }
    for name, value := range map[string]string{
        "input IR digest": r.InputIRDigest,
        "rewrite set digest": r.RewriteSetDigest,
        "candidate digest": r.CandidateDigest,
        "counterexample digest": r.CounterexampleDigest,
        "reverse observation digest": r.ReverseObservationDigest,
    } {
        if value != "" && !validRewriteCandidateDigestPart01(value) {
            return fmt.Errorf("invalid %s", name)
        }
    }
    if !validRewriteCandidateDigestPart01(r.InputIRDigest) || !validRewriteCandidateDigestPart01(r.RewriteSetDigest) || !validRewriteCandidateDigestPart01(r.CandidateDigest) || r.ExtractionCost < 0 {
        return errors.New("invalid rewrite candidate evidence")
    }
    switch r.Saturation {
    case RewriteCandidateSaturatedPart01, RewriteCandidateTimeoutPart01, RewriteCandidateResourceLimitPart01, RewriteCandidateSaturationUnknownPart01:
    default:
        return errors.New("invalid saturation status")
    }
    switch r.Status {
    case RewriteCandidatePassPart01:
        if r.Saturation != RewriteCandidateSaturatedPart01 || r.MissingStageIndex != -1 {
            return errors.New("PASS requires saturated candidate and missing stage index -1")
        }
    case RewriteCandidateUnknownPart01:
        if r.MissingStageIndex < 0 || r.Reason == "" {
            return errors.New("UNKNOWN requires reason and missing stage index")
        }
    default:
        return errors.New("invalid rewrite candidate status")
    }
    return nil
}

func (r RewriteCandidateProvenancePart01) evidencePrefixPart01() any {
    return struct {
        Schema                    string                           `json:"schema"`
        InputIRDigest             string                           `json:"input_ir_digest"`
        RewriteSetDigest          string                           `json:"rewrite_set_digest"`
        CandidateDigest           string                           `json:"candidate_digest"`
        RuleIdentity              string                           `json:"rule_identity"`
        SynthesisPolicyIdentity   string                           `json:"synthesis_policy_identity"`
        Saturation                RewriteCandidateSaturationPart01 `json:"saturation"`
        Status                    RewriteCandidateStatusPart01    `json:"status"`
        Reason                    string                           `json:"reason,omitempty"`
        ExtractionCost            float64                          `json:"extraction_cost"`
        CounterexampleDigest      string                           `json:"counterexample_digest,omitempty"`
        ReverseObservationDigest string                           `json:"reverse_observation_digest,omitempty"`
        MissingStageIndex         int                              `json:"missing_stage_index"`
    }{r.Schema, r.InputIRDigest, r.RewriteSetDigest, r.CandidateDigest, r.RuleIdentity, r.SynthesisPolicyIdentity, r.Saturation, r.Status, r.Reason, r.ExtractionCost, r.CounterexampleDigest, r.ReverseObservationDigest, r.MissingStageIndex}
}

func (r RewriteCandidateProvenancePart01) receiptViewPart01() any {
    return struct {
        Prefix string `json:"evidence_prefix_digest"`
        Input  any    `json:"input"`
    }{r.EvidencePrefixDigest, r.evidencePrefixPart01()}
}

func digestRewriteCandidatePart01(value any) string {
    encoded, _ := json.Marshal(value)
    sum := sha256.Sum256(encoded)
    return "sha256:" + hex.EncodeToString(sum[:])
}

func validRewriteCandidateDigestPart01(value string) bool {
    if len(value) != len("sha256:")+64 || value[:7] != "sha256:" {
        return false
    }
    _, err := hex.DecodeString(value[7:])
    return err == nil
}