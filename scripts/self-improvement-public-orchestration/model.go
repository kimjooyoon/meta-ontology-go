package main

import "github.com/kimjooyoon/meta-ontology-go/internal/meta/publicorchestration"

type orchestrationInput struct {
	Policy        string
	Source        string
	Contract      string
	Gooo          string
	TestReuse     string
	ProjectTest   string
	RepoRoot      string
	Output        string
	Handoff       string
	Candidate     string
	Authorization string
}

type evidenceInput struct {
	Schema                  string   `json:"schema"`
	Policy                  string   `json:"policy"`
	Source                  string   `json:"source"`
	ProjectTest             string   `json:"project_test"`
	GeneratedProgram        string   `json:"generated_program"`
	GeneratedManifest       string   `json:"generated_manifest"`
	Certificate             string   `json:"certificate"`
	Candidate               string   `json:"candidate"`
	Handoff                 string   `json:"handoff"`
	Authorization           string   `json:"authorization"`
	RejectedAuthorization   string   `json:"rejected_authorization"`
	PrepareReport           string   `json:"prepare_report"`
	ResumeReport            string   `json:"resume_report"`
	Receipt                 string   `json:"receipt"`
	BaselineReuse           string   `json:"baseline_reuse"`
	ReplayReuse             string   `json:"replay_reuse"`
	MissingAuthorization    string   `json:"missing_authorization"`
	MalformedContinuation   string   `json:"malformed_continuation"`
	ContradictoryCandidate  string   `json:"contradictory_candidate"`
	MismatchedAuthorization string   `json:"mismatched_authorization"`
	RuntimeMeasurements     string   `json:"runtime_measurements"`
	RepositoryStatus        string   `json:"repository_status"`
	PublishedRoot           string   `json:"published_root"`
	PublishedArtifacts      []string `json:"published_artifacts"`
}

type validationCheck struct {
	Schema              string `json:"schema"`
	Decision            string `json:"decision"`
	Passed              bool   `json:"passed"`
	BuildExecutions     int    `json:"build_executions"`
	TestExecutions      int    `json:"test_executions"`
	GeneratedFileCount  int    `json:"generated_file_count"`
	GeneratedGoFiles    int    `json:"generated_go_files"`
	WallMS              int64  `json:"wall_ms"`
	PeakRSSKib          int64  `json:"peak_rss_kib"`
	RepositoryWrites    int    `json:"repository_writes"`
	LocalTestExecutions int    `json:"local_test_executions"`
}

type timingEvidence struct {
	Schema                 string `json:"schema"`
	ManualWallMS           int64  `json:"manual_wall_ms"`
	ManualPeakRSSKib       int64  `json:"manual_peak_rss_kib"`
	OrchestratedWallMS     int64  `json:"orchestrated_wall_ms"`
	OrchestratedPeakRSSKib int64  `json:"orchestrated_peak_rss_kib"`
}

type reportAlias = publicorchestration.Report
