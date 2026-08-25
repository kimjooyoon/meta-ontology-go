package main

const (
	evidenceSchema   = "gooo.authorized-rewrite-write-set.v2"
	densitySchema    = "gooo.line-density-rewrite.v1"
	extractionSchema = "gooo.function-extraction.v1"
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
