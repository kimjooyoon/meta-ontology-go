package languagesemantic

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesemantic/replay"
)

type UpstreamEvidence struct {
	CaseID           string   `json:"case_id"`
	ObservedDecision string   `json:"observed_decision"`
	Diagnostics      []string `json:"diagnostics,omitempty"`
}

type LawEvidence struct {
	Law         string                `json:"law"`
	Satisfied   bool                  `json:"satisfied"`
	Observation replay.LawObservation `json:"observation"`
}

type CaseEvidence struct {
	Source   *replay.Observation `json:"source,omitempty"`
	Law      *LawEvidence        `json:"law,omitempty"`
	Upstream *UpstreamEvidence   `json:"upstream,omitempty"`
	Error    string              `json:"error,omitempty"`
}

type CaseResult struct {
	Definition Definition   `json:"definition"`
	Evidence   CaseEvidence `json:"evidence"`
	Status     CaseStatus   `json:"status"`
	Digest     string       `json:"evidence_digest"`
}
