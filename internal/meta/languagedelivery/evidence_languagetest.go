package languagedelivery

const languageTestReportSchema = "gooo/language-test-experiment-report/v1"

type LanguageTestReceipt struct {
	Schema            string              `json:"schema"`
	SubjectSHA        string              `json:"subject_sha"`
	Decision          string              `json:"decision"`
	Resolution        string              `json:"resolution"`
	Summary           LanguageTestSummary `json:"summary"`
	Views             []LanguageTestView  `json:"views"`
	RepositoryWrites  int                 `json:"repository_writes"`
	MutationAuthority bool                `json:"mutation_authority"`
}

type LanguageTestSummary struct {
	Coordinates             LanguageTestCoordinates `json:"coordinates"`
	DeclaredTests           int                     `json:"declared_tests"`
	ExecutedTests           int                     `json:"executed_tests"`
	PassedTests             int                     `json:"passed_tests"`
	ReceiptDigestVariants   int                     `json:"receipt_digest_variants"`
	ExecutionDigestVariants int                     `json:"execution_digest_variants"`
	AssertionRejections     int                     `json:"assertion_rejections"`
	MissingTestRejections   int                     `json:"missing_test_rejections"`
	Unknowns                int                     `json:"unknowns"`
	NonClaims               int                     `json:"non_claims"`
	Compiler                LanguageTestCompiler    `json:"compiler"`
	Effects                 LanguageTestEffects     `json:"effects"`
}

type LanguageTestCoordinates struct {
	Satisfied   int `json:"satisfied"`
	Total       int `json:"total"`
	BasisPoints int `json:"basis_points"`
}

type LanguageTestCompiler struct {
	ExecutableDigest string `json:"executable_digest"`
	Go127Runtimes    int    `json:"go127_runtimes"`
}

type LanguageTestEffects struct {
	RepositoryWrites  int  `json:"repository_writes"`
	MutationAuthority bool `json:"mutation_authority"`
}

type LanguageTestView struct {
	Audience    string `json:"audience"`
	Resolution  string `json:"resolution"`
	Satisfied   int    `json:"satisfied"`
	Total       int    `json:"total"`
	BasisPoints int    `json:"basis_points"`
}
