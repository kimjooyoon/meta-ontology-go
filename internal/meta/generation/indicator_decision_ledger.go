package generation

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"sort"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

const IndicatorDecisionLedgerSchemaVersion = "gooo/indicator-decision-ledger/v1"

type TrilemmaRoute string

const (
	TrilemmaRouteFoundation TrilemmaRoute = "FOUNDATION"
	TrilemmaRouteCoherence  TrilemmaRoute = "COHERENCE"
	TrilemmaRouteRegression TrilemmaRoute = "REGRESSION"
)

type IndicatorDisposition string

const (
	IndicatorDispositionExempt         IndicatorDisposition = "EXEMPT"
	IndicatorDispositionConforming     IndicatorDisposition = "CONFORMING"
	IndicatorDispositionRepairSelected IndicatorDisposition = "REPAIR_SELECTED"
)

// IndicatorDecisionLedgerEntry binds one measured indicator to exactly one
// logical closure route and, for violations, one generated meta operation.
type IndicatorDecisionLedgerEntry struct {
	IndicatorID      string                        `json:"indicator_id"`
	SourceIndicator  sourcepolicy.Indicator        `json:"source_indicator"`
	IndicatorOutcome sourcepolicy.IndicatorOutcome `json:"indicator_outcome"`
	TrilemmaRoute    TrilemmaRoute                 `json:"trilemma_route"`
	Disposition      IndicatorDisposition          `json:"disposition"`
	Action           *Action                       `json:"action,omitempty"`
}

// IndicatorDecisionLedger is a deterministic, replayable proof that every
// source indicator either closes, is explicitly exempt, or selects a repair.
type IndicatorDecisionLedger struct {
	SchemaVersion     string                         `json:"schema_version"`
	IndicatorCount    int                            `json:"indicator_count"`
	SelectedCount     int                            `json:"selected_count"`
	FoundationalCount int                            `json:"foundational_count"`
	CoherenceCount    int                            `json:"coherence_count"`
	RegressiveCount   int                            `json:"regressive_count"`
	Entries           []IndicatorDecisionLedgerEntry `json:"entries"`
	Digest            string                         `json:"digest"`
}

func BuildIndicatorDecisionLedger(indicators []sourcepolicy.Indicator, actions []Action) (IndicatorDecisionLedger, error) {
	actionsByIndicator, err := indexLedgerActions(actions)
	if err != nil {
		return IndicatorDecisionLedger{}, err
	}
	entries, selectedCount, err := buildLedgerEntries(indicators, actionsByIndicator)
	if err != nil {
		return IndicatorDecisionLedger{}, err
	}
	sort.Slice(entries, func(left, right int) bool {
		return entries[left].IndicatorID < entries[right].IndicatorID
	})
	ledger := IndicatorDecisionLedger{
		SchemaVersion:  IndicatorDecisionLedgerSchemaVersion,
		IndicatorCount: len(entries),
		SelectedCount:  selectedCount,
		Entries:        entries,
	}
	for _, entry := range entries {
		switch entry.TrilemmaRoute {
		case TrilemmaRouteFoundation:
			ledger.FoundationalCount++
		case TrilemmaRouteCoherence:
			ledger.CoherenceCount++
		case TrilemmaRouteRegression:
			ledger.RegressiveCount++
		}
	}
	digest, err := indicatorDecisionLedgerDigest(ledger)
	if err != nil {
		return IndicatorDecisionLedger{}, err
	}
	ledger.Digest = digest
	return ledger, nil
}

func indexLedgerActions(actions []Action) (map[string]Action, error) {
	indexed := make(map[string]Action, len(actions))
	for _, action := range actions {
		if !actionMatchesSourceIndicator(action) {
			return nil, fmt.Errorf("action %q does not match its source indicator", action.IndicatorID)
		}
		if !reflect.DeepEqual(action.IndicatorOutcome, action.SourceIndicator.Outcome()) {
			return nil, fmt.Errorf("action %q carries a forged indicator outcome", action.IndicatorID)
		}
		if _, exists := indexed[action.IndicatorID]; exists {
			return nil, fmt.Errorf("duplicate action for indicator %q", action.IndicatorID)
		}
		indexed[action.IndicatorID] = action
	}
	return indexed, nil
}

func buildLedgerEntries(indicators []sourcepolicy.Indicator, actions map[string]Action) ([]IndicatorDecisionLedgerEntry, int, error) {
	entries := make([]IndicatorDecisionLedgerEntry, 0, len(indicators))
	seen := make(map[string]struct{}, len(indicators))
	selectedCount := 0
	for _, indicator := range indicators {
		id := indicatorID(indicator)
		if _, exists := seen[id]; exists {
			return nil, 0, fmt.Errorf("duplicate source indicator %q", id)
		}
		seen[id] = struct{}{}
		action, hasAction := actions[id]
		entry, selected, err := buildLedgerEntry(id, indicator, action, hasAction)
		if err != nil {
			return nil, 0, err
		}
		if selected {
			selectedCount++
		}
		entries = append(entries, entry)
	}
	if selectedCount != len(actions) {
		return nil, 0, fmt.Errorf("%d actions do not belong to the indicator set", len(actions)-selectedCount)
	}
	return entries, selectedCount, nil
}

func buildLedgerEntry(id string, indicator sourcepolicy.Indicator, action Action, hasAction bool) (IndicatorDecisionLedgerEntry, bool, error) {
	route, err := indicatorTrilemmaRoute(indicator.Proof)
	if err != nil {
		return IndicatorDecisionLedgerEntry{}, false, fmt.Errorf("indicator %q: %w", id, err)
	}
	entry := IndicatorDecisionLedgerEntry{IndicatorID: id, SourceIndicator: indicator,
		IndicatorOutcome: indicator.Outcome(), TrilemmaRoute: route}
	switch indicator.Applicability {
	case sourcepolicy.ApplicabilityNotApplicable:
		if !indicator.Satisfied {
			return IndicatorDecisionLedgerEntry{}, false, fmt.Errorf("not-applicable indicator %q must be closed", id)
		}
		if hasAction {
			return IndicatorDecisionLedgerEntry{}, false, fmt.Errorf("not-applicable indicator %q selected an action", id)
		}
		entry.Disposition = IndicatorDispositionExempt
		return entry, false, nil
	case sourcepolicy.ApplicabilityApplicable:
		return buildApplicableLedgerEntry(entry, action, hasAction)
	default:
		return IndicatorDecisionLedgerEntry{}, false, fmt.Errorf("indicator %q has unknown applicability %q", id, indicator.Applicability)
	}
}

func buildApplicableLedgerEntry(entry IndicatorDecisionLedgerEntry, action Action, hasAction bool) (IndicatorDecisionLedgerEntry, bool, error) {
	if entry.SourceIndicator.Satisfied {
		if hasAction {
			return IndicatorDecisionLedgerEntry{}, false, fmt.Errorf("conforming indicator %q selected an action", entry.IndicatorID)
		}
		entry.Disposition = IndicatorDispositionConforming
		return entry, false, nil
	}
	if !hasAction {
		return IndicatorDecisionLedgerEntry{}, false, fmt.Errorf("violating indicator %q has no selected repair", entry.IndicatorID)
	}
	actionCopy := action
	entry.Disposition = IndicatorDispositionRepairSelected
	entry.Action = &actionCopy
	return entry, true, nil
}

func indicatorTrilemmaRoute(proof sourcepolicy.ProofChoice) (TrilemmaRoute, error) {
	switch proof {
	case sourcepolicy.ProofFoundation:
		return TrilemmaRouteFoundation, nil
	case sourcepolicy.ProofCoherence:
		return TrilemmaRouteCoherence, nil
	case sourcepolicy.ProofRegression:
		return TrilemmaRouteRegression, nil
	default:
		return "", fmt.Errorf("unknown proof choice %q", proof)
	}
}

func (ledger IndicatorDecisionLedger) Validate() error {
	indicators := make([]sourcepolicy.Indicator, 0, len(ledger.Entries))
	actions := make([]Action, 0, ledger.SelectedCount)
	for _, entry := range ledger.Entries {
		indicators = append(indicators, entry.SourceIndicator)
		if entry.Action != nil {
			actions = append(actions, *entry.Action)
		}
	}
	rebuilt, err := BuildIndicatorDecisionLedger(indicators, actions)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(ledger, rebuilt) {
		return fmt.Errorf("indicator decision ledger does not match its canonical replay")
	}
	return nil
}

func (ledger *IndicatorDecisionLedger) UnmarshalJSON(data []byte) error {
	type wire IndicatorDecisionLedger
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var candidate wire
	if err := decoder.Decode(&candidate); err != nil {
		return err
	}
	if err := ensureIndicatorLedgerEOF(decoder); err != nil {
		return err
	}
	decoded := IndicatorDecisionLedger(candidate)
	if err := decoded.Validate(); err != nil {
		return err
	}
	*ledger = decoded
	return nil
}

func indicatorDecisionLedgerDigest(ledger IndicatorDecisionLedger) (string, error) {
	ledger.Digest = ""
	payload, err := json.Marshal(ledger)
	if err != nil {
		return "", fmt.Errorf("marshal indicator decision ledger digest material: %w", err)
	}
	digest := sha256.Sum256(payload)
	return "sha256:" + hex.EncodeToString(digest[:]), nil
}

func ensureIndicatorLedgerEOF(decoder *json.Decoder) error {
	var trailing json.RawMessage
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return fmt.Errorf("unexpected trailing JSON value")
		}
		return err
	}
	return nil
}
