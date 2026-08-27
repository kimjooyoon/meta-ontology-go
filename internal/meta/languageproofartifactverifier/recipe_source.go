package languageproofartifactverifier

import (
	"fmt"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

// recipeFromSource is an independent consumer projection of the recipe
// meta-code. It deliberately shares only the core syntax boundary with the
// producer; the case-envelope reduction policy is reconstructed separately
// from the same raw source by policy.go.
func recipeFromSource(raw []byte) (Recipe, error) {
	file, diagnostics := syntax.ParseFile("proof-carrying-recipe-consumer.gooo", string(raw))
	if file == nil || diagnostics.HasErrors() || file.Package == nil || file.Namespace == nil {
		return Recipe{}, fmt.Errorf("recipe source is not parseable")
	}
	result := Recipe{Schema: RecipeSchema, Consumer: ConsumerID, Roles: []RecipeRole{}, Steps: []RecipeStep{}, Dependencies: []RecipeDependency{}}
	declarations := file.Declarations
	if declarations == nil {
		declarations = file.Decls
	}
	for _, declaration := range declarations {
		activity, ok := declaration.(*syntax.ActivityDecl)
		if !ok || !activity.ValueProgramPresent || !strings.HasPrefix(activity.ValueProgram, "proof.recipe.") {
			continue
		}
		fields, err := parseRecipeFields(activity.ValueProgram)
		if err != nil {
			return Recipe{}, err
		}
		kind := strings.TrimPrefix(strings.SplitN(activity.ValueProgram, ";", 2)[0], "proof.recipe.")
		switch kind {
		case "role":
			result.Roles = append(result.Roles, RecipeRole{ID: fields["claim"], Proposition: fields["proposition"], Target: fields["target"], ProofChoice: fields["proof"], Step: fields["step"], MetaOperation: fields["operation"], Dependencies: splitRecipeCSV(fields["requires"])})
		case "meta":
			result.Version = parseRecipeInt(fields["version"])
			result.ID, result.Consumer, result.SourceEntry = fields["id"], fields["consumer"], fields["entry"]
			result.Authority = RecipeAuthority{Capability: fields["authority"], Requires: splitRecipeCSV(fields["requires"]), Mutation: fields["mutation"] == "true", Promotion: fields["promotion"] == "true", Semantic: fields["semantic"] == "true"}
		case "dependency":
			result.Dependencies = append(result.Dependencies, RecipeDependency{From: fields["from"], To: fields["to"], Relation: fields["relation"]})
		default:
			return Recipe{}, fmt.Errorf("unknown recipe meta-code %q", activity.ValueProgram)
		}
	}
	for _, role := range result.Roles {
		result.Steps = append(result.Steps, RecipeStep{ID: role.Step, Input: role.Target, MetaOperation: role.MetaOperation, ProofChoice: role.ProofChoice, Role: role.ID})
	}
	if result.Version == 0 || result.ID == "" || result.Consumer != ConsumerID || result.SourceEntry == "" || len(result.Roles) != ClaimTemplateTotal || len(result.Steps) != ClaimTemplateTotal || len(result.Dependencies) != 4 {
		return Recipe{}, fmt.Errorf("recipe meta-code is incomplete")
	}
	return result, nil
}

func parseRecipeFields(program string) (map[string]string, error) {
	parts := strings.SplitN(program, ";", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("recipe meta-code has no fields")
	}
	fields := map[string]string{}
	for _, field := range strings.Split(parts[1], ";") {
		key, value, ok := strings.Cut(field, "=")
		if !ok || key == "" || value == "" {
			return nil, fmt.Errorf("recipe meta-code field %q is invalid", field)
		}
		fields[key] = value
	}
	return fields, nil
}

func splitRecipeCSV(value string) []string {
	if value == "" {
		return []string{}
	}
	return strings.Split(value, ",")
}

func parseRecipeInt(value string) int {
	result := 0
	for _, char := range value {
		if char < '0' || char > '9' {
			return 0
		}
		result = result*10 + int(char-'0')
	}
	return result
}
