package semanticdeltareceiptconsumer

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

type metaContract struct {
	Version            string
	Digest             string
	SourcePath         string
	Layers             []string
	ComponentKinds     []string
	ClaimKinds         []string
	Policies           []string
	Recipes            []string
	ClaimIdentity      string
	TransitionIdentity string
	CaseRecipes        []caseRecipe
	DenominatorVersion string
	DenominatorCases   int
}

const (
	metaSourcePath        = "examples/semantic-delta-receipt/main.gooo"
	denominatorVersion    = "v2"
	modeledComponentCount = 5
	totalComponentCount   = 5
)

type caseRecipe struct {
	ID         string
	BeforePath string
	AfterPath  string
}

func readMetaContract() (metaContract, error) {
	metaPath := metaSourcePath
	if _, err := os.Stat(metaPath); err != nil {
		_, filename, _, ok := runtime.Caller(0)
		if !ok {
			return metaContract{}, fmt.Errorf("locate meta source")
		}
		metaPath = filepath.Join(filepath.Dir(filename), "../../../..", metaSourcePath)
	}
	raw, err := os.ReadFile(metaPath)
	if err != nil {
		return metaContract{}, err
	}
	file, diagnostics := syntax.ParseFile(metaSourcePath, string(raw))
	if diagnostics.Error() != nil || file == nil {
		return metaContract{}, fmt.Errorf("consumer meta syntax rejected: %v", diagnostics.Error())
	}
	document, err := bidir.DocumentFromSyntax(file)
	if err != nil {
		return metaContract{}, fmt.Errorf("consumer meta adaptation: %w", err)
	}
	ir, err := bidir.LowerDocument(document)
	if err != nil {
		return metaContract{}, fmt.Errorf("consumer meta lowering rejected: %w", err)
	}
	contract := metaContract{Version: denominatorVersion, Digest: "sha256:" + ir.StableHash(), SourcePath: metaSourcePath}
	for _, node := range ir.Graph.Nodes() {
		id := node.ID.String()
		if strings.Contains(id, "/component/") {
			contract.ComponentKinds = append(contract.ComponentKinds, id[strings.LastIndex(id, "/")+1:])
		}
		if strings.Contains(id, "/claim/") {
			contract.ClaimKinds = append(contract.ClaimKinds, id[strings.LastIndex(id, "/")+1:])
		}
		if value, ok := strings.CutPrefix(node.ValueProgram, "meta.semantic-delta."); ok {
			switch {
			case strings.HasPrefix(value, "layers:"):
				contract.Layers = splitCSVConsumer(strings.TrimPrefix(value, "layers:"))
			case strings.HasPrefix(value, "components:"):
				contract.ComponentKinds = splitCSVConsumer(strings.TrimPrefix(value, "components:"))
			case strings.HasPrefix(value, "policy:"):
				contract.Policies = splitSemiConsumer(strings.TrimPrefix(value, "policy:"))
			case strings.HasPrefix(value, "ledger:"):
				contract.Recipes = splitCSVConsumer(strings.TrimPrefix(value, "ledger:"))
			case strings.HasPrefix(value, "claim-identity:"):
				contract.ClaimIdentity = strings.TrimPrefix(value, "claim-identity:")
			case strings.HasPrefix(value, "transition-identity:"):
				contract.TransitionIdentity = strings.TrimPrefix(value, "transition-identity:")
			case strings.HasPrefix(value, "denominator:"):
				parts := strings.Split(strings.TrimPrefix(value, "denominator:"), ":")
				if len(parts) == 2 {
					contract.Version = parts[0]
					contract.DenominatorCases, _ = strconv.Atoi(parts[1])
				}
			case strings.HasPrefix(value, "case:"):
				if recipe, ok := parseCaseRecipeConsumer(strings.TrimPrefix(value, "case:")); ok {
					contract.CaseRecipes = append(contract.CaseRecipes, recipe)
				}
			}
		}
	}
	sort.Strings(contract.ComponentKinds)
	sort.Strings(contract.ClaimKinds)
	if contract.Version != denominatorVersion || contract.DenominatorCases != 5 || len(contract.ComponentKinds) != totalComponentCount || len(contract.Policies) != 3 || len(contract.Recipes) != 4 || contract.ClaimIdentity != "v3:object=proposition-kind|canonical-semantic-fact-target-address|stable-relation-role;evidence=source-path|raw-digest|semantic-digest;preservation=before-proposition-id|canonical-pair-target-address|stable-relation-role;inventory=set-canonical" || contract.TransitionIdentity != "v2:claim-id|from|to|stage|step|reason|target-semantic-digest;sort-by-claim-id" || !sameCaseRecipesConsumer(contract.CaseRecipes) {
		return metaContract{}, fmt.Errorf("consumer meta contract incomplete")
	}
	return contract, nil
}

func parseCaseRecipeConsumer(value string) (caseRecipe, bool) {
	parts := strings.Split(value, "|")
	if len(parts) != 3 || !strings.HasPrefix(parts[1], "before:") || !strings.HasPrefix(parts[2], "after:") || parts[0] == "" {
		return caseRecipe{}, false
	}
	return caseRecipe{ID: parts[0], BeforePath: strings.TrimPrefix(parts[1], "before:"), AfterPath: strings.TrimPrefix(parts[2], "after:")}, true
}

func sameCaseRecipesConsumer(actual []caseRecipe) bool {
	expected := []caseRecipe{
		{ID: "equivalent", BeforePath: "examples/semantic-delta-receipt/before.gooo", AfterPath: "examples/semantic-delta-receipt/equivalent-after.gooo"},
		{ID: "semantic-change", BeforePath: "examples/semantic-delta-receipt/before.gooo", AfterPath: "examples/semantic-delta-receipt/semantic-after.gooo"},
		{ID: "value-program-change", BeforePath: "examples/semantic-delta-receipt/value-program-before.gooo", AfterPath: "examples/semantic-delta-receipt/value-program-after.gooo"},
		{ID: "indeterminate", BeforePath: "examples/semantic-delta-receipt/before.gooo", AfterPath: "examples/semantic-delta-receipt/indeterminate-after.gooo"},
		{ID: "ambiguous-match", BeforePath: "examples/semantic-delta-receipt/ambiguous-before.gooo", AfterPath: "examples/semantic-delta-receipt/ambiguous-after.gooo"},
	}
	if len(actual) != len(expected) {
		return false
	}
	byID := make(map[string]caseRecipe, len(actual))
	for _, recipe := range actual {
		if _, exists := byID[recipe.ID]; exists {
			return false
		}
		byID[recipe.ID] = recipe
	}
	for _, recipe := range expected {
		if byID[recipe.ID] != recipe {
			return false
		}
	}
	return true
}

func splitCSVConsumer(value string) []string {
	parts := strings.Split(value, ",")
	sort.Strings(parts)
	return parts
}

func splitSemiConsumer(value string) []string {
	parts := strings.Split(value, ";")
	sort.Strings(parts)
	return parts
}
