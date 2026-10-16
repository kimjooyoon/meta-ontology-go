package syntaxregistration

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesyntax"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesyntax/replay"
)

const ProfileMigrationSchema = "gooo/syntax-profile-migration-request/v1"
const profileContractPath = "internal/meta/syntaxregistration/profile_migration.gooo"
const profileSourceLimit = 128 << 10

type ProfileMigrationRequest struct {
	Schema            string                        `json:"schema"`
	Case              languagesyntax.CaseDefinition `json:"case"`
	TargetProfile     string                        `json:"target_entity_fields_profile"`
	BaseVersion       int                           `json:"base_version"`
	SnapshotDigest    string                        `json:"snapshot_digest"`
	SourceDigest      string                        `json:"source_digest"`
	ReplacementDigest string                        `json:"replacement_digest"`
	Toolchain         string                        `json:"toolchain"`
	ExecutionIdentity ExecutionIdentity             `json:"execution_identity"`
}

type ProfileMigrationMember struct {
	Path          string `json:"path"`
	ActivityID    string `json:"activity_id"`
	OutputID      string `json:"output_id"`
	BeforeDigest  string `json:"before_digest"`
	AfterDigest   string `json:"after_digest"`
	BeforeContent []byte `json:"before_content"`
	Content       []byte `json:"content"`
}

type ProfileMigrationCandidate struct {
	Schema           string                        `json:"schema"`
	Operation        string                        `json:"operation"`
	ActivityID       string                        `json:"activity_id"`
	ContractDigest   string                        `json:"contract_digest"`
	SemanticDigest   string                        `json:"semantic_digest"`
	InputDigest      string                        `json:"input_digest"`
	RequestDigest    string                        `json:"request_digest"`
	BeforeCase       languagesyntax.CaseDefinition `json:"before_case"`
	AfterCase        languagesyntax.CaseDefinition `json:"after_case"`
	SourceChecks     [2]replay.Result              `json:"source_checks"`
	CasePopulation   int                           `json:"case_population"`
	RequiredMembers  int                           `json:"required_members"`
	Members          [3]ProfileMigrationMember     `json:"members"`
	ExecutionBinding ExecutionBinding              `json:"execution_binding"`
	Admission        string                        `json:"semantic_admission"`
	State            string                        `json:"state"`
	RepositoryWrites int                           `json:"repository_writes"`
	ApplyAuthorized  bool                          `json:"apply_authorized"`
	PromotionAllowed bool                          `json:"promotion_authorized"`
}

type ProfileMigrationPlan struct {
	request      ProfileMigrationRequest
	inputs       map[string][]byte
	replacement  []byte
	registry     languagesyntax.Registry
	index        int
	binding      profileContract
	sourceChecks [2]replay.Result
}

func profileIssue(state, step, code string) error {
	class := ""
	if state == "UNKNOWN" {
		class = "UNRESOLVED_SOURCE"
	}
	return &Failure{state, "SYNTAX_PROFILE_MIGRATION", step, code, class,
		"inspect-profile-migration-source", []string{}}
}
