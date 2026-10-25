package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kimjooyoon/meta-ontology-go/internal/completeness"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/workspacesplit"
)

// This is a read projection of the CLI receipt. The CLI's existing strict
// decoder and re-preparation remain authoritative for the complete proposal.
type goooSplitReceipt struct {
	RawSHA256    string                            `json:"-"`
	ReplaySHA256 string                            `json:"-"`
	Schema       string                            `json:"schema"`
	Decision     string                            `json:"decision"`
	ManifestPath string                            `json:"manifest_path"`
	Error        string                            `json:"error"`
	Result       *packageruntime.SourceSplit       `json:"result"`
	Selection    *packageruntime.SourceSplitChoice `json:"selection"`
	Discovery    *struct {
		Root       string `json:"root"`
		Candidates string `json:"candidates"`
	} `json:"discovery"`
	Registration *struct {
		Root      string                                    `json:"root"`
		Operation generation.OperationInputContractEvidence `json:"operation"`
		Outputs   []struct {
			Path   string `json:"path"`
			SHA256 string `json:"sha256"`
			Source string `json:"source"`
		} `json:"outputs"`
	} `json:"syntax_registration"`
	Application *struct {
		State          string `json:"state"`
		ProposalSHA256 string `json:"proposal_sha256"`
	} `json:"application"`
}

func readGoooSplitReceipt(raw []byte) (goooSplitReceipt, error) {
	var receipt goooSplitReceipt
	if err := completeness.CheckJSON(raw, 32<<20); err != nil {
		return receipt, err
	}
	if err := json.Unmarshal(raw, &receipt); err != nil {
		return receipt, err
	}
	receipt.RawSHA256 = digestBytes(raw)
	if receipt.Schema != "gooo/workspace-source-split-receipt/v1" || receipt.Error != "" ||
		receipt.Result == nil || len(receipt.Result.Files) != 2 || receipt.Selection == nil ||
		receipt.Selection.Boundary == nil || receipt.Selection.State != "SELECTED" ||
		receipt.Registration == nil || receipt.Discovery == nil {
		return receipt, fmt.Errorf("complete registered and scoped source split receipt is required")
	}
	return receipt, nil
}

func verifyGoooSplitProposal(root string, request workspacesplit.Request, action generation.Action, receipt goooSplitReceipt) error {
	result, selection, registration := receipt.Result, receipt.Selection, receipt.Registration
	if receipt.Decision != "PROPOSED" || receipt.Application != nil ||
		receipt.ManifestPath != filepath.Join(root, filepath.FromSlash(request.Workspace)) ||
		result.PackagePath != request.Package || result.InputFilename != request.Source || result.Files[1].Filename != request.Into ||
		selection.LineLimit != request.LineLimit || selection.InputLines != action.SourceIndicator.Value ||
		selection.InputSHA256 != result.InputSHA256 || selection.Boundary.BeforeID != result.BeforeID ||
		registration.Root != root || receipt.Discovery.Root != filepath.Join(root, filepath.FromSlash(request.DiscoveryRoot)) ||
		receipt.Discovery.Candidates != "recursive-json-files" {
		return fmt.Errorf("split receipt differs from the exact metric and scoped request")
	}
	operation := registration.Operation
	if operation.SourceDigest != action.InputContractSourceDigest || operation.SemanticDigest != action.InputContractSemanticDigest ||
		operation.Operation != action.Operation || operation.ActivityID == "" || !operation.UsedInputFact || !operation.GeneratedOutputFact {
		return fmt.Errorf("split receipt differs from the source-owned operation binding")
	}
	before, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(request.Subject())))
	if err != nil {
		return err
	}
	return verifyGoooSplitOrigins(before, *result)
}

func verifyGoooSplitOrigins(before []byte, result packageruntime.SourceSplit) error {
	if digestBytes(before) != result.InputSHA256 {
		return fmt.Errorf("split source digest differs")
	}
	covered := make([]bool, len(before))
	for _, file := range result.Files {
		if digestBytes([]byte(file.Source)) != file.SHA256 {
			return fmt.Errorf("split output digest differs")
		}
		position := 0
		for _, origin := range file.Origins {
			if origin.InputStart < 0 || origin.InputEnd <= origin.InputStart || origin.InputEnd > len(before) ||
				origin.OutputStart != position || origin.OutputEnd <= position || origin.OutputEnd > len(file.Source) ||
				!bytes.Equal(before[origin.InputStart:origin.InputEnd], []byte(file.Source[position:origin.OutputEnd])) {
				return fmt.Errorf("split origin does not reproduce its exact output bytes")
			}
			for i := origin.InputStart; i < origin.InputEnd; i++ {
				covered[i] = true
			}
			position = origin.OutputEnd
		}
		if position != len(file.Source) {
			return fmt.Errorf("split output has bytes without an origin")
		}
	}
	for _, present := range covered {
		if !present {
			return fmt.Errorf("split lost source bytes")
		}
	}
	return nil
}
