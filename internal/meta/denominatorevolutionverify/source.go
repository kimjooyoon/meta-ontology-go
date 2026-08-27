package denominatorevolutionverify

import (
	"fmt"
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

var requiredEntities = []string{"FixedDenominator", "DenominatorVersion", "ChangeReason", "PredecessorEvidence", "MigrationReceipt"}
var requiredActivities = []string{"DeclareFixedDenominator", "ProposeDenominatorChange", "BindPredecessorDigest", "RecordChangeReasons", "IssueMigrationReceipt"}

type sourceProposal struct {
	ID          string
	Predecessor Ref
	Successor   string
	ReceiptID   string
}

type sourceReceiptMaterial struct {
	ID          string
	Predecessor Ref
	Successor   Ref
	Additions   []Change
	Deletions   []Change
	BoundPrev   Ref
	BoundNext   Ref
}

type sourceWire struct {
	Version                   string
	Obligations               []Obligation
	Proposals                 []sourceProposal
	Additions                 []Change
	Deletions                 []Change
	KnownPredecessorVersion   string
	UnknownPredecessorVersion string
	SuccessorVersion          string
	UnknownSuccessorVersion   string
	Receipt                   sourceReceiptMaterial
}

// parseSource is an independent ParseFile -> bidir.Lower implementation. It
// reconstructs input material, not any producer report labels.
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
	projection.ObligationCount, projection.CaseCount = len(wire.Obligations), len(wire.Proposals)
	projection.ForbiddenPropositionPresent = sourceHasForbiddenProposition(wire.Obligations)
	projection.WireDigest = digestValue(wire)
	projection.Exact = len(entities) == len(requiredEntities) && len(activities) == len(requiredActivities) && len(projection.RequiredEntities) == 0 && len(projection.RequiredActivities) == 0 && wire.Exact() && projection.ForbiddenPropositionPresent
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
	if err = decodeProposals(activities["ProposeDenominatorChange"], &wire); err != nil {
		return sourceWire{}, err
	}
	if err = decodePredecessors(activities["BindPredecessorDigest"], &wire); err != nil {
		return sourceWire{}, err
	}
	if err = decodeReceipt(activities["RecordChangeReasons"], &wire); err != nil {
		return sourceWire{}, err
	}
	if err = decodeReceiptBind(activities["IssueMigrationReceipt"], &wire); err != nil {
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

func decodeProposals(value string, wire *sourceWire) error {
	parts, err := payload(value, "proposals")
	if err != nil {
		return err
	}
	for _, part := range parts {
		key, value, ok := strings.Cut(part, "=")
		if !ok {
			return fmt.Errorf("GOOO_SOURCE_PROPOSAL_FIELD_INVALID")
		}
		switch key {
		case "proposal":
			fields := strings.Split(value, "^")
			if len(fields) != 5 || fields[0] == "" || fields[1] == "" || fields[2] == "" || fields[3] == "" || fields[4] == "" {
				return fmt.Errorf("GOOO_SOURCE_PROPOSAL_ARITY")
			}
			wire.Proposals = append(wire.Proposals, sourceProposal{ID: fields[0], Predecessor: Ref{Version: fields[1], Digest: fields[2]}, Successor: fields[3], ReceiptID: fields[4]})
		case "addition":
			change, err := parseAddition(value)
			if err != nil {
				return err
			}
			wire.Additions = append(wire.Additions, change)
		case "deletion":
			change, err := parseDeletion(value)
			if err != nil {
				return err
			}
			wire.Deletions = append(wire.Deletions, change)
		default:
			return fmt.Errorf("GOOO_SOURCE_PROPOSAL_KEY_INVALID: %s", key)
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
	parts, err := payload(value, "receipt-material")
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
			wire.Receipt.ID = value
		case "predecessor":
			ref, err := parseRef(value)
			if err != nil {
				return err
			}
			wire.Receipt.Predecessor = ref
		case "successor":
			ref, err := parseRef(value)
			if err != nil {
				return err
			}
			wire.Receipt.Successor = ref
		case "addition":
			change, err := parseDeletion(value)
			if err != nil {
				return err
			}
			wire.Receipt.Additions = append(wire.Receipt.Additions, change)
		case "deletion":
			change, err := parseDeletion(value)
			if err != nil {
				return err
			}
			wire.Receipt.Deletions = append(wire.Receipt.Deletions, change)
		default:
			return fmt.Errorf("GOOO_SOURCE_RECEIPT_KEY_INVALID: %s", key)
		}
	}
	return nil
}

func decodeReceiptBind(value string, wire *sourceWire) error {
	parts, err := payload(value, "receipt-bind")
	if err != nil {
		return err
	}
	seen := map[string]bool{}
	for _, part := range parts {
		key, value, ok := strings.Cut(part, "=")
		if !ok || (key != "predecessor" && key != "successor") || seen[key] {
			return fmt.Errorf("GOOO_SOURCE_RECEIPT_BIND_INVALID")
		}
		ref, err := parseRef(value)
		if err != nil {
			return err
		}
		seen[key] = true
		if key == "predecessor" {
			wire.Receipt.BoundPrev = ref
		} else {
			wire.Receipt.BoundNext = ref
		}
	}
	if !seen["predecessor"] || !seen["successor"] {
		return fmt.Errorf("GOOO_SOURCE_RECEIPT_BIND_INCOMPLETE")
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
	for _, field := range fields {
		if field == "" {
			return Obligation{}, fmt.Errorf("GOOO_SOURCE_OBLIGATION_EMPTY")
		}
	}
	return Obligation{ID: fields[0], Claim: fields[1], Class: fields[2], ProofChoice: fields[3], MetaOperation: fields[4], Stage: fields[5], Step: fields[6], Reason: fields[7]}, nil
}

func parseAddition(value string) (Change, error) {
	fields := strings.Split(value, "^")
	if len(fields) != 9 {
		return Change{}, fmt.Errorf("GOOO_SOURCE_ADDITION_ARITY")
	}
	for _, field := range fields {
		if field == "" {
			return Change{}, fmt.Errorf("GOOO_SOURCE_ADDITION_EMPTY")
		}
	}
	member := &Obligation{ID: fields[0], Claim: fields[2], Class: fields[3], ProofChoice: fields[4], MetaOperation: fields[5], Stage: fields[6], Step: fields[7], Reason: fields[8]}
	return Change{ObligationID: fields[0], Reason: fields[1], Member: member}, nil
}

func parseDeletion(value string) (Change, error) {
	fields := strings.Split(value, "^")
	if len(fields) != 2 || fields[0] == "" || fields[1] == "" {
		return Change{}, fmt.Errorf("GOOO_SOURCE_CHANGE_ARITY")
	}
	return Change{ObligationID: fields[0], Reason: fields[1]}, nil
}

func parseRef(value string) (Ref, error) {
	fields := strings.Split(value, "^")
	if len(fields) != 2 || fields[0] == "" || fields[1] == "" {
		return Ref{}, fmt.Errorf("GOOO_SOURCE_REF_ARITY")
	}
	return Ref{Version: fields[0], Digest: fields[1]}, nil
}

func sourceHasForbiddenProposition(values []Obligation) bool {
	for _, value := range values {
		claim := strings.ToLower(value.ID + " " + value.Claim)
		if strings.Contains(claim, "improvement rate") || strings.Contains(claim, "aggregate estimate") || strings.Contains(claim, "projected coverage") {
			return true
		}
	}
	return false
}

func (wire sourceWire) Exact() bool {
	if wire.Version != DenominatorVersion || len(wire.Obligations) != DenominatorSize || len(wire.Proposals) != CaseCount || len(wire.Additions) != 1 || len(wire.Deletions) != 1 {
		return false
	}
	ids := map[string]bool{}
	for _, obligation := range wire.Obligations {
		if obligation.ID == "" || ids[obligation.ID] {
			return false
		}
		ids[obligation.ID] = true
	}
	proposalIDs := map[string]bool{}
	for _, proposal := range wire.Proposals {
		if proposal.ID == "" || proposal.Predecessor.Version == "" || proposal.Predecessor.Digest == "" || proposal.Successor == "" || proposal.ReceiptID == "" || proposalIDs[proposal.ID] {
			return false
		}
		proposalIDs[proposal.ID] = true
	}
	addition := wire.Additions[0]
	if addition.Member == nil || addition.Member.ID != addition.ObligationID || ids[addition.ObligationID] {
		return false
	}
	deletion := wire.Deletions[0]
	if !ids[deletion.ObligationID] || addition.ObligationID == deletion.ObligationID {
		return false
	}
	if wire.KnownPredecessorVersion != DenominatorVersion || wire.UnknownPredecessorVersion == "" || wire.SuccessorVersion != SuccessorVersion || wire.UnknownSuccessorVersion == "" {
		return false
	}
	receipt := wire.Receipt
	if receipt.ID == "" || receipt.Predecessor.Version == "" || receipt.Predecessor.Digest == "" || receipt.Successor.Version == "" || receipt.Successor.Digest == "" || len(receipt.Additions) != 1 || len(receipt.Deletions) != 1 || receipt.BoundPrev.Version == "" || receipt.BoundPrev.Digest == "" || receipt.BoundNext.Version == "" || receipt.BoundNext.Digest == "" {
		return false
	}
	return sameChangeSet(wire.Additions, []Change{{ObligationID: receipt.Additions[0].ObligationID, Reason: receipt.Additions[0].Reason, Member: addition.Member}}) && sameChangeSet(wire.Deletions, receipt.Deletions)
}

func sameChangeSet(left, right []Change) bool {
	if len(left) != len(right) {
		return false
	}
	leftCopy := append([]Change(nil), left...)
	rightCopy := append([]Change(nil), right...)
	sort.Slice(leftCopy, func(i, j int) bool { return leftCopy[i].ObligationID < leftCopy[j].ObligationID })
	sort.Slice(rightCopy, func(i, j int) bool { return rightCopy[i].ObligationID < rightCopy[j].ObligationID })
	for index := range leftCopy {
		if leftCopy[index].ObligationID != rightCopy[index].ObligationID || leftCopy[index].Reason != rightCopy[index].Reason {
			return false
		}
	}
	return true
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
