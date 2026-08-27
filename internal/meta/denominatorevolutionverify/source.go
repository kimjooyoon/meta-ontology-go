package denominatorevolutionverify

import (
	"fmt"
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

var requiredEntities = []string{"FixedDenominator", "DenominatorVersion", "ChangeReason", "PredecessorEvidence", "MigrationReceipt", "ClaimTransition", "IndependentDecision"}
var requiredActivities = []string{"DeclareFixedDenominator", "ProposeDenominatorChange", "BindPredecessorDigest", "RecordChangeReasons", "IssueMigrationReceipt", "TransitionClaim", "IndependentlyDecide"}

type sourceCase struct {
	Spec               CaseSpec
	PredecessorVersion string
	SuccessorVersion   string
}
type sourceMutation struct {
	CaseID     string
	Obligation Obligation
}
type sourceDecision struct{ ID, Decision, Resolution, Reason string }
type sourceWire struct {
	Version                   string
	Obligations               []Obligation
	Cases                     []sourceCase
	Additions                 []Change
	Deletions                 []Change
	Mutations                 []sourceMutation
	KnownPredecessorVersion   string
	UnknownPredecessorVersion string
	SuccessorVersion          string
	UnknownSuccessorVersion   string
	ReceiptID                 string
	ReceiptDecision           string
	ReceiptReason             string
	ReceiptCoordinate         Coordinate
	Claims                    []ClaimLedgerEntry
	EmittedClaims             []EmittedClaim
	Decisions                 []sourceDecision
}

// parseSource is intentionally duplicated from the producer package. The
// independent consumer owns its own parser, wire model, and decision inputs.
func parseSource(raw []byte) (SourceProjection, sourceWire, error) {
	file, diagnostics := syntax.ParseFile("main.gooo", string(raw))
	if diagnostics.HasErrors() {
		return SourceProjection{}, sourceWire{}, fmt.Errorf("GOOO_SOURCE_SYNTAX_INVALID")
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return SourceProjection{}, sourceWire{}, fmt.Errorf("GOOO_SOURCE_LOWER_INVALID: %w", err)
	}
	entities := []string{}
	activities := map[string]string{}
	for _, node := range ir.Graph.Nodes() {
		switch node.Kind {
		case semantic.Entity:
			entities = append(entities, node.Name)
		case semantic.Activity:
			if _, exists := activities[node.Name]; exists {
				return SourceProjection{}, sourceWire{}, fmt.Errorf("GOOO_SOURCE_DUPLICATE_ACTIVITY")
			}
			activities[node.Name] = node.ValueProgram
		}
	}
	projection := SourceProjection{EntityCount: len(entities), ActivityCount: len(activities), RequiredEntities: missing(requiredEntities, entities), RequiredActivities: missing(requiredActivities, keys(activities)), SemanticDigest: ir.StableHash()}
	wire, err := decodeActivities(activities)
	if err != nil {
		return projection, sourceWire{}, err
	}
	projection.ObligationCount, projection.CaseCount = len(wire.Obligations), len(wire.Cases)
	projection.WireDigest = digestValue(wire)
	projection.Exact = len(entities) == len(requiredEntities) && len(activities) == len(requiredActivities) && len(projection.RequiredEntities) == 0 && len(projection.RequiredActivities) == 0 && wire.Exact()
	return projection, wire, nil
}

func decodeActivities(activities map[string]string) (sourceWire, error) {
	for _, name := range requiredActivities {
		if _, ok := activities[name]; !ok {
			return sourceWire{}, fmt.Errorf("GOOO_SOURCE_ACTIVITY_MISSING: %s", name)
		}
	}
	var wire sourceWire
	var err error
	if wire, err = decodeDenominator(activities["DeclareFixedDenominator"]); err != nil {
		return sourceWire{}, err
	}
	if err = decodeCases(activities["ProposeDenominatorChange"], &wire); err != nil {
		return sourceWire{}, err
	}
	if err = decodePredecessors(activities["BindPredecessorDigest"], &wire); err != nil {
		return sourceWire{}, err
	}
	if err = decodeReceipt(activities["RecordChangeReasons"], &wire); err != nil {
		return sourceWire{}, err
	}
	if err = decodeReceiptBind(activities["IssueMigrationReceipt"]); err != nil {
		return sourceWire{}, err
	}
	if err = decodeClaims(activities["TransitionClaim"], &wire); err != nil {
		return sourceWire{}, err
	}
	if err = decodeDecisions(activities["IndependentlyDecide"], &wire); err != nil {
		return sourceWire{}, err
	}
	if !wire.Exact() {
		return sourceWire{}, fmt.Errorf("GOOO_SOURCE_WIRE_INCOMPLETE")
	}
	return wire, nil
}

func decodeDenominator(value string) (sourceWire, error) {
	parts, err := payload(value, "denominator")
	if err != nil {
		return sourceWire{}, err
	}
	var wire sourceWire
	for _, part := range parts {
		key, value, ok := strings.Cut(part, "=")
		if !ok || (key != "version" && key != "obligation") {
			return sourceWire{}, fmt.Errorf("GOOO_SOURCE_DENOMINATOR_FIELD_INVALID")
		}
		if key == "version" {
			if wire.Version != "" {
				return sourceWire{}, fmt.Errorf("GOOO_SOURCE_DUPLICATE_VERSION")
			}
			wire.Version = value
			continue
		}
		obligation, err := parseObligation(value)
		if err != nil {
			return sourceWire{}, err
		}
		wire.Obligations = append(wire.Obligations, obligation)
	}
	return wire, nil
}

func decodeCases(value string, wire *sourceWire) error {
	parts, err := payload(value, "cases")
	if err != nil {
		return err
	}
	for _, part := range parts {
		key, value, ok := strings.Cut(part, "=")
		if !ok {
			return fmt.Errorf("GOOO_SOURCE_CASE_FIELD_INVALID")
		}
		switch key {
		case "case":
			fields := strings.Split(value, "^")
			if len(fields) != 14 {
				return fmt.Errorf("GOOO_SOURCE_CASE_ARITY")
			}
			wire.Cases = append(wire.Cases, sourceCase{Spec: CaseSpec{ID: fields[0], Kind: fields[1], ExpectedDecision: fields[2], ExpectedResolution: fields[3], ExpectedReason: fields[4], FromClaim: fields[5], ToClaim: fields[6], ProofChoice: fields[7], MetaOperation: fields[8], Stage: fields[9], Step: fields[10], Reason: fields[11]}, PredecessorVersion: fields[12], SuccessorVersion: fields[13]})
		case "addition", "deletion":
			change, err := parseChange(value)
			if err != nil {
				return err
			}
			if key == "addition" {
				wire.Additions = append(wire.Additions, change)
			} else {
				wire.Deletions = append(wire.Deletions, change)
			}
		case "mutation":
			fields := strings.Split(value, "^")
			if len(fields) != 9 {
				return fmt.Errorf("GOOO_SOURCE_MUTATION_ARITY")
			}
			obligation, err := parseObligation(strings.Join(fields[1:], "^"))
			if err != nil {
				return err
			}
			wire.Mutations = append(wire.Mutations, sourceMutation{CaseID: fields[0], Obligation: obligation})
		default:
			return fmt.Errorf("GOOO_SOURCE_CASE_KEY_INVALID: %s", key)
		}
	}
	return nil
}

func decodePredecessors(value string, wire *sourceWire) error {
	parts, err := payload(value, "predecessors")
	if err != nil {
		return err
	}
	for _, part := range parts {
		key, value, ok := strings.Cut(part, "=")
		if !ok {
			return fmt.Errorf("GOOO_SOURCE_PREDECESSOR_FIELD_INVALID")
		}
		switch key {
		case "known":
			wire.KnownPredecessorVersion = value
		case "unknown":
			wire.UnknownPredecessorVersion = value
		case "successor":
			wire.SuccessorVersion = value
		case "unknown-successor":
			wire.UnknownSuccessorVersion = value
		default:
			return fmt.Errorf("GOOO_SOURCE_PREDECESSOR_KEY_INVALID: %s", key)
		}
	}
	return nil
}

func decodeReceipt(value string, wire *sourceWire) error {
	parts, err := payload(value, "receipt")
	if err != nil {
		return err
	}
	for _, part := range parts {
		key, value, ok := strings.Cut(part, "=")
		if !ok {
			return fmt.Errorf("GOOO_SOURCE_RECEIPT_FIELD_INVALID")
		}
		switch key {
		case "id":
			wire.ReceiptID = value
		case "decision":
			wire.ReceiptDecision = value
		case "reason":
			wire.ReceiptReason = value
		case "coordinate":
			fields := strings.Split(value, "^")
			if len(fields) != 3 {
				return fmt.Errorf("GOOO_SOURCE_RECEIPT_COORDINATE_ARITY")
			}
			wire.ReceiptCoordinate = Coordinate{Stage: fields[0], Step: fields[1], Reason: fields[2]}
		default:
			return fmt.Errorf("GOOO_SOURCE_RECEIPT_KEY_INVALID: %s", key)
		}
	}
	return nil
}

func decodeReceiptBind(value string) error {
	parts, err := payload(value, "receipt-bind")
	if err != nil {
		return err
	}
	seen := map[string]bool{}
	for _, part := range parts {
		key, _, ok := strings.Cut(part, "=")
		if !ok || (key != "predecessor" && key != "successor") || seen[key] {
			return fmt.Errorf("GOOO_SOURCE_RECEIPT_BIND_INVALID")
		}
		seen[key] = true
	}
	if !seen["predecessor"] || !seen["successor"] {
		return fmt.Errorf("GOOO_SOURCE_RECEIPT_BIND_INCOMPLETE")
	}
	return nil
}

func decodeClaims(value string, wire *sourceWire) error {
	parts, err := payload(value, "claims")
	if err != nil {
		return err
	}
	for _, part := range parts {
		key, value, ok := strings.Cut(part, "=")
		if !ok {
			return fmt.Errorf("GOOO_SOURCE_CLAIM_FIELD_INVALID")
		}
		switch key {
		case "event":
			fields := strings.Split(value, "^")
			if len(fields) != 7 {
				return fmt.Errorf("GOOO_SOURCE_CLAIM_ARITY")
			}
			var sequence int
			if _, err := fmt.Sscanf(fields[0], "%d", &sequence); err != nil {
				return fmt.Errorf("GOOO_SOURCE_CLAIM_SEQUENCE_INVALID")
			}
			wire.Claims = append(wire.Claims, ClaimLedgerEntry{Sequence: sequence, ClaimID: fields[1], PriorState: fields[2], NextState: fields[3], Stage: fields[4], Step: fields[5], Reason: fields[6]})
		case "emitted":
			fields := strings.Split(value, "^")
			if len(fields) != 3 {
				return fmt.Errorf("GOOO_SOURCE_EMITTED_ARITY")
			}
			wire.EmittedClaims = append(wire.EmittedClaims, EmittedClaim{ID: fields[0], Class: fields[1], State: fields[2]})
		default:
			return fmt.Errorf("GOOO_SOURCE_CLAIM_KEY_INVALID: %s", key)
		}
	}
	return nil
}

func decodeDecisions(value string, wire *sourceWire) error {
	parts, err := payload(value, "decisions")
	if err != nil {
		return err
	}
	for _, part := range parts {
		key, value, ok := strings.Cut(part, "=")
		if !ok || key != "case" {
			return fmt.Errorf("GOOO_SOURCE_DECISION_FIELD_INVALID")
		}
		fields := strings.Split(value, "^")
		if len(fields) != 4 {
			return fmt.Errorf("GOOO_SOURCE_DECISION_ARITY")
		}
		wire.Decisions = append(wire.Decisions, sourceDecision{ID: fields[0], Decision: fields[1], Resolution: fields[2], Reason: fields[3]})
	}
	return nil
}

func payload(value, want string) ([]string, error) {
	parts := strings.Split(value, "|")
	if len(parts) == 0 || parts[0] != want {
		return nil, fmt.Errorf("GOOO_SOURCE_PAYLOAD_KIND_MISMATCH: %s", want)
	}
	return parts[1:], nil
}
func parseObligation(value string) (Obligation, error) {
	fields := strings.Split(value, "^")
	if len(fields) != 8 {
		return Obligation{}, fmt.Errorf("GOOO_SOURCE_OBLIGATION_ARITY")
	}
	return Obligation{ID: fields[0], Claim: fields[1], Class: fields[2], ProofChoice: fields[3], MetaOperation: fields[4], Stage: fields[5], Step: fields[6], Reason: fields[7]}, nil
}
func parseChange(value string) (Change, error) {
	fields := strings.Split(value, "^")
	if len(fields) != 2 || fields[0] == "" || fields[1] == "" {
		return Change{}, fmt.Errorf("GOOO_SOURCE_CHANGE_ARITY")
	}
	return Change{ObligationID: fields[0], Reason: fields[1]}, nil
}
func (wire sourceWire) Exact() bool {
	if wire.Version == "" || len(wire.Obligations) != DenominatorSize || len(wire.Cases) != CaseCount || len(wire.Additions) != 1 || len(wire.Deletions) != 1 || len(wire.Mutations) != 2 || len(wire.Claims) != CaseCount || len(wire.EmittedClaims) == 0 || len(wire.Decisions) != CaseCount {
		return false
	}
	for index, value := range wire.Cases {
		if value.Spec.ID == "" || value.PredecessorVersion == "" || value.SuccessorVersion == "" || (index > 0 && value.Spec.ID == wire.Cases[index-1].Spec.ID) {
			return false
		}
	}
	for index, value := range wire.Claims {
		if value.Sequence != index+1 || value.ClaimID == "" || value.ClaimID != wire.Cases[index].Spec.ID || value.PriorState == "" || value.NextState == "" || value.Stage == "" || value.Step == "" || value.Reason == "" {
			return false
		}
	}
	for index, value := range wire.Decisions {
		caseValue := wire.Cases[index].Spec
		if value.ID != caseValue.ID || value.Decision != caseValue.ExpectedDecision || value.Resolution != caseValue.ExpectedResolution || value.Reason != caseValue.ExpectedReason {
			return false
		}
	}
	return wire.KnownPredecessorVersion != "" && wire.UnknownPredecessorVersion != "" && wire.SuccessorVersion != "" && wire.UnknownSuccessorVersion != "" && wire.ReceiptID != "" && wire.ReceiptDecision != "" && wire.ReceiptReason != ""
}
func keys(values map[string]string) []string {
	result := make([]string, 0, len(values))
	for key := range values {
		result = append(result, key)
	}
	sort.Strings(result)
	return result
}
func missing(required, actual []string) []string {
	result := []string{}
	for _, want := range required {
		found := false
		for _, got := range actual {
			if got == want {
				found = true
				break
			}
		}
		if !found {
			result = append(result, want)
		}
	}
	return result
}
