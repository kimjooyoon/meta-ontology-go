package languageproofartifactverifier

const (
	ReportSchema    = "gooo/language-proof-carrying-artifact-verifier/v1"
	ContractSchema  = "gooo/language-proof-carrying-artifact-contract/v1"
	RecipeSchema    = "gooo/language-proof-carrying-recipe/v1"
	ArtifactSchema  = "gooo/language-proof-carrying-artifact/v1"
	ProducerID      = "gooo://producer/language-proof-carrying-artifact"
	ConsumerID      = "gooo://consumer/language-proof-carrying-artifact-verifier"
	CaseTotal       = 5
	EvidenceTotal   = 3
	TransitionTotal = 4
)

type Coordinate struct {
	Stage  string `json:"stage"`
	Step   string `json:"step"`
	Reason string `json:"reason"`
}

type IndependenceEvidence struct {
	Schema               string `json:"schema"`
	ProducerDependencies int    `json:"producer_dependencies"`
}

type Input struct {
	Contract         Contract
	HeadSHA          string
	ValidArtifact    []byte
	TamperedArtifact []byte
	MissingArtifact  []byte
	ByteOnlyArtifact []byte
	WrongRecipe      []byte
	Source           []byte
	Operation        []byte
	Recipe           []byte
	Independence     IndependenceEvidence
}

type RecipeStep struct {
	ID            string `json:"id"`
	Input         string `json:"input"`
	MetaOperation string `json:"meta_operation"`
	ProofChoice   string `json:"proof_choice"`
}

type Recipe struct {
	Schema   string       `json:"schema"`
	Version  int          `json:"version"`
	ID       string       `json:"id"`
	Consumer string       `json:"consumer"`
	Steps    []RecipeStep `json:"steps"`
}

type Evidence struct {
	Kind                            string     `json:"kind"`
	ClaimID                         string     `json:"claim_id"`
	ProofChoice                     string     `json:"proof_choice"`
	MetaOperation                   string     `json:"meta_operation"`
	Coordinate                      Coordinate `json:"coordinate"`
	SourceDigest                    string     `json:"source_digest"`
	ReceiptDigest                   string     `json:"receipt_digest,omitempty"`
	Activity                        string     `json:"activity,omitempty"`
	Predicate                       string     `json:"predicate,omitempty"`
	RepositoryWrites                int        `json:"repository_writes"`
	MutationAuthority               bool       `json:"mutation_authority"`
	AuthorityGranted                bool       `json:"authority_granted"`
	IndependentVerificationRequired bool       `json:"independent_verification_required"`
	EvidenceDigest                  string     `json:"evidence_digest"`
}

type Authority struct {
	Granted bool   `json:"granted"`
	Basis   string `json:"basis"`
}

type Effects struct {
	RepositoryWrites  int  `json:"repository_writes"`
	MutationAuthority bool `json:"mutation_authority"`
}

type Artifact struct {
	Schema        string     `json:"schema"`
	HeadSHA       string     `json:"head_sha"`
	Producer      string     `json:"producer"`
	Consumer      string     `json:"consumer"`
	MetaOperation string     `json:"meta_operation"`
	Decision      string     `json:"decision"`
	Resolution    string     `json:"resolution"`
	Reason        string     `json:"reason"`
	SourcePath    string     `json:"source_path"`
	Evidence      []Evidence `json:"evidence"`
	Recipe        Recipe     `json:"recipe"`
	Authority     Authority  `json:"authority"`
	Effects       Effects    `json:"effects"`
	NotClaimed    []string   `json:"not_claimed"`
	Digest        string     `json:"digest"`
}

type CaseSpec struct {
	ID                 string `json:"id"`
	InputKind          string `json:"input_kind"`
	ExpectedDecision   string `json:"expected_decision"`
	ExpectedResolution string `json:"expected_resolution"`
	ExpectedReason     string `json:"expected_reason"`
	ProofChoice        string `json:"proof_choice"`
	MetaOperation      string `json:"meta_operation"`
}

type Contract struct {
	Schema  string     `json:"schema"`
	Version int        `json:"version"`
	Cases   []CaseSpec `json:"cases"`
}

type ClaimTransition struct {
	ClaimID        string     `json:"claim_id"`
	From           string     `json:"from"`
	To             string     `json:"to"`
	Producer       string     `json:"producer"`
	Consumer       string     `json:"consumer"`
	ProofChoice    string     `json:"proof_choice"`
	MetaOperation  string     `json:"meta_operation"`
	Coordinate     Coordinate `json:"coordinate"`
	Reason         string     `json:"reason"`
	EvidenceDigest []string   `json:"evidence_digests"`
}
