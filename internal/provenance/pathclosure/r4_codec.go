package pathclosure

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"sort"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

type r4WireRecord struct {
	ID             string `json:"id"`
	SubjectID      string `json:"subject_id"`
	ObjectID       string `json:"object_id"`
	ProviderID     string `json:"provider_id"`
	ProviderDigest string `json:"provider_digest"`
	Phase          string `json:"phase"`
	PhaseDigest    string `json:"phase_digest"`
	Label          string `json:"label"`
	PredecessorID  string `json:"predecessor_id"`
	ReceiptID      string `json:"receipt_id"`
	Writes         bool   `json:"writes"`
	Effect         string `json:"effect"`
}

type r4WireReceipt struct {
	ID             string `json:"id"`
	EventID        string `json:"event_id"`
	RecordID       string `json:"record_id"`
	ProviderID     string `json:"provider_id"`
	ProviderDigest string `json:"provider_digest"`
	Phase          string `json:"phase"`
	PhaseDigest    string `json:"phase_digest"`
	ObserverID     string `json:"observer_id"`
	Writes         bool   `json:"writes"`
	Effect         string `json:"effect"`
}

type r4WirePath struct {
	ID             string   `json:"id"`
	StartID        string   `json:"start_id"`
	EndID          string   `json:"end_id"`
	RecordIDs      []string `json:"record_ids"`
	RecordBytes    []string `json:"record_bytes"`
	ExpectedLabels []string `json:"expected_labels"`
}

type r4WireBoundary struct {
	RequiredPathIDs []string `json:"required_path_ids"`
	Exhausted       bool     `json:"exhausted"`
	OpenWorld       bool     `json:"open_world"`
}

type r4WireInput struct {
	Schema   string          `json:"schema"`
	Boundary r4WireBoundary  `json:"boundary"`
	Records  []r4WireRecord  `json:"records"`
	Receipts []r4WireReceipt `json:"receipts"`
	Paths    []r4WirePath    `json:"paths"`
}

func marshalR4Record(value r4WireRecord) ([]byte, error) { return json.Marshal(value) }

func wireR4Input(value R4Input) r4WireInput {
	records := make([]r4WireRecord, 0, len(value.Records))
	for _, record := range value.Records {
		records = append(records, record.canonicalFields())
	}
	sort.Slice(records, func(i, j int) bool { return records[i].ID < records[j].ID })
	receipts := make([]r4WireReceipt, 0, len(value.Receipts))
	for _, receipt := range value.Receipts {
		receipts = append(receipts, r4WireReceipt{ID: receipt.ID.String(), EventID: receipt.EventID.String(), RecordID: receipt.RecordID.String(), ProviderID: receipt.ProviderID.String(), ProviderDigest: receipt.ProviderDigest, Phase: string(receipt.Phase), PhaseDigest: receipt.PhaseDigest, ObserverID: receipt.ObserverID.String(), Writes: receipt.Writes, Effect: receipt.Effect})
	}
	sort.Slice(receipts, func(i, j int) bool { return receipts[i].ID < receipts[j].ID })
	paths := make([]r4WirePath, 0, len(value.Paths))
	for _, path := range value.Paths {
		ids := make([]string, 0, len(path.RecordIDs))
		for _, id := range path.RecordIDs {
			ids = append(ids, id.String())
		}
		paths = append(paths, r4WirePath{ID: path.ID.String(), StartID: path.StartID.String(), EndID: path.EndID.String(), RecordIDs: ids, RecordBytes: append([]string(nil), path.RecordBytes...), ExpectedLabels: append([]string(nil), path.ExpectedLabels...)})
	}
	sort.Slice(paths, func(i, j int) bool { return paths[i].ID < paths[j].ID })
	ids := make([]string, 0, len(value.Boundary.RequiredPathIDs))
	for _, id := range sortedR4IDs(value.Boundary.RequiredPathIDs) {
		ids = append(ids, id.String())
	}
	return r4WireInput{Schema: value.Schema, Boundary: r4WireBoundary{RequiredPathIDs: ids, Exhausted: value.Boundary.Exhausted, OpenWorld: value.Boundary.OpenWorld}, Records: records, Receipts: receipts, Paths: paths}
}

func r4InputFromWire(value r4WireInput) R4Input {
	input := R4Input{Schema: value.Schema, Boundary: R4Boundary{Exhausted: value.Boundary.Exhausted, OpenWorld: value.Boundary.OpenWorld}}
	for _, id := range value.Boundary.RequiredPathIDs {
		input.Boundary.RequiredPathIDs = append(input.Boundary.RequiredPathIDs, semantic.ID(id))
	}
	for _, record := range value.Records {
		input.Records = append(input.Records, R4Record{ID: semantic.ID(record.ID), SubjectID: semantic.ID(record.SubjectID), ObjectID: semantic.ID(record.ObjectID), ProviderID: semantic.ID(record.ProviderID), ProviderDigest: record.ProviderDigest, Phase: R4Phase(record.Phase), PhaseDigest: record.PhaseDigest, Label: record.Label, PredecessorID: semantic.ID(record.PredecessorID), ReceiptID: semantic.ID(record.ReceiptID), Writes: record.Writes, Effect: record.Effect})
	}
	for _, receipt := range value.Receipts {
		input.Receipts = append(input.Receipts, R4Receipt{ID: semantic.ID(receipt.ID), EventID: semantic.ID(receipt.EventID), RecordID: semantic.ID(receipt.RecordID), ProviderID: semantic.ID(receipt.ProviderID), ProviderDigest: receipt.ProviderDigest, Phase: R4Phase(receipt.Phase), PhaseDigest: receipt.PhaseDigest, ObserverID: semantic.ID(receipt.ObserverID), Writes: receipt.Writes, Effect: receipt.Effect})
	}
	for _, path := range value.Paths {
		converted := R4Path{ID: semantic.ID(path.ID), StartID: semantic.ID(path.StartID), EndID: semantic.ID(path.EndID), RecordBytes: append([]string(nil), path.RecordBytes...), ExpectedLabels: append([]string(nil), path.ExpectedLabels...)}
		for _, id := range path.RecordIDs {
			converted.RecordIDs = append(converted.RecordIDs, semantic.ID(id))
		}
		input.Paths = append(input.Paths, converted)
	}
	return input
}

func walkR4JSON(decoder *json.Decoder) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	if delimiter, ok := token.(json.Delim); ok {
		switch delimiter {
		case '{':
			seen := map[string]struct{}{}
			for decoder.More() {
				key, tokenErr := decoder.Token()
				if tokenErr != nil {
					return tokenErr
				}
				name, ok := key.(string)
				if !ok {
					return fmt.Errorf("JSON object key is not a string")
				}
				if _, exists := seen[name]; exists {
					return fmt.Errorf("duplicate JSON field %q", name)
				}
				seen[name] = struct{}{}
				if err := walkR4JSON(decoder); err != nil {
					return err
				}
			}
			_, err = decoder.Token()
			return err
		case '[':
			for decoder.More() {
				if err := walkR4JSON(decoder); err != nil {
					return err
				}
			}
			_, err = decoder.Token()
			return err
		default:
			return fmt.Errorf("unexpected JSON delimiter %q", delimiter)
		}
	}
	return nil
}

func requireR4EOF(decoder *json.Decoder) error {
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return fmt.Errorf("strict JSON: trailing value")
		}
		return fmt.Errorf("strict JSON: trailing data: %w", err)
	}
	return nil
}

func decodeStrictR4(data []byte, target any) error {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 || trimmed[0] != '{' {
		return fmt.Errorf("strict JSON: top-level object is required")
	}
	check := json.NewDecoder(bytes.NewReader(data))
	if err := walkR4JSON(check); err != nil {
		return fmt.Errorf("strict JSON: %w", err)
	}
	if err := requireR4EOF(check); err != nil {
		return err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("strict JSON: %w", err)
	}
	return requireR4EOF(decoder)
}

// EncodeR4Input emits strict canonical JSON with stable top-level ordering.
func EncodeR4Input(value R4Input) ([]byte, error) { return json.Marshal(wireR4Input(value)) }

// DecodeR4Input rejects duplicate keys, unknown fields, trailing values, and
// non-canonical whitespace/order. Semantic validity is checked by EvaluateR4.
func DecodeR4Input(data []byte) (R4Input, error) {
	var wire r4WireInput
	if err := decodeStrictR4(data, &wire); err != nil {
		return R4Input{}, err
	}
	value := r4InputFromWire(wire)
	canonical, err := EncodeR4Input(value)
	if err != nil {
		return R4Input{}, err
	}
	if !bytes.Equal(bytes.TrimSpace(data), canonical) {
		return R4Input{}, fmt.Errorf("strict JSON: non-canonical encoding")
	}
	return value, nil
}

func (value R4Input) MarshalJSON() ([]byte, error) { return EncodeR4Input(value) }

func (value *R4Input) UnmarshalJSON(data []byte) error {
	decoded, err := DecodeR4Input(data)
	if err != nil {
		return err
	}
	*value = decoded
	return nil
}
