package languageproofartifact

import (
	"fmt"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

type RecipeRole struct {
	ID            string   `json:"id"`
	Proposition   string   `json:"proposition"`
	Target        string   `json:"target"`
	ProofChoice   string   `json:"proof_choice"`
	Step          string   `json:"step"`
	MetaOperation string   `json:"meta_operation"`
	Dependencies  []string `json:"dependencies"`
}

type RecipeDependency struct {
	From     string `json:"from"`
	To       string `json:"to"`
	Relation string `json:"relation"`
}

type RecipeAuthority struct {
	Capability string   `json:"capability"`
	Requires   []string `json:"requires"`
	Mutation   bool     `json:"mutation"`
	Promotion  bool     `json:"promotion"`
	Semantic   bool     `json:"semantic"`
}

type RecipeStep struct {
	ID            string `json:"id"`
	Input         string `json:"input"`
	MetaOperation string `json:"meta_operation"`
	ProofChoice   string `json:"proof_choice"`
	Role          string `json:"role"`
}

type Recipe struct {
	Schema       string             `json:"schema"`
	Version      int                `json:"version"`
	ID           string             `json:"id"`
	Consumer     string             `json:"consumer"`
	SourceEntry  string             `json:"source_entry"`
	Roles        []RecipeRole       `json:"roles"`
	Steps        []RecipeStep       `json:"steps"`
	Dependencies []RecipeDependency `json:"dependencies"`
	Authority    RecipeAuthority    `json:"authority"`
}

const recipePrefix = "proof.recipe."

// RecipeFromSource is the producer's only recipe authority. The JSON recipe
// is a projection that must equal this result; it cannot introduce a rule.
func RecipeFromSource(raw []byte) (Recipe, error) {
	file, diagnostics := syntax.ParseFile("language-proof-carrying-artifact.gooo", string(raw))
	if file == nil || diagnostics.HasErrors() || file.Package == nil || file.Namespace == nil {
		return Recipe{}, fmt.Errorf("proof recipe source is not parseable")
	}
	result := Recipe{Schema: RecipeSchema, Consumer: ConsumerID, Roles: []RecipeRole{}, Steps: []RecipeStep{}, Dependencies: []RecipeDependency{}}
	declarations := file.Declarations
	if declarations == nil {
		declarations = file.Decls
	}
	for _, declaration := range declarations {
		activity, ok := declaration.(*syntax.ActivityDecl)
		if !ok || !activity.ValueProgramPresent || !strings.HasPrefix(activity.ValueProgram, recipePrefix) {
			continue
		}
		fields, err := recipeFields(activity.ValueProgram)
		if err != nil {
			return Recipe{}, fmt.Errorf("recipe activity %q: %w", activity.Name, err)
		}
		switch strings.TrimPrefix(activity.ValueProgram, recipePrefix) {
		case "role":
			role := RecipeRole{ID: fields["claim"], Proposition: fields["proposition"], Target: fields["target"], ProofChoice: fields["proof"], Step: fields["step"], MetaOperation: fields["operation"], Dependencies: splitCSV(fields["requires"])}
			if role.ID == "" || role.Proposition == "" || role.Target == "" || role.ProofChoice == "" || role.Step == "" || role.MetaOperation == "" {
				return Recipe{}, fmt.Errorf("recipe role %q is incomplete", activity.Name)
			}
			result.Roles = append(result.Roles, role)
		case "meta":
			result.Version = atoi(fields["version"])
			result.ID = fields["id"]
			result.Consumer = fields["consumer"]
			result.SourceEntry = fields["entry"]
			result.Authority = RecipeAuthority{Capability: fields["authority"], Requires: splitCSV(fields["requires"]), Mutation: fields["mutation"] == "true", Promotion: fields["promotion"] == "true", Semantic: fields["semantic"] == "true"}
		case "dependency":
			result.Dependencies = append(result.Dependencies, RecipeDependency{From: fields["from"], To: fields["to"], Relation: fields["relation"]})
		case "entry":
			result.SourceEntry = fields["entry"]
		default:
			return Recipe{}, fmt.Errorf("unknown recipe declaration %q", activity.ValueProgram)
		}
	}
	for _, role := range result.Roles {
		result.Steps = append(result.Steps, RecipeStep{ID: role.Step, Input: role.Target, MetaOperation: role.MetaOperation, ProofChoice: role.ProofChoice, Role: role.ID})
	}
	if result.Version == 0 || result.ID == "" || result.Consumer != ConsumerID || result.SourceEntry == "" || len(result.Roles) != 5 || len(result.Dependencies) != 4 || result.Authority.Capability != "READ_ONLY_CONSUMPTION" {
		return Recipe{}, fmt.Errorf("proof recipe source is incomplete")
	}
	return result, nil
}

func recipeFields(program string) (map[string]string, error) {
	parts := strings.SplitN(program, ";", 2)
	if len(parts) != 2 || !strings.HasPrefix(parts[0], recipePrefix) {
		return nil, fmt.Errorf("invalid recipe program")
	}
	fields := map[string]string{}
	for _, part := range strings.Split(parts[1], ";") {
		key, value, ok := strings.Cut(part, "=")
		if !ok || key == "" || value == "" {
			return nil, fmt.Errorf("invalid recipe field %q", part)
		}
		fields[key] = value
	}
	return fields, nil
}

func splitCSV(value string) []string {
	if value == "" {
		return []string{}
	}
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}

func atoi(value string) int {
	var result int
	for _, char := range value {
		if char < '0' || char > '9' {
			return 0
		}
		result = result*10 + int(char-'0')
	}
	return result
}
