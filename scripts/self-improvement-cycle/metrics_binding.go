package main

type MetricsBinding struct {
	Schema             string          `json:"schema"`
	LogicalRoot        MetricsSnapshot `json:"logical_root"`
	StorageRoot        MetricsSnapshot `json:"storage_root"`
	RootTopologyExempt bool            `json:"root_topology_exempt"`
	RootREADMEExempt   bool            `json:"root_readme_exempt"`
	RootREADMEValue    int             `json:"root_readme_value"`
	RootREADMEOntology string          `json:"root_readme_ontology"`
	RootWitnessDigest  string          `json:"root_witness_digest"`
	RootWitnessCount   int             `json:"root_witness_count"`
	SemanticDigest     string          `json:"semantic_digest"`
}
