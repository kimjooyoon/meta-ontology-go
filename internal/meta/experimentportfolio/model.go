package experimentportfolio

type Input struct {
	SubjectSHA string    `json:"subject_sha"`
	Contract   Contract  `json:"contract"`
	Receipts   []Receipt `json:"receipts"`
}

type Coordinate struct {
	ID            string `json:"id"`
	Producer      string `json:"producer"`
	Consumer      string `json:"consumer"`
	MetaOperation string `json:"meta_operation"`
	ProofChoice   string `json:"proof_choice"`
	Numerator     int    `json:"numerator"`
	Denominator   int    `json:"denominator"`
	Status        string `json:"status"`
	Stage         string `json:"stage"`
	Step          string `json:"step"`
	Reason        string `json:"reason"`
}

type Counterexample struct {
	ID       string `json:"id"`
	Location string `json:"location"`
	Claim    string `json:"claim"`
	Stage    string `json:"stage"`
	Step     string `json:"step"`
	Reason   string `json:"reason"`
}

type UnknownLocation struct {
	ID     string `json:"id"`
	Path   string `json:"path"`
	Stage  string `json:"stage"`
	Step   string `json:"step"`
	Reason string `json:"reason"`
}

type ExtensionEvidence struct {
	ID       string `json:"id"`
	Claim    string `json:"claim"`
	Status   string `json:"status"`
	Evidence string `json:"evidence"`
	Stage    string `json:"stage"`
	Step     string `json:"step"`
	Reason   string `json:"reason"`
}

type Receipt struct {
	Schema            string              `json:"schema"`
	SubjectSHA        string              `json:"subject_sha"`
	CandidateID       string              `json:"candidate_id"`
	SourcePath        string              `json:"source_path"`
	SourceDigest      string              `json:"source_digest"`
	Producer          string              `json:"producer"`
	Consumer          string              `json:"consumer"`
	MetaOperation     string              `json:"meta_operation"`
	ProofChoice       string              `json:"proof_choice"`
	CoordinateVector  []Coordinate        `json:"coordinate_vector"`
	Counterexamples   []Counterexample    `json:"counterexamples"`
	UnknownLocations  []UnknownLocation   `json:"unknown_locations"`
	ExtensionEvidence []ExtensionEvidence `json:"extension_evidence"`
	RepositoryWrites  int                 `json:"repository_writes"`
	MutationAuthority bool                `json:"mutation_authority"`
	FactsDigest       string              `json:"facts_digest"`
	Digest            string              `json:"digest"`
}

type CandidateComparison struct {
	CandidateID          string              `json:"candidate_id"`
	SourcePath           string              `json:"source_path"`
	MetaOperation        string              `json:"meta_operation"`
	Producer             string              `json:"producer"`
	Consumer             string              `json:"consumer"`
	ProofChoice          string              `json:"proof_choice"`
	Receipt              Receipt             `json:"receipt"`
	CoordinateVector     []Coordinate        `json:"coordinate_vector"`
	CounterexampleCount  int                 `json:"counterexample_count"`
	Counterexamples      []Counterexample    `json:"counterexamples"`
	UnknownLocationCount int                 `json:"unknown_location_count"`
	UnknownLocations     []UnknownLocation   `json:"unknown_locations"`
	ExtensionEvidence    []ExtensionEvidence `json:"extension_evidence"`
}

type Summary struct {
	Candidates                int                 `json:"candidates"`
	CoordinatesPerCandidate   int                 `json:"coordinates_per_candidate"`
	CounterexampleCounts      map[string]int      `json:"counterexample_counts"`
	UnknownLocationIDs        map[string][]string `json:"unknown_location_ids"`
	ExtensionEvidenceStatuses map[string][]string `json:"extension_evidence_statuses"`
	RepositoryWrites          int                 `json:"repository_writes"`
	MutationAuthority         bool                `json:"mutation_authority"`
	Unknowns                  int                 `json:"unknowns"`
}

type Proof struct {
	Choice         string `json:"choice"`
	MetaOperation  string `json:"meta_operation"`
	Stage          string `json:"stage"`
	Step           string `json:"step"`
	Reason         string `json:"reason"`
	EvidenceDigest string `json:"evidence_digest"`
	Passed         bool   `json:"passed"`
}

type Report struct {
	Schema            string                `json:"schema"`
	Decision          string                `json:"decision"`
	Resolution        string                `json:"resolution"`
	Reason            string                `json:"reason"`
	Interpretation    string                `json:"interpretation"`
	SubjectSHA        string                `json:"subject_sha"`
	ContractID        string                `json:"contract_id"`
	Candidates        []CandidateComparison `json:"candidates"`
	Summary           Summary               `json:"summary"`
	Proofs            []Proof               `json:"proofs"`
	NotClaimed        []string              `json:"not_claimed"`
	RepositoryWrites  int                   `json:"repository_writes"`
	MutationAuthority bool                  `json:"mutation_authority"`
	FactsDigest       string                `json:"facts_digest"`
	Digest            string                `json:"digest"`
}
