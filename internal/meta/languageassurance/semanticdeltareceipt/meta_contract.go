package semanticdeltareceipt

import (
	"fmt"
	"os"
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
	DenominatorVersion string
	DenominatorCases   int
}

func ReadMetaContract() (MetaContract, error) {
	raw, err := os.ReadFile(MetaSourcePath)
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
			case strings.HasPrefix(value, "denominator:"):
				parts := strings.Split(strings.TrimPrefix(value, "denominator:"), ":")
				if len(parts) == 2 {
					contract.Version = parts[0]
					contract.DenominatorCases, _ = strconv.Atoi(parts[1])
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
	if len(contract.ComponentKinds) != TotalComponentCount || len(contract.Policies) != 3 || len(contract.Recipes) != 4 {
		return fmt.Errorf("meta semantic contract coverage is incomplete")
	}
	return nil
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
