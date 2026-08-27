package semanticdeltareceipt

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

// MetaContract is reconstructed from the checked-in Gooo meta source. The
// Go constants above are validator expectations, never the contract authority.
type MetaContract struct {
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
	CaseRecipes        []CaseRecipe
	DenominatorVersion string
	DenominatorCases   int
}

func ReadMetaContract() (MetaContract, error) {
	metaPath := MetaSourcePath
	if _, err := os.Stat(metaPath); err != nil {
		_, filename, _, ok := runtime.Caller(0)
		if !ok {
			return MetaContract{}, fmt.Errorf("locate meta source")
		}
		metaPath = filepath.Join(filepath.Dir(filename), "../../../..", MetaSourcePath)
	}
	raw, err := os.ReadFile(metaPath)
	if err != nil {
		return MetaContract{}, err
	}
	file, diagnostics := syntax.ParseFile(MetaSourcePath, string(raw))
	if diagnostics.Error() != nil || file == nil {
		return MetaContract{}, fmt.Errorf("meta syntax rejected: %v", diagnostics.Error())
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return MetaContract{}, fmt.Errorf("meta lowering rejected: %w", err)
	}
	contract := MetaContract{Version: DenominatorVersion, Digest: "sha256:" + ir.StableHash(), SourcePath: MetaSourcePath}
	for _, node := range ir.Graph.Nodes() {
		id := node.ID.String()
		switch {
		case strings.Contains(id, "/component/"):
			contract.ComponentKinds = append(contract.ComponentKinds, id[strings.LastIndex(id, "/")+1:])
		case strings.Contains(id, "/claim/"):
			contract.ClaimKinds = append(contract.ClaimKinds, id[strings.LastIndex(id, "/")+1:])
		}
		if value, ok := strings.CutPrefix(node.ValueProgram, "meta.semantic-delta."); ok {
			switch {
			case strings.HasPrefix(value, "layers:"):
				contract.Layers = splitCSV(strings.TrimPrefix(value, "layers:"))
			case strings.HasPrefix(value, "components:"):
				contract.ComponentKinds = splitCSV(strings.TrimPrefix(value, "components:"))
			case strings.HasPrefix(value, "policy:"):
				contract.Policies = splitSemi(strings.TrimPrefix(value, "policy:"))
			case strings.HasPrefix(value, "ledger:"):
				contract.Recipes = splitCSV(strings.TrimPrefix(value, "ledger:"))
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
				if recipe, ok := parseCaseRecipe(strings.TrimPrefix(value, "case:")); ok {
					contract.CaseRecipes = append(contract.CaseRecipes, recipe)
				}
			}
		}
	}
	sort.Strings(contract.ComponentKinds)
	sort.Strings(contract.ClaimKinds)
	if err := validateMetaContract(contract); err != nil {
		return MetaContract{}, err
	}
	return contract, nil
}

func validateMetaContract(contract MetaContract) error {
	if contract.Version != DenominatorVersion || contract.DenominatorCases != len(Denominator()) {
		return fmt.Errorf("meta denominator contract mismatch: version=%s cases=%d", contract.Version, contract.DenominatorCases)
	}
	if strings.Join(contract.Layers, ",") != "semantic,structural,textual" {
		return fmt.Errorf("meta delta layers are incomplete")
	}
	if len(contract.ComponentKinds) != TotalComponentCount || len(contract.Policies) != 3 || len(contract.Recipes) != 4 || contract.ClaimIdentity != "object=normalized|target|raw-digest|semantic-digest;preservation=before-id|after-id|target|before-proposition|after-raw|after-semantic;inventory=set-canonical" || contract.TransitionIdentity != "v2:claim-id|from|to|stage|step|reason|target-semantic-digest;sort-by-claim-id" || !sameCaseRecipes(contract.CaseRecipes, Denominator()) {
		return fmt.Errorf("meta semantic contract coverage is incomplete")
	}
	return nil
}

func parseCaseRecipe(value string) (CaseRecipe, bool) {
	parts := strings.Split(value, "|")
	if len(parts) != 3 {
		return CaseRecipe{}, false
	}
	id := parts[0]
	if !strings.HasPrefix(parts[1], "before:") || !strings.HasPrefix(parts[2], "after:") {
		return CaseRecipe{}, false
	}
	return CaseRecipe{ID: id, BeforePath: strings.TrimPrefix(parts[1], "before:"), AfterPath: strings.TrimPrefix(parts[2], "after:")}, id != ""
}

func sameCaseRecipes(actual []CaseRecipe, expected []CaseDefinition) bool {
	if len(actual) != len(expected) {
		return false
	}
	byID := make(map[string]CaseRecipe, len(actual))
	for _, recipe := range actual {
		if _, exists := byID[recipe.ID]; exists {
			return false
		}
		byID[recipe.ID] = recipe
	}
	for _, definition := range expected {
		recipe, ok := byID[definition.ID]
		if !ok || recipe.BeforePath != definition.BeforePath || recipe.AfterPath != definition.AfterPath {
			return false
		}
	}
	return true
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	sort.Strings(parts)
	return parts
}

func splitSemi(value string) []string {
	parts := strings.Split(value, ";")
	sort.Strings(parts)
	return parts
}
