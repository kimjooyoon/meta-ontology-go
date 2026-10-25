package main

import (
	"bytes"
	"fmt"
	"maps"
	"path"
	"path/filepath"
	"slices"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/syntaxregistration"
	"github.com/kimjooyoon/meta-ontology-go/internal/sourcebundle"
)

func describeSplitRegistration(root, directory string, partition sourcebundle.Partition, candidate syntaxregistration.PartitionRegistrationCandidate) (*workspaceSplitRegistration, error) {
	operation, err := generation.SplitGoooInputContractEvidence()
	if err != nil {
		return nil, err
	}
	inputs, err := splitRegistrationEmbedInputs(root, candidate)
	if err != nil {
		return nil, err
	}
	if err := captureSplitOperationInputs(root, operation, inputs); err != nil {
		return nil, err
	}
	maps.Copy(inputs, candidate.Inputs)
	result := &workspaceSplitRegistration{Root: root, Directory: directory, Operation: operation, Profile: *partition.EntityFieldsSupport}
	for _, name := range slices.Sorted(maps.Keys(inputs)) {
		result.Inputs = append(result.Inputs, workspaceSplitRegistrationInput{Path: name, SHA256: workspaceDigest(inputs[name])})
	}
	for _, name := range slices.Sorted(maps.Keys(candidate.Outputs)) {
		// Source fragments already have one wire owner in receipt.Result.Files.
		if name == path.Join(directory, partition.Files[0].Filename) || name == path.Join(directory, partition.Files[1].Filename) {
			continue
		}
		raw := candidate.Outputs[name]
		result.Outputs = append(result.Outputs, workspaceSplitRegistrationOutput{Path: name, SHA256: workspaceDigest(raw), Source: string(raw)})
	}
	return result, nil
}

func captureSplitOperationInputs(root string, operation generation.OperationInputContractEvidence, inputs map[string][]byte) error {
	names, err := filepath.Glob(filepath.Join(root, "internal/meta/generation/operation-input-contract*.gooo"))
	if err != nil {
		return err
	}
	if len(names) != len(operation.SourceOrigins.Members) {
		return fmt.Errorf("repository operation source membership differs from this compiler")
	}
	for _, member := range operation.SourceOrigins.Members {
		name := "internal/meta/generation/" + member.Filename
		target, err := splitRegistrationTarget(root, name)
		if err != nil {
			return err
		}
		raw, err := readRegularFile(target, maxInputBytes)
		if err != nil {
			return err
		}
		if len(raw) != member.ByteLength || workspaceDigest(raw) != "sha256:"+member.SHA256 {
			return fmt.Errorf("repository operation contract differs from this compiler: %s", name)
		}
		inputs[name] = raw
	}
	return nil
}

func appendSplitRegistrationWrites(writes []atomicWrite, registration *workspaceSplitRegistration) ([]atomicWrite, error) {
	if registration == nil {
		return writes, nil
	}
	inputs := make(map[string]string, len(registration.Inputs))
	for _, input := range registration.Inputs {
		inputs[input.Path] = input.SHA256
	}
	for _, output := range registration.Outputs {
		write, err := splitRegistrationWrite(registration.Root, output, inputs[output.Path])
		if err != nil {
			return nil, err
		}
		duplicate, err := identicalSplitWrite(writes, write)
		if err != nil {
			return nil, err
		}
		if !duplicate {
			writes = append(writes, write)
		}
	}
	return writes, nil
}

func splitRegistrationWrite(root string, output workspaceSplitRegistrationOutput, beforeDigest string) (atomicWrite, error) {
	target, err := splitRegistrationTarget(root, output.Path)
	if err != nil {
		return atomicWrite{}, err
	}
	before, err := captureOutputSnapshot(target)
	if err != nil {
		return atomicWrite{}, err
	}
	if (beforeDigest == "" && before.exists) || (beforeDigest != "" && (!before.exists || workspaceDigest(before.data) != beforeDigest)) {
		return atomicWrite{}, fmt.Errorf("syntax registration input changed: %s", output.Path)
	}
	if workspaceDigest([]byte(output.Source)) != output.SHA256 {
		return atomicWrite{}, fmt.Errorf("syntax registration output digest differs: %s", output.Path)
	}
	return atomicWrite{path: target, data: []byte(output.Source), before: &before}, nil
}

func identicalSplitWrite(writes []atomicWrite, candidate atomicWrite) (bool, error) {
	for _, write := range writes {
		shared, err := samePublicationEntry(write.path, candidate.path)
		if err != nil {
			return false, err
		}
		if write.path != candidate.path && !shared {
			continue
		}
		if write.path != candidate.path || !bytes.Equal(write.data, candidate.data) || !sameOutputSnapshot(*write.before, *candidate.before) {
			return false, fmt.Errorf("syntax registration and workspace disagree on publication target %s", candidate.path)
		}
		return true, nil
	}
	return false, nil
}

func verifySplitRegistrationOutput(registration *workspaceSplitRegistration, reader SourceReader) error {
	if registration == nil {
		return nil
	}
	for _, output := range registration.Outputs {
		raw, err := reader.ReadFile(filepath.Join(registration.Root, filepath.FromSlash(output.Path)))
		if err != nil {
			return err
		}
		if workspaceDigest(raw) != output.SHA256 {
			return fmt.Errorf("installed syntax registration output differs: %s", output.Path)
		}
	}
	return nil
}
