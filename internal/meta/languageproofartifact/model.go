package languageproofartifact

const (
	ArtifactSchema     = "gooo/language-proof-carrying-artifact/v1"
	RecipeSchema       = "gooo/language-proof-carrying-recipe/v1"
	ProducerID         = "gooo://producer/language-proof-carrying-artifact"
	ConsumerID         = "gooo://consumer/language-proof-carrying-artifact-verifier"
	ArtifactDecision   = "CARRIED"
	ArtifactResolution = "EVIDENCE_ATTACHED"
)

type Coordinate struct {
	Stage  string `json:"stage"`
	Step   string `json:"step"`
	Reason string `json:"reason"`
}

type Effects struct {
	RepositoryWrites  int  `json:"repository_writes"`
	MutationAuthority bool `json:"mutation_authority"`
}

type Authority struct {
	Granted bool   `json:"granted"`
	Basis   string `json:"basis"`
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

type Input struct {
	HeadSHA    string
	SourcePath string
	Source     []byte
	Operation  []byte
	Recipe     Recipe
}
