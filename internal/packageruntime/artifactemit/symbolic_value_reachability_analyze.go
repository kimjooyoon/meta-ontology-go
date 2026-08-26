package artifactemit

import (
	"slices"
	"strings"
)

func analyzeSymbolicValueReachability(artifact Artifact, contract SymbolicValueContract) symbolicValueReachabilityAnalysis {
	supported := symbolicValueSchemaProfileSupported(artifact.JSONSchema)
	readyRule := findSymbolicValueRule(contract.Rules, "complete-symbolic-invocation")
	missingRule := findSymbolicValueRule(contract.Rules, "missing-activity")
	entailsReady := supported && symbolicValueCompleteRuleRecognized(readyRule)
	readyWitness := entailsReady && symbolicValueSchemaProvidesReadyWitness(artifact.JSONSchema)
	rules := make([]SymbolicValueRuleReachability, 0, len(contract.Rules))
	for _, rule := range contract.Rules {
		switch rule.ID {
		case "complete-symbolic-invocation":
			if readyWitness {
				rules = append(rules, symbolicValueRuleReachability(rule, "REACHABLE", true, "NORMAL_PATH", "GENERATED_SCHEMA_PROVIDES_READY_WITNESS"))
			} else {
				rules = append(rules, symbolicValueRuleReachability(rule, "UNKNOWN", false, "UNCLASSIFIED", "SCHEMA_TO_RULE_RELATION_UNSUPPORTED"))
			}
		case "missing-activity":
			if supported && symbolicValueMissingRuleRecognized(missingRule) {
				rules = append(rules, symbolicValueRuleReachability(rule, "UNREACHABLE", false, "DEFENSE_IN_DEPTH", "GENERATED_SCHEMA_REQUIRES_NON_EMPTY_ACTIVITY"))
			} else {
				rules = append(rules, symbolicValueRuleReachability(rule, "UNKNOWN", false, "UNCLASSIFIED", "SCHEMA_TO_RULE_RELATION_UNSUPPORTED"))
			}
		default:
			rules = append(rules, symbolicValueRuleReachability(rule, "UNKNOWN", false, "UNCLASSIFIED", "VALUE_RULE_NOT_RECOGNIZED"))
		}
	}
	defaultReachability := symbolicValueUnknownDefault(contract.Default)
	if entailsReady && symbolicValueDefaultRecognized(contract.Default) {
		defaultReachability = SymbolicValueDefaultReachability{
			Reachability: "UNREACHABLE", Role: "DEFENSE_IN_DEPTH",
			Reason: "GENERATED_SCHEMA_ENTAILS_COMPLETE_READY_RULE",
			ProofChoice: contract.Default.ProofChoice, MetaOperation: contract.Default.MetaOperation,
		}
	}
	return symbolicValueReachabilityAnalysis{
		SchemaProfileSupported: supported,
		SchemaEntailsReady:      entailsReady,
		Rules:                   rules,
		Default:                 defaultReachability,
		Summary:                 summarizeSymbolicValueReachability(rules, defaultReachability),
	}
}

func symbolicValueSchemaProfileSupported(schema *InvocationSchema) bool {
	if schema == nil || schema.Dialect != JSONSchemaDraft202012 || schema.Type != "object" || schema.AdditionalProperties ||
		len(schema.Required) != 2 || !slices.Contains(schema.Required, "activity") || !slices.Contains(schema.Required, "inputs") ||
		strings.TrimSpace(schema.Properties.Activity.Const) == "" {
		return false
	}
	inputs := schema.Properties.Inputs
	if inputs.Type != "array" || inputs.Items || len(inputs.PrefixItems) == 0 ||
		inputs.MinItems != len(inputs.PrefixItems) || inputs.MaxItems != len(inputs.PrefixItems) {
		return false
	}
	for _, item := range inputs.PrefixItems {
		if strings.TrimSpace(item.Const) == "" {
			return false
		}
	}
	return true
}

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
