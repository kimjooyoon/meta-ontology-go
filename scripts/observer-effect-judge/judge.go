package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/observereffect"
)

type Check struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	Reason string `json:"reason"`
}

type Judgment struct {
	Schema              string                         `json:"schema"`
	Producer            string                         `json:"producer"`
	Consumer            string                         `json:"consumer"`
	MetaOperation       string                         `json:"meta_operation"`
	ProofChoice         string                         `json:"proof_choice"`
	Decision            string                         `json:"decision"`
	SubjectDecision     string                         `json:"subject_decision"`
	Resolution          string                         `json:"resolution"`
	RepositoryWrites    int                            `json:"repository_writes"`
	MutationAuthority   bool                           `json:"mutation_authority"`
	PromotionAuthorized bool                           `json:"promotion_authorized"`
	Unknown             observereffect.Unknown         `json:"unknown"`
	Coordinate          observereffect.Unknown         `json:"coordinate"`
	Reason              string                         `json:"reason"`
	ClaimTransition     observereffect.ClaimTransition `json:"claim_transition"`
	Metrics             observereffect.Metrics         `json:"metrics"`
	Checks              []Check                        `json:"checks"`
	Digest              string                         `json:"digest"`
}

func judge(report observereffect.Report, observationReceipt, effectReceipt observereffect.Receipt) (Judgment, error) {
	if report.Schema != observereffect.LedgerSchema || report.Experiment != observereffect.ExperimentName {
		return Judgment{}, fmt.Errorf("unexpected observer-effect ledger identity")
	}
	if !report.Source.GoooSource || !strings.HasSuffix(report.Source.Path, ".gooo") || report.Source.Digest == "" || len(report.Effects) != 4 || len(report.Indicators) != observereffect.FixedDenominator {
		return Judgment{}, fmt.Errorf("ledger does not contain the fixed experiment surface")
	}
	if report.MutationAuthority || report.PromotionAuthorized || report.Authority.MutationAuthority || report.Authority.PromotionAuthorized {
		return Judgment{}, fmt.Errorf("ledger grants mutation or promotion authority")
	}
	if report.Coordinate != report.Unknown || report.Reason != report.Unknown.Reason {
		return Judgment{}, fmt.Errorf("unknown coordinate is not persistent")
	}
	if report.RepositoryWrites != report.Authority.RepositoryWrites {
		return Judgment{}, fmt.Errorf("repository write count is not authority-bound")
	}
	if report.Authority.OutputWrites != 3 {
		return Judgment{}, fmt.Errorf("observer output effect count is not fixed")
	}
	expectedSubject, expectedResolution := independentDecision(report)
	if report.Decision != expectedSubject || report.Resolution != expectedResolution {
		return Judgment{}, fmt.Errorf("decision is not derived from observed effects")
	}
	if err := validateEffects(report); err != nil {
		return Judgment{}, err
	}
	if err := validateIndicators(report); err != nil {
		return Judgment{}, err
	}
	if err := validateReceipts(report, observationReceipt, effectReceipt); err != nil {
		return Judgment{}, err
	}
	if report.Digest != independentReportDigest(report) {
		return Judgment{}, fmt.Errorf("ledger digest does not replay")
	}
	if report.EvidenceDigest != independentValueDigest([]any{report.Source, report.Observation, report.Effects, report.Unknown, report.ClaimTransition}) {
		return Judgment{}, fmt.Errorf("ledger evidence digest does not replay")
	}
	if report.ClaimTransition.CurrentState != claimState(report.Decision) || !report.ClaimTransition.Persistent || report.ClaimTransition.Sequence != 2 {
		return Judgment{}, fmt.Errorf("persistent claim transition is inconsistent")
	}
	judgment := Judgment{
		Schema: observereffect.JudgmentSchema, Producer: "observer-effect-judge",
		Consumer: "ci-proof", MetaOperation: "independently-judge-effect-ledger",
		ProofChoice: "REGRESSION", Decision: "PASS", SubjectDecision: report.Decision,
		Resolution: report.Resolution, RepositoryWrites: report.RepositoryWrites,
		MutationAuthority: report.MutationAuthority, PromotionAuthorized: report.PromotionAuthorized,
		Unknown: report.Unknown, Coordinate: report.Coordinate, Reason: report.Reason,
		ClaimTransition: report.ClaimTransition, Metrics: report.Metrics,
		Checks: []Check{
			{ID: "judge.decision-recomputed", Status: "PASS", Reason: "EFFECTS_DERIVE_SUBJECT_DECISION"},
			{ID: "judge.fixed-denominator", Status: "PASS", Reason: "TWELVE_INDICATORS_RETAINED"},
			{ID: "judge.receipt-chain", Status: "PASS", Reason: "RECEIPTS_AND_LEDGER_DIGEST_BOUND"},
			{ID: "judge.authority", Status: "PASS", Reason: "MUTATION_AND_PROMOTION_DENIED"},
		},
	}
	judgment.Digest = independentJudgmentDigest(judgment)
	return judgment, nil
}

func independentDecision(report observereffect.Report) (string, string) {
	if report.Unknown.Reason != "NONE" {
		return "UNKNOWN", "LOWER_RESOLUTION"
	}
	if report.RepositoryWrites != 0 || report.Observation.RepositoryStorage.Changed {
		return "FAIL_CLOSED", "EXACT"
	}
	return "OBSERVED", "EXACT"
}

func validateEffects(report observereffect.Report) error {
	seen := make(map[string]bool, len(report.Effects))
	for _, effect := range report.Effects {
		if seen[effect.Domain] || effect.Producer != "observer-effect-ledger" || effect.Consumer != "observer-effect-judge" || effect.MetaOperation == "" || effect.ProofChoice == "" {
			return fmt.Errorf("effect metadata is not unique and bound")
		}
		seen[effect.Domain] = true
	}
	for _, domain := range []string{"REPOSITORY_STORAGE", "ENVIRONMENT", "LOGICAL_TIME", "OUTPUT"} {
		if !seen[domain] {
			return fmt.Errorf("effect domain %s is missing", domain)
		}
	}
	if effect := effectByDomain(report.Effects, "OUTPUT"); effect.WriteCount != 3 || !effect.ObservedChanged || !effect.MutationAttempted {
		return fmt.Errorf("observer output effect is not recorded")
	}
	repository := effectByDomain(report.Effects, "REPOSITORY_STORAGE")
	if repository.WriteCount != report.RepositoryWrites || repository.ObservedChanged != (report.RepositoryWrites != 0) || repository.MutationAttempted != (report.RepositoryWrites != 0) {
		return fmt.Errorf("repository storage effect is inconsistent")
	}
	return nil
}

func validateIndicators(report observereffect.Report) error {
	expectedIDs := map[string]bool{
		"OEL-OBS-01": true, "OEL-OBS-02": true, "OEL-OBS-03": true, "OEL-OBS-04": true,
		"OEL-OBS-05": true, "OEL-OBS-06": true, "OEL-EFF-01": true, "OEL-EFF-02": true,
		"OEL-EFF-03": true, "OEL-EFF-04": true, "OEL-GOV-01": true, "OEL-GOV-02": true,
	}
	ids := make(map[string]bool, len(report.Indicators))
	pass, observations, effects, guardrails := 0, 0, 0, 0
	for _, indicator := range report.Indicators {
		if ids[indicator.ID] || !expectedIDs[indicator.ID] || indicator.Producer == "" || indicator.Consumer == "" || indicator.MetaOperation == "" || indicator.ProofChoice == "" {
			return fmt.Errorf("indicator metadata is not bound")
		}
		ids[indicator.ID] = true
		if indicator.Status == "PASS" {
			pass++
		} else if indicator.Status != "FAIL" && indicator.Status != "UNKNOWN" {
			return fmt.Errorf("indicator %s has invalid status", indicator.ID)
		}
		switch indicator.Class {
		case "OBSERVATION":
			observations += boolInt(indicator.Status == "PASS")
		case "EFFECT":
			effects += boolInt(indicator.Status == "PASS")
		case "GUARDRAIL":
			guardrails += boolInt(indicator.Status == "PASS")
		default:
			return fmt.Errorf("indicator %s has invalid class", indicator.ID)
		}
	}
	if report.Metrics.FixedDenominator != observereffect.FixedDenominator || report.Metrics.Satisfied != pass || report.Metrics.CoverageBasisPoints != pass*10000/observereffect.FixedDenominator {
		return fmt.Errorf("fixed denominator metrics are not recomputed")
	}
	if report.Metrics.ObservationTotal != 6 || report.Metrics.EffectTotal != 4 || report.Metrics.GuardrailTotal != 2 {
		return fmt.Errorf("indicator denominators changed")
	}
	if report.Metrics.ObservationSatisfied != observations || report.Metrics.EffectSatisfied != effects || report.Metrics.GuardrailSatisfied != guardrails {
		return fmt.Errorf("indicator class metrics are not recomputed")
	}
	if report.Decision == "OBSERVED" && pass != observereffect.FixedDenominator {
		return fmt.Errorf("observed result did not satisfy all indicators")
	}
	return nil
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func validateReceipts(report observereffect.Report, observationReceipt, effectReceipt observereffect.Receipt) error {
	for _, receipt := range []observereffect.Receipt{observationReceipt, effectReceipt} {
		if receipt.Schema != observereffect.ReceiptSchema || receipt.Producer != "observer-effect-ledger" || receipt.Consumer != "observer-effect-judge" || receipt.SubjectDigest != report.Source.Digest || receipt.Decision != report.Decision || receipt.Resolution != report.Resolution || receipt.RepositoryWrites != report.RepositoryWrites || receipt.MutationAuthority || receipt.Coordinate != report.Coordinate || receipt.Reason != report.Reason {
			return fmt.Errorf("receipt is not bound to the report")
		}
		if receipt.Digest != independentReceiptDigest(receipt) {
			return fmt.Errorf("receipt digest does not replay")
		}
	}
	if len(report.ReceiptDigests) != 2 || report.ReceiptDigests[0] != observationReceipt.Digest || report.ReceiptDigests[1] != effectReceipt.Digest {
		return fmt.Errorf("receipt digest list is not canonical")
	}
	if observationReceipt.Kind != "OBSERVATION_RESULT" || effectReceipt.Kind != "OBSERVER_EFFECT" {
		return fmt.Errorf("observation and effect receipts are not separated")
	}
	if observationReceipt.EvidenceDigest != report.EvidenceDigest || effectReceipt.EvidenceDigest != independentValueDigest(report.Effects) {
		return fmt.Errorf("receipt evidence is not bound to its role")
	}
	if observationReceipt.ClaimTransition != report.ClaimTransition || effectReceipt.ClaimTransition != report.ClaimTransition {
		return fmt.Errorf("receipt claim transition is not persistent")
	}
	return nil
}

func effectByDomain(effects []observereffect.Effect, domain string) observereffect.Effect {
	for _, effect := range effects {
		if effect.Domain == domain {
			return effect
		}
	}
	return observereffect.Effect{}
}

func claimState(decision string) string {
	switch decision {
	case "FAIL_CLOSED":
		return "REFUTED"
	case "UNKNOWN":
		return "UNKNOWN"
	default:
		return "SUPPORTED"
	}
}

func independentReceiptDigest(receipt observereffect.Receipt) string {
	receipt.Digest = ""
	return independentValueDigest(receipt)
}

func independentReportDigest(report observereffect.Report) string {
	report.Digest = ""
	return independentValueDigest(report)
}

func independentJudgmentDigest(judgment Judgment) string {
	judgment.Digest = ""
	return independentValueDigest(judgment)
}

func independentValueDigest(value any) string {
	payload, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	digest := sha256.Sum256(payload)
	return hex.EncodeToString(digest[:])
}

func bytesReader(payload []byte) *bytes.Reader { return bytes.NewReader(payload) }
