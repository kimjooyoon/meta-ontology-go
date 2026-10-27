package languagesyntax

import "slices"

func (planner *partitionRegistrationPlanner) projectMembership(registry *Registry, source, into string) ([]PartitionConsumerReplay, error) {
	var consumers []PartitionConsumerReplay
	for _, definition := range registry.Cases {
		names := caseSourcesFor(*registry, definition.ID)
		if len(names) == 0 {
			names = []string{definition.Path}
		}
		if !slices.Contains(names, source) {
			continue
		}
		after := append(slices.Clone(names), into)
		check, err := planner.replayConsumer("case", definition, names, after)
		if err != nil {
			return nil, err
		}
		consumers = append(consumers, check)
		setCaseSourceMembers(registry, definition.ID, after)
	}
	for i, unit := range registry.PackageUnits {
		if !slices.Contains(unit.Members, source) {
			continue
		}
		after := append(slices.Clone(unit.Members), into)
		definition := CaseDefinition{ID: unit.ID, Path: unit.Members[0], Kind: KindValid}
		check, err := planner.replayConsumer("package-syntax", definition, unit.Members, after)
		if err != nil {
			return nil, err
		}
		consumers = append(consumers, check)
		registry.PackageUnits[i].Members = after
	}
	if slices.Contains(registry.MetaSources, source) {
		registry.MetaSources = append(registry.MetaSources, into)
	}
	return consumers, nil
}

func setCaseSourceMembers(registry *Registry, id string, members []string) {
	for i, group := range registry.CaseSources {
		if group.CaseID == id {
			registry.CaseSources[i].Members = members
			return
		}
	}
	registry.CaseSources = append(registry.CaseSources, CaseSourceSet{CaseID: id, Members: members})
}
