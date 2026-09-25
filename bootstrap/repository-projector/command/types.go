package main

import projectionevidence "github.com/kimjooyoon/meta-ontology-go/bootstrap/repository-projector/evidence"

type manifestEntry = projectionevidence.Entry

type manifest struct {
	Schema    string          `json:"schema"`
	SourceSHA string          `json:"source_sha"`
	Proof     string          `json:"proof_choice"`
	Authority string          `json:"proof_authority"`
	Entries   []manifestEntry `json:"entries"`
}
