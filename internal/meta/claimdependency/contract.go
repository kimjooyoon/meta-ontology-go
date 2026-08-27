package claimdependency

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const (
	ValidatorContractSchema  = "gooo.meta.claim-dependency-validator-contract/v2"
	StructuralManifestSchema = "gooo.meta.claim-dependency-structural-manifest/v1"
	FailureReceiptSchema     = "gooo.meta.claim-dependency-failure-receipt/v2"
	failureProcedure         = "CI_EDGE_SPECIFIC_FAILURE_COMPARATOR_V2"
)

type sourceGraph struct {
	IR    semantic.IR
	Graph Graph
}

// GraphFromSource reconstructs executable propositions and typed dependency
// edges from syntax.ParseFile -> bidir.Lower canonical IR. Case names do not
// participate in the graph or in the final state decision.
func GraphFromSource(source []byte, sourcePath string) (Graph, error) {
	parsed, err := graphFromSource(source, sourcePath)
	if err != nil {
		return Graph{}, err
	}
	return parsed.Graph, nil
}

func graphFromSource(source []byte, sourcePath string) (sourceGraph, error) {
	file, diagnostics := syntax.ParseFile(sourcePath, string(source))
	if file == nil || diagnostics.HasErrors() {
		return sourceGraph{}, fmt.Errorf("source parse failed: %s", diagnostics.Error())
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return sourceGraph{}, fmt.Errorf("source lower failed: %w", err)
	}
	if err := ir.Validate(); err != nil {
		return sourceGraph{}, fmt.Errorf("canonical IR invalid: %w", err)
	}
	activities := map[string]semantic.Node{}
	for _, node := range ir.Graph.Nodes() {
		if node.Kind == semantic.Activity {
			activities[node.Name] = node
		}
	}
	generatedBy, usedBy := map[string]string{}, map[string][]string{}
	for _, fact := range ir.Graph.AllFacts() {
		switch fact.Predicate {
		case semantic.WasGeneratedBy:
			generatedBy[fact.Subject.String()] = fact.Object.String()
		case semantic.Used:
			usedBy[fact.Subject.String()] = append(usedBy[fact.Subject.String()], fact.Object.String())
		}
	}
	claims, activityIndex := make([]Claim, 0, ClaimTotal), map[string]int{}
	activityNames := map[string]bool{}
	for _, declaration := range file.Declarations {
		activity, ok := declaration.(*syntax.ActivityDecl)
		if !ok {
			continue
		}
		if activityNames[activity.Name] {
			return sourceGraph{}, fmt.Errorf("duplicate activity %q is not a unique semantic occurrence", activity.Name)
		}
		activityNames[activity.Name] = true
		node, ok := activities[activity.Name]
		if !ok || node.ValueProgram == "" {
			return sourceGraph{}, fmt.Errorf("activity %q is not a semantic value claim", activity.Name)
		}
		inputs := append([]string(nil), usedBy[node.ID.String()]...)
		sort.Strings(inputs)
		output := ""
		for entityID, activityID := range generatedBy {
			if activityID == node.ID.String() {
				output = entityID
				break
			}
		}
		artifact := "gooo://claim-dependency/artifact/" + strings.ToLower(activity.Name)
		proposition := normalizedProposition(node.ID.String(), inputs, output, artifact, node.ValueProgram)
		activityIndex[node.ID.String()] = len(claims)
		claims = append(claims, Claim{Ordinal: len(claims) + 1, Axis: strings.ToLower(activity.Name), ClaimID: node.ID.String(), ActivityID: node.ID.String(), ActivityName: activity.Name, Proposition: proposition, PropositionDigest: digestBytes([]byte(proposition)), Target: TargetAddress{Inputs: inputs, Output: output, Artifact: artifact}, ValueProgram: node.ValueProgram, Producer: ProducerID, Consumer: ConsumerID, MetaOperation: MetaOperationID, ProofChoice: ProofChoice, Coordinate: Coordinate{Stage: "CLAIM", Step: activity.Name, Reason: "NORMALIZED_EXECUTION_PROPOSITION"}})
	}
	if len(claims) != ClaimTotal {
		return sourceGraph{}, fmt.Errorf("source must contain exactly %d activity claims, got %d", ClaimTotal, len(claims))
	}
	digests := map[string]bool{}
	for _, claim := range claims {
		digests[claim.PropositionDigest] = true
	}
	if len(digests) != ClaimTotal {
		return sourceGraph{}, fmt.Errorf("claims do not contain %d distinct observed propositions", ClaimTotal)
	}
	type edgeCandidate struct {
		from, to int
		kind     EdgeKind
	}
	candidates := make([]edgeCandidate, 0, EdgeTotal)
	for downstreamID, entities := range usedBy {
		to, ok := activityIndex[downstreamID]
		if !ok {
			continue
		}
		for _, entityID := range entities {
			upstreamID, ok := generatedBy[entityID]
			if !ok {
				continue
			}
			from, ok := activityIndex[upstreamID]
			if !ok || from == to {
				continue
			}
			kind, ok := edgeKind(claims[to].ValueProgram)
			if !ok {
				return sourceGraph{}, fmt.Errorf("activity %q does not declare a typed dependency edge", claims[to].ActivityName)
			}
			candidates = append(candidates, edgeCandidate{from: from, to: to, kind: kind})
		}
	}
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].from != candidates[j].from {
			return candidates[i].from < candidates[j].from
		}
		if candidates[i].to != candidates[j].to {
			return candidates[i].to < candidates[j].to
		}
		return candidates[i].kind < candidates[j].kind
	})
	if len(candidates) != EdgeTotal {
		return sourceGraph{}, fmt.Errorf("semantic activity relations must reconstruct exactly %d edges, got %d", EdgeTotal, len(candidates))
	}
	edges := make([]Edge, len(candidates))
	for i, c := range candidates {
		edges[i] = Edge{EdgeID: fmt.Sprintf("E%02d", i+1), FromClaimID: claims[c.from].ClaimID, ToClaimID: claims[c.to].ClaimID, Kind: c.kind, ActivationPredicate: activationPredicate(c.kind), SemanticBasis: "prov:wasGeneratedBy + prov:used + source-derived value-program edge predicate"}
	}
	graph := Graph{Schema: GraphSchema, Authority: "CANONICAL_IR_FROM_SYNTAX_PARSE_AND_BIDIR_LOWER", Completeness: "CLOSED_WORLD_SOURCE_RECONSTRUCTED", CanonicalIRDigest: prefixedDigest(ir.StableHash()), NodeTotal: len(claims), EdgeTotal: len(edges), Nodes: claims, Edges: edges}
	graph.Digest, err = graphDigest(graph)
	if err != nil {
		return sourceGraph{}, err
	}
	return sourceGraph{IR: ir, Graph: graph}, nil
}

func activationPredicate(kind EdgeKind) string {
	switch kind {
	case Requires:
		return "UPSTREAM_DISCHARGED_AND_LOCAL_EVIDENCE"
	case Contradicts:
		return "UPSTREAM_DISCHARGED_PROPOSITION"
	case FailureEntailment:
		return "OBSERVED_FAILURE_ANTECEDENT_AND_UPSTREAM_REFUTED"
	default:
		return "UPSTREAM_UNKNOWN_BLOCKS_ONLY"
	}
}

func normalizedProposition(activityID string, inputs []string, output, artifact, value string) string {
	return fmt.Sprintf("execute(activity=%s,inputs=[%s],output=%s,artifact=%s,value=%s)", activityID, strings.Join(inputs, ","), output, artifact, value)
}

func edgeKind(program string) (EdgeKind, bool) {
	const prefix = "claim.edge:"
	if !strings.HasPrefix(program, prefix) {
		return "", false
	}
	value := strings.TrimPrefix(program, prefix)
	if pipe := strings.IndexByte(value, '|'); pipe >= 0 {
		value = value[:pipe]
	}
	switch value {
	case "supports":
		return Supports, true
	case "requires":
		return Requires, true
	case "contradicts":
		return Contradicts, true
	case "failure-entailment":
		return FailureEntailment, true
	}
	return "", false
}

func TruthTableCases() []TruthTableCase {
	return []TruthTableCase{
		{Schema: TruthTableSchema, CaseID: "SUPPORTS-POSITIVE", Kind: Supports, Direction: "direction-matching", UpstreamState: "DISCHARGED", LocalPredicate: "EVIDENCE_ACCEPTED", ExpectedState: "OPEN", Positive: true, SemanticBasis: "support never discharges or refutes by itself"},
		{Schema: TruthTableSchema, CaseID: "SUPPORTS-REVERSED", Kind: Supports, Direction: "reversed-direction", UpstreamState: "REFUTED", LocalPredicate: "UNKNOWN", ExpectedState: "OPEN", Positive: false, SemanticBasis: "support does not refute"},
		{Schema: TruthTableSchema, CaseID: "REQUIRES-POSITIVE", Kind: Requires, Direction: "direction-matching", UpstreamState: "DISCHARGED", LocalPredicate: "EVIDENCE_ACCEPTED", ExpectedState: "DISCHARGED", Positive: true, SemanticBasis: "upstream and local requirement hold"},
		{Schema: TruthTableSchema, CaseID: "REQUIRES-UNKNOWN", Kind: Requires, Direction: "direction-matching", UpstreamState: "DISCHARGED", LocalPredicate: "UNKNOWN", ExpectedState: "OPEN", Positive: true, SemanticBasis: "local requirement evidence is unknown"},
		{Schema: TruthTableSchema, CaseID: "CONTRADICTS-POSITIVE", Kind: Contradicts, Direction: "direction-matching", UpstreamState: "DISCHARGED", LocalPredicate: "UNKNOWN", ExpectedState: "REFUTED", Positive: true, ContradictionObserved: true, SemanticBasis: "established proposition and observed contradiction agree in the declared direction"},
		{Schema: TruthTableSchema, CaseID: "CONTRADICTS-REVERSED", Kind: Contradicts, Direction: "reversed-direction", UpstreamState: "DISCHARGED", LocalPredicate: "UNKNOWN", ExpectedState: "OPEN", Positive: false, ContradictionObserved: true, SemanticBasis: "reversed contradiction direction cannot refute"},
		{Schema: TruthTableSchema, CaseID: "FAILURE_ENTAILMENT-POSITIVE", Kind: FailureEntailment, Direction: "direction-matching", UpstreamState: "REFUTED", LocalPredicate: "UNKNOWN", ExpectedState: "REFUTED", Positive: true, FailureAntecedentObserved: true, SemanticBasis: "declared failure antecedent is observed and entails target failure"},
		{Schema: TruthTableSchema, CaseID: "FAILURE_ENTAILMENT-UNKNOWN", Kind: FailureEntailment, Direction: "direction-matching", UpstreamState: "REFUTED", LocalPredicate: "UNKNOWN", ExpectedState: "OPEN", Positive: true, FailureAntecedentObserved: false, SemanticBasis: "upstream refutation alone does not activate failure entailment"},
	}
}

type relationOutcome string

const (
	relationOpen       relationOutcome = "OPEN"
	relationDischarged relationOutcome = "DISCHARGED"
	relationRefuted    relationOutcome = "REFUTED"
)

// edgeRelation is the single state relation used by both the executable
// truth-table cases and the graph classifier. directionMatches is explicit so
// the negative cases exercise the same relation with a reversed edge.
func edgeRelation(kind EdgeKind, upstreamState string, local ObservationPredicate, directionMatches, contradictionObserved, failureAntecedentObserved bool) relationOutcome {
	if !directionMatches {
		return relationOpen
	}
	switch kind {
	case Requires:
		if upstreamState == "DISCHARGED" && local == ObservationEvidence {
			return relationDischarged
		}
	case Contradicts:
		if upstreamState == "DISCHARGED" && contradictionObserved {
			return relationRefuted
		}
	case FailureEntailment:
		if upstreamState == "REFUTED" && failureAntecedentObserved {
			return relationRefuted
		}
	case Supports:
		// SUPPORTS can block an OPEN claim when its upstream is unresolved,
		// but it can never discharge or refute the target by itself.
	}
	return relationOpen
}

// validateTruthTable executes the closed edge algebra over each positive and
// negative case before the table is emitted. The table is therefore an
// executable counterexample set, not a list of unchecked labels.
func validateTruthTable(cases []TruthTableCase) error {
	if len(cases) != 2*len(EdgeKinds()) {
		return fmt.Errorf("truth table has %d cases, want %d", len(cases), 2*len(EdgeKinds()))
	}
	seen := map[EdgeKind]int{}
	for _, test := range cases {
		actual := edgeRelation(test.Kind, test.UpstreamState, ObservationPredicate(test.LocalPredicate), test.Positive, test.ContradictionObserved, test.FailureAntecedentObserved)
		if string(actual) != test.ExpectedState {
			return fmt.Errorf("truth table case %q computed %s, expected %s", test.CaseID, actual, test.ExpectedState)
		}
		seen[test.Kind]++
	}
	for _, kind := range EdgeKinds() {
		if seen[kind] != 2 {
			return fmt.Errorf("truth table edge kind %s has %d cases", kind, seen[kind])
		}
	}
	return nil
}

func graphDigest(graph Graph) (string, error)       { graph.Digest = ""; return digestJSON(graph) }
func receiptDigest(receipt Receipt) (string, error) { receipt.Digest = ""; return digestJSON(receipt) }
func evidenceReceiptDigest(receipt EvidenceReceipt) (string, error) {
	receipt.Digest = ""
	return digestJSON(receipt)
}
func observationReceiptDigest(receipt ObservationReceipt) (string, error) {
	receipt.Digest = ""
	return digestJSON(receipt)
}
func observationBundleDigest(bundle ObservationBundle) (string, error) {
	bundle.Digest = ""
	// Profile is a fixture label only.  It is deliberately excluded from the
	// semantic digest so relabelling the same raw observations cannot alter a
	// decision.
	bundle.Profile = ""
	return digestJSON(bundle)
}
func failureReceiptDigest(receipt FailureReceipt) (string, error) {
	receipt.Digest = ""
	return digestJSON(receipt)
}
func readValidatorContract(path string) (ValidatorContract, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return ValidatorContract{}, fmt.Errorf("validator contract: %w", err)
	}
	var contract ValidatorContract
	if err := decodeStrictJSON(data, &contract); err != nil {
		return ValidatorContract{}, fmt.Errorf("validator contract decode: %w", err)
	}
	if err := validateValidatorContract(contract); err != nil {
		return ValidatorContract{}, err
	}
	return contract, nil
}

func readStructuralInventoryManifest(path string) (StructuralInventoryManifest, []byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return StructuralInventoryManifest{}, nil, fmt.Errorf("structural inventory manifest: %w", err)
	}
	var manifest StructuralInventoryManifest
	if err := decodeStrictJSON(data, &manifest); err != nil {
		return StructuralInventoryManifest{}, nil, fmt.Errorf("structural inventory manifest decode: %w", err)
	}
	return manifest, data, nil
}

func validateStructuralInventoryManifest(manifest StructuralInventoryManifest, contract ValidatorContract, graph Graph) error {
	if manifest.Schema != StructuralManifestSchema || manifest.ManifestID == "" || manifest.ContractID != contract.ContractID || len(manifest.EligibleClaimIDs) == 0 {
		return fmt.Errorf("STRUCTURAL_MANIFEST_IDENTITY_INVALID")
	}
	known := map[string]bool{}
	for _, claim := range graph.Nodes {
		known[claim.ClaimID] = true
	}
	eligible := map[string]bool{}
	for _, claimID := range manifest.EligibleClaimIDs {
		if claimID == "" || !known[claimID] || eligible[claimID] {
			return fmt.Errorf("STRUCTURAL_MANIFEST_ELIGIBLE_CLAIMS_INVALID: claim=%s", claimID)
		}
		eligible[claimID] = true
	}
	expected := map[string]bool{}
	for _, claimID := range manifest.ExpectedContradictionClaimIDs {
		if claimID == "" || !eligible[claimID] || expected[claimID] {
			return fmt.Errorf("STRUCTURAL_MANIFEST_EXPECTED_CLAIMS_INVALID: claim=%s", claimID)
		}
		expected[claimID] = true
	}
	return nil
}

func validateStructuralInventoryManifestRows(observed []StructuralContradiction, manifest StructuralInventoryManifest) error {
	expected := map[string]bool{}
	for _, claimID := range manifest.ExpectedContradictionClaimIDs {
		expected[claimID] = true
	}
	actual := map[string]bool{}
	for _, finding := range observed {
		if actual[finding.ClaimID] {
			return fmt.Errorf("STRUCTURAL_INVENTORY_DUPLICATE: claim=%s", finding.ClaimID)
		}
		actual[finding.ClaimID] = true
	}
	if len(actual) < len(expected) {
		return fmt.Errorf("STRUCTURAL_INVENTORY_MISSING: observed=%d expected=%d", len(actual), len(expected))
	}
	if len(actual) > len(expected) {
		return fmt.Errorf("STRUCTURAL_INVENTORY_ADDITIONAL: observed=%d expected=%d", len(actual), len(expected))
	}
	for claimID := range expected {
		if !actual[claimID] {
			return fmt.Errorf("STRUCTURAL_INVENTORY_REPLACEMENT: claim=%s", claimID)
		}
	}
	for claimID := range actual {
		if !expected[claimID] {
			return fmt.Errorf("STRUCTURAL_INVENTORY_REPLACEMENT: claim=%s", claimID)
		}
	}
	return nil
}

func decodeStrictJSON(data []byte, value any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(value); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return fmt.Errorf("trailing JSON value")
		}
		return fmt.Errorf("trailing JSON data: %w", err)
	}
	return nil
}

func decodeStrictJSONBytes(data []byte, value any) error { return decodeStrictJSON(data, value) }
func validateValidatorContract(contract ValidatorContract) error {
	if contract.Schema != ValidatorContractSchema || contract.ContractID == "" || contract.ExpectedArtifactPath == "" || contract.ExpectedArtifactDigest == "" || len(contract.Claims) != ClaimTotal {
		return fmt.Errorf("validator contract identity or denominator is invalid")
	}
	seen := map[string]bool{}
	claimIDs := map[string]bool{}
	targets := map[string]bool{}
	procedures := map[string]bool{}
	rowDigests := map[string]bool{}
	for _, claim := range contract.Claims {
		if claim.ClaimID == "" || claim.PropositionDigest == "" || claim.ProcedureID == "" || claim.TargetRowDigest == "" || claim.ExpectedMaterialDigest == "" || claim.ActivityName == "" || claim.ExpectedValueProgram == "" || claim.ExpectedTarget.Artifact == "" || claim.ExpectedTarget.Output == "" || seen[claim.ActivityName] || claimIDs[claim.ClaimID] || procedures[claim.ProcedureID] {
			return fmt.Errorf("validator contract claim material is invalid")
		}
		seen[claim.ActivityName] = true
		claimIDs[claim.ClaimID] = true
		procedures[claim.ProcedureID] = true
		if claim.ProcedureID != validatorProcedureID(claim.ActivityName) {
			return fmt.Errorf("VALIDATOR_CONTRACT_PROCEDURE_RELABEL: activity=%s", claim.ActivityName)
		}
		if claim.ExpectedMaterialDigest != validatorExpectedMaterialDigest(claim) || rowDigests[claim.TargetRowDigest] {
			return fmt.Errorf("validator contract expected material digest is invalid for %q", claim.ActivityName)
		}
		rowDigests[claim.TargetRowDigest] = true
		targetKey := strings.Join(claim.ExpectedTarget.Inputs, ",") + "|" + claim.ExpectedTarget.Output + "|" + claim.ExpectedTarget.Artifact
		if targets[targetKey] {
			return fmt.Errorf("validator contract targets are not distinct")
		}
		targets[targetKey] = true
	}
	if len(seen) != ClaimTotal {
		return fmt.Errorf("validator contract claims are not distinct")
	}
	return nil
}

func validatorProcedureID(activity string) string {
	return map[string]string{"Root": "CI_CLAIM_TARGET_ROW_ROOT_V1", "Derived": "CI_CLAIM_TARGET_ROW_DERIVED_V1", "SupportCheck": "CI_CLAIM_TARGET_ROW_SUPPORT_V1", "RequirementCheck": "CI_CLAIM_TARGET_ROW_REQUIREMENT_V1", "ContradictionCheck": "CI_CLAIM_TARGET_ROW_CONTRADICTION_V1", "FailureEntailmentCheck": "CI_CLAIM_TARGET_ROW_FAILURE_V1"}[activity]
}

func validatorExpectedMaterialDigest(claim ValidatorClaim) string {
	return digestBytes([]byte(fmt.Sprintf("claim-contract|claim_id=%s|proposition_digest=%s|procedure_id=%s|target_row_digest=%s|target_inputs=%s|target_output=%s|target_artifact=%s|expected_value_program=%s", claim.ClaimID, claim.PropositionDigest, claim.ProcedureID, claim.TargetRowDigest, strings.Join(claim.ExpectedTarget.Inputs, ","), claim.ExpectedTarget.Output, claim.ExpectedTarget.Artifact, claim.ExpectedValueProgram)))
}
func contractClaim(contract ValidatorContract, activityName string) (ValidatorClaim, bool) {
	for _, claim := range contract.Claims {
		if claim.ActivityName == activityName {
			return claim, true
		}
	}
	return ValidatorClaim{}, false
}

// canonicalTargetOccurrence identifies an activity in the target by the
// syntax AST and lowered semantic graph.  The AST span is used only to hash
// the raw declaration; it is never used to decide which activity exists.
func canonicalTargetOccurrence(artifact []byte, artifactPath string, sourceClaim Claim) (TargetOccurrence, []byte, error) {
	file, diagnostics := syntax.ParseFile(artifactPath, string(artifact))
	if file == nil || diagnostics.HasErrors() {
		return TargetOccurrence{}, nil, fmt.Errorf("TARGET_SYNTAX_OR_LOWER_INVALID: %s", diagnostics.Error())
	}
	var activities []*syntax.ActivityDecl
	for _, declaration := range file.Declarations {
		activity, ok := declaration.(*syntax.ActivityDecl)
		if ok && activity.Name == sourceClaim.ActivityName {
			activities = append(activities, activity)
		}
	}
	if len(activities) != 1 {
		return TargetOccurrence{}, nil, fmt.Errorf("TARGET_ACTIVITY_OCCURRENCE_CARDINALITY: activity=%s count=%d", sourceClaim.ActivityName, len(activities))
	}
	target, err := graphFromSource(artifact, artifactPath)
	if err != nil {
		return TargetOccurrence{}, nil, fmt.Errorf("TARGET_SYNTAX_OR_LOWER_INVALID: %w", err)
	}
	var targetClaim *Claim
	for i := range target.Graph.Nodes {
		if target.Graph.Nodes[i].ActivityName == sourceClaim.ActivityName {
			if targetClaim != nil {
				return TargetOccurrence{}, nil, fmt.Errorf("TARGET_ACTIVITY_OCCURRENCE_CARDINALITY: activity=%s semantic_count>1", sourceClaim.ActivityName)
			}
			targetClaim = &target.Graph.Nodes[i]
		}
	}
	if targetClaim == nil || targetClaim.ClaimID != sourceClaim.ClaimID || targetClaim.PropositionDigest != sourceClaim.PropositionDigest || !reflect.DeepEqual(targetClaim.Target, sourceClaim.Target) {
		return TargetOccurrence{}, nil, fmt.Errorf("TARGET_OCCURRENCE_SEMANTIC_ADDRESS_MISMATCH: activity=%s", sourceClaim.ActivityName)
	}
	span := activities[0].Span
	if span.Start.Offset < 0 || span.End.Offset < span.Start.Offset || span.End.Offset > len(artifact) {
		return TargetOccurrence{}, nil, fmt.Errorf("TARGET_ACTIVITY_OCCURRENCE_SPAN_INVALID: activity=%s", sourceClaim.ActivityName)
	}
	raw := append([]byte(nil), artifact[span.Start.Offset:span.End.Offset]...)
	occurrence := TargetOccurrence{
		Address:           "activity:" + targetClaim.ClaimID,
		ActivityName:      targetClaim.ActivityName,
		ClaimID:           targetClaim.ClaimID,
		PropositionDigest: targetClaim.PropositionDigest,
		Target:            targetClaim.Target,
		ValueProgram:      targetClaim.ValueProgram,
		RawSpanStart:      span.Start.Offset,
		RawSpanEnd:        span.End.Offset,
		RawRowDigest:      digestBytes(raw),
		SemanticDigest:    targetOccurrenceSemanticDigest(*targetClaim),
		ContextDigest:     target.Graph.CanonicalIRDigest,
	}
	return occurrence, raw, nil
}

func targetOccurrenceMaterial(value TargetOccurrence) string {
	return fmt.Sprintf("address=%s|activity=%s|claim_id=%s|proposition_digest=%s|target_inputs=%s|target_output=%s|target_artifact=%s|value_program=%s|semantic_digest=%s", value.Address, value.ActivityName, value.ClaimID, value.PropositionDigest, strings.Join(value.Target.Inputs, ","), value.Target.Output, value.Target.Artifact, value.ValueProgram, value.SemanticDigest)
}

func targetOccurrenceSemanticDigest(claim Claim) string {
	payload := fmt.Sprintf("target-occurrence|activity=%s|claim_id=%s|proposition=%s|target_inputs=%s|target_output=%s|target_artifact=%s|value_program=%s", claim.ActivityName, claim.ClaimID, claim.Proposition, strings.Join(claim.Target.Inputs, ","), claim.Target.Output, claim.Target.Artifact, claim.ValueProgram)
	return digestBytes([]byte(payload))
}

func observationProcedureDigest(procedure, claimID, propositionDigest, edgeID string, occurrence TargetOccurrence) string {
	payload := fmt.Sprintf("procedure|procedure_id=%s|claim_id=%s|proposition_digest=%s|edge_id=%s|%s", procedure, claimID, propositionDigest, edgeID, targetOccurrenceMaterial(occurrence))
	return digestBytes([]byte(payload))
}

func rawProvenanceDigest(occurrence TargetOccurrence, artifactDigest string) string {
	payload := fmt.Sprintf("raw-provenance|artifact_digest=%s|address=%s|raw_span_start=%d|raw_span_end=%d|raw_row_digest=%s", artifactDigest, occurrence.Address, occurrence.RawSpanStart, occurrence.RawSpanEnd, occurrence.RawRowDigest)
	return digestBytes([]byte(payload))
}

func structuralContradictionDigest(value StructuralContradiction) (string, error) {
	value.Digest = ""
	return digestJSON(value)
}

func deriveStructuralInventory(graph Graph, contract ValidatorContract, artifact []byte, artifactPath string) ([]StructuralContradiction, error) {
	result := []StructuralContradiction{}
	for _, claim := range graph.Nodes {
		material, ok := contractClaim(contract, claim.ActivityName)
		if !ok {
			return nil, fmt.Errorf("STRUCTURAL_INVENTORY_CONTRACT_CLAIM_MISSING: %s", claim.ActivityName)
		}
		if claim.ValueProgram == material.ExpectedValueProgram {
			continue
		}
		occurrence, _, err := canonicalTargetOccurrence(artifact, artifactPath, claim)
		if err != nil {
			return nil, err
		}
		finding := StructuralContradiction{ClaimID: claim.ClaimID, PropositionDigest: claim.PropositionDigest, ExpectedValue: material.ExpectedValueProgram, DeclaredValue: claim.ValueProgram, ProcedureID: material.ProcedureID, Occurrence: occurrence}
		finding.Digest, err = structuralContradictionDigest(finding)
		if err != nil {
			return nil, err
		}
		result = append(result, finding)
	}
	return result, nil
}

func validateStructuralInventory(observed, expected []StructuralContradiction) error {
	expectedByClaim := map[string]StructuralContradiction{}
	for _, finding := range expected {
		expectedByClaim[finding.ClaimID] = finding
	}
	seen := map[string]bool{}
	for _, finding := range observed {
		if seen[finding.ClaimID] {
			return fmt.Errorf("STRUCTURAL_INVENTORY_DUPLICATE: claim=%s", finding.ClaimID)
		}
		seen[finding.ClaimID] = true
	}
	if len(observed) < len(expected) {
		return fmt.Errorf("STRUCTURAL_INVENTORY_MISSING: observed=%d expected=%d", len(observed), len(expected))
	}
	if len(observed) > len(expected) {
		return fmt.Errorf("STRUCTURAL_INVENTORY_ADDITIONAL: observed=%d expected=%d", len(observed), len(expected))
	}
	for _, finding := range observed {
		want, ok := expectedByClaim[finding.ClaimID]
		if !ok || !reflect.DeepEqual(finding, want) {
			return fmt.Errorf("STRUCTURAL_INVENTORY_REPLACEMENT: claim=%s", finding.ClaimID)
		}
	}
	return nil
}

func claimEvidenceDigest(value EvidenceClaim) (string, error) {
	value.Digest = ""
	return digestJSON(value)
}
func capabilityDigest(value CapabilityEvidence) (string, error) {
	value.Digest = ""
	return digestJSON(value)
}
func transitionDigest(transition Transition) (string, error) {
	transition.TransitionDigest = ""
	return digestJSON(transition)
}
func digestJSON(value any) (string, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:]), nil
}
func digestBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}

// DigestBytesForObservation is the provider boundary's byte digest primitive.
// The observer uses it without parsing source or importing the evidence
// classifier's implementation details.
func DigestBytesForObservation(data []byte) string { return digestBytes(data) }
func prefixedDigest(raw string) string {
	if strings.HasPrefix(raw, "sha256:") {
		return raw
	}
	return "sha256:" + raw
}
func EdgeKinds() []EdgeKind { return []EdgeKind{Supports, Requires, Contradicts, FailureEntailment} }
