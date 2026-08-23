package toolchainlsp

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/lsp"
)

type runtimeStats struct {
	Capabilities, ReadFeatures, DiagnosticPaths, NavigationPaths int
	SymbolPaths, SemanticTokenPaths, UTF16Replays, TranscriptReplays int
	FailClosedPaths, UnexpectedProtocolErrors, DiagnosticGaps int
	NonstandardWireFields, StaleLeaks, UnknownLeaks, FailClosedLeaks int
}

type observation struct { Observed string; Satisfied bool }

func observeServer(raw []byte, messages []rpcMessage, replay []byte) (map[string]observation, runtimeStats) {
	result := map[string]observation{}
	stats := runtimeStats{NonstandardWireFields: forbiddenWireFields(raw)}
	responses := responseMap(messages)
	capabilities, capabilityCount := observeCapabilities(responses["1"])
	stats.Capabilities = capabilityCount
	result["initialize-capabilities"] = observation{"8_CAPABILITIES", capabilities}
	features, featureCount := observeFeatures(responses)
	for id, value := range features { result[id] = value }
	stats.ReadFeatures = featureCount
	diagnostics, paths := observeDiagnostics(messages)
	for id, value := range diagnostics { result[id] = value }
	stats.DiagnosticPaths = paths
	stats.DiagnosticGaps = 3 - paths
	unsupported := responses["9"].Error != nil && responses["9"].Error.Code == -32601
	result["unsupported-method"] = observation{"METHOD_NOT_FOUND", unsupported}
	shutdown := responses["10"].Error == nil && string(responses["10"].Result) == "null"
	result["shutdown-exit"] = observation{"SHUTDOWN_NULL_EXIT_SILENT", shutdown}
	result["server-lifecycle"] = observation{"13_STANDARD_MESSAGES", len(messages) == 13}
	replayEqual := bytes.Equal(raw, replay)
	result["transcript-replay"] = observation{"BYTE_EQUAL", replayEqual}
	if replayEqual { stats.TranscriptReplays = 1 }
	utf := observeUTF16()
	result["utf16-roundtrip"] = observation{"UTF16_EXACT", utf}
	if utf { stats.UTF16Replays = 1 }
	stats.SymbolPaths = boolCount(result["document-symbol"].Satisfied, result["workspace-symbol"].Satisfied)
	stats.SemanticTokenPaths = boolCount(result["semantic-tokens"].Satisfied)
	stats.NavigationPaths = boolCount(result["definition"].Satisfied, result["references"].Satisfied)
	stats.FailClosedPaths = boolCount(unsupported)
	for id, response := range responses { if response.Error != nil && id != "9" { stats.UnexpectedProtocolErrors++ } }
	return result, stats
}

func observeUTF16() bool {
	source := "a😀b\r\nΩ"
	position, err := lsp.OffsetToPosition(source, len("a😀"))
	if err != nil || position.Character != 3 { return false }
	offset, err := lsp.PositionToOffset(source, position)
	return err == nil && offset == len("a😀")
}

func boolCount(values ...bool) int { count := 0; for _, value := range values { if value { count++ } }; return count }

var _ = json.RawMessage{}
var _ = strings.Contains
