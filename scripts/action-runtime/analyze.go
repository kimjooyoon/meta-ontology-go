package main

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
)

func analyze(workflow string, source []byte, commit string) Report {
	policy := runtimePolicy()
	index := make(map[string]Rule, len(policy))
	for _, rule := range policy {
		index[rule.Action] = rule
	}
	report := Report{
		Schema:       reportSchema,
		Metaprogram:  metaprogramPath,
		CommitSHA:    commit,
		Workflow:     workflow,
		SourceSHA256: digest(source),
		PolicySHA256: policyDigest(policy),
		Policy:       policy,
	}
	for _, site := range parseWorkflow(source) {
		observation := observe(site, index)
		report.Observations = append(report.Observations, observation)
		report.ActionsTotal++
		if _, known := index[site.Action]; known {
			report.ActionsKnown++
		}
		if observation.RuntimeVerdict == "PASS" {
			report.ActionsCompliant++
		}
		report.InvalidInputsTotal += len(observation.InvalidInputs)
	}
	return report
}

func observe(site useSite, index map[string]Rule) Observation {
	rule, known := index[site.Action]
	major, parsed := parseMajor(site.Ref)
	runtimeVerdict := verdict(known && parsed && major >= rule.MinimumMajor)
	invalid := invalidInputs(site.Inputs, rule.AllowedInputs)
	inputVerdict := verdict(known && len(invalid) == 0)
	return Observation{
		Action: site.Action, Reference: site.Ref, Line: site.Line,
		MinimumMajor: rule.MinimumMajor, Runtime: rule.Runtime,
		RuntimeVerdict: runtimeVerdict, Inputs: site.Inputs,
		InvalidInputs: invalid, InputVerdict: inputVerdict,
	}
}

func invalidInputs(observed, allowed []string) []string {
	if allowed == nil {
		return nil
	}
	valid := make(map[string]bool, len(allowed))
	for _, input := range allowed {
		valid[input] = true
	}
	invalid := make([]string, 0)
	for _, input := range observed {
		if !valid[input] {
			invalid = append(invalid, input)
		}
	}
	sort.Strings(invalid)
	return invalid
}

func digest(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
