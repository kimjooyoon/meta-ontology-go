package pathclosure

import (
	"encoding/hex"
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"strings"
)

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
