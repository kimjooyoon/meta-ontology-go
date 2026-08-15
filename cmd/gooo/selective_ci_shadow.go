package main

import (
	"errors"
	"fmt"
	"io"
	"strings"
)

const selectiveCIShadowUsage = "usage: gooo selective-ci shadow --base-snapshot FILE --head-snapshot FILE --plan-input FILE --evidence-input FILE --lane-input FILE"

const selectiveCIShadowSchemaVersion = "gooo/selective-ci-shadow/v1"

type selectiveCIShadowOptions struct {
	baseSnapshot  string
	headSnapshot  string
	planInput     string
	evidenceInput string
	laneInput     string
}

type shadowInputFiles struct {
	baseSnapshot  []byte
	headSnapshot  []byte
	planInput     []byte
	evidenceInput []byte
	laneInput     []byte
}

type shadowCommandSpec struct {
	ID   string   `json:"id"`
	Argv []string `json:"argv"`
}

type shadowResourceReceipt struct {
	CommandID    string `json:"command_id"`
	CPUWorkUnits uint64 `json:"cpu_work_units"`
	MemoryBytes  uint64 `json:"memory_bytes"`
}

type shadowLaneReceipt struct {
	Decision       string `json:"decision"`
	Reason         string `json:"reason"`
	RegistryDigest string `json:"registry_digest"`
	BaseSHA        string `json:"base_sha"`
	LaneHeadSHA    string `json:"lane_head_sha"`
	LaneID         string `json:"lane_id"`
}

// selectiveCIShadowOutput is intentionally a receipt, not an execution plan.
// It contains argv as data only; no field authorizes or represents execution.
type selectiveCIShadowOutput struct {
	SchemaVersion       string                  `json:"schema_version"`
	Command             string                  `json:"command"`
	Status              string                  `json:"status"`
	Stage               string                  `json:"stage"`
	Component           string                  `json:"component"`
	Reason              string                  `json:"reason"`
	ExecutionAuthorized bool                    `json:"execution_authorized"`
	ShadowOnly          bool                    `json:"shadow_only"`
	BaseSourceDigest    string                  `json:"base_source_digest"`
	HeadSourceDigest    string                  `json:"head_source_digest"`
	BaseSemanticDigest  string                  `json:"base_semantic_digest"`
	HeadSemanticDigest  string                  `json:"head_semantic_digest"`
	RegistryDigest      string                  `json:"registry_digest"`
	PlanDigest          string                  `json:"plan_digest"`
	ProofStatus         string                  `json:"proof_status"`
	ProofCode           string                  `json:"proof_code"`
	ChangedSemanticIDs  []string                `json:"changed_semantic_ids"`
	SelectedCommands    []shadowCommandSpec     `json:"selected_commands"`
	SelectedGuards      []shadowCommandSpec     `json:"selected_guard_commands"`
	SelectedWorkIDs     []string                `json:"selected_work_ids"`
	ResourceReceipts    []shadowResourceReceipt `json:"resource_receipts"`
	Lane                shadowLaneReceipt       `json:"lane"`
	CanonicalDigest     string                  `json:"canonical_digest"`
}

func parseSelectiveCIShadowArguments(args []string) (selectiveCIShadowOptions, error) {
	var options selectiveCIShadowOptions
	seen := map[string]bool{}
	for index := 0; index < len(args); index++ {
		flag := args[index]
		var target *string
		switch flag {
		case "--base-snapshot":
			target = &options.baseSnapshot
		case "--head-snapshot":
			target = &options.headSnapshot
		case "--plan-input":
			target = &options.planInput
		case "--evidence-input":
			target = &options.evidenceInput
		case "--lane-input":
			target = &options.laneInput
		default:
			return selectiveCIShadowOptions{}, errors.New(selectiveCIShadowUsage)
		}
		if seen[flag] || index+1 >= len(args) || args[index+1] == "" || strings.HasPrefix(args[index+1], "-") {
			return selectiveCIShadowOptions{}, errors.New(selectiveCIShadowUsage)
		}
		seen[flag] = true
		*target = args[index+1]
		index++
	}
	if options.baseSnapshot == "" || options.headSnapshot == "" || options.planInput == "" || options.evidenceInput == "" || options.laneInput == "" {
		return selectiveCIShadowOptions{}, errors.New(selectiveCIShadowUsage)
	}
	return options, nil
}

func runSelectiveCI(args []string, reader SourceReader, stdout, stderr io.Writer) int {
	if len(args) == 0 || args[0] != "shadow" {
		fmt.Fprintln(stderr, selectiveCIShadowUsage)
		return exitUsage
	}
	options, err := parseSelectiveCIShadowArguments(args[1:])
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitUsage
	}
	files, missing := readSelectiveCIShadowFiles(options, reader)
	if missing != "" {
		// File availability is a CLI contract error. Do not emit a partial
		// semantic receipt when one of the explicit inputs cannot be read.
		fmt.Fprintf(stderr, "gooo: cli.usage: missing %s input file\n%s\n", missing, selectiveCIShadowUsage)
		return exitUsage
	}
	output := evaluateSelectiveCIShadow(files)
	data, err := encodeSelectiveCIShadowOutput(output)
	if err != nil {
		fmt.Fprintf(stderr, "gooo: selective-ci shadow: output encoding failed: %v\n", err)
		return exitFailure
	}
	if _, err := stdout.Write(data); err != nil {
		return exitFailure
	}
	return exitOK
}

func readSelectiveCIShadowFiles(options selectiveCIShadowOptions, reader SourceReader) (shadowInputFiles, string) {
	read := func(name, filename string) ([]byte, string) {
		data, err := reader.ReadFile(filename)
		if err != nil {
			return nil, name
		}
		return data, ""
	}
	base, missing := read("base_snapshot", options.baseSnapshot)
	if missing != "" {
		return shadowInputFiles{}, missing
	}
	head, missing := read("head_snapshot", options.headSnapshot)
	if missing != "" {
		return shadowInputFiles{}, missing
	}
	plan, missing := read("plan_input", options.planInput)
	if missing != "" {
		return shadowInputFiles{}, missing
	}
	evidence, missing := read("evidence_input", options.evidenceInput)
	if missing != "" {
		return shadowInputFiles{}, missing
	}
	lane, missing := read("lane_input", options.laneInput)
	if missing != "" {
		return shadowInputFiles{}, missing
	}
	return shadowInputFiles{baseSnapshot: base, headSnapshot: head, planInput: plan, evidenceInput: evidence, laneInput: lane}, ""
}

func newSelectiveCIShadowOutput() selectiveCIShadowOutput {
	return selectiveCIShadowOutput{
		SchemaVersion:       selectiveCIShadowSchemaVersion,
		Command:             "selective-ci shadow",
		ExecutionAuthorized: false,
		ShadowOnly:          true,
		ChangedSemanticIDs:  []string{},
		SelectedCommands:    []shadowCommandSpec{},
		SelectedGuards:      []shadowCommandSpec{},
		SelectedWorkIDs:     []string{},
		ResourceReceipts:    []shadowResourceReceipt{},
	}
}
