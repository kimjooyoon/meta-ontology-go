package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/compatibilitypolicy"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/publictrust"
)

type derivedCandidate struct {
	Rule   derivedOutputRule
	Before []byte
	After  []byte
	Mode   os.FileMode
}

func deriveCandidate(root string, rule derivedOutputRule, projected map[string][]byte) (derivedCandidate, error) {
	if !allowedDerivedRule(rule) {
		return derivedCandidate{}, fmt.Errorf("unrecognized derivation rule %s", rule.RuleID)
	}
	candidateSource, ok := projected[rule.SourcePath]
	if !ok {
		return derivedCandidate{}, fmt.Errorf("derived source was not projected: %s", rule.SourcePath)
	}
	originalSource, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rule.SourcePath)))
	if err != nil || digest(originalSource) != rule.SourceSHA256 {
		return derivedCandidate{}, fmt.Errorf("derived source preimage changed: %s", rule.SourcePath)
	}
	target := filepath.Join(root, filepath.FromSlash(rule.TargetPath))
	info, err := os.Lstat(target)
	if err != nil || !info.Mode().IsRegular() {
		return derivedCandidate{}, fmt.Errorf("derived target is not a regular file: %s", rule.TargetPath)
	}
	before, err := os.ReadFile(target)
	if err != nil {
		return derivedCandidate{}, err
	}
	if digest(before) != rule.TargetSHA256 {
		return derivedCandidate{}, fmt.Errorf("derived target preimage changed: %s", rule.TargetPath)
	}
	var expected, after []byte
	switch rule.RuleID {
	case "urn:gooo:toolchain-projection:compatibility-evaluator":
		_, expected, err = compatibilitypolicy.GenerateNamed(rule.SourcePath, originalSource)
		if err == nil {
			_, after, err = compatibilitypolicy.GenerateNamed(rule.SourcePath, candidateSource)
		}
	case "urn:gooo:toolchain-projection:public-trust-evaluator":
		_, expected, err = publictrust.GenerateNamed(rule.SourcePath, originalSource)
		if err == nil {
			_, after, err = publictrust.GenerateNamed(rule.SourcePath, candidateSource)
		}
	case "urn:gooo:toolchain-projection:public-trust-badges":
		originalPolicy, _, compileErr := publictrust.GenerateNamed(rule.SourcePath, originalSource)
		if compileErr != nil {
			return derivedCandidate{}, compileErr
		}
		candidatePolicy, _, compileErr := publictrust.GenerateNamed(rule.SourcePath, candidateSource)
		if compileErr != nil {
			return derivedCandidate{}, compileErr
		}
		after, err = replaceBadgeBlock(before, publictrust.RenderBadgeBlock(originalPolicy),
			publictrust.RenderBadgeBlock(candidatePolicy))
		if err != nil {
			return derivedCandidate{}, err
		}
		if bytes.Equal(before, after) {
			return derivedCandidate{}, fmt.Errorf("derived badge output did not change")
		}
		return derivedCandidate{Rule: rule, Before: before, After: after, Mode: info.Mode().Perm()}, nil
	}
	if err != nil {
		return derivedCandidate{}, err
	}
	if !bytes.Equal(before, expected) {
		return derivedCandidate{}, fmt.Errorf("derived target no longer matches its Gooo source: %s", rule.TargetPath)
	}
	if bytes.Equal(before, after) {
		return derivedCandidate{}, fmt.Errorf("derived evaluator output did not change: %s", rule.TargetPath)
	}
	return derivedCandidate{Rule: rule, Before: before, After: after, Mode: info.Mode().Perm()}, nil
}

func allowedDerivedRule(rule derivedOutputRule) bool {
	if rule.RuleVersion != 1 {
		return false
	}
	switch rule.RuleID {
	case "urn:gooo:toolchain-projection:compatibility-evaluator":
		return rule.SourcePath == compatibilitySourcePath && rule.TargetPath == compatibilityTargetPath
	case "urn:gooo:toolchain-projection:public-trust-evaluator":
		return rule.SourcePath == publicTrustSourcePath && rule.TargetPath == publicTrustTargetPath
	case "urn:gooo:toolchain-projection:public-trust-badges":
		return rule.SourcePath == publicTrustSourcePath && rule.TargetPath == readmeTargetPath
	}
	return false
}

func replaceBadgeBlock(readme []byte, before, after string) ([]byte, error) {
	const begin = "<!-- PUBLIC-TRUST-BADGES:BEGIN -->"
	const end = "<!-- PUBLIC-TRUST-BADGES:END -->"
	if bytes.Count(readme, []byte(begin)) != 1 || bytes.Count(readme, []byte(end)) != 1 ||
		bytes.Count(readme, []byte(before)) != 1 {
		return nil, fmt.Errorf("README public trust block no longer matches its Gooo source")
	}
	return bytes.Replace(readme, []byte(before), []byte(after), 1), nil
}
