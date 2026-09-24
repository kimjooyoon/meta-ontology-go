package main

import (
	"reflect"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/lsp"
)

func TestRunLSPInitializeAdvertisesExactSupportedCapabilities(t *testing.T) {
	input := lspTranscript(
		lspRequest(1, "initialize", nil),
		lspNotification("initialized", nil),
		lspRequest(2, "shutdown", nil),
		lspNotification("exit", nil),
	)
	first, code, stderr := runLSPTranscript(t, input)
	if code != exitOK || stderr != "" {
		t.Fatalf("initialize lifecycle = code %d, stderr=%q, output=%q", code, stderr, first)
	}
	messages := readLSPFrames(t, first)
	if len(messages) != 2 {
		t.Fatalf("lifecycle messages = %d, want 2", len(messages))
	}
	var initialize struct {
		JSONRPC string               `json:"jsonrpc"`
		ID      int                  `json:"id"`
		Result  lsp.InitializeResult `json:"result"`
	}
	decodeLSPJSON(t, messages[0], &initialize)
	if initialize.JSONRPC != "2.0" || initialize.ID != 1 || initialize.Result.ServerInfo.Name != "gooo-lsp" || initialize.Result.ServerInfo.Version != "current-ddaf" {
		t.Fatalf("initialize envelope = %#v", initialize)
	}
	want := lsp.ServerCapabilities{
		TextDocumentSync:       lsp.TextDocumentSyncOptions{OpenClose: true, Change: 2},
		HoverProvider:          true,
		CompletionProvider:     &lsp.CompletionOptions{},
		DefinitionProvider:     true,
		DocumentSymbolProvider: true,
		ReferencesProvider:     true,
		RenameProvider:         true,
		WorkspaceSymbolProvider: &lsp.WorkspaceSymbolOptions{
			Schema: lsp.WorkspaceSymbolProtocolSchema,
		},
		SemanticTokensProvider: &lsp.SemanticTokensOptions{
			Schema: lsp.SemanticTokensProtocolSchema,
			Legend: lsp.SemanticTokensLegend{
				TokenTypes:     []string{"entity", "activity", "reference", "symbol"},
				TokenModifiers: []string{},
			},
			Full: true,
		},
		Experimental: map[string]any{
			"goooDocumentProvenance": map[string]any{
				"method": "gooo/documentProvenance",
				"schema": "gooo/lsp-document-provenance/v3",
			},
			"goooSelfImprovementProvenance": map[string]any{
				"method": "gooo/selfImprovementProvenance",
				"schema": "gooo/lsp-self-improvement-provenance-chain/v1",
			},
			"goooReferencesProvenance": map[string]any{
				"method": "gooo/referencesProvenance",
				"schema": "gooo/lsp-references-provenance/v1",
			},
			"goooDiagnosticProvenance": map[string]any{
				"method": "gooo/diagnosticProvenance",
				"schema": "gooo/lsp-diagnostic-provenance/v1",
			},
			"goooCompletionProvenance": map[string]any{
				"method": "gooo/completionProvenance",
				"schema": "gooo/lsp-completion-provenance/v1",
			},
			"goooExecutionPlanProvenance": map[string]any{
				"method": "gooo/executionPlanProvenance",
				"schema": "gooo/execution-plan-provenance-binding/v1",
			},
			"goooStoryProvenance": map[string]any{
				"method": "gooo/storyProvenance",
				"schema": "gooo/lsp-story-provenance/v1",
			},
		},
	}
	if !reflect.DeepEqual(initialize.Result.Capabilities, want) {
		t.Fatalf("capabilities = %#v, want %#v", initialize.Result.Capabilities, want)
	}
	assertLSPResponseID(t, messages[1], 2)
}
