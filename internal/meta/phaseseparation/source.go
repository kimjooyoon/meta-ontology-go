package phaseseparation

import (
	"fmt"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

// ParseAndLower is the producer's source boundary. It deliberately passes
// through both the syntax AST and the canonical semantic lowerer before a
// computes payload can become a phase record.
func ParseAndLower(filename string, source []byte) (ParsedFile, error) {
	file, diagnostics := syntax.ParseFile(filename, string(source))
	if file == nil || diagnostics.HasErrors() {
		return ParsedFile{}, fmt.Errorf("%s: syntax.ParseFile: %v", filename, diagnostics.Error())
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return ParsedFile{}, fmt.Errorf("%s: bidir.Lower: %w", filename, err)
	}
	parsed := ParsedFile{
		Filename: filename, Source: append([]byte(nil), source...), File: file,
		IR: ir, SemanticHash: ir.StableHash(), EntityIDs: make(map[string]string),
	}
	for _, declaration := range file.Declarations {
		entity, ok := declaration.(*syntax.EntityDecl)
		if ok {
			if entity.ID == "" {
				return ParsedFile{}, fmt.Errorf("%s: entity %q has no stable ID", filename, entity.Name)
			}
			if _, exists := parsed.EntityIDs[entity.Name]; exists {
				return ParsedFile{}, fmt.Errorf("%s: duplicate entity name %q", filename, entity.Name)
			}
			parsed.EntityIDs[entity.Name] = entity.ID
		}
	}
	for _, declaration := range file.Declarations {
		activity, ok := declaration.(*syntax.ActivityDecl)
		if !ok {
			continue
		}
		if len(activity.Inputs) != 1 || activity.Output == "" {
			return ParsedFile{}, fmt.Errorf("%s: activity %q must have one input and one output", filename, activity.Name)
		}
		fromID, fromOK := parsed.EntityIDs[activity.Inputs[0].Name]
		toID, toOK := parsed.EntityIDs[activity.Output]
		if !fromOK || !toOK {
			return ParsedFile{}, fmt.Errorf("%s: activity %q is not entity-bound", filename, activity.Name)
		}
		activityNode, ok := activityNode(ir, activity.Name)
		if !ok || activityNode.ValueProgram != activity.ValueProgram {
			return ParsedFile{}, fmt.Errorf("%s: activity %q was not retained by bidir.Lower", filename, activity.Name)
		}
		record, err := decodeActivity(filename, activity, fromID, toID, activityNode.ID.String())
		if err != nil {
			return ParsedFile{}, err
		}
		parsed.Activities = append(parsed.Activities, record)
	}
	if len(parsed.Activities) == 0 {
		return ParsedFile{}, fmt.Errorf("%s: no phase activities", filename)
	}
	return parsed, nil
}

func activityNode(ir semantic.IR, name string) (semantic.Node, bool) {
	for _, node := range ir.Graph.Nodes() {
		if node.Kind == semantic.Activity && node.Name == name {
			return node, true
		}
	}
	return semantic.Node{}, false
}

func decodeActivity(filename string, activity *syntax.ActivityDecl, fromID, toID, activityID string) (SourceRecord, error) {
	fields, err := parseComputes(activity.ValueProgram)
	if err != nil {
		return SourceRecord{}, fmt.Errorf("%s: activity %q: %w", filename, activity.Name, err)
	}
	record := SourceRecord{
		ActivityName: activity.Name, ActivityID: activityID, Program: activity.ValueProgram,
		CaseKey: fields["case"], TransferID: fields["transfer_id"], ValueID: fields["value_id"],
		FromValueID: fields["from_value_id"], ToValueID: fields["to_value_id"],
		LiteralClass: fields["literal_class"], FromLiteralClass: fields["from_literal_class"],
		ToLiteralClass: fields["to_literal_class"], FromPhase: fields["from_phase"], ToPhase: fields["to_phase"],
		PayloadClass: fields["payload_class"], ClaimDigest: fields["claim_digest"], TargetDigest: fields["target_digest"],
		Provenance: fields["provenance"], ClaimStateFrom: fields["claim_state_from"], ClaimStateTo: fields["claim_state_to"],
		Stage: fields["stage"], Step: fields["step"], DeclaredReason: fields["reason"],
	}
	if record.CaseKey == "" || record.TransferID == "" || record.ValueID == "" || record.FromValueID == "" || record.ToValueID == "" {
		return SourceRecord{}, fmt.Errorf("%s: activity %q: incomplete phase value identity", filename, activity.Name)
	}
	if record.ValueID != record.FromValueID || record.FromValueID != fromID || record.ToValueID != toID {
		return SourceRecord{}, fmt.Errorf("%s: activity %q: computes value IDs disagree with entity endpoints", filename, activity.Name)
	}
	fromPhase, err := phaseOfID(fromID)
	if err != nil {
		return SourceRecord{}, fmt.Errorf("%s: activity %q: %w", filename, activity.Name, err)
	}
	toPhase, err := phaseOfID(toID)
	if err != nil {
		return SourceRecord{}, fmt.Errorf("%s: activity %q: %w", filename, activity.Name, err)
	}
	if record.FromPhase != fromPhase || record.ToPhase != toPhase || record.LiteralClass != record.FromLiteralClass {
		return SourceRecord{}, fmt.Errorf("%s: activity %q: computes phase or literal class disagrees with declaration", filename, activity.Name)
	}
	if expectedLiteralClass[fromPhase] != record.FromLiteralClass || expectedLiteralClass[toPhase] != record.ToLiteralClass {
		return SourceRecord{}, fmt.Errorf("%s: activity %q: literal class is not phase-local", filename, activity.Name)
	}
	if record.PayloadClass != PayloadClaim && record.PayloadClass != PayloadValue && record.PayloadClass != PayloadAuthority && record.PayloadClass != PayloadEvidence {
		return SourceRecord{}, fmt.Errorf("%s: activity %q: unknown transfer payload class %q", filename, activity.Name, record.PayloadClass)
	}
	if record.Provenance == "" || record.Stage == "" || record.Step == "" || record.DeclaredReason == "" {
		return SourceRecord{}, fmt.Errorf("%s: activity %q: evidence coordinate or provenance is missing", filename, activity.Name)
	}
	if record.ClaimStateFrom != StateOpen || (record.ClaimStateTo != StateOpen && record.ClaimStateTo != StateDischarged && record.ClaimStateTo != StateRefuted) {
		return SourceRecord{}, fmt.Errorf("%s: activity %q: invalid claim lifecycle", filename, activity.Name)
	}
	if record.PayloadClass == PayloadClaim {
		if !isDigest(record.ClaimDigest) || !isDigest(record.TargetDigest) {
			return SourceRecord{}, fmt.Errorf("%s: activity %q: explicit claim lacks digests", filename, activity.Name)
		}
	} else if record.ClaimDigest != "none" || record.TargetDigest != "none" {
		return SourceRecord{}, fmt.Errorf("%s: activity %q: non-claim payload carries claim authority", filename, activity.Name)
	}
	return record, nil
}

func parseComputes(program string) (map[string]string, error) {
	if strings.TrimSpace(program) == "" {
		return nil, fmt.Errorf("computes value is required")
	}
	fields := make(map[string]string)
	for _, part := range strings.Split(program, ";") {
		key, value, ok := strings.Cut(part, "=")
		key, value = strings.TrimSpace(key), strings.TrimSpace(value)
		if !ok || key == "" || value == "" || fields[key] != "" {
			return nil, fmt.Errorf("invalid computes field %q", part)
		}
		fields[key] = value
	}
	for _, key := range []string{"case", "transfer_id", "value_id", "from_value_id", "to_value_id", "literal_class", "from_literal_class", "to_literal_class", "from_phase", "to_phase", "payload_class", "claim_digest", "target_digest", "provenance", "claim_state_from", "claim_state_to", "stage", "step", "reason"} {
		if fields[key] == "" {
			return nil, fmt.Errorf("missing computes field %q", key)
		}
	}
	return fields, nil
}

func phaseOfID(id string) (string, error) {
	_, path, ok := strings.Cut(id, "://")
	if !ok {
		return "", fmt.Errorf("phase value ID %q has no URI path", id)
	}
	parts := strings.Split(path, "/")
	if len(parts) != 2 || !validPhase(parts[0]) || parts[1] == "" {
		return "", fmt.Errorf("phase value ID %q is not phase-local", id)
	}
	return parts[0], nil
}

func validPhase(phase string) bool {
	for _, candidate := range phases {
		if phase == candidate {
			return true
		}
	}
	return false
}

func isDigest(value string) bool {
	return len(value) == len("sha256:")+64 && strings.HasPrefix(value, "sha256:")
}
