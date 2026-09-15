package main

type fixture struct {
	Schema              string   `json:"schema"`
	ProjectSource       string   `json:"project_source"`
	ProjectTest         string   `json:"project_test"`
	ContractSource      string   `json:"contract_source"`
	PublishedRoot       string   `json:"published_root"`
	PublishedArtifacts  []string `json:"published_artifacts"`
	ProjectInputFiles   []string `json:"project_input_files"`
	FirstReport         string   `json:"first_report"`
	SecondReport        string   `json:"second_report"`
	Ledger              string   `json:"ledger"`
	Candidate           string   `json:"candidate"`
	AcceptedDecision    string   `json:"accepted_decision"`
	RejectedDecision    string   `json:"rejected_decision"`
	Certificate         string   `json:"certificate"`
	CertificationReport string   `json:"certification_report"`
	ConsumptionReport   string   `json:"consumption_report"`
	BaselineSource      string   `json:"baseline_source"`
	BaselineManifest    string   `json:"baseline_manifest"`
	LearnedSource       string   `json:"learned_source"`
	LearnedManifest     string   `json:"learned_manifest"`
	StaleReport         string   `json:"stale_report"`
	StaleHuman          string   `json:"stale_human"`
	TamperedReport      string   `json:"tampered_report"`
	TamperedHuman       string   `json:"tampered_human"`
	BaselineBuild       string   `json:"baseline_build"`
	BaselineTest        string   `json:"baseline_test"`
	LearnedBuild        string   `json:"learned_build"`
	LearnedTest         string   `json:"learned_test"`
	Measurements        string   `json:"measurements"`
	RepositoryStatus    string   `json:"repository_status"`
}

type buildTestRecord struct {
	Schema              string `json:"schema"`
	Mode                string `json:"mode"`
	Command             string `json:"command"`
	Decision            string `json:"decision"`
	GeneratedFileCount  int    `json:"generated_file_count"`
	GeneratedGoFiles    int    `json:"generated_go_files"`
	Executions          int    `json:"executions"`
	Passed              bool   `json:"passed"`
	WallMS              int64  `json:"wall_ms"`
	PeakRSSKib          int64  `json:"peak_rss_kib"`
	RepositoryWrites    int    `json:"repository_writes"`
	LocalTestExecutions int    `json:"local_test_executions"`
}

type measurementObservation struct {
	Index                   int    `json:"index"`
	WallMS                  int64  `json:"wall_ms"`
	PeakRSSKib              int64  `json:"peak_rss_kib"`
	SourceDigest            string `json:"source_digest"`
	InputSemanticDigest     string `json:"input_semantic_digest"`
	CompilerDigest          string `json:"compiler_digest"`
	ToolchainDigest         string `json:"toolchain_digest"`
	ContractDigest          string `json:"contract_digest"`
	EvaluatorDigest         string `json:"evaluator_digest"`
	GeneratedOutputDigest   string `json:"generated_output_digest"`
	GeneratedManifestDigest string `json:"generated_manifest_digest"`
}

type measurementReport struct {
	Schema            string                   `json:"schema"`
	RuntimeComparable bool                     `json:"runtime_comparable"`
	UnknownStage      string                   `json:"unknown_stage"`
	UnknownStep       string                   `json:"unknown_step"`
	UnknownReason     string                   `json:"unknown_reason"`
	UnknownClass      string                   `json:"unknown_class"`
	NextOperation     string                   `json:"next_operation"`
	Baseline          []measurementObservation `json:"baseline"`
	Learned           []measurementObservation `json:"learned"`
}

type repositoryStatus struct {
	Schema string `json:"schema"`
	Before string `json:"before"`
	After  string `json:"after"`
}

type utilityCase struct {
	ID                           string   `json:"id"`
	ExpectedDecision             string   `json:"expected_decision"`
	ObservedDecision             string   `json:"observed_decision"`
	Reason                       string   `json:"reason"`
	CandidateDigest              string   `json:"candidate_digest,omitempty"`
	UnknownStage                 string   `json:"unknown_stage,omitempty"`
	UnknownStep                  string   `json:"unknown_step,omitempty"`
	UnknownReason                string   `json:"unknown_reason,omitempty"`
	UnknownClass                 string   `json:"unknown_class,omitempty"`
	NextOperation                string   `json:"next_operation,omitempty"`
	BlockedBy                    []string `json:"blocked_by,omitempty"`
	ByteMismatches               int      `json:"byte_mismatches"`
	NormalizedSemanticMismatches int      `json:"normalized_semantic_mismatches"`
	DigestEdgesExpected          int      `json:"digest_edges_expected"`
	DigestEdgesObserved          int      `json:"digest_edges_observed"`
	BuildExecutions              int      `json:"build_executions"`
	TestExecutions               int      `json:"test_executions"`
	CertificateHits              int      `json:"certificate_hits"`
	CertificateMisses            int      `json:"certificate_misses"`
	ArtifactsProduced            int      `json:"artifacts_produced"`
}

type utilityReport struct {
	Schema               string                   `json:"schema"`
	Decision             string                   `json:"decision"`
	Reason               string                   `json:"reason"`
	ContractSourceDigest string                   `json:"contract_source_digest"`
	ProjectSourceDigest  string                   `json:"project_source_digest"`
	Project              projectSemantics         `json:"project"`
	Generated            generatedEvidence        `json:"generated"`
	OperationCounts      operationCounts          `json:"operation_counts"`
	Continuity           continuityEvidence       `json:"continuity"`
	Comparisons          comparisonEvidence       `json:"comparisons"`
	BuildTest            buildTestEvidence        `json:"build_test"`
	Measurements         measurementEvidence      `json:"measurements"`
	CertificateCache     certificateCacheEvidence `json:"certificate_cache"`
	Journey              []journeyStep            `json:"journey"`
	Cases                []utilityCase            `json:"cases"`
	CaseDenominator      int                      `json:"case_denominator"`
	ClosedCases          int                      `json:"closed_cases"`
	UnknownCases         int                      `json:"unknown_cases"`
	RefutedCases         int                      `json:"refuted_cases"`
	ArtifactDenominator  int                      `json:"artifact_denominator"`
	ArtifactCount        int                      `json:"artifact_count"`
	RepositoryWrites     int                      `json:"repository_writes"`
	LocalTestExecutions  int                      `json:"local_test_executions"`
}

type projectSemantics struct {
	Package                string `json:"package"`
	Namespace              string `json:"namespace"`
	InputDescendantDirs    int    `json:"input_descendant_dirs"`
	InputRegularFiles      int    `json:"input_regular_files"`
	InputPhysicalLines     int    `json:"input_physical_lines"`
	InputGoFiles           int    `json:"input_go_files"`
	InputGoPhysicalLines   int    `json:"input_go_physical_lines"`
	InputGoooFiles         int    `json:"input_gooo_files"`
	InputGoooPhysicalLines int    `json:"input_gooo_physical_lines"`
	Entities               int    `json:"entities"`
	Activities             int    `json:"activities"`
	Relations              int    `json:"relations"`
	RootReadmeExcluded     bool   `json:"root_readme_excluded"`
}

type generatedEvidence struct {
	BaselineFiles           int `json:"baseline_files"`
	LearnedFiles            int `json:"learned_files"`
	BaselineGoFiles         int `json:"baseline_go_files"`
	LearnedGoFiles          int `json:"learned_go_files"`
	BaselineGoPhysicalLines int `json:"baseline_go_physical_lines"`
	LearnedGoPhysicalLines  int `json:"learned_go_physical_lines"`
	BaselineGoBytes         int `json:"baseline_go_bytes"`
	LearnedGoBytes          int `json:"learned_go_bytes"`
}

type operationCounts struct {
	Baseline operationSet `json:"baseline"`
	Learned  operationSet `json:"learned"`
}

type operationSet struct {
	Semantic   int `json:"semantic"`
	Lowering   int `json:"lowering"`
	Generation int `json:"generation"`
}

type continuityEvidence struct {
	CandidateToDecision      int      `json:"candidate_to_decision"`
	DecisionToCertificate    int      `json:"decision_to_certificate"`
	CertificateToConsumption int      `json:"certificate_to_consumption"`
	EdgesExpected            int      `json:"edges_expected"`
	EdgesObserved            int      `json:"edges_observed"`
	EdgeNames                []string `json:"edge_names"`
}

type comparisonEvidence struct {
	GeneratedSourceByteMismatches         int  `json:"generated_source_byte_mismatches"`
	GeneratedManifestNormalizedMismatches int  `json:"generated_manifest_normalized_mismatches"`
	CandidateCertificateByteMismatches    int  `json:"candidate_certificate_byte_mismatches"`
	GeneratedSourceBytesEqual             bool `json:"generated_source_bytes_equal"`
	NormalizedSemanticEqual               bool `json:"normalized_semantic_equal"`
}

type buildTestEvidence struct {
	BuildExecutions      int64 `json:"build_executions"`
	TestExecutions       int64 `json:"test_executions"`
	BaselineBuildMS      int64 `json:"baseline_build_ms"`
	BaselineBuildPeakRSS int64 `json:"baseline_build_peak_rss_kib"`
	BaselineTestMS       int64 `json:"baseline_test_ms"`
	BaselineTestPeakRSS  int64 `json:"baseline_test_peak_rss_kib"`
	LearnedBuildMS       int64 `json:"learned_build_ms"`
	LearnedBuildPeakRSS  int64 `json:"learned_build_peak_rss_kib"`
	LearnedTestMS        int64 `json:"learned_test_ms"`
	LearnedTestPeakRSS   int64 `json:"learned_test_peak_rss_kib"`
}

type measurementEvidence struct {
	Baseline          []measurementObservation `json:"baseline"`
	Learned           []measurementObservation `json:"learned"`
	RuntimeComparable bool                     `json:"runtime_comparable"`
	UnknownStage      string                   `json:"unknown_stage"`
	UnknownReason     string                   `json:"unknown_reason"`
	NextOperation     string                   `json:"next_operation"`
}

type certificateCacheEvidence struct {
	BaselineHits   int `json:"baseline_hits"`
	BaselineMisses int `json:"baseline_misses"`
	LearnedHits    int `json:"learned_hits"`
	LearnedMisses  int `json:"learned_misses"`
	NegativeHits   int `json:"negative_hits"`
	NegativeMisses int `json:"negative_misses"`
}

type journeyStep struct {
	Ordinal  string `json:"ordinal"`
	Action   string `json:"action"`
	Decision string `json:"decision"`
}
