package main

import (
	"errors"
	plannersci "github.com/kimjooyoon/meta-ontology-go/internal/detection/selectiveci"
	"sort"
)

func selectedShadowCommands(plan plannersci.PlanResult, registry plannersci.Registry) ([]shadowCommandSpec, []shadowCommandSpec, []shadowResourceReceipt, error) {
	commands := make(map[string]plannersci.Command, len(registry.Commands))
	for _, command := range registry.Commands {
		commands[command.ID] = command
	}
	guards := make(map[string]plannersci.Command, len(registry.GlobalGuardCommands))
	for _, command := range registry.GlobalGuardCommands {
		guards[command.ID] = command
	}
	makeSpecs := func(ids []string, source map[string]plannersci.Command) ([]shadowCommandSpec, []shadowResourceReceipt, error) {
		specs := make([]shadowCommandSpec, 0, len(ids))
		receipts := make([]shadowResourceReceipt, 0, len(ids))
		for _, id := range ids {
			command, ok := source[id]
			if !ok {
				return nil, nil, errors.New("selected command is not registered")
			}
			specs = append(specs, shadowCommandSpec{ID: command.ID, Argv: append([]string{}, command.Argv...)})
			receipts = append(receipts, shadowResourceReceipt{CommandID: command.ID, CPUWorkUnits: command.CPUWorkUnits, MemoryBytes: command.MemoryBytes})
		}
		sort.Slice(specs, func(i, j int) bool { return specs[i].ID < specs[j].ID })
		sort.Slice(receipts, func(i, j int) bool { return receipts[i].CommandID < receipts[j].CommandID })
		return specs, receipts, nil
	}
	commandSpecs, commandReceipts, err := makeSpecs(plan.SelectedCommandIDs, commands)
	if err != nil {
		return nil, nil, nil, err
	}
	guardSpecs, guardReceipts, err := makeSpecs(plan.SelectedGuardCommandIDs, guards)
	if err != nil {
		return nil, nil, nil, err
	}
	receipts := append(commandReceipts, guardReceipts...)
	sort.Slice(receipts, func(i, j int) bool { return receipts[i].CommandID < receipts[j].CommandID })
	return commandSpecs, guardSpecs, receipts, nil
}
