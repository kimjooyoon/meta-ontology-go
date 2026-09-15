package selectiveci

import (
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

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
