package pathclosure

import (
	"encoding/hex"
	"fmt"
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

// R4SchemaVersion identifies the finite, receipt-bound path contract. It is
// deliberately separate from the semantic inference-path schema.
const R4SchemaVersion = "gooo-path-closure-r4/v1"

type R4Phase string

const (
	R4CompilePhase R4Phase = "compile"
	R4RuntimePhase R4Phase = "runtime"
)

const (
	CodeInvalidPath            = "INVALID_PATH"
	CodeConflictingReceipt     = "CONFLICTING_RECEIPT"
	CodeMissingProvider        = "MISSING_PROVIDER_BINDING"
	CodeMissingProviderBinding = CodeMissingProvider
	CodePhaseMismatch          = "PHASE_MISMATCH"
	CodeMissingObserver        = "MISSING_OBSERVER"
	CodeOpenWorld              = "OPEN_WORLD"
	CodeUnexhaustedBoundary    = "UNEXHAUSTED_BOUNDARY"
	CodeMissingRequiredPaths   = "MISSING_REQUIRED_PATHS"
)

// R4Record is one canonical producer record. Its ReceiptID is an explicit
// binding to the observer/provider receipt; it is never inferred by lookup.
type R4Record struct {
	ID             semantic.ID
	SubjectID      semantic.ID
	ObjectID       semantic.ID
	ProviderID     semantic.ID
	ProviderDigest string
	Phase          R4Phase
	PhaseDigest    string
	PredecessorID  semantic.ID
	ReceiptID      semantic.ID
	Writes         bool
	Effect         string
}

// R4Receipt is append-only evidence for exactly one record. EventID is kept
// distinct from ReceiptID so replay cannot silently reuse a physical event.
type R4Receipt struct {
	ID             semantic.ID
	EventID        semantic.ID
	RecordID       semantic.ID
	ProviderID     semantic.ID
	ProviderDigest string
	Phase          R4Phase
	PhaseDigest    string
	ObserverID     semantic.ID
	Writes         bool
	Effect         string
}

// R4Path is an ordered, finite claim. RecordBytes are the exact canonical JSON
// bytes that must be recomputed for the ordered RecordIDs.
type R4Path struct {
	ID          semantic.ID
	StartID     semantic.ID
	EndID       semantic.ID
	RecordIDs   []semantic.ID
	RecordBytes []string
}

type R4Boundary struct {
	RequiredPathIDs []semantic.ID
	Exhausted       bool
	OpenWorld       bool
}

type R4Input struct {
	Schema   string
	Boundary R4Boundary
	Records  []R4Record
	Receipts []R4Receipt
	Paths    []R4Path
}

type R4Result struct {
	Status              Status
	Code                string
	Reason              string
	RequiredPathIDs     []semantic.ID
	CoveredPathIDs      []semantic.ID
	ProofValid          bool
	PromotionAuthorized bool
	Cost                int
}

func (r R4Record) canonicalFields() r4WireRecord {
	return r4WireRecord{
		ID: r.ID.String(), SubjectID: r.SubjectID.String(), ObjectID: r.ObjectID.String(),
		ProviderID: r.ProviderID.String(), ProviderDigest: r.ProviderDigest,
		Phase: string(r.Phase), PhaseDigest: r.PhaseDigest,
		PredecessorID: r.PredecessorID.String(), ReceiptID: r.ReceiptID.String(),
		Writes: r.Writes, Effect: r.Effect,
	}
}

// CanonicalRecordBytes returns the canonical record object without a trailing
// newline. It does not provide authenticity; the path evaluator rechecks it.
func (r R4Record) CanonicalRecordBytes() ([]byte, error) { return marshalR4Record(r.canonicalFields()) }

func validR4Digest(value string) bool {
	value = strings.TrimSpace(value)
	if len(value) != 64 || strings.ToLower(value) != value {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}

func normalizeR4ID(value semantic.ID, label string) (semantic.ID, error) {
	id, err := semantic.ParseIdentity(value.String())
	if err != nil {
		return "", fmt.Errorf("%s: %w", label, err)
	}
	return id, nil
}

func normalizeR4Record(raw R4Record) (R4Record, error) {
	var err error
	out := raw
	if out.ID, err = normalizeR4ID(raw.ID, "record ID"); err != nil {
		return R4Record{}, err
	}
	if out.SubjectID, err = normalizeR4ID(raw.SubjectID, "subject ID"); err != nil {
		return R4Record{}, err
	}
	if out.ObjectID, err = normalizeR4ID(raw.ObjectID, "object ID"); err != nil {
		return R4Record{}, err
	}
	if raw.ProviderID != "" {
		if out.ProviderID, err = normalizeR4ID(raw.ProviderID, "provider ID"); err != nil {
			return R4Record{}, err
		}
	}
	if raw.PredecessorID != "" {
		if out.PredecessorID, err = normalizeR4ID(raw.PredecessorID, "predecessor ID"); err != nil {
			return R4Record{}, err
		}
	}
	if raw.ReceiptID != "" {
		if out.ReceiptID, err = normalizeR4ID(raw.ReceiptID, "receipt ID"); err != nil {
			return R4Record{}, err
		}
	}
	out.ProviderDigest, out.PhaseDigest, out.Effect = strings.TrimSpace(raw.ProviderDigest), strings.TrimSpace(raw.PhaseDigest), strings.TrimSpace(raw.Effect)
	return out, nil
}

func normalizeR4Receipt(raw R4Receipt) (R4Receipt, error) {
	var err error
	out := raw
	for _, entry := range []struct {
		value  semantic.ID
		label  string
		target *semantic.ID
	}{
		{raw.ID, "receipt ID", &out.ID}, {raw.EventID, "event ID", &out.EventID}, {raw.RecordID, "receipt record ID", &out.RecordID}, {raw.ProviderID, "receipt provider ID", &out.ProviderID}, {raw.ObserverID, "observer ID", &out.ObserverID},
	} {
		if entry.value == "" {
			continue
		}
		if *entry.target, err = normalizeR4ID(entry.value, entry.label); err != nil {
			return R4Receipt{}, err
		}
	}
	out.ProviderDigest, out.PhaseDigest, out.Effect = strings.TrimSpace(raw.ProviderDigest), strings.TrimSpace(raw.PhaseDigest), strings.TrimSpace(raw.Effect)
	return out, nil
}

func normalizeR4Path(raw R4Path) (R4Path, error) {
	var err error
	out := raw
	if out.ID, err = normalizeR4ID(raw.ID, "path ID"); err != nil {
		return R4Path{}, err
	}
	if out.StartID, err = normalizeR4ID(raw.StartID, "path start ID"); err != nil {
		return R4Path{}, err
	}
	if out.EndID, err = normalizeR4ID(raw.EndID, "path end ID"); err != nil {
		return R4Path{}, err
	}
	out.RecordIDs = append([]semantic.ID(nil), raw.RecordIDs...)
	for i, value := range raw.RecordIDs {
		if out.RecordIDs[i], err = normalizeR4ID(value, "path record ID"); err != nil {
			return R4Path{}, err
		}
	}
	out.RecordBytes = append([]string(nil), raw.RecordBytes...)
	return out, nil
}

func sortedR4IDs(values []semantic.ID) []semantic.ID {
	out := append([]semantic.ID(nil), values...)
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

// Canonical is the decision-only representation. It contains no expected or
// display label, so metadata outside the R4 contract cannot affect evidence.
func (r R4Result) Canonical() string {
	return fmt.Sprintf("status=%s|code=%s|reason=%s|required=%v|covered=%v|proof=%t|promotion=%t|cost=%d", r.Status, r.Code, r.Reason, sortedR4IDs(r.RequiredPathIDs), sortedR4IDs(r.CoveredPathIDs), r.ProofValid, r.PromotionAuthorized, r.Cost)
}

// CanonicalDigest seals only the deterministic decision fields. It is not a
// signature and carries no external-authenticity or promotion authority.
func (r R4Result) CanonicalDigest() string {
	return semantic.StableHashString("gooo-path-closure-r4-result/v1\x00" + r.Canonical())
}
