package artifactemit

import "slices"

func symbolicValueSchemaProvidesReadyWitness(schema *InvocationSchema) bool {
	if schema == nil || len(schema.Examples) != 1 {
		return false
	}
	expectedInputs := make([]string, len(schema.Properties.Inputs.PrefixItems))
	for index, item := range schema.Properties.Inputs.PrefixItems {
		expectedInputs[index] = item.Const
	}
	example := schema.Examples[0]
	return example.Activity == schema.Properties.Activity.Const && slices.Equal(example.Inputs, expectedInputs)
}
