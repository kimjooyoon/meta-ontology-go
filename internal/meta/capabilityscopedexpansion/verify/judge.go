package verify

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
)

const (
	validSchema        = "gooo/capability-scoped-expansion/v1"
	validMetaOperation = "expand-capability-scoped-meta-code"
	validProducer      = "capabilityscopedexpansion.Evaluate"
	validConsumer      = "ci-capability-expansion-gate"
	validGoVersion     = "go1.27.0"
	validStage         = "expand"
	validStep          = "authorize-before-expand"
)

type Verdict struct {
	Status        string `json:"status"`
	Decision      string `json:"decision"`
	Resolution    string `json:"resolution"`
	Reason        string `json:"reason"`
	ReceiptDigest string `json:"receipt_digest"`
}

type receipt struct {
	Schema             string        `json:"schema"`
	MetaOperation      string        `json:"meta_operation"`
	Producer           string        `json:"producer"`
	Consumer           string        `json:"consumer"`
	GoVersion          string        `json:"go_version"`
	SourceDigest       string        `json:"source_digest"`
	Stage              string        `json:"stage"`
	Step               string        `json:"step"`
	Decision           string        `json:"decision"`
	Resolution         string        `json:"resolution"`
	EnforcementEffect  string        `json:"enforcement_effect"`
	Reason             string        `json:"reason"`
	Declarations       []declaration `json:"declarations"`
	Capabilities       []capability  `json:"capabilities"`
	Evidence           []evidence    `json:"evidence"`
	Unknown            *unknown      `json:"unknown,omitempty"`
	Authority          authority     `json:"authority"`
	Claims             []claim       `json:"claims"`
	Indicators         []indicator   `json:"indicators"`
	RepositoryWrites   int           `json:"repository_writes"`
	MutationAuthority  bool          `json:"mutation_authority"`
	PromotionAuthority bool          `json:"promotion_authority"`
	ReportDigest       string        `json:"report_digest"`
}

type declaration struct {
	Kind      string `json:"kind"`
	Operation string `json:"operation"`
	Target    string `json:"target"`
	Declared  bool   `json:"declared"`
}
type capability struct {
	ValueID   string `json:"value_id"`
	Kind      string `json:"kind"`
	Operation string `json:"operation"`
	Target    string `json:"target"`
}
type evidence struct {
	ValueID        string `json:"value_id"`
	Observed       string `json:"observed"`
	EvidenceDigest string `json:"evidence_digest"`
}
type unknown struct {
	Stage  string `json:"stage"`
	Step   string `json:"step"`
	Reason string `json:"reason"`
}
type authority struct {
	CapabilitiesRequested       int  `json:"capabilities_requested"`
	CapabilitiesDeclared        int  `json:"capabilities_declared"`
	CapabilitiesAuthorized      int  `json:"capabilities_authorized"`
	CapabilitiesDenied          int  `json:"capabilities_denied"`
	CapabilitiesUnknown         int  `json:"capabilities_unknown"`
	RequestedRepositoryWrites   int  `json:"requested_repository_writes"`
	RequestedMutationAuthority  bool `json:"requested_mutation_authority"`
	RequestedPromotionAuthority bool `json:"requested_promotion_authority"`
	RepositoryWrites            int  `json:"repository_writes"`
	MutationAuthority           bool `json:"mutation_authority"`
	PromotionAuthority          bool `json:"promotion_authority"`
}
type claim struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}
type indicator struct {
	ID            string `json:"id"`
	Class         string `json:"class"`
	Producer      string `json:"producer"`
	Consumer      string `json:"consumer"`
	MetaOperation string `json:"meta_operation"`
	ProofChoice   string `json:"proof_choice"`
}

// Judge checks a receipt from raw JSON and the source independently of Evaluate.
// Its duplicated boundary table is intentional: a green producer cannot certify
// its own conclusion without a second relation check.
func Judge(source, raw []byte) Verdict {
	var observed receipt
	if err := json.Unmarshal(raw, &observed); err != nil {
		return invalid("receipt is not JSON")
	}
	if err := shape(source, observed); err != nil {
		return invalid(err.Error())
	}
	expectedDecision, expectedResolution, reason := expected(source, observed)
	if observed.Decision != expectedDecision || observed.Resolution != expectedResolution {
		return invalid(fmt.Sprintf("producer decision disagrees with independent expectation: %s/%s", expectedDecision, expectedResolution))
	}
	if observed.Reason != reason {
		return invalid("producer reason disagrees with independent expectation")
	}
	if observed.ReportDigest == "" || canonicalDigest(raw) != observed.ReportDigest {
		return invalid("receipt digest does not bind raw evidence")
	}
	if observed.Decision == "UNKNOWN" {
		if observed.Unknown == nil || observed.Unknown.Stage == "" || observed.Unknown.Step == "" || observed.Unknown.Reason == "" {
			return invalid("UNKNOWN receipt lacks stage, step, or reason")
		}
	} else if observed.Unknown != nil {
		return invalid("known receipt carries UNKNOWN evidence")
	}
	wantEffect := "BLOCK"
	if expectedDecision == "ALLOW" {
		wantEffect = "NONE"
	}
	if observed.EnforcementEffect != wantEffect {
		return invalid("enforcement effect does not match independent decision")
	}
	if err := authorityCounters(observed, expectedDecision); err != nil {
		return invalid(err.Error())
	}
	if err := claims(observed.Claims, observed.Decision); err != nil {
		return invalid(err.Error())
	}
	return Verdict{Status: "PASS", Decision: observed.Decision, Resolution: observed.Resolution,
		Reason: reason, ReceiptDigest: observed.ReportDigest}
}

func shape(source []byte, observed receipt) error {
	if observed.Schema != validSchema || observed.MetaOperation != validMetaOperation ||
		observed.Producer != validProducer || observed.Consumer != validConsumer || observed.GoVersion != validGoVersion {
		return fmt.Errorf("receipt identity or Go toolchain is not pinned")
	}
	if observed.SourceDigest != digestBytes(source) || len(observed.Indicators) != 12 {
		return fmt.Errorf("source digest or fixed indicator denominator mismatch")
	}
	if err := sourceShape(source); err != nil {
		return err
	}
	if err := declarationsExact(observed.Declarations); err != nil {
		return err
	}
	if observed.Authority.RepositoryWrites != 0 || observed.Authority.MutationAuthority || observed.Authority.PromotionAuthority ||
		observed.RepositoryWrites != 0 || observed.MutationAuthority || observed.PromotionAuthority {
		return fmt.Errorf("receipt exceeds the zero-effect ceiling")
	}
	wantIndicators := map[string]string{
		"CSE-authority-ceiling":               "GUARDRAIL/REGRESSION",
		"CSE-default-deny":                    "GUARDRAIL/REGRESSION",
		"CSE-environment-capability-declared": "DRIVER/FOUNDATION",
		"CSE-expansion-stage-order":           "OUTCOME/COHERENCE",
		"CSE-file-capability-declared":        "DRIVER/FOUNDATION",
		"CSE-network-capability-declared":     "DRIVER/FOUNDATION",
		"CSE-receipt-seal":                    "DRIVER/COHERENCE",
		"CSE-source-binding":                  "OUTCOME/COHERENCE",
		"CSE-source-shape":                    "DRIVER/FOUNDATION",
		"CSE-time-capability-declared":        "DRIVER/FOUNDATION",
		"CSE-toolchain-1.27":                  "DRIVER/FOUNDATION",
		"CSE-value-evidence-relation":         "OUTCOME/COHERENCE",
	}
	indicatorIDs := make(map[string]bool, len(observed.Indicators))
	for _, item := range observed.Indicators {
		if item.Producer != validProducer || item.Consumer != validConsumer || item.MetaOperation != validMetaOperation || item.ProofChoice == "" {
			return fmt.Errorf("indicator provenance is incomplete")
		}
		if indicatorIDs[item.ID] || wantIndicators[item.ID] != item.Class+"/"+item.ProofChoice {
			return fmt.Errorf("indicator denominator is duplicated or malformed")
		}
		indicatorIDs[item.ID] = true
	}
	if len(indicatorIDs) != len(wantIndicators) {
		return fmt.Errorf("indicator denominator is incomplete")
	}
	if observed.Authority.CapabilitiesRequested != len(observed.Capabilities) {
		return fmt.Errorf("capability request count is inconsistent")
	}
	return nil
}

func sourceShape(source []byte) error {
	for _, marker := range []string{
		"package capabilityscopedexpansion", "namespace capabilityscopedexpansion",
		"activity DeclareFileReadCapability", "activity DeclareLogicalTimeCapability",
		"activity DeclareEnvironmentReadCapability", "activity DeclarePinnedNetworkCapability",
		"activity BindCapabilityEvidence", "activity ExpandWithCapabilityEvidence", "activity DenyByDefault",
	} {
		if !strings.Contains(string(source), marker) {
			return fmt.Errorf("source declaration %q is not independently observable", marker)
		}
	}
	return nil
}

func declarationsExact(values []declaration) error {
	if len(values) != 4 {
		return fmt.Errorf("capability declaration denominator mismatch")
	}
	want := map[string]bool{
		"file|read|source":                                        true,
		"time|read|logical-clock":                                 true,
		"environment|read|GOOO_EXPANSION_PROFILE":                 true,
		"network|read|https://example.invalid/gooo/pinned-schema": true,
	}
	seen := make(map[string]bool, len(values))
	for _, item := range values {
		key := item.Kind + "|" + item.Operation + "|" + item.Target
		if !item.Declared || !want[key] || seen[key] {
			return fmt.Errorf("capability declarations are not exact")
		}
		seen[key] = true
	}
	return nil
}

func expected(source []byte, observed receipt) (string, string, string) {
	if observed.Stage == "" || observed.Step == "" || observed.GoVersion == "" {
		return "UNKNOWN", "UNKNOWN", "EXPANSION_INPUT_UNOBSERVED"
	}
	if observed.Stage != validStage || observed.Step != validStep || observed.GoVersion != validGoVersion {
		return "DENY", "INVARIANT_ONLY", "EXPANSION_BOUNDARY_REJECTED"
	}
	declared := 0
	for _, item := range observed.Capabilities {
		if known(item) && declaredIn(item, observed.Declarations) {
			declared++
		}
	}
	if declared != len(observed.Capabilities) {
		return "DENY", "INVARIANT_ONLY", "CAPABILITY_NOT_DECLARED"
	}
	if !evidenceMatches(source, observed.Capabilities, observed.Evidence) {
		return "UNKNOWN", "UNKNOWN", "EVIDENCE_UNOBSERVED"
	}
	if observed.Authority.RequestedRepositoryWrites != 0 || observed.Authority.RequestedMutationAuthority || observed.Authority.RequestedPromotionAuthority {
		return "DENY", "INVARIANT_ONLY", "EFFECT_CEILING_REJECTED"
	}
	return "ALLOW", "EXACT", "CAPABILITY_SCOPE_EXACT"
}

func known(item capability) bool {
	return (item.Kind == "file" && item.Operation == "read" && item.Target == "source") ||
		(item.Kind == "time" && item.Operation == "read" && item.Target == "logical-clock") ||
		(item.Kind == "environment" && item.Operation == "read" && item.Target == "GOOO_EXPANSION_PROFILE") ||
		(item.Kind == "network" && item.Operation == "read" && item.Target == "https://example.invalid/gooo/pinned-schema")
}

func declaredIn(item capability, declarations []declaration) bool {
	for _, declaration := range declarations {
		if declaration.Declared && declaration.Kind == item.Kind && declaration.Operation == item.Operation && declaration.Target == item.Target {
			return true
		}
	}
	return false
}

func evidenceMatches(source []byte, values []capability, evidence []evidence) bool {
	if len(values) != len(evidence) {
		return false
	}
	byID := make(map[string]evidence, len(evidence))
	for _, item := range evidence {
		if _, exists := byID[item.ValueID]; exists {
			return false
		}
		byID[item.ValueID] = item
	}
	for _, value := range values {
		item, ok := byID[value.ValueID]
		if !ok {
			return false
		}
		observed := item.Observed
		switch value.ValueID {
		case "file-read":
			if observed != digestBytes(source) {
				return false
			}
		case "logical-time-read":
			if observed != "logical-clock:0" {
				return false
			}
		case "environment-read":
			if observed != "GOOO_EXPANSION_PROFILE=deterministic" {
				return false
			}
		case "pinned-network-read":
			if observed != digestBytes([]byte("pinned-schema-v1")) {
				return false
			}
		default:
			return false
		}
		if item.EvidenceDigest != digestBytes([]byte(item.ValueID+"="+item.Observed)) {
			return false
		}
	}
	return true
}

func claims(values []claim, decision string) error {
	if len(values) != 3 {
		return fmt.Errorf("claim denominator mismatch")
	}
	statuses := make(map[string]string, len(values))
	for _, item := range values {
		if statuses[item.ID] != "" {
			return fmt.Errorf("duplicate claim")
		}
		statuses[item.ID] = item.Status
	}
	wantScope := "DISCHARGED"
	if decision == "DENY" {
		wantScope = "REFUTED"
	}
	if decision == "UNKNOWN" {
		wantScope = "OPEN"
	}
	wantOther := "DISCHARGED"
	if decision == "UNKNOWN" {
		wantOther = "OPEN"
	}
	if statuses["capability-scope-exact"] != wantScope || statuses["default-deny"] != wantOther || statuses["effect-ceiling"] != wantOther {
		return fmt.Errorf("claim lifecycle does not match decision")
	}
	return nil
}

func authorityCounters(observed receipt, decision string) error {
	a := observed.Authority
	if a.CapabilitiesDeclared < 0 || a.CapabilitiesDeclared > a.CapabilitiesRequested || a.CapabilitiesAuthorized < 0 || a.CapabilitiesAuthorized > a.CapabilitiesRequested || a.CapabilitiesDenied < 0 || a.CapabilitiesDenied > a.CapabilitiesRequested || a.CapabilitiesUnknown < 0 || a.CapabilitiesUnknown > a.CapabilitiesRequested {
		return fmt.Errorf("capability permission counters are inconsistent")
	}
	switch decision {
	case "ALLOW":
		if a.CapabilitiesDeclared != a.CapabilitiesRequested || a.CapabilitiesAuthorized != a.CapabilitiesRequested || a.CapabilitiesDenied != 0 || a.CapabilitiesUnknown != 0 {
			return fmt.Errorf("ALLOW authority counters are inconsistent")
		}
	case "DENY":
		if a.CapabilitiesAuthorized != 0 || a.CapabilitiesUnknown != 0 || a.CapabilitiesDenied != a.CapabilitiesRequested {
			return fmt.Errorf("DENY authority counters are inconsistent")
		}
	case "UNKNOWN":
		if a.CapabilitiesAuthorized != 0 || a.CapabilitiesDenied != 0 || a.CapabilitiesUnknown != a.CapabilitiesRequested {
			return fmt.Errorf("UNKNOWN authority counters are inconsistent")
		}
	default:
		return fmt.Errorf("unknown decision")
	}
	return nil
}

func invalid(reason string) Verdict { return Verdict{Status: "FAIL", Reason: reason} }

func digestBytes(value []byte) string {
	sum := sha256.Sum256(value)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func canonicalDigest(raw []byte) string {
	var value map[string]any
	if json.Unmarshal(raw, &value) != nil {
		return ""
	}
	value["report_digest"] = ""
	canonical, _ := json.Marshal(value)
	return digestBytes(canonical)
}
