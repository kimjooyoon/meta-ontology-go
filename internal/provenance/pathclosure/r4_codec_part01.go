package pathclosure

import (
	"encoding/json"
)

type r4WireRecord struct {
	ID             string `json:"id"`
	SubjectID      string `json:"subject_id"`
	ObjectID       string `json:"object_id"`
	ProviderID     string `json:"provider_id"`
	ProviderDigest string `json:"provider_digest"`
	Phase          string `json:"phase"`
	PhaseDigest    string `json:"phase_digest"`
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
	ID          string   `json:"id"`
	StartID     string   `json:"start_id"`
	EndID       string   `json:"end_id"`
	RecordIDs   []string `json:"record_ids"`
	RecordBytes []string `json:"record_bytes"`
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
