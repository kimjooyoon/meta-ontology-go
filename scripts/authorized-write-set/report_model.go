package main

type coordinates struct {
	SourceReceipts      int `json:"source_receipts"`
	SourceReceiptsTotal int `json:"source_receipts_total"`
	DensityPaths        int `json:"density_paths"`
	ExtractionPaths     int `json:"extraction_paths"`
	OverlapPaths        int `json:"overlap_paths"`
	ExpectedPaths       int `json:"expected_paths"`
	ObservedPaths       int `json:"observed_paths"`
	CreatedPaths        int `json:"created_paths"`
	UntrackedPaths      int `json:"untracked_paths"`
	UnclassifiedPaths   int `json:"unclassified_untracked_paths"`
	Unknowns            int `json:"unknowns"`
}
type indicator struct {
	ID            string `json:"id"`
	Class         string `json:"class"`
	ProofChoice   string `json:"proof_choice"`
	MetaOperation string `json:"meta_operation"`
	Observed      int    `json:"observed"`
	Expected      int    `json:"expected"`
	Satisfied     bool   `json:"satisfied"`
}
type proof struct {
	Choice        string `json:"choice"`
	MetaOperation string `json:"meta_operation"`
	Passed        bool   `json:"passed"`
}
type effects struct {
	RepositoryWrites  int  `json:"repository_writes"`
	MutationAuthority bool `json:"mutation_authority"`
}
type evidence struct {
	Schema          string      `json:"schema"`
	Decision        string      `json:"decision"`
	Resolution      string      `json:"resolution"`
	Reason          string      `json:"reason"`
	Audience        string      `json:"audience"`
	SourceSHA       string      `json:"source_sha"`
	MetaOperation   string      `json:"meta_operation"`
	Expected        []string    `json:"expected"`
	Observed        []string    `json:"observed"`
	ExpectedCreated []string    `json:"expected_created"`
	ObservedCreated []string    `json:"observed_created"`
	Coordinates     coordinates `json:"coordinates"`
	Indicators      []indicator `json:"indicators"`
	Proofs          []proof     `json:"proofs"`
	Effects         effects     `json:"effects"`
	Exact           bool        `json:"exact"`
	Digest          string      `json:"digest"`
}
