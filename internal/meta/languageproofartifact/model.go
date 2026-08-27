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
	ArtifactUseAuthority string `json:"artifact_use_authority"`
	MutationAuthority    bool   `json:"mutation_authority"`
	PromotionAuthority   bool   `json:"promotion_authority"`
	SemanticAuthority    bool   `json:"semantic_authority"`
	Basis                string `json:"basis"`
}

type Evidence struct {
	Kind                            string     `json:"kind"`
	ClaimID                         string     `json:"claim_id"`
	ProofChoice                     string     `json:"proof_choice"`
	MetaOperation                   string     `json:"meta_operation"`
	Coordinate                      Coordinate `json:"coordinate"`
	SourceDigest                    string     `json:"source_digest"`
	SemanticDigest                  string     `json:"semantic_digest"`
	ReceiptDigest                   string     `json:"receipt_digest,omitempty"`
	Activity                        string     `json:"activity,omitempty"`
	Predicate                       string     `json:"predicate,omitempty"`
	RepositoryWrites                int        `json:"repository_writes"`
	MutationAuthority               bool       `json:"mutation_authority"`
	ArtifactUseAuthority            string     `json:"artifact_use_authority"`
	IndependentVerificationRequired bool       `json:"independent_verification_required"`
	EvidenceDigest                  string     `json:"evidence_digest"`
}

type ClaimStatement struct {
	ID             string     `json:"id"`
	Proposition    string     `json:"proposition"`
	TargetDigest   string     `json:"target_digest"`
	Dependencies   []string   `json:"dependencies"`
	ProofChoice    string     `json:"proof_choice"`
	MetaOperation  string     `json:"meta_operation"`
	Coordinate     Coordinate `json:"coordinate"`
	EvidenceDigest []string   `json:"evidence_digests"`
	State          string     `json:"state"`
	Digest         string     `json:"digest"`
}

type WriteSetEntry struct {
	Kind   string `json:"kind"`
	Path   string `json:"path"`
	Digest string `json:"digest"`
}

type WriteSetChange struct {
	Path         string `json:"path"`
	BeforeDigest string `json:"before_digest,omitempty"`
	AfterDigest  string `json:"after_digest,omitempty"`
	BeforeKind   string `json:"before_kind,omitempty"`
	AfterKind    string `json:"after_kind,omitempty"`
}

type WriteSetObservation struct {
	Schema            string           `json:"schema"`
	Version           int              `json:"version"`
	Before            []WriteSetEntry  `json:"before"`
	After             []WriteSetEntry  `json:"after"`
	Changed           []WriteSetChange `json:"changed"`
	BeforeDigest      string           `json:"before_digest"`
	AfterDigest       string           `json:"after_digest"`
	RepositoryWrites  int              `json:"repository_writes"`
	MutationAuthority bool             `json:"mutation_authority"`
	ObservedScope     string           `json:"observed_scope"`
	NetUnchanged      bool             `json:"net_repository_state_unchanged"`
	TransientUnknown  bool             `json:"transient_writes_unknown"`
	AuthorityBasis    string           `json:"authority_observation"`
	Digest            string           `json:"digest"`
}

type LedgerEntry struct {
	ClaimID        string     `json:"claim_id"`
	Proposition    string     `json:"proposition"`
	TargetDigest   string     `json:"target_digest"`
	Dependencies   []string   `json:"dependencies"`
	Status         string     `json:"status"`
	Resolution     string     `json:"resolution"`
	Producer       string     `json:"producer"`
	Consumer       string     `json:"consumer"`
	ProofChoice    string     `json:"proof_choice"`
	MetaOperation  string     `json:"meta_operation"`
	Coordinate     Coordinate `json:"coordinate"`
	Reason         string     `json:"reason"`
	EvidenceDigest []string   `json:"evidence_digests"`
	Provenance     string     `json:"provenance"`
	PreviousDigest string     `json:"previous_digest"`
	Digest         string     `json:"digest"`
}

type ClaimLedger struct {
	Schema  string        `json:"schema"`
	Version int           `json:"version"`
	Entries []LedgerEntry `json:"entries"`
	Digest  string        `json:"digest"`
}

type Artifact struct {
	Schema         string              `json:"schema"`
	HeadSHA        string              `json:"head_sha"`
	Producer       string              `json:"producer"`
	Consumer       string              `json:"consumer"`
	MetaOperation  string              `json:"meta_operation"`
	Decision       string              `json:"decision"`
	Resolution     string              `json:"resolution"`
	Reason         string              `json:"reason"`
	SourcePath     string              `json:"source_path"`
	SourceDigest   string              `json:"source_digest"`
	SemanticDigest string              `json:"semantic_digest"`
	Evidence       []Evidence          `json:"evidence"`
	Recipe         Recipe              `json:"recipe"`
	RecipeDigest   string              `json:"recipe_digest"`
	PriorLedger    ClaimLedger         `json:"prior_ledger"`
	WriteSet       WriteSetObservation `json:"write_set"`
	Authority      Authority           `json:"authority"`
	Effects        Effects             `json:"effects"`
	NotClaimed     []string            `json:"not_claimed"`
	Claims         []ClaimStatement    `json:"claims"`
	BundleDigest   string              `json:"bundle_digest"`
	Digest         string              `json:"digest"`
}

type Input struct {
	HeadSHA    string
	SourcePath string
	Source     []byte
	Operation  []byte
	Recipe     Recipe
	WriteSet   WriteSetObservation
}
