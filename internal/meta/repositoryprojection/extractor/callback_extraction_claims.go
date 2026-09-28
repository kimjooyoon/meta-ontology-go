package extractor

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"slices"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
)

type CallbackExtractionClaim struct {
	ID            string                              `json:"id"`
	State         string                              `json:"state"`
	Stage         string                              `json:"stage"`
	Step          string                              `json:"step"`
	Reason        string                              `json:"reason"`
	UnknownClass  string                              `json:"unknown_class"`
	NextOperation string                              `json:"next_operation"`
	BlockedBy     []string                            `json:"blocked_by"`
	Record        generation.CallbackExtractionRecord `json:"record"`
}

func bindCallbackExtractionClaims(proposal *CallbackExtractionProposal) error {
	digests := []string{proposal.SourceDigest, proposal.StructureProof.ProofDigest, proposal.PartitionDigest, proposal.PackageDigest, "", ""}
	observed := []int{1, 1, len(proposal.Artifacts), 1, 0, 4}
	required := []int{1, 1, len(proposal.Artifacts), 1, 1, 5}
	for index, step := range proposal.Contract.Steps {
		claim := CallbackExtractionClaim{ID: "gooo://callback-extraction/claim/" + step.ID, State: "CLOSED",
			Stage: "CALLBACK_EXTRACTION", Step: step.Activity, Reason: "OBSERVED_PREDICATE_SATISFIED", BlockedBy: []string{}}
		if index == 4 {
			claim.State, claim.Reason, claim.UnknownClass = "UNKNOWN", "CALLBACK_OBSERVER_EVIDENCE_MISSING", "DIRECT_MISSING"
			claim.NextOperation = "CAPTURE_SOURCE_AND_FINAL_PACKAGE_OBSERVER_EVIDENCE"
		}
		if index == 5 {
			claim.State, claim.Reason, claim.UnknownClass = "UNKNOWN", "CALLBACK_OBSERVER_EVIDENCE_UNPROVEN", "DEPENDENCY_BLOCKED"
			claim.NextOperation = "RESOLVE_CALLBACK_OBSERVER_EVIDENCE"
			claim.BlockedBy = []string{"gooo://callback-extraction/claim/observers"}
		}
		record, err := proposal.Contract.BuildRecord(index, claim.State, digests[index], observed[index], required[index])
		if err != nil {
			return err
		}
		claim.Record = record
		proposal.Claims = append(proposal.Claims, claim)
		if claim.State == "CLOSED" {
			proposal.Closed++
		} else {
			proposal.Unknown++
		}
	}
	return nil
}

// Partitioning may change files and import headers, but may not silently alter
// any non-import declaration. This proof covers the final generated units.
func callbackExtractionPartitionDigest(intermediate []byte, generated map[string][]byte) (string, error) {
	before, err := callbackExtractionDeclarationDigests(intermediate)
	if err != nil {
		return "", err
	}
	var after []string
	for _, source := range generated {
		declarations, err := callbackExtractionDeclarationDigests(source)
		if err != nil {
			return "", err
		}
		after = append(after, declarations...)
	}
	slices.Sort(before)
	slices.Sort(after)
	if !slices.Equal(before, after) {
		return "", fail("plan-callback-extraction", "prove-final-partition", "CALLBACK_DECLARATION_PARTITION_REFUTED", "KNOWN_CONTRADICTION", "report-counterexample", nil)
	}
	var payload bytes.Buffer
	for _, digest := range before {
		payload.WriteString(digest)
		payload.WriteByte(0)
	}
	return proofDigest(payload.Bytes()), nil
}

func callbackExtractionDeclarationDigests(source []byte) ([]string, error) {
	file, err := parser.ParseFile(token.NewFileSet(), "partition.go", source, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("callback extraction partition syntax: %w", err)
	}
	declarations := make([]string, 0, len(file.Decls))
	for _, declaration := range file.Decls {
		if group, ok := declaration.(*ast.GenDecl); ok && group.Tok == token.IMPORT {
			continue
		}
		declarations = append(declarations, proofDigest(closureASTEncoding(declaration)))
	}
	return declarations, nil
}
