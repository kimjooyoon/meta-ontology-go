package pathclosure

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"strings"
)

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
