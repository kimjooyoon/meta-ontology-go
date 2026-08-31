package main

const (
	evidenceSchema   = "gooo.authorized-rewrite-write-set.v3"
	densitySchema    = "gooo.line-density-rewrite.v1"
	extractionSchema = "gooo.function-extraction.v2"
	splitSchema      = "gooo.logical-source-split.v1"
	metaOperation    = "union-ci-selected-rewrite-receipts"
)

type densityReport struct {
	Schema    string           `json:"schema"`
	SourceSHA string           `json:"source_sha"`
	Subjects  []densitySubject `json:"subjects"`
}
type densitySubject struct {
	Logical string `json:"logical"`
	Status  string `json:"status"`
}
type extractionReport struct {
	Schema    string              `json:"schema"`
	SourceSHA string              `json:"source_sha"`
	Subjects  []extractionSubject `json:"subjects"`
	Unhandled []string            `json:"unhandled"`
}
type extractionSubject struct {
	Files   []string `json:"changed_files"`
	Created []string `json:"created_files"`
}

type splitReport struct {
	Schema      string              `json:"schema"`
	SourceSHA   string              `json:"source_sha"`
	Decision    string              `json:"decision"`
	Resolution  string              `json:"resolution"`
	Subjects    []extractionSubject `json:"subjects"`
	Unhandled   []string            `json:"unhandled"`
	Coordinates splitCoordinates    `json:"coordinates"`
	Exact       bool                `json:"exact"`
}

type splitCoordinates struct {
	Selected int `json:"selected_subjects"`
	Applied  int `json:"applied_subjects"`
	Changed  int `json:"changed_paths"`
	Created  int `json:"created_paths"`
	Unknowns int `json:"unknowns"`
}
