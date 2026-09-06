// Package syntaxregistration builds source-bound, caller-reviewed registration
// candidates. It never edits its input filesystem or grants application authority.
package syntaxregistration

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesyntax"
)

const Operation = "RegisterSyntaxCapability"
const RequiredMembers = 9
const corpusPath = "examples/language-syntax-roundtrip/corpus.json"
const syntaxRoot = "internal/meta/languagereadiness/languagesyntax/"
const closureRoot = "internal/meta/languageassurance/verticalsliceclosureshadow/"

type Request struct {
	Case           languagesyntax.CaseDefinition `json:"case"`
	BaseVersion    int                           `json:"base_version"`
	SnapshotDigest string                        `json:"snapshot_digest"`
	SourceDigest   string                        `json:"source_digest"`
	Toolchain      string                        `json:"toolchain"`
}

type Failure struct {
	State         string   `json:"state"`
	Stage         string   `json:"stage"`
	Step          string   `json:"step"`
	Reason        string   `json:"reason"`
	UnknownClass  string   `json:"unknown_class"`
	NextOperation string   `json:"next_operation"`
	BlockedBy     []string `json:"blocked_by"`
}

func (failure *Failure) Error() string { return failure.State + "/" + failure.Reason }

func failure(state, step, reason, class, next string) error {
	return &Failure{state, "SYNTAX_REGISTRATION", step, reason, class, next, []string{}}
}

type Member struct {
	Path         string `json:"path"`
	ActivityID   string `json:"activity_id"`
	BeforeDigest string `json:"before_digest"`
	AfterDigest  string `json:"after_digest"`
	Content      []byte `json:"content"`
}

type Candidate struct {
	Operation        string   `json:"operation"`
	ActivityID       string   `json:"activity_id"`
	ContractDigest   string   `json:"contract_digest"`
	SemanticDigest   string   `json:"semantic_digest"`
	InputDigest      string   `json:"input_digest"`
	RequestDigest    string   `json:"request_digest"`
	Toolchain        string   `json:"toolchain"`
	State            string   `json:"state"`
	Admission        string   `json:"semantic_admission"`
	Members          []Member `json:"members"`
	Required         int      `json:"required_members"`
	Emitted          int      `json:"emitted_members"`
	RepositoryWrites int      `json:"repository_writes"`
	ApplyAuthorized  bool     `json:"apply_authorized"`
	PromotionAllowed bool     `json:"promotion_authorized"`
}

type Plan struct {
	request Request
	inputs  map[string][]byte
	digest  string
	binding contractBinding
}

func digest(raw []byte) string {
	sum := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func digestValue(value any) string {
	raw, _ := json.Marshal(value)
	return digest(raw)
}

func denominatorPath(version int) string {
	return fmt.Sprintf("%sevidence/denominator-v%d.json", closureRoot, version)
}

func memberPaths(version int) []string {
	return []string{corpusPath, syntaxRoot + "registry.go", syntaxRoot + "model.go",
		syntaxRoot + "conformance/evaluate_test.go", closureRoot + "denominator.go",
		closureRoot + "evidence.go", closureRoot + "contract.go", denominatorPath(version),
		closureRoot + "denominator_migration_test.go"}
}

func validPath(path string) bool {
	return fs.ValidPath(path) && path != "." && len(path) > 0
}
