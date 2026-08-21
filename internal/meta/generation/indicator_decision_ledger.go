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
	actionsByIndicator := make(map[string]Action, len(actions))
	for _, action := range actions {
		if !actionMatchesSourceIndicator(action) {
			return IndicatorDecisionLedger{}, fmt.Errorf("action %q does not match its source indicator", action.IndicatorID)
		}
		if !reflect.DeepEqual(action.IndicatorOutcome, action.SourceIndicator.Outcome()) {
			return IndicatorDecisionLedger{}, fmt.Errorf("action %q carries a forged indicator outcome", action.IndicatorID)
		}
		if _, exists := actionsByIndicator[action.IndicatorID]; exists {
			return IndicatorDecisionLedger{}, fmt.Errorf("duplicate action for indicator %q", action.IndicatorID)
		}
		actionsByIndicator[action.IndicatorID] = action
	}

	entries := make([]IndicatorDecisionLedgerEntry, 0, len(indicators))
	seenIndicators := make(map[string]struct{}, len(indicators))
	selected := make(map[string]struct{}, len(actions))
	for _, indicator := range indicators {
		id := indicatorID(indicator)
		if _, exists := seenIndicators[id]; exists {
			return IndicatorDecisionLedger{}, fmt.Errorf("duplicate source indicator %q", id)
		}
		seenIndicators[id] = struct{}{}

		action, hasAction := actionsByIndicator[id]
		route, err := indicatorTrilemmaRoute(indicator.Proof)
		if err != nil {
			return IndicatorDecisionLedger{}, fmt.Errorf("indicator %q: %w", id, err)
		}
		entry := IndicatorDecisionLedgerEntry{
			IndicatorID:      id,
			SourceIndicator:  indicator,
			IndicatorOutcome: indicator.Outcome(),
			TrilemmaRoute:    route,
		}
		switch indicator.Applicability {
		case sourcepolicy.ApplicabilityNotApplicable:
			if !indicator.Satisfied {
				return IndicatorDecisionLedger{}, fmt.Errorf("not-applicable indicator %q must be closed", id)
			}
			if hasAction {
				return IndicatorDecisionLedger{}, fmt.Errorf("not-applicable indicator %q selected an action", id)
			}
			entry.Disposition = IndicatorDispositionExempt
		case sourcepolicy.ApplicabilityApplicable:
			if indicator.Satisfied {
				if hasAction {
					return IndicatorDecisionLedger{}, fmt.Errorf("conforming indicator %q selected an action", id)
				}
				entry.Disposition = IndicatorDispositionConforming
				break
			}
			if !hasAction {
				return IndicatorDecisionLedger{}, fmt.Errorf("violating indicator %q has no selected repair", id)
			}
			actionCopy := action
			entry.Disposition = IndicatorDispositionRepairSelected
			entry.Action = &actionCopy
			selected[id] = struct{}{}
		default:
			return IndicatorDecisionLedger{}, fmt.Errorf("indicator %q has unknown applicability %q", id, indicator.Applicability)
		}
		entries = append(entries, entry)
	}
	if len(selected) != len(actionsByIndicator) {
		return IndicatorDecisionLedger{}, fmt.Errorf("%d actions do not belong to the indicator set", len(actionsByIndicator)-len(selected))
	}

	sort.Slice(entries, func(left, right int) bool {
		return entries[left].IndicatorID < entries[right].IndicatorID
	})
	ledger := IndicatorDecisionLedger{
		SchemaVersion:  IndicatorDecisionLedgerSchemaVersion,
		IndicatorCount: len(entries),
		SelectedCount:  len(selected),
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
