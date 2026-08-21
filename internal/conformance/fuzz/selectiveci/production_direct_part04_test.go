package selectiveci

import (
	productionsci "github.com/kimjooyoon/meta-ontology-go/internal/detection/selectiveci"
)

func mapArgv(result productionsci.PlanResult, fixture directFixture, commands map[string]productionsci.Command) map[string][]string {
	argv := make(map[string][]string)
	for _, id := range append(result.SelectedCommandIDs, result.SelectedGuardCommandIDs...) {
		key := id
		if mapped, ok := fixture.ProdToOracle[id]; ok {
			key = mapped
		}
		argv[key] = append([]string(nil), commands[id].Argv...)
	}
	return argv
}
