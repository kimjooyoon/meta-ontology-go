package semantic

import (
	"fmt"
	"sort"
	"strings"
)

func (p InferencePathV1) Normalized() (InferencePathV1, error) {
	issues := &InferencePathErrors{}
	version := strings.TrimSpace(p.Version)
	if version != InferencePathSchemaVersion {
		issues.add("unknown-version", "", fmt.Sprintf("got %q", p.Version))
	}
	out := normalizeInferenceRecords(p, version, issues)
	validateInferencePathRecords(out, issues)
	if len(issues.Issues) != 0 {
		sortInferenceIssues(issues)
		return InferencePathV1{}, issues
	}
	return out, nil
}

func normalizeInferenceRecords(p InferencePathV1, version string, issues *InferencePathErrors) InferencePathV1 {
	out := InferencePathV1{
		Version: version, Edges: make([]InferenceEdge, 0, len(p.Edges)),
		Claims:   make([]SemanticChangeClaim, 0, len(p.Claims)),
		Evidence: make([]InferenceEvidence, 0, len(p.Evidence)),
	}
	for _, raw := range p.Evidence {
		normalized, err := raw.normalized()
		if err != nil {
			issues.add("evidence", raw.ID, err.Error())
			continue
		}
		out.Evidence = append(out.Evidence, normalized)
	}
	for _, raw := range p.Edges {
		normalized, err := raw.normalized()
		if err != nil {
			issues.add("edge", raw.RecordID, err.Error())
			continue
		}
		out.Edges = append(out.Edges, normalized)
	}
	for _, raw := range p.Claims {
		normalized, err := raw.normalized()
		if err != nil {
			issues.add("claim", raw.RecordID, err.Error())
			continue
		}
		out.Claims = append(out.Claims, normalized)
	}
	sort.Slice(out.Evidence, func(i, j int) bool { return out.Evidence[i].ID < out.Evidence[j].ID })
	sort.Slice(out.Edges, func(i, j int) bool { return out.Edges[i].RecordID < out.Edges[j].RecordID })
	sort.Slice(out.Claims, func(i, j int) bool { return out.Claims[i].RecordID < out.Claims[j].RecordID })
	return out
}

func validateInferencePathRecords(out InferencePathV1, issues *InferencePathErrors) {
	seen := make(map[ID]string, len(out.Evidence)+len(out.Edges)+len(out.Claims))
	for _, record := range out.Evidence {
		registerInferenceID(seen, record.ID, "evidence", issues)
	}
	for _, edge := range out.Edges {
		registerInferenceID(seen, edge.RecordID, "edge", issues)
	}
	for _, claim := range out.Claims {
		registerInferenceID(seen, claim.RecordID, "claim", issues)
	}
	evidence := make(map[ID]InferenceEvidence, len(out.Evidence))
	for _, record := range out.Evidence {
		evidence[record.ID] = record
	}
	for _, edge := range out.Edges {
		validateInferenceEvidence(
			edge.InferenceRecord, edge.AcceptanceReceipt, edge.Kind == InferenceAcceptedLift, evidence, issues,
		)
		if edge.Kind == InferenceIndependentVerification && !hasIndependentEvidence(edge.Evidence, evidence) {
			issues.add("independent-evidence", edge.RecordID, "verification requires independent evidence")
		}
	}
	for _, claim := range out.Claims {
		validateInferenceEvidence(claim.InferenceRecord, "", false, evidence, issues)
	}
}

func sortInferenceIssues(issues *InferencePathErrors) {
	sort.Slice(issues.Issues, func(i, j int) bool {
		if issues.Issues[i].Code != issues.Issues[j].Code {
			return issues.Issues[i].Code < issues.Issues[j].Code
		}
		if issues.Issues[i].Record != issues.Issues[j].Record {
			return issues.Issues[i].Record < issues.Issues[j].Record
		}
		return issues.Issues[i].Detail < issues.Issues[j].Detail
	})
}

func registerInferenceID(seen map[ID]string, id ID, kind string, issues *InferencePathErrors) {
	if previous, exists := seen[id]; exists {
		issues.add("stable-id-collision", id, fmt.Sprintf("%s also used by %s", kind, previous))
		return
	}
	seen[id] = kind
}

func validateInferenceEvidence(
	binding InferenceRecord, receipt ID, requireReceipt bool,
	records map[ID]InferenceEvidence, issues *InferencePathErrors,
) {
	refs := make(map[ID]struct{}, len(binding.Evidence))
	for _, ref := range binding.Evidence {
		if _, duplicate := refs[ref.ID]; duplicate {
			issues.add("duplicate-evidence", binding.RecordID, ref.ID.String())
			continue
		}
		refs[ref.ID] = struct{}{}
		record, ok := records[ref.ID]
		if !ok {
			issues.add("orphan-evidence", binding.RecordID, ref.ID.String())
			continue
		}
		if record.Digest != ref.Digest || record.Before != binding.Before ||
			record.After != binding.After || record.Controls != binding.Controls {
			issues.add("stale-evidence", binding.RecordID, ref.ID.String())
		}
	}
	if requireReceipt {
		if receipt == "" {
			issues.add("missing-acceptance-receipt", binding.RecordID, "empty receipt")
		} else if _, ok := refs[receipt]; !ok {
			issues.add("orphan-acceptance-receipt", binding.RecordID, receipt.String())
		} else if !records[receipt].SourceBacked || records[receipt].Before.Source == "" ||
			records[receipt].After.Source == "" {
			issues.add("unbacked-acceptance-receipt", binding.RecordID, receipt.String())
		}
	}
}

func hasIndependentEvidence(refs []EvidenceReference, records map[ID]InferenceEvidence) bool {
	for _, ref := range refs {
		if records[ref.ID].Independent {
			return true
		}
	}
	return false
}

func (p InferencePathV1) Validate() error { _, err := p.Normalized(); return err }

func (p InferencePathV1) StableHash() string { return StableHashString(p.Canonical()) }

// InferencePathChain proves one finite, unambiguous chain over typed edges.
type InferencePathChain struct{ Edges []InferenceEdge }

func NewInferencePathChain(edges ...InferenceEdge) (InferencePathChain, error) {
	if len(edges) == 0 {
		return InferencePathChain{}, fmt.Errorf("%w: path_orphan: empty path", ErrInferencePath)
	}
	normalized := make([]InferenceEdge, 0, len(edges))
	bySubject := make(map[ID][]InferenceEdge, len(edges))
	objects := make(map[ID]struct{}, len(edges))
	seen := make(map[ID]struct{}, len(edges))
	for _, raw := range edges {
		edge, err := raw.normalized()
		if err != nil {
			return InferencePathChain{}, fmt.Errorf("%w: path_orphan: %v", ErrInferencePath, err)
		}
		if _, exists := seen[edge.RecordID]; exists {
			return InferencePathChain{}, fmt.Errorf("%w: path_ambiguity: duplicate edge %s", ErrInferencePath, edge.RecordID)
		}
		seen[edge.RecordID] = struct{}{}
		normalized = append(normalized, edge)
		bySubject[edge.SubjectID] = append(bySubject[edge.SubjectID], edge)
		objects[edge.ObjectID] = struct{}{}
	}
	starts := make([]ID, 0, len(bySubject))
	for subject := range bySubject {
		if _, hasIncoming := objects[subject]; !hasIncoming {
			starts = append(starts, subject)
		}
	}
	if len(starts) != 1 {
		return InferencePathChain{}, fmt.Errorf("%w: path_ambiguity: want one start, got %d", ErrInferencePath, len(starts))
	}
	ordered := make([]InferenceEdge, 0, len(edges))
	visited := make(map[ID]struct{}, len(edges))
	current := starts[0]
	for {
		outgoing := bySubject[current]
		if len(outgoing) == 0 {
			break
		}
		if len(outgoing) != 1 {
			return InferencePathChain{}, fmt.Errorf(
				"%w: path_ambiguity: %s has %d outgoing edges", ErrInferencePath, current, len(outgoing),
			)
		}
		edge := outgoing[0]
		if _, exists := visited[edge.RecordID]; exists {
			return InferencePathChain{}, fmt.Errorf("%w: path_orphan: cycle at %s", ErrInferencePath, edge.RecordID)
		}
		visited[edge.RecordID] = struct{}{}
		ordered = append(ordered, edge)
		current = edge.ObjectID
	}
	if len(ordered) != len(normalized) {
		return InferencePathChain{}, fmt.Errorf(
			"%w: path_orphan: %d edges are disconnected", ErrInferencePath, len(normalized)-len(ordered),
		)
	}
	return InferencePathChain{Edges: ordered}, nil
}

func (c InferencePathChain) Validate() error {
	_, err := NewInferencePathChain(c.Edges...)
	return err
}

func (c InferencePathChain) Canonical() string {
	if normalized, err := NewInferencePathChain(c.Edges...); err == nil {
		c = normalized
	}
	var b strings.Builder
	b.WriteString("inference-chain\t")
	b.WriteString(fmt.Sprint(len(c.Edges)))
	b.WriteByte('\n')
	for _, edge := range c.Edges {
		b.WriteString(edge.Canonical())
		b.WriteByte('\n')
	}
	return b.String()
}
