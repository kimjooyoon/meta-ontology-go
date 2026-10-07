package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"

	jev "github.com/kimjooyoon/gooo-jev/gooo"
	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/completeness"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

type discoveryDomainContract struct {
	Path   string
	Source []byte
	IR     semantic.IR
}

func loadDiscoveryDomainContract(reader SourceReader, sourcePath string, source []byte, contractPath string) (*discoveryDomainContract, string, error) {
	if contractPath == "" {
		return nil, "", nil
	}
	if filepath.Clean(contractPath) == filepath.Clean(sourcePath) {
		return nil, "DOMAIN_CONTRACT_IS_SOURCE", fmt.Errorf("domain contract must be a separate Gooo file")
	}
	contractSource, err := reader.ReadFile(contractPath)
	if err != nil {
		return nil, "DOMAIN_CONTRACT_READ_FAILED", err
	}
	if bytes.Equal(source, contractSource) {
		return nil, "DOMAIN_CONTRACT_IDENTICAL_TO_SOURCE", fmt.Errorf("domain contract must define an independent declaration scope")
	}
	file, diagnostics := syntax.ParseFileWithEntityFieldsSupport(contractPath, string(contractSource), syntax.EntityFieldsV4Support())
	if diagnostics.HasErrors() || file == nil {
		if diagnosticErr := diagnostics.Error(); diagnosticErr != nil {
			return nil, "DOMAIN_CONTRACT_PARSE_FAILED", diagnosticErr
		}
		return nil, "DOMAIN_CONTRACT_PARSE_FAILED", fmt.Errorf("domain contract parser did not produce a declaration")
	}
	if len(file.Bindings) != 0 {
		return nil, "DOMAIN_CONTRACT_HAS_RUNTIME_BINDINGS", fmt.Errorf("domain contract may declare expected entities and activities but cannot declare runtime binds")
	}
	ir, err := bidir.LowerContextWithEntityFieldsSupport(context.Background(), file, syntax.EntityFieldsV4Support())
	if err != nil {
		return nil, "DOMAIN_CONTRACT_LOWER_FAILED", err
	}
	if len(domainDeclarations(ir)) == 0 {
		return nil, "DOMAIN_CONTRACT_EMPTY", fmt.Errorf("domain contract must declare at least one entity or activity")
	}
	return &discoveryDomainContract{Path: contractPath, Source: bytes.Clone(contractSource), IR: ir}, "", nil
}

func discoveryDeclarationCoverage(actual semantic.IR, contract *discoveryDomainContract) completeness.CompletenessDimension {
	if contract == nil {
		return completeness.CompletenessDimension{
			ID: "declaration_coverage", Status: "UNKNOWN", Numerator: 0, Denominator: 0,
			Unit:     "expected stable Gooo entity and activity declarations from a separate domain contract",
			Reason:   "The target source cannot define its own coverage denominator; no separate Gooo domain contract was supplied.",
			Evidence: []string{"domain_contract:absent", "target_declarations_are_not_a_denominator"},
		}
	}
	expected := domainDeclarations(contract.IR)
	actualByID := make(map[semantic.ID]semantic.Node)
	for _, node := range domainDeclarations(actual) {
		actualByID[node.ID] = node
	}
	covered := 0
	uncoveredIDs := make([]string, 0)
	for _, node := range expected {
		observed, ok := actualByID[node.ID]
		if ok && discoveryDeclarationFingerprint(contract.IR, node) == discoveryDeclarationFingerprint(actual, observed) {
			covered++
			continue
		}
		uncoveredIDs = append(uncoveredIDs, node.ID.String())
	}
	status := "PASS"
	reason := "Every stable declaration required by the separate Gooo domain contract is present with matching kind, namespace, fields, and typed activity ports."
	if covered != len(expected) {
		status = "PROGRESS"
		reason = "Some stable declarations required by the separate Gooo domain contract are missing or have a different semantic shape."
	}
	contractDigest := sha256Hex(contract.Source)
	contractSemanticDigest := contract.IR.StableHash()
	uncoveredJSON, _ := json.Marshal(uncoveredIDs)
	evidence := []string{
		"domain_contract_path:" + contract.Path,
		"domain_contract_digest:sha256:" + contractDigest,
		"domain_contract_semantic_digest:" + contractSemanticDigest,
		"covered_declarations:" + fmt.Sprintf("%d", covered),
		"expected_declarations:" + fmt.Sprintf("%d", len(expected)),
		"uncovered_ids:" + string(uncoveredJSON),
		"implementation_bodies_are_outside_this_axis",
	}
	return completeness.CompletenessDimension{
		ID: "declaration_coverage", Status: status, Numerator: covered, Denominator: len(expected),
		Unit:   "stable Gooo entity and activity declarations in the supplied domain contract",
		Reason: reason, Evidence: evidence,
	}
}

func domainDeclarations(ir semantic.IR) []semantic.Node {
	declarations := make([]semantic.Node, 0)
	for _, node := range ir.Graph.Nodes() {
		if node.Kind == semantic.Entity || node.Kind == semantic.Activity {
			declarations = append(declarations, node)
		}
	}
	return declarations
}

func discoveryDeclarationFingerprint(ir semantic.IR, node semantic.Node) string {
	type declaration struct {
		ID        string   `json:"id"`
		Kind      string   `json:"kind"`
		Namespace string   `json:"namespace"`
		Fields    []string `json:"fields"`
		Inputs    []string `json:"inputs"`
		Outputs   []string `json:"outputs"`
	}
	value := declaration{ID: node.ID.String(), Kind: node.Kind.String(), Namespace: node.Namespace.String(),
		Fields: make([]string, 0, len(node.Fields)), Inputs: make([]string, 0), Outputs: make([]string, 0)}
	for _, field := range node.Fields {
		value.Fields = append(value.Fields, field.SemanticCanonical())
	}
	if node.Kind == semantic.Activity {
		for _, fact := range ir.Graph.Facts() {
			switch {
			case fact.Subject == node.ID && fact.Predicate == semantic.Used:
				value.Inputs = append(value.Inputs, fact.Object.String())
			case fact.Object == node.ID && fact.Predicate == semantic.WasGeneratedBy:
				value.Outputs = append(value.Outputs, fact.Subject.String())
			}
		}
		sort.Strings(value.Inputs)
		sort.Strings(value.Outputs)
	}
	canonical, _ := json.Marshal(value)
	digest := sha256.Sum256(canonical)
	return hex.EncodeToString(digest[:])
}

func discoveryDomainContractScope(contract *discoveryDomainContract) map[string]any {
	if contract == nil {
		return map[string]any{"provided": false, "expected_declarations": nil}
	}
	return map[string]any{
		"provided": true, "path": contract.Path,
		"source_digest":         "sha256:" + sha256Hex(contract.Source),
		"semantic_digest":       contract.IR.StableHash(),
		"expected_declarations": len(domainDeclarations(contract.IR)),
		"shape_rule":            "stable ID, kind, namespace, entity fields, activity inputs and outputs; implementation body excluded",
	}
}

func discoveryProvenanceEvidence(sourceDigest, semanticDigest string, trail jev.CapabilityQueryTrail, contract *discoveryDomainContract) []string {
	evidence := []string{
		"source_digest:" + sourceDigest, "semantic_digest:" + semanticDigest,
		"jev_declaration_observation_digest:" + trail.Response.Declaration.SourceDigest,
		"jev_query_digest:" + trail.Response.QueryDigest, "jev_evidence_digest:" + trail.EvidenceDigest,
	}
	if contract != nil {
		evidence = append(evidence,
			"domain_contract_digest:sha256:"+sha256Hex(contract.Source),
			"domain_contract_semantic_digest:"+contract.IR.StableHash())
	}
	return evidence
}

func sha256Hex(value []byte) string {
	digest := sha256.Sum256(value)
	return hex.EncodeToString(digest[:])
}
