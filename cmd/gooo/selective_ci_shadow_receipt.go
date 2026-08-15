package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"
)

func shadowFallback(output selectiveCIShadowOutput, stage, component, reason string) selectiveCIShadowOutput {
	output.Status = "FULL_SUITE_FALLBACK"
	output.Stage = stage
	output.Component = component
	output.Reason = reason
	output.ExecutionAuthorized = false
	output.ShadowOnly = true
	output.ChangedSemanticIDs = []string{}
	output.SelectedCommands = []shadowCommandSpec{}
	output.SelectedGuards = []shadowCommandSpec{}
	output.SelectedWorkIDs = []string{}
	output.ResourceReceipts = []shadowResourceReceipt{}
	return sealSelectiveCIShadowOutput(output)
}

func (output selectiveCIShadowOutput) canonicalJSON() ([]byte, error) {
	copy := normalizeSelectiveCIShadowOutput(output)
	copy.CanonicalDigest = ""
	return json.Marshal(copy)
}

func (output selectiveCIShadowOutput) stableDigest() string {
	data, err := output.canonicalJSON()
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func sealSelectiveCIShadowOutput(output selectiveCIShadowOutput) selectiveCIShadowOutput {
	output = normalizeSelectiveCIShadowOutput(output)
	output.CanonicalDigest = output.stableDigest()
	return output
}

func normalizeSelectiveCIShadowOutput(output selectiveCIShadowOutput) selectiveCIShadowOutput {
	if output.ChangedSemanticIDs == nil {
		output.ChangedSemanticIDs = []string{}
	}
	if output.SelectedCommands == nil {
		output.SelectedCommands = []shadowCommandSpec{}
	}
	if output.SelectedGuards == nil {
		output.SelectedGuards = []shadowCommandSpec{}
	}
	if output.SelectedWorkIDs == nil {
		output.SelectedWorkIDs = []string{}
	}
	if output.ResourceReceipts == nil {
		output.ResourceReceipts = []shadowResourceReceipt{}
	}
	sort.Strings(output.ChangedSemanticIDs)
	output.ChangedSemanticIDs = uniqueStrings(output.ChangedSemanticIDs)
	sort.Slice(output.SelectedCommands, func(i, j int) bool { return output.SelectedCommands[i].ID < output.SelectedCommands[j].ID })
	sort.Slice(output.SelectedGuards, func(i, j int) bool { return output.SelectedGuards[i].ID < output.SelectedGuards[j].ID })
	sort.Strings(output.SelectedWorkIDs)
	output.SelectedWorkIDs = uniqueStrings(output.SelectedWorkIDs)
	sort.Slice(output.ResourceReceipts, func(i, j int) bool {
		return output.ResourceReceipts[i].CommandID < output.ResourceReceipts[j].CommandID
	})
	return output
}

func encodeSelectiveCIShadowOutput(output selectiveCIShadowOutput) ([]byte, error) {
	output = sealSelectiveCIShadowOutput(output)
	data, err := json.Marshal(output)
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}
