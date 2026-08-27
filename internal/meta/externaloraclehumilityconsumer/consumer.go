// Package externaloraclehumilityconsumer is an independent wire consumer and
// judge. It intentionally repeats the source parse/lower boundary instead of
// importing the producer's types or implementation.
package externaloraclehumilityconsumer

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const (
	ContractSchema = "gooo/external-oracle-humility-contract/v1"
	CapsuleSchema  = "gooo/external-oracle-reference-capsule/v2"
	ReceiptSchema  = "gooo/external-oracle-humility-receipt/v2"
	ReportSchema   = "gooo/external-oracle-humility-report/v2"
	SuiteSchema    = "gooo/external-oracle-humility-suite/v2"
)

type Claim struct {
	ID    string `json:"id"`
	State string `json:"state"`
}

type SourcePolicy struct {
	SourceAuthority           string  `json:"source_authority"`
	ExternalEvidenceRelation  string  `json:"external_evidence_relation"`
	ExternalEvidenceAuthority string  `json:"external_evidence_authority"`
	Claims                    []Claim `json:"claims"`
}

type Declaration struct {
	ID           string `json:"id"`
	Kind         string `json:"kind"`
	Name         string `json:"name"`
	ValueProgram string `json:"value_program,omitempty"`
}

type SourceReceipt struct {
	Schema         string        `json:"schema"`
	SubjectSHA     string        `json:"subject_sha"`
	SourcePath     string        `json:"source_path"`
	SourceSHA256   string        `json:"source_sha256"`
	SemanticSHA256 string        `json:"semantic_sha256"`
	Producer       string        `json:"producer"`
	Consumer       string        `json:"consumer"`
	MetaOperation  string        `json:"meta_operation"`
	ProofChoice    string        `json:"proof_choice"`
	Stage          string        `json:"stage"`
	Step           string        `json:"step"`
	Reason         string        `json:"reason"`
	LowerPipeline  []string      `json:"lower_pipeline"`
	Declarations   []Declaration `json:"declarations"`
	SourcePolicy   SourcePolicy  `json:"source_policy"`
}

type Input struct {
	Subject      string
	SourcePath   string
	Source       []byte
	Contract     []byte
	Receipt      []byte
	Capsule      []byte
	Current      []byte
	Effects      []byte
	Independence []byte
	Conformance  bool
}

type PolicyPredicate struct {
	SourceAuthority           string  `json:"source_authority"`
	ExternalEvidenceRelation  string  `json:"external_evidence_relation"`
	ExternalEvidenceAuthority string  `json:"external_evidence_authority"`
	Claims                    []Claim `json:"claims"`
}

type sourceContract struct {
	Path            string          `json:"path"`
	SHA256          string          `json:"sha256"`
	PolicyPredicate PolicyPredicate `json:"policy_predicate"`
}

type referenceContract struct {
	ID            string `json:"id"`
	Kind          string `json:"kind"`
	URL           string `json:"url"`
	Revision      string `json:"revision"`
	Locator       string `json:"locator"`
	ContentSHA256 string `json:"content_sha256"`
	ClaimID       string `json:"claim_id"`
	Signal        string `json:"signal"`
}

type caseContract struct {
	ID                 string `json:"id"`
	ExpectedDecision   string `json:"expected_decision"`
	ExpectedResolution string `json:"expected_resolution"`
	ExpectedAuthority  string `json:"expected_authority"`
	ExpectedEffect     string `json:"expected_effect"`
}

type contract struct {
	Schema           string         `json:"schema"`
	Version          int            `json:"version"`
	Source           sourceContract `json:"source"`
	FixedDenominator struct {
		Version         string `json:"version"`
		Total           int    `json:"total"`
		BasisPointsGoal int    `json:"basis_points_required"`
	} `json:"fixed_denominator"`
	References []referenceContract `json:"references"`
	Cases      []caseContract      `json:"cases"`
}

type capsuleProposition struct {
	ClaimID string `json:"claim_id"`
	Signal  string `json:"signal"`
}

type capsuleProvenance struct {
	Kind      string `json:"kind"`
	Role      string `json:"role"`
	Authority string `json:"authority"`
}

type capsuleReference struct {
	ID            string             `json:"id"`
	URL           string             `json:"url"`
	Revision      string             `json:"revision"`
	Locator       string             `json:"locator"`
	ContentSHA256 string             `json:"content_sha256"`
	Proposition   capsuleProposition `json:"proposition"`
	Provenance    capsuleProvenance  `json:"provenance"`
}

type capsule struct {
	Schema       string             `json:"schema"`
	CapsuleState string             `json:"capsule_state"`
	CapturedAt   string             `json:"captured_at"`
	References   []capsuleReference `json:"references"`
}

type currentObservation struct {
	ID            string `json:"id"`
	URL           string `json:"url"`
	HTTPStatus    int    `json:"http_status"`
	Bytes         int    `json:"bytes"`
	ContentSHA256 string `json:"content_sha256"`
	Origin        string `json:"origin"`
	CapturedAt    string `json:"captured_at"`
}

type currentSet struct {
	Schema           string               `json:"schema"`
	ObservationState string               `json:"observation_state"`
	References       []currentObservation `json:"references"`
}

type effects struct {
	BeforeStatus      string `json:"before_status"`
	AfterStatus       string `json:"after_status"`
	HeadBefore        string `json:"head_before"`
	HeadAfter         string `json:"head_after"`
	OfficialMutations int    `json:"official_mutations"`
	RepositoryWrites  int    `json:"repository_writes"`
	PromotionCount    int    `json:"promotion_count"`
}

type independence struct {
	ProducerToConsumer int  `json:"producer_to_consumer"`
	ConsumerToProducer int  `json:"consumer_to_producer"`
	Snapshot           bool `json:"-"`
}

type Metric struct {
	Completed   int  `json:"completed"`
	Total       int  `json:"total"`
	BasisPoints int  `json:"basis_points"`
	Satisfied   bool `json:"satisfied"`
}

type FixedDenominator struct {
	SourcePolicy                 Metric `json:"source_policy"`
	ProducerImports              Metric `json:"producer_imports"`
	CurrentReferenceObservations Metric `json:"current_reference_observations"`
	HistoricalFixtures           Metric `json:"historical_fixtures"`
	SemanticCausality            Metric `json:"semantic_causality"`
	NonsemanticPreservation      Metric `json:"nonsemantic_preservation"`
}

type ReferenceResult struct {
	ID             string `json:"id"`
	State          string `json:"state"`
	Agreement      string `json:"agreement"`
	EvidenceDigest string `json:"evidence_digest"`
	Provenance     string `json:"provenance"`
	Stage          string `json:"stage"`
	Step           string `json:"step"`
	Reason         string `json:"reason"`
}

type CurrentResult struct {
	ID             string `json:"id"`
	State          string `json:"state"`
	Resolution     string `json:"resolution"`
	EvidenceDigest string `json:"evidence_digest"`
	Stage          string `json:"stage"`
	Step           string `json:"step"`
	Reason         string `json:"reason"`
}

type PersistentClaim struct {
	ID             string `json:"id"`
	Status         string `json:"status"`
	EvidenceDigest string `json:"evidence_digest"`
	Provenance     string `json:"provenance"`
	Stage          string `json:"stage"`
	Step           string `json:"step"`
	Reason         string `json:"reason"`
}

type ClaimTransition struct {
	ClaimID        string `json:"claim_id"`
	Before         string `json:"before"`
	After          string `json:"after"`
	EvidenceDigest string `json:"evidence_digest"`
	Provenance     string `json:"provenance"`
	Stage          string `json:"stage"`
	Step           string `json:"step"`
	Reason         string `json:"reason"`
	Persisted      bool   `json:"persisted"`
}

type Indicator struct {
	ID            string `json:"id"`
	Class         string `json:"class"`
	ProofChoice   string `json:"proof_choice"`
	Producer      string `json:"producer"`
	Consumer      string `json:"consumer"`
	MetaOperation string `json:"meta_operation"`
	Stage         string `json:"stage"`
	Step          string `json:"step"`
	Status        string `json:"status"`
	Value         int    `json:"value"`
	Target        int    `json:"target"`
	Reason        string `json:"reason"`
}

type Report struct {
	Schema                       string            `json:"schema"`
	SubjectSHA                   string            `json:"subject_sha"`
	Mode                         string            `json:"mode"`
	SourceContractBinding        bool              `json:"source_contract_binding"`
	Decision                     string            `json:"decision"`
	Resolution                   string            `json:"resolution"`
	Reason                       string            `json:"reason"`
	ReferenceAgreement           string            `json:"reference_agreement"`
	SemanticAuthority            string            `json:"semantic_authority"`
	AuthorityGrant               string            `json:"authority_grant"`
	EnforcementEffect            string            `json:"enforcement_effect"`
	SourcePolicy                 SourcePolicy      `json:"source_policy"`
	HistoricalReferences         []ReferenceResult `json:"historical_references"`
	CurrentReferences            []CurrentResult   `json:"current_references"`
	CurrentReferenceObservations int               `json:"current_reference_observations"`
	CurrentReferenceTotal        int               `json:"current_reference_total"`
	CurrentResolution            string            `json:"current_resolution"`
	Completed                    int               `json:"completed"`
	Total                        int               `json:"total"`
	BasisPoints                  int               `json:"basis_points"`
	UnknownIndicators            int               `json:"unknown_indicators"`
	OfficialMutations            int               `json:"official_mutations"`
	RepositoryWrites             int               `json:"repository_writes"`
	PromotionCount               int               `json:"promotion_count"`
	ProducerToConsumerImports    int               `json:"producer_to_consumer_imports"`
	ConsumerToProducerImports    int               `json:"consumer_to_producer_imports"`
	ReadOnly                     bool              `json:"read_only"`
	FixedDenominator             FixedDenominator  `json:"fixed_denominator"`
	Producer                     string            `json:"producer"`
	Consumer                     string            `json:"consumer"`
	MetaOperation                string            `json:"meta_operation"`
	ProofChoice                  string            `json:"proof_choice"`
	Stage                        string            `json:"stage"`
	Step                         string            `json:"step"`
	PersistentClaims             []PersistentClaim `json:"persistent_claims"`
	ClaimTransitions             []ClaimTransition `json:"claim_transitions"`
	Indicators                   []Indicator       `json:"indicators"`
	Receipt                      SourceReceipt     `json:"source_receipt"`
	ReportDigest                 string            `json:"report_digest"`
}

type SuiteCase struct {
	ID                 string `json:"id"`
	ExpectedDecision   string `json:"expected_decision"`
	ExpectedResolution string `json:"expected_resolution"`
	ActualDecision     string `json:"actual_decision"`
	ActualResolution   string `json:"actual_resolution"`
	Authority          string `json:"authority"`
	Effect             string `json:"effect"`
	Passed             bool   `json:"passed"`
}

type Suite struct {
	Schema                  string      `json:"schema"`
	SubjectSHA              string      `json:"subject_sha"`
	Decision                string      `json:"decision"`
	Resolution              string      `json:"resolution"`
	Reason                  string      `json:"reason"`
	CasesTotal              int         `json:"cases_total"`
	CasesSatisfied          int         `json:"cases_satisfied"`
	CoverageBPS             int         `json:"coverage_bps"`
	SemanticCausality       Metric      `json:"semantic_causality"`
	NonsemanticPreservation Metric      `json:"nonsemantic_preservation"`
	Cases                   []SuiteCase `json:"cases"`
	SuiteDigest             string      `json:"suite_digest"`
}

type loweredSource struct {
	SourceDigest   string
	SemanticDigest string
	Declarations   []Declaration
	Policy         SourcePolicy
}

type historicalState struct {
	Results   []ReferenceResult
	Agreement string
	Reason    string
	Complete  int
}

func Judge(input Input) (Report, error) {
	c, err := decodeContract(input.Contract)
	if err != nil {
		return Report{}, err
	}
	lowered, err := lowerSource(input.SourcePath, input.Source)
	if err != nil {
		return Report{}, err
	}
	var receipt SourceReceipt
	if err := json.Unmarshal(input.Receipt, &receipt); err != nil {
		return Report{}, fmt.Errorf("decode producer receipt: %w", err)
	}
	var cap capsule
	if err := json.Unmarshal(input.Capsule, &cap); err != nil {
		return Report{}, fmt.Errorf("decode reference capsule: %w", err)
	}
	var current currentSet
	if len(input.Current) > 0 {
		if err := json.Unmarshal(input.Current, &current); err != nil {
			return Report{}, fmt.Errorf("decode current observations: %w", err)
		}
	}
	var fx effects
	if len(input.Effects) > 0 {
		if err := json.Unmarshal(input.Effects, &fx); err != nil {
			return Report{}, fmt.Errorf("decode effects snapshot: %w", err)
		}
	}
	var indep independence
	if len(input.Independence) > 0 {
		if err := json.Unmarshal(input.Independence, &indep); err != nil {
			return Report{}, fmt.Errorf("decode independence snapshot: %w", err)
		}
		indep.Snapshot = true
	}

	policyExact := samePolicy(lowered.Policy, c.Source.PolicyPredicate)
	sourceBinding := input.SourcePath == c.Source.Path && lowered.SourceDigest == c.Source.SHA256
	receiptBinding := receiptExact(receipt, input.Subject, input.SourcePath, lowered)
	rolesExact := receiptRolesExact(receipt)
	historical := inspectHistorical(c, cap, lowered.Policy)
	currentResults, currentObserved, currentResolution := inspectCurrent(c, current)
	if input.Conformance {
		currentResults, currentObserved, currentResolution = nil, 0, "EXACT"
	}

	decision, resolution, reason, effect := decide(policyExact, receiptBinding && rolesExact, historical, currentResolution, input.Conformance)
	semanticAuthority := lowered.Policy.SourceAuthority
	if semanticAuthority == "" {
		semanticAuthority = "UNKNOWN"
	}
	claims, transitions := makeClaims(lowered, receipt, cap, policyExact, receiptBinding && rolesExact, historical, currentObserved, DigestBytes(input.Current))
	readOnly := fx.RepositoryWrites == 0 && fx.PromotionCount == 0 && fx.BeforeStatus == fx.AfterStatus && fx.HeadBefore == fx.HeadAfter
	fixed := makeFixedDenominator(policyExact, indep, currentObserved, historical.Complete, false, false, readOnly)
	indicators := makeIndicators(policyExact, receipt, indep, currentObserved, historical.Complete, historical.Agreement, claims, false, false, readOnly)
	report := Report{
		Schema: ReportSchema, SubjectSHA: input.Subject, Mode: modeFor(input.Conformance),
		SourceContractBinding: sourceBinding,
		Decision:              decision, Resolution: resolution, Reason: reason,
		ReferenceAgreement: historical.Agreement, SemanticAuthority: semanticAuthority,
		AuthorityGrant: "NONE", EnforcementEffect: effect, SourcePolicy: lowered.Policy,
		HistoricalReferences: historical.Results, CurrentReferences: currentResults,
		CurrentReferenceObservations: currentObserved, CurrentReferenceTotal: len(c.References),
		CurrentResolution: currentResolution, OfficialMutations: fx.OfficialMutations,
		RepositoryWrites: fx.RepositoryWrites, PromotionCount: fx.PromotionCount,
		ProducerToConsumerImports: indep.ProducerToConsumer, ConsumerToProducerImports: indep.ConsumerToProducer,
		ReadOnly: readOnly, FixedDenominator: fixed,
		Producer: "independent-judge", Consumer: "semantic-authority-governor",
		MetaOperation: "separate-reference-agreement-from-authority", ProofChoice: "REGRESSION",
		Stage: "govern", Step: "authority-boundary", PersistentClaims: claims,
		ClaimTransitions: transitions, Indicators: indicators, Receipt: receipt,
	}
	for _, indicator := range report.Indicators {
		if indicator.Status == "SATISFIED" {
			report.Completed++
		}
		if indicator.Status == "OPEN" || indicator.Status == "UNKNOWN" {
			report.UnknownIndicators++
		}
	}
	report.Total = len(report.Indicators)
	if report.Total > 0 {
		report.BasisPoints = report.Completed * 10000 / report.Total
	}
	report.ReportDigest = Digest(reportWithoutDigest(report))
	return report, nil
}

func FinalizeCausality(base, intervention, comment Report) Report {
	semantic := Metric{Total: 1}
	if base.Receipt.SemanticSHA256 != intervention.Receipt.SemanticSHA256 &&
		base.Decision != intervention.Decision && claimStatus(base, "source-intent-authority") != claimStatus(intervention, "source-intent-authority") {
		semantic.Completed, semantic.Satisfied = 1, true
	}
	semantic.BasisPoints = metricBPS(semantic)
	presentation := Metric{Total: 1}
	if base.Receipt.SemanticSHA256 == comment.Receipt.SemanticSHA256 &&
		base.Decision == comment.Decision && base.ReferenceAgreement == comment.ReferenceAgreement &&
		base.AuthorityGrant == comment.AuthorityGrant {
		presentation.Completed, presentation.Satisfied = 1, true
	}
	presentation.BasisPoints = metricBPS(presentation)
	base.FixedDenominator.SemanticCausality = semantic
	base.FixedDenominator.NonsemanticPreservation = presentation
	for index := range base.Indicators {
		switch base.Indicators[index].ID {
		case "semantic-causality":
			base.Indicators[index].Status = statusBool(semantic.Satisfied)
		case "nonsemantic-preservation":
			base.Indicators[index].Status = statusBool(presentation.Satisfied)
		}
		base.Indicators[index].Value = indicatorValue(base.Indicators[index].Status, base.Indicators[index].Target)
	}
	base.Completed, base.UnknownIndicators, base.Total = 0, 0, len(base.Indicators)
	for _, indicator := range base.Indicators {
		if indicator.Status == "SATISFIED" {
			base.Completed++
		}
		if indicator.Status == "OPEN" || indicator.Status == "UNKNOWN" {
			base.UnknownIndicators++
		}
	}
	base.BasisPoints = base.Completed * 10000 / base.Total
	base.ReportDigest = Digest(reportWithoutDigest(base))
	return base
}

func BuildSuite(contractRaw []byte, subject string, agreement, mismatch, absence, intervention, comment Report) (Suite, error) {
	c, err := decodeContract(contractRaw)
	if err != nil {
		return Suite{}, err
	}
	lookup := map[string]Report{
		"reference-agreement": agreement,
		"reference-mismatch":  mismatch,
		"reference-absence":   absence,
	}
	suite := Suite{Schema: SuiteSchema, SubjectSHA: subject, CasesTotal: len(c.Cases)}
	for _, expected := range c.Cases {
		actual, ok := lookup[expected.ID]
		passed := ok && actual.Decision == expected.ExpectedDecision && actual.Resolution == expected.ExpectedResolution &&
			actual.SemanticAuthority == expected.ExpectedAuthority && actual.EnforcementEffect == expected.ExpectedEffect && actual.Decision != "PASS"
		if passed {
			suite.CasesSatisfied++
		}
		suite.Cases = append(suite.Cases, SuiteCase{ID: expected.ID, ExpectedDecision: expected.ExpectedDecision,
			ExpectedResolution: expected.ExpectedResolution, ActualDecision: actual.Decision,
			ActualResolution: actual.Resolution, Authority: actual.SemanticAuthority,
			Effect: actual.EnforcementEffect, Passed: passed})
	}
	if suite.CasesTotal > 0 {
		suite.CoverageBPS = suite.CasesSatisfied * 10000 / suite.CasesTotal
	}
	suite.SemanticCausality = causalMetric(agreement, intervention)
	suite.NonsemanticPreservation = preservationMetric(agreement, comment)
	if suite.CasesSatisfied == suite.CasesTotal && suite.SemanticCausality.Satisfied && suite.NonsemanticPreservation.Satisfied {
		suite.Decision, suite.Resolution, suite.Reason = "HUMILITY_MODEL_BOUND", "EXACT", "CASES_AND_CAUSAL_BOUNDARIES_REPLAYED"
	} else {
		suite.Decision, suite.Resolution, suite.Reason = "FAIL_CLOSED", "INVARIANT_ONLY", "BOUNDARY_OR_CAUSALITY_MISMATCH"
	}
	suite.SuiteDigest = Digest(suiteWithoutDigest(suite))
	return suite, nil
}

func causalMetric(base, intervention Report) Metric {
	metric := Metric{Total: 1}
	if base.Receipt.SemanticSHA256 != intervention.Receipt.SemanticSHA256 && base.Decision != intervention.Decision && claimStatus(base, "source-intent-authority") != claimStatus(intervention, "source-intent-authority") {
		metric.Completed, metric.Satisfied = 1, true
	}
	metric.BasisPoints = metricBPS(metric)
	return metric
}

func preservationMetric(base, comment Report) Metric {
	metric := Metric{Total: 1}
	if base.Receipt.SemanticSHA256 == comment.Receipt.SemanticSHA256 && base.Decision == comment.Decision && base.ReferenceAgreement == comment.ReferenceAgreement && base.AuthorityGrant == comment.AuthorityGrant {
		metric.Completed, metric.Satisfied = 1, true
	}
	metric.BasisPoints = metricBPS(metric)
	return metric
}

func decodeContract(raw []byte) (contract, error) {
	var c contract
	if err := json.Unmarshal(raw, &c); err != nil {
		return contract{}, fmt.Errorf("decode contract: %w", err)
	}
	if c.Schema != ContractSchema || c.Version != 1 || c.FixedDenominator.Total != 12 || c.FixedDenominator.BasisPointsGoal != 10000 || len(c.References) != 3 || len(c.Cases) != 3 {
		return contract{}, fmt.Errorf("contract shape mismatch")
	}
	return c, nil
}

func lowerSource(path string, source []byte) (loweredSource, error) {
	file, diagnostics := syntax.ParseFile(path, string(source))
	if diagnostics.HasErrors() {
		return loweredSource{}, fmt.Errorf("consumer parse %s: syntax diagnostics", path)
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return loweredSource{}, fmt.Errorf("consumer lower %s: %w", path, err)
	}
	declarations := make([]Declaration, 0, len(ir.Graph.Nodes()))
	for _, node := range ir.Graph.Nodes() {
		declarations = append(declarations, Declaration{ID: node.ID.String(), Kind: node.Kind.String(), Name: node.Name, ValueProgram: node.ValueProgram})
	}
	policy, err := policyFromDeclarations(declarations)
	if err != nil {
		return loweredSource{}, err
	}
	return loweredSource{SourceDigest: DigestBytes(source), SemanticDigest: DigestString(ir.SemanticCanonical()), Declarations: declarations, Policy: policy}, nil
}

func policyFromDeclarations(declarations []Declaration) (SourcePolicy, error) {
	var policy SourcePolicy
	for _, declaration := range declarations {
		if declaration.Kind != "Activity" || declaration.ValueProgram == "" {
			continue
		}
		values, err := parseProgram(declaration.ValueProgram)
		if err != nil {
			return SourcePolicy{}, fmt.Errorf("activity %s: %w", declaration.Name, err)
		}
		if declaration.Name == "ComputeSourceAuthorityPolicy" {
			policy.SourceAuthority = values["source_authority"]
			policy.ExternalEvidenceRelation = values["external_evidence_relation"]
			policy.ExternalEvidenceAuthority = values["external_evidence_authority"]
			continue
		}
		if id, okID := values["claim_id"]; okID {
			if state, okState := values["state"]; okState {
				policy.Claims = append(policy.Claims, Claim{ID: id, State: state})
			}
		}
	}
	if policy.SourceAuthority == "" || policy.ExternalEvidenceRelation == "" || policy.ExternalEvidenceAuthority == "" {
		return SourcePolicy{}, fmt.Errorf("source policy computes values are incomplete")
	}
	sort.Slice(policy.Claims, func(i, j int) bool { return policy.Claims[i].ID < policy.Claims[j].ID })
	return policy, nil
}

func parseProgram(program string) (map[string]string, error) {
	values := make(map[string]string)
	for _, item := range strings.Split(program, ";") {
		parts := strings.SplitN(item, "=", 2)
		if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" || strings.TrimSpace(parts[1]) == "" {
			return nil, fmt.Errorf("invalid computes field %q", item)
		}
		key, value := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
		if _, exists := values[key]; exists {
			return nil, fmt.Errorf("duplicate computes field %q", key)
		}
		values[key] = value
	}
	return values, nil
}

func inspectHistorical(c contract, cap capsule, policy SourcePolicy) historicalState {
	state := historicalState{Agreement: "UNKNOWN", Reason: "HISTORICAL_CAPSULE_UNAVAILABLE"}
	if cap.Schema != CapsuleSchema || cap.CapsuleState != "HISTORICAL_FIXTURE" || cap.CapturedAt == "" {
		return state
	}
	byID := make(map[string]capsuleReference, len(cap.References))
	for _, ref := range cap.References {
		byID[ref.ID] = ref
	}
	anyOpen, anyRefuted := false, false
	for _, expected := range c.References {
		actual, ok := byID[expected.ID]
		if !ok {
			anyOpen = true
			state.Results = append(state.Results, ReferenceResult{ID: expected.ID, State: "OPEN", Agreement: "UNKNOWN", Provenance: "HISTORICAL_FIXTURE", Stage: "compare", Step: "capsule-proposition", Reason: "REFERENCE_ABSENT_FROM_CAPSULE"})
			continue
		}
		complete := actual.URL == expected.URL && actual.Revision == expected.Revision && actual.Locator == expected.Locator && actual.ContentSHA256 == expected.ContentSHA256 && actual.Proposition.ClaimID != "" && actual.Proposition.Signal != "" && actual.Provenance.Kind == "PRIMARY_SOURCE" && actual.Provenance.Role == "COMPARATIVE_EVIDENCE" && actual.Provenance.Authority == "NOT_AUTHORITY"
		if complete {
			state.Complete++
		}
		claimOK := policyHasClaim(policy, actual.Proposition.ClaimID)
		agreement := complete && claimOK && actual.Proposition.ClaimID == expected.ClaimID && actual.Proposition.Signal == expected.Signal && policy.ExternalEvidenceRelation == actual.Provenance.Role && policy.ExternalEvidenceAuthority == actual.Provenance.Authority
		result := ReferenceResult{ID: expected.ID, EvidenceDigest: actual.ContentSHA256, Provenance: actual.Provenance.Kind + "/" + actual.Provenance.Role, Stage: "compare", Step: "capsule-proposition"}
		if agreement {
			result.State, result.Agreement, result.Reason = "SATISFIED", "AGREES", "CAPSULE_PROPOSITION_MATCHES_GOOO_POLICY_PREDICATE"
		} else if complete || actual.Proposition.ClaimID != "" {
			anyRefuted = true
			result.State, result.Agreement, result.Reason = "REFUTED", "DISAGREES", "CAPSULE_PROPOSITION_DOES_NOT_MATCH_GOOO_POLICY_PREDICATE"
		} else {
			anyOpen = true
			result.State, result.Agreement, result.Reason = "OPEN", "UNKNOWN", "CAPSULE_PROPOSITION_UNRESOLVED"
		}
		state.Results = append(state.Results, result)
	}
	switch {
	case anyRefuted:
		state.Agreement, state.Reason = "DISAGREES", "EXTERNAL_REFERENCE_DISAGREEMENT_DERIVED"
	case anyOpen || len(state.Results) != len(c.References):
		state.Agreement, state.Reason = "UNKNOWN", "EXTERNAL_REFERENCE_CAPSULE_OPEN"
	case state.Complete == len(c.References):
		state.Agreement, state.Reason = "AGREES", "EXTERNAL_REFERENCE_AGREEMENT_DERIVED_FROM_CAPSULE"
	}
	return state
}

func inspectCurrent(c contract, current currentSet) ([]CurrentResult, int, string) {
	if current.Schema != "gooo/external-oracle-current-observations/v1" || current.ObservationState != "ACTIONS_RETRIEVAL" {
		current.References = nil
	}
	byID := make(map[string]currentObservation, len(current.References))
	for _, observation := range current.References {
		byID[observation.ID] = observation
	}
	results := make([]CurrentResult, 0, len(c.References))
	observed := 0
	for _, expected := range c.References {
		observation, ok := byID[expected.ID]
		result := CurrentResult{ID: expected.ID, State: "OPEN", Resolution: "LOWER_RESOLUTION", Stage: "retrieve", Step: "current-reference"}
		if !ok {
			result.Reason = "CURRENT_REFERENCE_NOT_OBSERVED"
		} else if observation.Origin != "ACTIONS_RETRIEVAL" || !strings.HasPrefix(observation.CapturedAt, "actions-head:") || observation.HTTPStatus != 200 || observation.Bytes <= 0 {
			result.EvidenceDigest = observation.ContentSHA256
			result.Reason = "CURRENT_REFERENCE_RETRIEVAL_UNAVAILABLE"
		} else if observation.URL != expected.URL || observation.ContentSHA256 != expected.ContentSHA256 {
			result.EvidenceDigest = observation.ContentSHA256
			result.Reason = "CURRENT_REFERENCE_CONTENT_DIGEST_CHANGED"
		} else {
			observed++
			result.State, result.Resolution, result.EvidenceDigest, result.Reason = "CURRENT_EVIDENCE", "EXACT", observation.ContentSHA256, "ACTIONS_RETRIEVAL_MATCHES_HISTORICAL_CAPSULE"
		}
		results = append(results, result)
	}
	resolution := "LOWER_RESOLUTION"
	if observed == len(c.References) {
		resolution = "EXACT"
	}
	return results, observed, resolution
}

func decide(policyExact, receiptExact bool, historical historicalState, currentResolution string, conformance bool) (string, string, string, string) {
	if !policyExact {
		return "FAIL_CLOSED", "EXACT", "SOURCE_AUTHORITY_POLICY_CHANGED", "BLOCK"
	}
	if !receiptExact {
		return "FAIL_CLOSED", "EXACT", "SOURCE_RECEIPT_REPLAY_MISMATCH", "BLOCK"
	}
	if historical.Agreement == "DISAGREES" {
		return "FAIL_CLOSED", "EXACT", historical.Reason, "BLOCK"
	}
	if historical.Agreement == "UNKNOWN" {
		return "FAIL_CLOSED", "UNKNOWN", historical.Reason, "BLOCK"
	}
	if conformance {
		return "REFERENCE_AGREEMENT_OBSERVED", "EXACT", "EXTERNAL_AGREEMENT_IS_COMPARATIVE_ONLY", "NO_EFFECT"
	}
	if currentResolution != "EXACT" {
		return "REFERENCE_AGREEMENT_OBSERVED", "LOWER_RESOLUTION", "HISTORICAL_AGREEMENT_CURRENT_OBSERVATION_OPEN", "NO_EFFECT"
	}
	return "REFERENCE_AGREEMENT_OBSERVED", "EXACT", "EXTERNAL_AGREEMENT_IS_COMPARATIVE_ONLY", "NO_EFFECT"
}

func makeClaims(source loweredSource, receipt SourceReceipt, cap capsule, policyExact, receiptExact bool, historical historicalState, currentObserved int, currentDigest string) ([]PersistentClaim, []ClaimTransition) {
	claims := []PersistentClaim{
		{ID: "source-intent-authority", Status: statusFor(policyExact, "REFUTED", "DISCHARGED"), EvidenceDigest: source.SemanticDigest, Provenance: "GOOO_SOURCE/independent-consumer", Stage: "govern", Step: "source-policy", Reason: reasonFor(policyExact, "SOURCE_POLICY_PREDICATE_SATISFIED", "SOURCE_AUTHORITY_POLICY_CHANGED")},
		{ID: "reference-comparison-only", Status: historicalStatus(historical.Agreement), EvidenceDigest: DigestCapsule(cap), Provenance: "HISTORICAL_FIXTURE/PRIMARY_SOURCE", Stage: "compare", Step: "capsule-proposition", Reason: historical.Reason},
		{ID: "receipt-replayability", Status: statusFor(receiptExact, "REFUTED", "DISCHARGED"), EvidenceDigest: Digest(receipt), Provenance: "producer-receipt/independent-consumer", Stage: "replay", Step: "receipt-binding", Reason: reasonFor(receiptExact, "RECEIPT_REPLAY_BOUND", "RECEIPT_REPLAY_MISMATCH")},
		{ID: "external-semantic-authority", Status: "REFUTED", EvidenceDigest: DigestCapsule(cap), Provenance: "GOOO_POLICY_GOVERNOR", Stage: "govern", Step: "refuse-external-authority", Reason: "COMPARISON_CANNOT_PROMOTE_SEMANTIC_AUTHORITY"},
	}
	if currentObserved < 3 {
		claims = append(claims, PersistentClaim{ID: "current-reference-observation", Status: "OPEN", EvidenceDigest: currentDigest, Provenance: "ACTIONS_RETRIEVAL", Stage: "retrieve", Step: "current-reference", Reason: "CURRENT_REFERENCE_OBSERVATION_INCOMPLETE"})
	} else {
		claims = append(claims, PersistentClaim{ID: "current-reference-observation", Status: "DISCHARGED", EvidenceDigest: "CURRENT_EVIDENCE", Provenance: "ACTIONS_RETRIEVAL", Stage: "retrieve", Step: "current-reference", Reason: "CURRENT_REFERENCES_OBSERVED"})
	}
	transitions := make([]ClaimTransition, 0, len(claims))
	for _, claim := range claims {
		transitions = append(transitions, ClaimTransition{ClaimID: claim.ID, Before: "OPEN", After: claim.Status, EvidenceDigest: claim.EvidenceDigest, Provenance: claim.Provenance, Stage: claim.Stage, Step: claim.Step, Reason: claim.Reason, Persisted: claim.ID != "" && claim.Status != "" && claim.Stage != "" && claim.Step != "" && claim.Reason != ""})
	}
	return claims, transitions
}

func makeFixedDenominator(policy bool, indep independence, currentObserved, historicalComplete int, causal, preservation, readOnly bool) FixedDenominator {
	return FixedDenominator{
		SourcePolicy:                 metricFromBool(policy, 1),
		ProducerImports:              metricZero(indep.ProducerToConsumer, indep.ConsumerToProducer, indep.Snapshot),
		CurrentReferenceObservations: metric(currentObserved, 3),
		HistoricalFixtures:           metric(historicalComplete, 3),
		SemanticCausality:            metricFromBool(causal, 1),
		NonsemanticPreservation:      metricFromBool(preservation, 1),
	}
}

func makeIndicators(policy bool, receipt SourceReceipt, indep independence, currentObserved, historicalComplete int, agreement string, claims []PersistentClaim, causal, preservation, readOnly bool) []Indicator {
	independenceOK := len(receipt.LowerPipeline) == 2 && receipt.LowerPipeline[0] == "syntax.ParseFile" && receipt.LowerPipeline[1] == "bidir.Lower" && indep.Snapshot && indep.ProducerToConsumer == 0 && indep.ConsumerToProducer == 0
	receiptOK := receipt.Schema == ReceiptSchema && receipt.Producer == "source-receipt-producer" && receipt.Consumer == "external-oracle-humility-consumer"
	claimOK := len(claims) >= 4 && allTransitionsBound(claims)
	statuses := []string{
		statusBool(policy), statusBool(independenceOK), statusBool(indep.Snapshot && indep.ProducerToConsumer == 0 && indep.ConsumerToProducer == 0), statusBool(receiptOK),
		statusCount(historicalComplete, 3), statusCount(currentObserved, 3), statusAgreement(agreement), statusBool(claimStatusByID(claims, "external-semantic-authority") == "REFUTED"),
		statusBool(claimOK), statusBool(causal), statusBool(preservation), statusBool(readOnly),
	}
	criteria := []struct {
		id, class, proof, producer, consumer, operation, stage, step, reason string
		target                                                               int
	}{
		{"source-policy-structured", "DRIVER", "FOUNDATION", "gooo-source", "independent-consumer", "derive-source-authority-policy", "observe", "source-policy", "SOURCE_POLICY_FROM_COMPUTES", 1},
		{"independent-lowering", "DRIVER", "FOUNDATION", "producer-and-consumer", "independent-consumer", "parse-and-lower-twice", "observe", "syntax-to-ir", "PRODUCER_AND_CONSUMER_USE_PARSEFILE_LOWER", 1},
		{"producer-imports-zero", "GUARDRAIL", "REGRESSION", "dependency-snapshot", "independent-consumer", "guard-package-closure", "guard", "producer-consumer-imports", "PRODUCER_CONSUMER_IMPORTS_ZERO", 0},
		{"source-receipt-replay", "DRIVER", "FOUNDATION", "source-receipt-producer", "independent-consumer", "replay-source-receipt", "replay", "receipt-binding", "RECEIPT_BOUND_TO_RELOWERED_SOURCE", 1},
		{"historical-fixtures", "DRIVER", "COHERENCE", "reference-capsule", "independent-consumer", "validate-historical-capsule", "compare", "capsule-metadata", "THREE_HISTORICAL_FIXTURES", 3},
		{"current-reference-observations", "OUTCOME", "COHERENCE", "actions-retrieval", "independent-consumer", "derive-current-status", "retrieve", "current-reference", "CURRENT_EVIDENCE_ONLY_AFTER_RETRIEVAL", 3},
		{"reference-agreement-derived", "OUTCOME", "COHERENCE", "historical-capsule", "independent-consumer", "compare-proposition-to-policy", "compare", "agreement", "AGREEMENT_DERIVED_NOT_READ_FROM_BOOL", 1},
		{"external-authority-refused", "GUARDRAIL", "REGRESSION", "gooo-policy", "semantic-authority-governor", "refuse-external-authority", "govern", "authority-boundary", "EXTERNAL_REFERENCE_IS_NOT_AUTHORITY", 1},
		{"claim-lifecycle-persisted", "OUTCOME", "COHERENCE", "evidence-ledger", "independent-consumer", "persist-claim-status", "persist", "claim-ledger", "STATUS_DIGEST_PROVENANCE_STAGE_STEP_REASON", 1},
		{"semantic-causality", "OUTCOME", "REGRESSION", "semantic-intervention", "independent-consumer", "compare-policy-intervention", "intervene", "semantic-policy", "POLICY_VALUE_CHANGES_DECISION", 1},
		{"nonsemantic-preservation", "GUARDRAIL", "REGRESSION", "comment-intervention", "independent-consumer", "compare-comment-intervention", "intervene", "comment-only", "COMMENT_PRESERVES_IR_AND_DECISION", 1},
		{"read-only-snapshot", "GUARDRAIL", "REGRESSION", "effects-snapshot", "independent-consumer", "observe-repository-effects", "guard", "read-only", "SNAPSHOT_OBSERVES_ZERO_WRITES_AND_PROMOTIONS", 1},
	}
	result := make([]Indicator, len(criteria))
	for i, criterion := range criteria {
		result[i] = Indicator{ID: criterion.id, Class: criterion.class, ProofChoice: criterion.proof, Producer: criterion.producer, Consumer: criterion.consumer, MetaOperation: criterion.operation, Stage: criterion.stage, Step: criterion.step, Status: statuses[i], Value: indicatorValue(statuses[i], criterion.target), Target: criterion.target, Reason: criterion.reason}
	}
	return result
}

func allTransitionsBound(claims []PersistentClaim) bool {
	for _, claim := range claims {
		if claim.ID == "" || claim.Status == "" || claim.EvidenceDigest == "" || claim.Provenance == "" || claim.Stage == "" || claim.Step == "" || claim.Reason == "" {
			return false
		}
	}
	return true
}

func claimStatus(report Report, id string) string {
	return claimStatusByID(report.PersistentClaims, id)
}
func claimStatusByID(claims []PersistentClaim, id string) string {
	for _, claim := range claims {
		if claim.ID == id {
			return claim.Status
		}
	}
	return ""
}
func historicalStatus(agreement string) string {
	if agreement == "AGREES" {
		return "DISCHARGED"
	}
	if agreement == "DISAGREES" {
		return "REFUTED"
	}
	return "OPEN"
}
func statusFor(ok bool, fail, success string) string {
	if ok {
		return success
	}
	return fail
}
func reasonFor(ok bool, success, fail string) string {
	if ok {
		return success
	}
	return fail
}
func statusBool(ok bool) string {
	if ok {
		return "SATISFIED"
	}
	return "UNSATISFIED"
}
func statusCount(value, target int) string {
	if value == target {
		return "SATISFIED"
	}
	if value < target {
		return "OPEN"
	}
	return "UNSATISFIED"
}
func statusAgreement(value string) string {
	if value == "AGREES" {
		return "SATISFIED"
	}
	if value == "UNKNOWN" {
		return "OPEN"
	}
	return "REFUTED"
}
func indicatorValue(status string, target int) int {
	if status == "SATISFIED" {
		return target
	}
	return 0
}
func samePolicy(a SourcePolicy, b PolicyPredicate) bool {
	return a.SourceAuthority == b.SourceAuthority && a.ExternalEvidenceRelation == b.ExternalEvidenceRelation && a.ExternalEvidenceAuthority == b.ExternalEvidenceAuthority && reflect.DeepEqual(a.Claims, b.Claims)
}
func policyHasClaim(policy SourcePolicy, id string) bool {
	for _, claim := range policy.Claims {
		if claim.ID == id {
			return true
		}
	}
	return false
}
func receiptExact(receipt SourceReceipt, subject, sourcePath string, source loweredSource) bool {
	return receipt.Schema == ReceiptSchema && receipt.SubjectSHA == subject && receipt.SourcePath == sourcePath && receipt.SourceSHA256 == source.SourceDigest && receipt.SemanticSHA256 == source.SemanticDigest && reflect.DeepEqual(receipt.Declarations, source.Declarations) && samePolicy(receipt.SourcePolicy, PolicyPredicate{SourceAuthority: source.Policy.SourceAuthority, ExternalEvidenceRelation: source.Policy.ExternalEvidenceRelation, ExternalEvidenceAuthority: source.Policy.ExternalEvidenceAuthority, Claims: source.Policy.Claims})
}
func receiptRolesExact(receipt SourceReceipt) bool {
	return receipt.Producer == "source-receipt-producer" && receipt.Consumer == "external-oracle-humility-consumer" && receipt.MetaOperation == "emit-source-receipt" && receipt.ProofChoice == "FOUNDATION" && receipt.Stage == "observe" && receipt.Step == "source-receipt" && receipt.Reason == "SOURCE_POLICY_BOUND"
}
func modeFor(conformance bool) string {
	if conformance {
		return "CONFORMANCE_SCENARIO"
	}
	return "SUBJECT_RESOLUTION"
}
func metric(completed, total int) Metric {
	result := Metric{Completed: completed, Total: total, Satisfied: completed == total}
	result.BasisPoints = metricBPS(result)
	return result
}
func metricFromBool(satisfied bool, total int) Metric {
	if satisfied {
		return metric(total, total)
	}
	return metric(0, total)
}
func metricZero(producer, consumer int, present bool) Metric {
	return Metric{Completed: 0, Total: 0, BasisPoints: 10000, Satisfied: present && producer == 0 && consumer == 0}
}
func metricBPS(metric Metric) int {
	if metric.Total == 0 {
		return 10000
	}
	return metric.Completed * 10000 / metric.Total
}
func DigestBytes(value []byte) string {
	sum := sha256.Sum256(value)
	return "sha256:" + hex.EncodeToString(sum[:])
}
func DigestString(value string) string        { return DigestBytes([]byte(value)) }
func Digest(value any) string                 { raw, _ := json.Marshal(value); return DigestBytes(raw) }
func DigestCapsule(value capsule) string      { return Digest(value) }
func reportWithoutDigest(value Report) Report { value.ReportDigest = ""; return value }
func suiteWithoutDigest(value Suite) Suite    { value.SuiteDigest = ""; return value }
func Encode(value any) ([]byte, error) {
	raw, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(raw, '\n'), nil
}
