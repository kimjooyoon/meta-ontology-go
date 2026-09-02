package authorization

import capability "github.com/kimjooyoon/meta-ontology-go/internal/meta/externalcapabilityexecution"

type EffectCeiling struct {
	RepositoryWrites         int `json:"repository_writes"`
	ExternalRepositoryWrites int `json:"external_repository_writes"`
	OfficialMutations        int `json:"official_mutations"`
	Promotions               int `json:"promotions"`
}

type Envelope struct {
	Schema                string        `json:"schema"`
	SubjectSHA            string        `json:"subject_sha"`
	Issuer                string        `json:"issuer"`
	Operation             string        `json:"operation"`
	Scope                 string        `json:"scope"`
	PolicySourceDigest    string        `json:"policy_source_digest"`
	PolicyGeneratedDigest string        `json:"policy_generated_digest"`
	SourceReportDigest    string        `json:"source_report_digest"`
	DefaultDecision       string        `json:"default_decision"`
	RunID                 string        `json:"run_id"`
	RunAttempt            int           `json:"run_attempt"`
	Nonce                 string        `json:"nonce"`
	EffectCeiling         EffectCeiling `json:"effect_ceiling"`
	EnvelopeDigest        string        `json:"envelope_digest"`
}

type PolicyEvidence struct {
	SourceAvailable    bool   `json:"source_available"`
	GeneratedAvailable bool   `json:"generated_available"`
	SourceDigest       string `json:"source_digest"`
	GeneratedDigest    string `json:"generated_digest"`
}

type Foundation struct {
	Schema                string `json:"schema"`
	Available             bool   `json:"available"`
	SubjectSHA            string `json:"subject_sha"`
	ProducerRunID         string `json:"producer_run_id"`
	ArtifactID            int64  `json:"artifact_id"`
	ArchiveDigest         string `json:"archive_digest"`
	PolicySourceDigest    string `json:"policy_source_digest"`
	PolicyGeneratedDigest string `json:"policy_generated_digest"`
}

type Invocation struct {
	SubjectSHA string
	RunID      string
	RunAttempt int
}

type Input struct {
	EnvelopeAvailable bool
	ReportAvailable   bool
	Envelope          Envelope
	Report            capability.Report
	Policy            PolicyEvidence
	Foundation        Foundation
	Invocation        Invocation
}
