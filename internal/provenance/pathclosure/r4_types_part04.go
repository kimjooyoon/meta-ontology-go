package pathclosure

import (
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"slices"
)

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
	slices.Sort(out)
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
