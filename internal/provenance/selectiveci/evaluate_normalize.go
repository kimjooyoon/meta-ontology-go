package selectiveci

import (
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

type issueClass uint8

const (
	issueNone issueClass = iota
	issueUnknown
	issueFailClosed
)

type issueState struct {
	class issueClass
	code  string
}

func (s *issueState) add(class issueClass, code string) {
	if class > s.class || class == s.class && code < s.code {
		s.class, s.code = class, code
	}
}

func normalizeSnapshot(raw semantic.SnapshotDigests, label string, state *issueState) semantic.SnapshotDigests {
	if raw.Source == "" || raw.Semantic == "" {
		state.add(issueUnknown, CodeMissing)
		return raw
	}
	source, sourceErr := normalizeDigest(raw.Source, label+" source snapshot")
	semanticDigest, semanticErr := normalizeDigest(raw.Semantic, label+" semantic snapshot")
	if sourceErr != nil || semanticErr != nil {
		state.add(issueFailClosed, CodeDigestMismatch)
	}
	return semantic.SnapshotDigests{Source: source, Semantic: semanticDigest}
}

func normalizeSequence(values []semantic.ID, label string) ([]semantic.ID, error) {
	if len(values) == 0 {
		return nil, fmt.Errorf("%s is empty", label)
	}
	out := make([]semantic.ID, len(values))
	seen := make(map[semantic.ID]struct{}, len(values))
	for i, value := range values {
		id, err := normalizeID(value, label)
		if err != nil {
			return nil, err
		}
		if _, exists := seen[id]; exists {
			return nil, fmt.Errorf("duplicate %s %s", label, id)
		}
		seen[id] = struct{}{}
		out[i] = id
	}
	return out, nil
}

func normalizePath(raw Path) (Path, error) {
	path := raw
	var err error
	if path.PathID, err = normalizeID(raw.PathID, "path ID"); err != nil {
		return Path{}, err
	}
	if path.RootID, err = normalizeID(raw.RootID, "path root ID"); err != nil {
		return Path{}, err
	}
	if path.ObligationID, err = normalizeID(raw.ObligationID, "path obligation ID"); err != nil {
		return Path{}, err
	}
	if path.CommandID, err = normalizeID(raw.CommandID, "path command ID"); err != nil {
		return Path{}, err
	}
	if path.ReceiptID, err = normalizeID(raw.ReceiptID, "path receipt ID"); err != nil {
		return Path{}, err
	}
	if len(raw.RecordIDs) == 0 || len(raw.RecordIDs) != len(raw.ExpectedKinds) {
		return Path{}, fmt.Errorf("path record and kind sequences are incomplete")
	}
	if path.RecordIDs, err = normalizeSequence(raw.RecordIDs, "path record ID"); err != nil {
		return Path{}, err
	}
	path.ExpectedKinds = append([]semantic.InferenceKind(nil), raw.ExpectedKinds...)
	for _, kind := range path.ExpectedKinds {
		if !kind.Valid() {
			return Path{}, fmt.Errorf("unknown path inference kind %q", kind)
		}
	}
	return path, nil
}

func normalizeCommandReceipt(raw CommandReceipt, input Input, state *issueState) CommandReceipt {
	receipt := raw
	var err error
	if receipt.CommandID, err = normalizeID(raw.CommandID, "command receipt command ID"); err != nil {
		state.add(issueFailClosed, CodeMalformed)
	}
	if receipt.ReceiptID, err = normalizeID(raw.ReceiptID, "command receipt ID"); err != nil {
		state.add(issueFailClosed, CodeMalformed)
	}
	for label, value := range map[string]string{
		"provider receipt": raw.ProviderReceiptDigest,
		"phase receipt":    raw.PhaseReceiptDigest,
		"resource receipt": raw.ResourceReceiptDigest,
		"registry":         raw.RegistryDigest,
		"plan":             raw.PlanDigest,
	} {
		if value == "" {
			state.add(issueUnknown, CodeMissing)
			continue
		}
		if _, err := normalizeDigest(value, label+" digest"); err != nil {
			state.add(issueFailClosed, CodeDigestMismatch)
		}
	}
	receipt.ProviderReceiptDigest, _ = normalizeDigest(raw.ProviderReceiptDigest, "provider receipt")
	receipt.PhaseReceiptDigest, _ = normalizeDigest(raw.PhaseReceiptDigest, "phase receipt")
	receipt.ResourceReceiptDigest, _ = normalizeDigest(raw.ResourceReceiptDigest, "resource receipt")
	receipt.RegistryDigest, _ = normalizeDigest(raw.RegistryDigest, "registry")
	receipt.PlanDigest, _ = normalizeDigest(raw.PlanDigest, "plan")
	if receipt.RegistryDigest != input.RegistryDigest || receipt.PlanDigest != input.PlanDigest {
		state.add(issueFailClosed, CodeDigestMismatch)
	}
	switch receipt.Status {
	case ReceiptVerified, ReceiptCandidate, ReceiptDeferred, ReceiptNotRun:
	default:
		state.add(issueFailClosed, CodeMalformed)
	}
	if receipt.Digest == "" {
		state.add(issueUnknown, CodeMissing)
	} else if _, err := normalizeDigest(receipt.Digest, "command receipt digest"); err != nil {
		state.add(issueFailClosed, CodeReceiptMismatch)
	}
	receipt.Digest, _ = normalizeDigest(raw.Digest, "command receipt digest")
	return receipt
}
