package main

type config struct {
	root         string
	coverage     string
	provenance   string
	output       string
	ciConclusion string
}

type provenanceEnvelope struct {
	SchemaVersion  string `json:"schema_version"`
	HeadSHA        string `json:"head_sha"`
	Decision       string `json:"decision"`
	Reason         string `json:"reason"`
	EnvelopeDigest string `json:"envelope_digest"`
	ReplayDigest   string `json:"replay_digest"`
}
