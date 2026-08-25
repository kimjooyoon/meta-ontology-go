package languagetestexperiment

import "github.com/kimjooyoon/meta-ontology-go/internal/languagetest"

const ReportSchema = "gooo/language-test-experiment-report/v1"

type Observation struct {
	Runtime string               `json:"runtime"`
	Receipt languagetest.Receipt `json:"receipt"`
}

type Input struct {
	SubjectSHA       string               `json:"subject_sha"`
	ExecutableDigest string               `json:"executable_digest"`
	Contract         Contract             `json:"contract"`
	First            Observation          `json:"first"`
	Replay           Observation          `json:"replay"`
	AssertionFailure languagetest.Receipt `json:"assertion_failure"`
	Missing          languagetest.Receipt `json:"missing"`
}

type Counter struct {
	Satisfied int `json:"satisfied"`
	Total     int `json:"total"`
	BasisPoints int `json:"basis_points"`
}

type Compiler struct {
	ExecutableDigest string `json:"executable_digest"`
	Go127Runtimes     int    `json:"go127_runtimes"`
}

type Effects struct {
	RepositoryWrites  int  `json:"repository_writes"`
	MutationAuthority bool `json:"mutation_authority"`
}

type Summary struct {
	Coordinates               Counter  `json:"coordinates"`
	Receipts                  int      `json:"receipts"`
	DeclaredTests             int      `json:"declared_tests"`
	ExecutedTests             int      `json:"executed_tests"`
	PassedTests               int      `json:"passed_tests"`
	SourceCoherence           int      `json:"source_coherence"`
	ReceiptDigestVariants     int      `json:"receipt_digest_variants"`
	ExecutionDigestVariants   int      `json:"execution_digest_variants"`
	AssertionRejections       int      `json:"assertion_rejections"`
	MissingTestRejections     int      `json:"missing_test_rejections"`
	NonClaims                 int      `json:"non_claims"`
	Unknowns                  int      `json:"unknowns"`
	Compiler                  Compiler `json:"compiler"`
	Effects                   Effects  `json:"effects"`
}

type Indicator struct {
	ID            string `json:"id"`
	Class         string `json:"class"`
	ProofChoice   string `json:"proof_choice"`
	MetaOperation string `json:"meta_operation"`
	Observed      int    `json:"observed"`
	Expected      int    `json:"expected"`
	Satisfied     bool   `json:"satisfied"`
}

type View struct {
	Audience     string   `json:"audience"`
	Resolution   string   `json:"resolution"`
	Satisfied    int      `json:"satisfied"`
	Total        int      `json:"total"`
	BasisPoints  int      `json:"basis_points"`
	IndicatorIDs []string `json:"indicator_ids"`
}

type Report struct {
	Schema            string      `json:"schema"`
	SubjectSHA        string      `json:"subject_sha"`
	Decision          string      `json:"decision"`
	Reason            string      `json:"reason"`
	Resolution        string      `json:"resolution"`
	Summary           Summary     `json:"summary"`
	Indicators        []Indicator `json:"indicators"`
	Views             []View      `json:"views"`
	RepositoryWrites  int         `json:"repository_writes"`
	MutationAuthority bool        `json:"mutation_authority"`
	Digest            string      `json:"digest"`
}
