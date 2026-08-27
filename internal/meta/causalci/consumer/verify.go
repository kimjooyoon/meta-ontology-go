package consumer

import (
	"bytes"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const (
	observationSchema            = "gooo/causal-ci-selection-observation/v2"
	receiptSchema                = "gooo/causal-ci-selection-receipt/v2"
	receiptScope                 = "CAUSAL_SELECTION_PLAN"
	conformancePass              = "PLAN_RECONSTRUCTION_CONFORMANCE_PASS"
	conformanceFailClosed        = "FAIL_CLOSED"
	planGatePass                 = "PLAN_FINAL_GATE_PASS"
	planGateFailClosed           = "PLAN_FINAL_GATE_FAIL_CLOSED"
	resolutionSelected           = "SELECTED"
	resolutionUnknown            = "UNKNOWN"
	resolutionFailClosed         = "FAIL_CLOSED"
	planSelective                = "SELECTIVE_PLAN"
	planFull                     = "DESCEND_TO_FULL_SUITE"
	planNone                     = "NO_PLAN"
	claimOpen                    = "OPEN"
	claimDischarged              = "DISCHARGED"
	claimRefuted                 = "REFUTED"
	proofCausalPath              = "CLAIM_IMPACT_REASON"
	proofFullDescent             = "FULL_SUITE_FALLBACK"
	fixedCheckDenominator        = 6
	fixedIndicatorDenominator    = 6
	executionUnknown             = "UNKNOWN"
	capabilityPlanOnly           = "PLAN_ONLY"
	observedStateUnchanged       = "NET_REPOSITORY_STATE_UNCHANGED"
	observedUnknown              = "UNKNOWN"
	reasonCompleteRoute          = "complete policy route reconstructed"
	reasonClaimDischarged        = "COMPLETE_POLICY_ROUTE_RECONSTRUCTED"
	reasonClaimLowered           = "OPEN_STATE_LOWERED_TO_FULL_SUITE_UNDER_UNKNOWN_PATH"
	reasonUnknownDischarged      = "DISCHARGED_STATE_PERSISTED_UNDER_UNKNOWN_PATH"
	reasonUnknownRefuted         = "REFUTED_STATE_PERSISTED_UNDER_UNKNOWN_PATH"
	reasonPlanOnlyOpen           = "PLAN_ONLY_EXECUTION_NOT_OBSERVED"
	reasonUnrelatedContradiction = "UNRELATED_POLICY_CONTRADICTION_CANNOT_REFUTE"
	reasonClaimRefuted           = "STRUCTURALLY_LINKED_POLICY_CONTRADICTION"
	reasonSourceBinding          = "SOURCE_BYTES_NOT_BOUND_TO_EXACT_HEAD"
)

var fixedCheckIDs = [...]string{"gofmt", "go-vet", "go-test", "go-test-race", "semantic-conformance", "ci-policy"}

const (
	changedFileSemanticID = "gooo://causal-ci-selection/surface/changed-file"
	claimSemanticID       = "gooo://causal-ci-selection/claim/causal-selection"
	surfaceSemanticID     = "gooo://causal-ci-selection/surface/semantic-policy"
	checkSemanticPrefix   = "gooo://causal-ci-selection/check/"
	programChangedFile    = "causal-ci.changed-file-to-claim/v2"
	programClaimSurface   = "causal-ci.claim-to-surface/v2"
	programSurfaceCheck   = "causal-ci.surface-to-check/v2"
	programPriorState     = "causal-ci.prior-claim-state/open/v2"
	programDischarge      = "causal-ci.claim-transition/discharged/v2"
	programLower          = "causal-ci.claim-transition/open-lower-resolution/v2"
	programRefute         = "causal-ci.claim-transition/refuted/v2"
)

// Verify is the independent consumer boundary. It starts with raw source and
// raw observation, then parses, lowers, reconstructs, and compares a receipt.
// This package intentionally does not import the producer package.
func Verify(observationRaw, sourcePath string, source, receiptRaw []byte) error {
	return verify(observationRaw, sourcePath, source, receiptRaw, false)
}

// VerifyIntervention is the same independent replay with an explicitly
// supplied intervention source. It does not pretend those bytes are HEAD.
func VerifyIntervention(observationRaw, sourcePath string, source, receiptRaw []byte) error {
	return verify(observationRaw, sourcePath, source, receiptRaw, true)
}

func verify(observationRaw, sourcePath string, source, receiptRaw []byte, intervention bool) error {
	observationBytes := []byte(observationRaw)
	observation, err := decodeObservation(observationBytes)
	if err != nil {
		return err
	}
	if observation.SourcePath != sourcePath || len(source) == 0 {
		return fmt.Errorf("source binding mismatch")
	}
	actual, err := decodeReceipt(receiptRaw)
	if err != nil {
		return err
	}
	policy, err := reconstructPolicy(sourcePath, source)
	if err != nil {
		return err
	}
	expected := evaluate(observationBytes, observation, policy, source, intervention)
	if expected.Digest, err = receiptDigest(expected); err != nil {
		return err
	}
	if actual.Digest != expected.Digest {
		return fmt.Errorf("receipt digest mismatch")
	}
	if err := validateReceipt(actual, expected, observationBytes, sourcePath, source); err != nil {
		return err
	}
	return nil
}

func decodeObservation(raw []byte) (Observation, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	var value Observation
	if err := decoder.Decode(&value); err != nil {
		return Observation{}, fmt.Errorf("decode raw observation: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return Observation{}, fmt.Errorf("raw observation has trailing JSON")
		}
		return Observation{}, fmt.Errorf("raw observation has trailing bytes: %w", err)
	}
	if value.Schema != observationSchema || value.Repository == "" || value.BaseSHA == "" || value.HeadSHA == "" || value.ObservedCheckoutSHA == "" || filepath.Ext(value.SourcePath) != ".gooo" || (value.ObjectFormat != "sha1" && value.ObjectFormat != "sha256") || value.HeadPathObjectID == "" || value.SourceBytesDigest == "" {
		return Observation{}, fmt.Errorf("malformed raw observation")
	}
	seen := map[string]struct{}{}
	for _, file := range value.ChangedFiles {
		if file.Path == "" || file.Status == "" {
			return Observation{}, fmt.Errorf("malformed changed-file observation")
		}
		if _, exists := seen[file.Path]; exists {
			return Observation{}, fmt.Errorf("duplicate changed path")
		}
		seen[file.Path] = struct{}{}
	}
	seenClaims := map[string]struct{}{}
	for _, claim := range value.PriorClaims {
		if claim.TemplateID == "" || claim.InstanceID == "" || claim.SubjectPath == "" || claim.Proposition == "" || claim.State == "" || claim.Provenance == "" {
			return Observation{}, fmt.Errorf("malformed prior claim observation")
		}
		if claim.InstanceID != claimInstanceID(claim.TemplateID, claim.SubjectPath, claim.Proposition) {
			return Observation{}, fmt.Errorf("content-addressed claim instance mismatch")
		}
		if _, exists := seenClaims[claim.InstanceID]; exists {
			return Observation{}, fmt.Errorf("duplicate claim instance")
		}
		seenClaims[claim.InstanceID] = struct{}{}
	}
	if err := validateSnapshot(value.Isolation.Before); err != nil {
		return Observation{}, err
	}
	if err := validateSnapshot(value.Isolation.After); err != nil {
		return Observation{}, err
	}
	return value, nil
}

func decodeReceipt(raw []byte) (Receipt, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	var value Receipt
	if err := decoder.Decode(&value); err != nil {
		return Receipt{}, fmt.Errorf("decode receipt: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return Receipt{}, fmt.Errorf("receipt has trailing JSON")
		}
		return Receipt{}, fmt.Errorf("receipt has trailing bytes: %w", err)
	}
	return value, nil
}

func validateSnapshot(value RepositorySnapshot) error {
	entries := make([]RepositoryEntry, len(value.Entries))
	copy(entries, value.Entries)
	sort.Slice(entries, func(i, j int) bool { return entries[i].Path < entries[j].Path })
	seen := map[string]struct{}{}
	for _, entry := range entries {
		if entry.Path == "" || entry.Kind == "" || entry.Mode == "" || entry.ContentDigest == "" || (entry.ObjectFormat != "sha1" && entry.ObjectFormat != "sha256") || entry.ObjectID == "" || (entry.Kind == "symlink" && entry.SymlinkTargetDigest == "") {
			return fmt.Errorf("malformed isolation snapshot entry")
		}
		if _, exists := seen[entry.Path]; exists {
			return fmt.Errorf("duplicate isolation snapshot path")
		}
		seen[entry.Path] = struct{}{}
	}
	digest, err := digestJSON(entries)
	if err != nil || digest != value.SnapshotDigest {
		return fmt.Errorf("isolation snapshot digest mismatch")
	}
	return nil
}

func reconstructPolicy(sourcePath string, source []byte) (PolicyGraph, error) {
	file, diagnostics := syntax.ParseFile(sourcePath, string(source))
	if diagnostics.HasErrors() || file == nil {
		return PolicyGraph{}, fmt.Errorf("cannot parse Gooo policy: %s", diagnostics.Error())
	}
	canonical, err := syntax.Format(file)
	if err != nil {
		return PolicyGraph{}, err
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return PolicyGraph{}, fmt.Errorf("cannot lower Gooo policy: %w", err)
	}
	policy := PolicyGraph{Source: SourceEvidence{Path: sourcePath, RawDigest: digestBytes(source), ParsedDigest: digestBytes([]byte(canonical)), SemanticDigest: digestBytes([]byte(ir.SemanticCanonical()))}, ClaimStateRules: map[string]string{}}
	for _, node := range ir.Graph.Nodes() {
		if node.Kind != semantic.Entity {
			continue
		}
		id := string(node.ID)
		switch {
		case id == changedFileSemanticID:
			policy.ChangedFileID = id
		case id == claimSemanticID:
			policy.ClaimID = id
		case id == surfaceSemanticID:
			policy.SurfaceID = id
		case strings.HasPrefix(id, checkSemanticPrefix):
			check, err := parseCheck(id)
			if err != nil {
				return PolicyGraph{}, err
			}
			policy.Checks = append(policy.Checks, check)
		}
	}
	sort.Slice(policy.Checks, func(i, j int) bool { return policy.Checks[i].Ordinal < policy.Checks[j].Ordinal })
	if policy.ChangedFileID == "" || policy.ClaimID == "" || policy.SurfaceID == "" || len(policy.Checks) != fixedCheckDenominator {
		return PolicyGraph{}, fmt.Errorf("Gooo policy entity contract mismatch")
	}
	for index, check := range policy.Checks {
		if check.Ordinal != index+1 || check.ID != fixedCheckIDs[index] {
			return PolicyGraph{}, fmt.Errorf("Gooo check catalog mismatch")
		}
	}

	used := map[string][]string{}
	outputs := map[string][]string{}
	for _, fact := range ir.Graph.Facts() {
		subject, object := string(fact.Subject), string(fact.Object)
		switch fact.Predicate {
		case semantic.Used:
			used[subject] = append(used[subject], object)
		case semantic.WasGeneratedBy:
			outputs[object] = append(outputs[object], subject)
		}
	}
	for _, node := range ir.Graph.Nodes() {
		kind, ok := programKind(node.ValueProgram)
		if !ok {
			continue
		}
		inputs := append([]string(nil), used[string(node.ID)]...)
		results := append([]string(nil), outputs[string(node.ID)]...)
		sort.Strings(inputs)
		sort.Strings(results)
		if len(inputs) != 1 || len(results) != 1 {
			return PolicyGraph{}, fmt.Errorf("policy activity has incomplete typed contract")
		}
		if kind == "claim-state" {
			state := stateFromProgram(node.ValueProgram)
			policy.PriorStates = append(policy.PriorStates, PriorStateRule{State: state, ActivityID: string(node.ID), ValueProgram: node.ValueProgram})
			policy.ClaimStateRules[node.ValueProgram] = state
			continue
		}
		edge := PolicyEdge{Kind: kind, From: inputs[0], To: results[0], ActivityID: string(node.ID), ValueProgram: node.ValueProgram}
		canonicalEdge := edge.Kind + "\x00" + edge.From + "\x00" + edge.To + "\x00" + edge.ActivityID + "\x00" + edge.ValueProgram
		edge.ID = "policy-edge:" + strings.TrimPrefix(digestBytes([]byte(canonicalEdge)), "sha256:")
		policy.Edges = append(policy.Edges, edge)
	}
	validatePolicy(&policy)
	return policy, nil
}

func parseCheck(id string) (Check, error) {
	value := strings.TrimPrefix(id, checkSemanticPrefix)
	if len(value) < 3 || value[2] != '-' {
		return Check{}, fmt.Errorf("malformed check identity")
	}
	ordinal, err := strconv.Atoi(value[:2])
	if err != nil {
		return Check{}, fmt.Errorf("malformed check ordinal")
	}
	return Check{ID: value[3:], Ordinal: ordinal, SemanticID: id}, nil
}

func programKind(value string) (string, bool) {
	switch value {
	case programChangedFile:
		return "changed-file-to-claim", true
	case programClaimSurface:
		return "claim-to-surface", true
	case programSurfaceCheck:
		return "surface-to-check", true
	case programPriorState, programDischarge, programLower, programRefute:
		return "claim-state", true
	default:
		return "", false
	}
}

func stateFromProgram(value string) string {
	switch {
	case value == programPriorState, value == programLower:
		return claimOpen
	case value == programDischarge:
		return claimDischarged
	case value == programRefute:
		return claimRefuted
	default:
		return ""
	}
}

func validatePolicy(policy *PolicyGraph) {
	for _, edge := range policy.Edges {
		if edge.Kind == "changed-file-to-claim" && (edge.From != policy.ChangedFileID || edge.To != policy.ClaimID) {
			policy.Contradictions = append(policy.Contradictions, PolicyContradiction{Stage: "CONFORMANCE", Step: "validate-policy", Reason: "CHANGED_FILE_CLAIM_ENDPOINT_MISMATCH", Edges: []string{edge.ID}})
		}
		if edge.Kind == "claim-to-surface" && (edge.From != policy.ClaimID || edge.To != policy.SurfaceID) {
			policy.Contradictions = append(policy.Contradictions, PolicyContradiction{Stage: "CONFORMANCE", Step: "validate-policy", Reason: "CLAIM_SURFACE_ENDPOINT_MISMATCH", Edges: []string{edge.ID}})
		}
		if edge.Kind == "surface-to-check" && !knownCheck(policy, edge.To) {
			policy.Contradictions = append(policy.Contradictions, PolicyContradiction{Stage: "CONFORMANCE", Step: "validate-policy", Reason: "SURFACE_CHECK_ENDPOINT_UNREGISTERED", Edges: []string{edge.ID}})
		}
	}
	var surfaceEdges []string
	for _, edge := range policy.Edges {
		if edge.Kind == "surface-to-check" && edge.From == policy.SurfaceID {
			surfaceEdges = append(surfaceEdges, edge.ID)
		}
	}
	if len(surfaceEdges) > 1 {
		policy.Contradictions = append(policy.Contradictions, PolicyContradiction{Stage: "CONFORMANCE", Step: "validate-policy", Reason: "CONTRADICTORY_POLICY_PATH", Edges: surfaceEdges})
	}
	requiredStatePrograms := []string{programPriorState, programDischarge, programLower, programRefute}
	for _, program := range requiredStatePrograms {
		if _, exists := policy.ClaimStateRules[program]; !exists {
			policy.Contradictions = append(policy.Contradictions, PolicyContradiction{Stage: "CONFORMANCE", Step: "validate-policy", Reason: "CLAIM_STATE_POLICY_INCOMPLETE"})
		}
	}
	if edgeCount(policy, "changed-file-to-claim") != 1 || edgeCount(policy, "claim-to-surface") != 1 || len(surfaceEdges) != 1 || len(policy.PriorStates) < 4 {
		policy.Contradictions = append(policy.Contradictions, PolicyContradiction{Stage: "CONFORMANCE", Step: "validate-policy", Reason: "REQUIRED_CAUSAL_POLICY_VALUE_MISSING"})
	}
}

func edgeCount(policy *PolicyGraph, kind string) int {
	count := 0
	for _, edge := range policy.Edges {
		if edge.Kind == kind {
			count++
		}
	}
	return count
}

func knownCheck(policy *PolicyGraph, id string) bool {
	for _, check := range policy.Checks {
		if check.SemanticID == id {
			return true
		}
	}
	return false
}

func evaluate(raw []byte, observation Observation, policy PolicyGraph, source []byte, intervention bool) Receipt {
	policy.Source.BindingMode = "HEAD"
	if intervention {
		policy.Source.BindingMode = "INTERVENTION"
	}
	policy.Source.ObservedCheckoutSHA = observation.ObservedCheckoutSHA
	policy.Source.HeadPathObjectID = observation.HeadPathObjectID
	policy.Source.ObjectFormat = observation.ObjectFormat
	policy.Source.ActualSourceObjectID = gitBlobIDForFormat(source, observation.ObjectFormat)
	policy.Source.SourceBytesDigest = digestBytes(source)
	if !intervention && (observation.ObservedCheckoutSHA != observation.HeadSHA || observation.SourceBytesDigest != digestBytes(source) || observation.HeadPathObjectID != policy.Source.ActualSourceObjectID) {
		policy.Contradictions = append([]PolicyContradiction{{Stage: "SOURCE_BINDING", Step: "validate-exact-head", Reason: reasonSourceBinding}}, policy.Contradictions...)
	}
	receipt := Receipt{Schema: receiptSchema, Scope: receiptScope, Source: policy.Source, ObservationDigest: digestBytes(raw), Operation: operation(observation), ExecutionMode: capabilityPlanOnly, Execution: ExecutionStatus{Result: executionUnknown, Capability: capabilityPlanOnly, Coordinate: Coordinate{Stage: "ADJUDICATION", Step: "await-consumer", Reason: "CONSUMER_PROCESS_NOT_RUN"}}, CheckInventory: ExactInventory{ExpectedIDs: fixedCheckIDs[:]}, IndicatorInventory: ExactInventory{ExpectedIDs: indicatorIDs()}, IndependentVerifier: IndependentVerifier{ID: "gooo://consumer/causal-ci-selection", Mode: "INDEPENDENT_RECONSTRUCTION", Required: true, Capability: "SEPARATE_PROCESS"}}
	attachContradictionTargets(&policy, observation)
	for _, file := range sortedFiles(observation.ChangedFiles) {
		if value, ok := targetedContradiction(policy, file.Path); ok {
			receipt.Subjects = append(receipt.Subjects, SubjectResolution{Path: file.Path, Resolution: resolutionFailClosed, Coordinate: Coordinate{Stage: value.Stage, Step: value.Step, Reason: value.Reason}, SelectedChecks: []CheckChoice{}})
		} else if file.Path != observation.SourcePath || file.Status == "D" {
			receipt.Subjects = append(receipt.Subjects, unknown(file, observation.SourcePath))
		} else {
			receipt.Subjects = append(receipt.Subjects, selected(file.Path, policy))
		}
	}
	sort.Slice(receipt.Subjects, func(i, j int) bool { return receipt.Subjects[i].Path < receipt.Subjects[j].Path })
	receipt.ClaimTransitions = transitions(observation, policy, receipt.Subjects, receipt.ObservationDigest)
	receipt.Metrics = metrics(observation, policy, receipt)
	receipt.Indicators = indicators(observation, policy, receipt)
	for _, indicator := range receipt.Indicators {
		if indicator.Satisfied {
			receipt.Metrics.FixedIndicatorSatisfied++
		}
	}
	receipt.CheckInventory.ObservedIDs = checkIDs(policy.Checks)
	receipt.IndicatorInventory.ObservedIDs = indicatorIDsFromPolicy(receipt.Indicators)
	receipt.Metrics.FixedIndicatorSatisfied = satisfiedIndicators(receipt.Indicators)
	receipt.PolicyContradictions = contradictionInventory(policy.Contradictions)
	receipt.PlanGate = derivePlanGate(observation, receipt)
	receipt.Conformance = conformanceFor(policy, receipt.PlanGate)
	receipt.PlanDigest, _ = planDigest(receipt)
	return receipt
}

func conformanceFor(policy PolicyGraph, gate PlanGate) Conformance {
	inventory := contradictionInventory(policy.Contradictions)
	inventoryDigest, _ := digestJSON(inventory)
	conformance := Conformance{RootContradictionInventory: inventory, RootContradictionInventoryDigest: inventoryDigest}
	if len(inventory) > 0 {
		value := inventory[0]
		conformance.Decision = conformanceFailClosed
		conformance.Coordinate = Coordinate{Stage: value.Stage, Step: value.Step, Reason: value.Reason}
		return conformance
	}
	if gate.Decision != planGatePass {
		conformance.Decision = conformanceFailClosed
		conformance.Coordinate = Coordinate{Stage: "CONFORMANCE", Step: "final-plan-gate", Reason: "PLAN_FINAL_GATE_FAILED"}
		return conformance
	}
	conformance.Decision = conformancePass
	conformance.Coordinate = Coordinate{Stage: "CONFORMANCE", Step: "lower", Reason: "GOOO_POLICY_SEMANTIC_GRAPH_RECONSTRUCTED"}
	return conformance
}

func contradictionInventory(values []PolicyContradiction) []PolicyContradiction {
	result := make([]PolicyContradiction, len(values))
	for index, value := range values {
		result[index] = value
		result[index].Edges = sortedUniqueStrings(value.Edges)
		result[index].ClaimInstanceIDs = sortedUniqueStrings(value.ClaimInstanceIDs)
	}
	sort.Slice(result, func(i, j int) bool { return contradictionKey(result[i]) < contradictionKey(result[j]) })
	return result
}

func contradictionKey(value PolicyContradiction) string {
	return strings.Join([]string{value.Stage, value.Step, value.Reason, value.SubjectPath, strings.Join(value.Edges, "\x00"), strings.Join(value.ClaimInstanceIDs, "\x00")}, "\x00")
}

func sortedUniqueStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	result := append([]string(nil), values...)
	sort.Strings(result)
	unique := result[:0]
	for _, value := range result {
		if value == "" || (len(unique) > 0 && unique[len(unique)-1] == value) {
			continue
		}
		unique = append(unique, value)
	}
	return unique
}

func derivePlanGate(observation Observation, receipt Receipt) PlanGate {
	inventoriesExact := sameIDs(receipt.CheckInventory.ExpectedIDs, receipt.CheckInventory.ObservedIDs) && sameIDs(receipt.IndicatorInventory.ExpectedIDs, receipt.IndicatorInventory.ObservedIDs)
	coverageExact := receipt.Metrics.SubjectUniverseCount == len(observation.ChangedFiles) && receipt.Metrics.SubjectCoverageNumerator == receipt.Metrics.SubjectCoverageDenominator && receipt.Metrics.SubjectCoverageDenominator == receipt.Metrics.SubjectUniverseCount && len(receipt.Subjects) == receipt.Metrics.SubjectUniverseCount
	indicatorsValid := validIndicators(receipt.Indicators) && receipt.Metrics.FixedIndicatorDenominator == fixedIndicatorDenominator
	claimsValid := validTransitions(receipt.ClaimTransitions)
	gate := PlanGate{Observed: boolInt(inventoriesExact && coverageExact && indicatorsValid && claimsValid), Denominator: 1, InventoryExact: inventoriesExact, SubjectCoverageExact: coverageExact, IndicatorsValid: indicatorsValid, ClaimTransitionsValid: claimsValid}
	if inventoriesExact && coverageExact && indicatorsValid && claimsValid {
		gate.Decision = planGatePass
	} else {
		gate.Decision = planGateFailClosed
	}
	return gate
}

func validIndicators(values []Indicator) bool {
	if len(values) != fixedIndicatorDenominator {
		return false
	}
	seen := map[string]struct{}{}
	for _, value := range values {
		if value.ID == "" || value.Denominator != 1 || value.Observed != boolInt(value.Satisfied) {
			return false
		}
		if _, exists := seen[value.ID]; exists {
			return false
		}
		seen[value.ID] = struct{}{}
	}
	return sameIDs(indicatorIDs(), indicatorIDsFromPolicy(values))
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func attachContradictionTargets(policy *PolicyGraph, observation Observation) {
	actualEdges := map[string]struct{}{}
	for _, edge := range policy.Edges {
		actualEdges[edge.ID] = struct{}{}
	}
	for index := range policy.Contradictions {
		contradiction := &policy.Contradictions[index]
		var parsedEdges []string
		for _, edgeID := range contradiction.Edges {
			if _, exists := actualEdges[edgeID]; exists {
				parsedEdges = append(parsedEdges, edgeID)
			}
		}
		contradiction.Edges = sortedUniqueStrings(parsedEdges)
		contradiction.ClaimInstanceIDs = nil
		if len(contradiction.Edges) == 0 {
			continue
		}
		if contradiction.SubjectPath == "" {
			contradiction.SubjectPath = observation.SourcePath
		}
		for _, claim := range observation.PriorClaims {
			if claim.SubjectPath == contradiction.SubjectPath && claim.Proposition == reasonCompleteRoute {
				contradiction.ClaimInstanceIDs = append(contradiction.ClaimInstanceIDs, claim.InstanceID)
			}
		}
		contradiction.ClaimInstanceIDs = sortedUniqueStrings(contradiction.ClaimInstanceIDs)
	}
	policy.Contradictions = contradictionInventory(policy.Contradictions)
}

func targetedContradiction(policy PolicyGraph, path string) (PolicyContradiction, bool) {
	for _, contradiction := range policy.Contradictions {
		if contradictionHasActualEdges(policy, contradiction) && contradiction.SubjectPath == path {
			return contradiction, true
		}
	}
	return PolicyContradiction{}, false
}

func contradictionTargetsClaim(policy PolicyGraph, claim PriorClaimObservation) (PolicyContradiction, bool) {
	for _, contradiction := range policy.Contradictions {
		if !contradictionHasActualEdges(policy, contradiction) || contradiction.SubjectPath != claim.SubjectPath {
			continue
		}
		for _, claimID := range contradiction.ClaimInstanceIDs {
			if claimID == claim.InstanceID {
				return contradiction, true
			}
		}
	}
	return PolicyContradiction{}, false
}

func contradictionHasActualEdges(policy PolicyGraph, contradiction PolicyContradiction) bool {
	if len(contradiction.Edges) == 0 {
		return false
	}
	actual := map[string]struct{}{}
	for _, edge := range policy.Edges {
		actual[edge.ID] = struct{}{}
	}
	for _, edgeID := range contradiction.Edges {
		if _, exists := actual[edgeID]; !exists {
			return false
		}
	}
	return true
}

func operation(observation Observation) Operation {
	return Operation{Producer: "gooo://producer/causal-ci-selection", Consumer: "gooo://consumer/causal-ci-selection", MetaOperation: "causal-ci-select", ProofChoice: proofCausalPath, DeclaredPlanCapability: capabilityPlanOnly, ObservedRepositoryState: repositoryState(observation.Isolation)}
}

func selected(path string, policy PolicyGraph) SubjectResolution {
	a := edges(policy, "changed-file-to-claim", policy.ChangedFileID)
	b := edges(policy, "claim-to-surface", policy.ClaimID)
	c := edges(policy, "surface-to-check", policy.SurfaceID)
	if len(a) != 1 || len(b) != 1 || len(c) != 1 || a[0].To != policy.ClaimID || b[0].To != policy.SurfaceID {
		return unknown(ChangedFileObservation{Path: path, Status: "M"}, policy.Source.Path)
	}
	checkID := checkID(&policy, c[0].To)
	if checkID == "" {
		return unknown(ChangedFileObservation{Path: path, Status: "M"}, policy.Source.Path)
	}
	pathEvidence := PathEvidence{SubjectPath: path, ClaimIDs: []string{policy.ClaimID}, Proposition: reasonCompleteRoute, SurfaceID: policy.SurfaceID, CheckID: checkID, PolicyEdgeIDs: []string{a[0].ID, b[0].ID, c[0].ID}, SemanticDigest: policy.Source.SemanticDigest, Explanation: "changed-file observation traverses claim-to-surface and surface-to-check semantic policy", ProofChoice: proofCausalPath}
	return SubjectResolution{Path: path, Resolution: resolutionSelected, Coordinate: Coordinate{Stage: "CAUSAL_SELECTION", Step: "select-checks", Reason: "COMPLETE_CLAIM_SURFACE_CHECK_PATH"}, Paths: []PathEvidence{pathEvidence}, SelectedChecks: []CheckChoice{{CheckID: checkID, ProofChoice: proofCausalPath, Reason: "COMPLETE_CLAIM_SURFACE_CHECK_PATH", ClaimIDs: []string{policy.ClaimID}, PathIDs: pathEvidence.PolicyEdgeIDs}}}
}

func unknown(file ChangedFileObservation, sourcePath string) SubjectResolution {
	reason := "SOURCE_NOT_BOUND_TO_POLICY"
	if file.Path == sourcePath && file.Status == "D" {
		reason = "SOURCE_OBJECT_NOT_AVAILABLE"
	}
	choices := make([]CheckChoice, 0, fixedCheckDenominator)
	for _, id := range fixedCheckIDs {
		choices = append(choices, CheckChoice{CheckID: id, ProofChoice: proofFullDescent, Reason: reason})
	}
	return SubjectResolution{Path: file.Path, Resolution: resolutionUnknown, Coordinate: Coordinate{Stage: "CAUSAL_SELECTION", Step: "descend-full-suite", Reason: reason}, UnknownCauses: []UnknownCause{{SubjectPath: file.Path, Coordinate: Coordinate{Stage: "CAUSAL_SELECTION", Step: "observe-subject", Reason: reason}, Provenance: "git://pull-request/changed-file/" + file.Path}}, SelectedChecks: choices}
}

func edges(policy PolicyGraph, kind, from string) []PolicyEdge {
	result := []PolicyEdge{}
	for _, edge := range policy.Edges {
		if edge.Kind == kind && edge.From == from {
			result = append(result, edge)
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result
}

func checkID(policy *PolicyGraph, id string) string {
	for _, check := range policy.Checks {
		if check.SemanticID == id {
			return check.ID
		}
	}
	return ""
}

func transitions(observation Observation, policy PolicyGraph, subjects []SubjectResolution, observationDigest string) []ClaimTransition {
	byPath := map[string]SubjectResolution{}
	for _, subject := range subjects {
		byPath[subject.Path] = subject
	}
	prior := sortedClaims(observation.PriorClaims)
	result := make([]ClaimTransition, 0, len(prior))
	previous := ""
	for index, claim := range prior {
		after, resolution, reason := claim.State, planNone, reasonClaimLowered
		stage, step := "CLAIM_LEDGER", "append-transition"
		if contradiction, linked := contradictionTargetsClaim(policy, claim); linked {
			after, resolution, reason, stage, step = claimRefuted, planNone, reasonClaimRefuted, contradiction.Stage, contradiction.Step
		} else if subject, exists := byPath[claim.SubjectPath]; exists {
			switch subject.Resolution {
			case resolutionSelected:
				resolution = planSelective
				if claim.Proposition == reasonCompleteRoute {
					after, reason = claimDischarged, reasonClaimDischarged
				} else {
					after, reason = claimOpen, reasonPlanOnlyOpen
				}
			case resolutionUnknown:
				resolution = planFull
				switch claim.State {
				case claimOpen:
					reason = reasonClaimLowered
				case claimDischarged:
					reason = reasonUnknownDischarged
				case claimRefuted:
					reason = reasonUnknownRefuted
				}
			case resolutionFailClosed:
				resolution = planNone
				reason = reasonUnrelatedContradiction
			}
		}
		if claim.State == claimRefuted {
			after = claimRefuted
		}
		evidence, _ := digestJSON(struct {
			Observation string                `json:"observation"`
			Source      string                `json:"source"`
			Claim       PriorClaimObservation `json:"claim"`
			After       string                `json:"after"`
			Resolution  string                `json:"resolution"`
		}{observationDigest, policy.Source.SemanticDigest, claim, after, resolution})
		value := ClaimTransition{Sequence: index + 1, TemplateID: claim.TemplateID, ClaimID: claim.InstanceID, SubjectPath: claim.SubjectPath, Proposition: claim.Proposition, Before: claim.State, After: after, Resolution: resolution, Stage: stage, Step: step, Reason: reason, EvidenceDigest: evidence, Provenance: claim.Provenance, PreviousDigest: previous}
		value.Digest, _ = transitionDigest(value)
		result = append(result, value)
		previous = value.Digest
	}
	return result
}

func metrics(observation Observation, policy PolicyGraph, receipt Receipt) Metrics {
	paths := make([]string, 0, len(observation.ChangedFiles))
	for _, file := range sortedFiles(observation.ChangedFiles) {
		paths = append(paths, file.Path)
	}
	universe, _ := digestJSON(paths)
	value := Metrics{SubjectUniverseDigest: universe, SubjectUniverseCount: len(paths), SubjectCoverageNumerator: len(receipt.Subjects), SubjectCoverageDenominator: len(paths), SubjectTotal: len(receipt.Subjects), FullSuiteCheckDenominator: len(policy.Checks), ClaimTransitionTotal: len(receipt.ClaimTransitions), FixedIndicatorDenominator: fixedIndicatorDenominator}
	for _, subject := range receipt.Subjects {
		value.SelectedCheckTotal += len(subject.SelectedChecks)
		switch subject.Resolution {
		case resolutionSelected:
			value.SelectedSubjectTotal++
		case resolutionUnknown:
			value.UnknownSubjectTotal++
		case resolutionFailClosed:
			value.FailClosedSubjectTotal++
		}
	}
	for _, transition := range receipt.ClaimTransitions {
		if transition.After == claimDischarged {
			value.DischargedClaimTotal++
		}
		if transition.After == claimRefuted {
			value.RefutedClaimTotal++
		}
		if transition.Resolution == planFull {
			value.LowerResolutionClaimTotal++
		}
	}
	return value
}

func indicators(observation Observation, policy PolicyGraph, receipt Receipt) []Indicator {
	state := receipt.Operation.ObservedRepositoryState
	values := []bool{
		policy.Source.ParsedDigest != "" && policy.Source.SemanticDigest != "",
		len(receipt.Subjects) == len(observation.ChangedFiles),
		unknownSubjectsDescendFully(receipt.Subjects),
		validTransitions(receipt.ClaimTransitions),
		state.NetState == observedStateUnchanged && state.ChangedPathCount == 0 && state.ChangedContentCount == 0 && state.TransientWrites == observedUnknown && state.GlobalMutationAuthority == observedUnknown,
		receipt.ExecutionMode == capabilityPlanOnly && receipt.Execution.Result == executionUnknown,
	}
	result := make([]Indicator, 0, len(values))
	ids := indicatorIDs()
	for index, id := range ids {
		observed := 0
		if values[index] {
			observed = 1
		}
		result = append(result, Indicator{ID: id, Observed: observed, Denominator: 1, Satisfied: values[index]})
	}
	return result
}

func unknownSubjectsDescendFully(subjects []SubjectResolution) bool {
	for _, subject := range subjects {
		if subject.Resolution != resolutionUnknown {
			continue
		}
		if len(subject.SelectedChecks) != fixedCheckDenominator {
			return false
		}
		for index, choice := range subject.SelectedChecks {
			if choice.CheckID != fixedCheckIDs[index] || choice.ProofChoice != proofFullDescent || choice.Reason == "" {
				return false
			}
		}
	}
	return true
}

func validTransitions(values []ClaimTransition) bool {
	previous := ""
	for index, value := range values {
		computed, err := transitionDigest(value)
		if err != nil || value.Sequence != index+1 || value.PreviousDigest != previous || computed != value.Digest || value.EvidenceDigest == "" || value.Provenance == "" || value.TemplateID == "" || value.Proposition == "" || value.ClaimID != claimInstanceID(value.TemplateID, value.SubjectPath, value.Proposition) {
			return false
		}
		previous = value.Digest
	}
	return true
}

func planDigest(value Receipt) (string, error) {
	projection := struct {
		ObservationDigest string                `json:"observation_digest"`
		Operation         Operation             `json:"operation"`
		ExecutionMode     string                `json:"execution_mode"`
		Conformance       Conformance           `json:"conformance"`
		PlanGate          PlanGate              `json:"plan_gate"`
		Contradictions    []PolicyContradiction `json:"policy_contradictions"`
		Subjects          []SubjectResolution   `json:"subjects"`
		ClaimTransitions  []ClaimTransition     `json:"claim_transitions"`
	}{value.ObservationDigest, value.Operation, value.ExecutionMode, value.Conformance, value.PlanGate, value.PolicyContradictions, value.Subjects, value.ClaimTransitions}
	return digestJSON(projection)
}

func receiptDigest(value Receipt) (string, error) {
	value.Digest = ""
	return digestJSON(value)
}

func transitionDigest(value ClaimTransition) (string, error) {
	value.Digest = ""
	return digestJSON(value)
}

func digestBytes(data []byte) string {
	var value [32]byte
	value = sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(value[:])
}

func digestJSON(value any) (string, error) {
	raw, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return digestBytes(raw), nil
}

func repositoryState(value IsolationObservation) RepositoryStateObservation {
	before := map[string]RepositoryEntry{}
	for _, entry := range value.Before.Entries {
		before[entry.Path] = entry
	}
	after := map[string]RepositoryEntry{}
	for _, entry := range value.After.Entries {
		after[entry.Path] = entry
	}
	paths := map[string]struct{}{}
	for path := range before {
		paths[path] = struct{}{}
	}
	for path := range after {
		paths[path] = struct{}{}
	}
	changedPaths, changedContents := 0, 0
	for path := range paths {
		left, leftOK := before[path]
		right, rightOK := after[path]
		if !leftOK || !rightOK {
			changedPaths++
			changedContents++
			continue
		}
		if !sameRepositoryEntry(left, right) {
			changedPaths++
			changedContents++
		}
	}
	state := "NET_REPOSITORY_STATE_CHANGED"
	if changedPaths == 0 && changedContents == 0 {
		state = observedStateUnchanged
	}
	return RepositoryStateObservation{NetState: state, ChangedPathCount: changedPaths, ChangedContentCount: changedContents, TransientWrites: observedUnknown, GlobalMutationAuthority: observedUnknown}
}

func sameRepositoryEntry(left, right RepositoryEntry) bool {
	return left.Path == right.Path && left.Tracked == right.Tracked && left.Kind == right.Kind && left.Mode == right.Mode && left.SymlinkTargetDigest == right.SymlinkTargetDigest && left.ContentDigest == right.ContentDigest && left.ObjectFormat == right.ObjectFormat && left.ObjectID == right.ObjectID
}

func sortedFiles(values []ChangedFileObservation) []ChangedFileObservation {
	result := append([]ChangedFileObservation(nil), values...)
	sort.Slice(result, func(i, j int) bool { return result[i].Path < result[j].Path })
	return result
}

func sortedClaims(values []PriorClaimObservation) []PriorClaimObservation {
	result := append([]PriorClaimObservation(nil), values...)
	sort.Slice(result, func(i, j int) bool {
		if result[i].SubjectPath != result[j].SubjectPath {
			return result[i].SubjectPath < result[j].SubjectPath
		}
		return result[i].InstanceID < result[j].InstanceID
	})
	return result
}

func validateReceipt(actual, expected Receipt, observationRaw []byte, sourcePath string, source []byte) error {
	if actual.Schema != receiptSchema || actual.Scope != receiptScope || actual.Source.Path != sourcePath || actual.Source.RawDigest != digestBytes(source) || actual.Source.SourceBytesDigest != digestBytes(source) || actual.Source.ObjectFormat == "" || actual.Source.ActualSourceObjectID != gitBlobIDForFormat(source, actual.Source.ObjectFormat) || actual.ObservationDigest != digestBytes(observationRaw) {
		return fmt.Errorf("receipt source binding mismatch")
	}
	if actual.ExecutionMode != capabilityPlanOnly || actual.Execution.Result != executionUnknown || actual.IndependentVerifier.ID != "gooo://consumer/causal-ci-selection" || actual.IndependentVerifier.Mode != "INDEPENDENT_RECONSTRUCTION" || !actual.IndependentVerifier.Required || actual.IndependentVerifier.Capability != "SEPARATE_PROCESS" || !sameIDs(actual.CheckInventory.ExpectedIDs, actual.CheckInventory.ObservedIDs) || !sameIDs(actual.IndicatorInventory.ExpectedIDs, actual.IndicatorInventory.ObservedIDs) {
		return fmt.Errorf("independent consumer declaration mismatch")
	}
	if !validRootContradictionInventory(actual.Conformance, expected.Conformance.RootContradictionInventory) {
		return fmt.Errorf("root contradiction inventory mismatch")
	}
	if !equalJSON(actual, expected) {
		return fmt.Errorf("independent reconstruction mismatch")
	}
	return nil
}

func validRootContradictionInventory(actual Conformance, expected []PolicyContradiction) bool {
	if len(actual.RootContradictionInventory) != len(expected) {
		return false
	}
	seen := map[string]struct{}{}
	for index, value := range actual.RootContradictionInventory {
		key := contradictionKey(value)
		if _, exists := seen[key]; exists || key != contradictionKey(expected[index]) {
			return false
		}
		seen[key] = struct{}{}
	}
	digest, err := digestJSON(actual.RootContradictionInventory)
	return err == nil && digest == actual.RootContradictionInventoryDigest
}

func indicatorIDs() []string {
	return []string{"semantic-policy-derived", "changed-file-observation-bound", "unknown-descends-to-full", "claim-transition-append-only", "isolation-state-observed", "plan-only-no-execution-claim"}
}

func checkIDs(values []Check) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		result = append(result, value.ID)
	}
	return result
}

func indicatorIDsFromPolicy(values []Indicator) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		result = append(result, value.ID)
	}
	return result
}

func sameIDs(expected, observed []string) bool {
	if len(expected) != len(observed) {
		return false
	}
	left, right := append([]string(nil), expected...), append([]string(nil), observed...)
	sort.Strings(left)
	sort.Strings(right)
	for index := range left {
		if left[index] == "" || right[index] == "" || left[index] != right[index] {
			return false
		}
		if index > 0 && left[index] == left[index-1] {
			return false
		}
	}
	return true
}

func satisfiedIndicators(values []Indicator) int {
	result := 0
	for _, value := range values {
		if value.Satisfied {
			result++
		}
	}
	return result
}

func claimInstanceID(templateID, subjectPath, proposition string) string {
	digest, _ := digestJSON(struct {
		TemplateID  string `json:"template_id"`
		SubjectPath string `json:"subject_path"`
		Proposition string `json:"proposition"`
	}{templateID, subjectPath, proposition})
	return "claim-instance:" + strings.TrimPrefix(digest, "sha256:")
}

func gitBlobID(data []byte) string {
	return gitBlobIDForFormat(data, "sha1")
}

func gitBlobIDForFormat(data []byte, format string) string {
	header := []byte(fmt.Sprintf("blob %d\x00", len(data)))
	payload := append(header, data...)
	if strings.EqualFold(format, "sha256") {
		sum := sha256.Sum256(payload)
		return hex.EncodeToString(sum[:])
	}
	sum := sha1.Sum(payload)
	return hex.EncodeToString(sum[:])
}

func equalJSON(left, right any) bool {
	a, errA := json.Marshal(left)
	b, errB := json.Marshal(right)
	return errA == nil && errB == nil && bytes.Equal(a, b)
}
