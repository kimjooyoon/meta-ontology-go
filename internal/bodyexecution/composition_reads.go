package bodyexecution

import "github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"

const readObservationFile = "gooo-native-read-observation.json"
const readObservationLimit = 65536

type CompositionReadSite struct {
	ExpressionID     string                      `json:"expression_id"`
	RootActivityID   string                      `json:"root_activity_id"`
	ActivityID       string                      `json:"activity_id"`
	Kind             string                      `json:"kind"`
	ElementType      string                      `json:"element_type"`
	NativeOrdinal    int                         `json:"native_ordinal"`
	Expression       string                      `json:"expression"`
	ProjectionSHA256 string                      `json:"projection_sha256"`
	Start            int                         `json:"start_byte"`
	End              int                         `json:"end_byte"`
	Source           *bodycodegen.BodyReadSource `json:"source"`
}

// Input collections name decoded external ports, never a value-matched array.
type CompositionReadInput struct {
	Presence    string `json:"presence,omitempty"`
	ActivityID  string `json:"activity_id"`
	InputKey    string `json:"input_key"`
	Port        string `json:"port"`
	EntityID    string `json:"entity_id"`
	FieldID     string `json:"field_id"`
	FieldName   string `json:"field_name"`
	ElementType string `json:"element_type"`
}

type CompositionReadOrigin struct {
	Input int `json:"input_collection"`
	Index int `json:"element_index"`
}

type CompositionReadObservation struct {
	SiteIndex    int                           `json:"site_index"`
	Index        int                           `json:"index"`
	Length       int                           `json:"length"`
	Origin       *CompositionReadOrigin        `json:"origin,omitempty"`
	Reason       string                        `json:"reason,omitempty"`
	Construction *CompositionConstructedOrigin `json:"construction,omitempty"`
}

type CompositionReadSummary struct {
	Schema                   string `json:"schema"`
	Status                   string `json:"status"`
	RecordedReads            int    `json:"recorded_reads"`
	ResolvedReads            int    `json:"resolved_reads"`
	ConstructedReads         int    `json:"constructed_reads,omitempty"`
	InterruptedConstructions int    `json:"interrupted_constructions,omitempty"`
	RecordedConstructions    int    `json:"recorded_constructions,omitempty"`
	ConstructionDropped      int    `json:"construction_dropped,omitempty"`
	UnresolvedReads          int    `json:"unresolved_reads"`
	DroppedReads             int    `json:"dropped_reads"`
	RegistrationDropped      int    `json:"registration_dropped"`
	SHA256                   string `json:"sha256"`
	Scope                    string `json:"scope"`
}
