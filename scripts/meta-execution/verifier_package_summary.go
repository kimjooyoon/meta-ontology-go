package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"
)

const verifierPackageSummarySchema = "gooo/meta-execution-verifier-package-summary-advisory/v1"
const verifierPackageSummarySidecarSuffix = ".verifier-package-summaries.json"
const verifierPackageSummaryModulePrefix = "github.com/kimjooyoon/meta-ontology-go"

const maxVerifierPackageSummaryStdoutBytes = 64 * 1024
const maxVerifierPackageSummaryLineBytes = 1024
const maxVerifierPackageSummaryRowsPerInvocation = 512
const maxVerifierPackageSummaryTotalRows = 2048
const maxVerifierPackageSummaryRecords = 64
const maxVerifierPackageSummaryFileBytes = 256 * 1024
const maxVerifierDurationNanoseconds int64 = 1<<63 - 1
const verifierNanosecondsPerSecond uint64 = 1_000_000_000

const verifierOutputMarkerAuthority = "OBSERVED_OUTPUT_MARKER"

const (
	verifierOutputMarkerCached       = "CACHED"
	verifierOutputMarkerNoTestFiles  = "NO_TEST_FILES"
	verifierOutputMarkerNoTestsToRun = "NO_TESTS_TO_RUN"
	verifierOutputMarkerBuildFailed  = "BUILD_FAILED"
	verifierOutputMarkerSetupFailed  = "SETUP_FAILED"
)

type verifierPackageSummaryDocument struct {
	Schema            string                         `json:"schema"`
	DiagnosticOnly    string                         `json:"diagnostic_only"`
	Authenticity      string                         `json:"authenticity"`
	Improvement       string                         `json:"improvement"`
	DiagnosticMarkers []string                       `json:"diagnostic_markers"`
	Truncated         bool                           `json:"truncated"`
	Records           []verifierPackageSummaryRecord `json:"records"`
}

type verifierPackageSummaryRecord struct {
	InvocationID          string                      `json:"invocation_id"`
	ActionIndicatorID     string                      `json:"action_indicator_id"`
	Activity              string                      `json:"activity"`
	MetaOperation         string                      `json:"meta_operation"`
	Subject               string                      `json:"subject"`
	OperationSequence     int                         `json:"operation_sequence"`
	Pass                  string                      `json:"pass"`
	CommandKind           string                      `json:"command_kind"`
	ExitCode              int                         `json:"exit_code"`
	StdoutBytes           int                         `json:"stdout_bytes"`
	RawStdoutDigest       string                      `json:"raw_stdout_digest"`
	StdoutDigest          string                      `json:"stdout_digest"`
	ParseStatus           string                      `json:"parse_status"`
	DiagnosticMarkers     []string                    `json:"diagnostic_markers"`
	Truncated             bool                        `json:"truncated"`
	UnrecognizedLineCount int                         `json:"unrecognized_line_count"`
	Packages              []verifierPackageSummaryRow `json:"packages"`
}

type verifierPackageSummaryRow struct {
	Package               string `json:"package"`
	Status                string `json:"status"`
	ElapsedToken          string `json:"elapsed_token,omitempty"`
	ElapsedStatus         string `json:"elapsed_status"`
	ElapsedUnit           string `json:"elapsed_unit,omitempty"`
	ElapsedNanoseconds    *int64 `json:"elapsed_nanoseconds,omitempty"`
	OutputMarker          string `json:"output_marker,omitempty"`
	OutputMarkerAuthority string `json:"output_marker_authority,omitempty"`
}

type verifierPackageSummaryParse struct {
	Rows                  []verifierPackageSummaryRow
	Truncated             bool
	UnrecognizedLineCount int
}

type verifierPackageSummaryCollector struct {
	path      string
	records   []verifierPackageSummaryRecord
	totalRows int
	truncated bool
}

func newVerifierPackageSummaryCollector(outputPath string) *verifierPackageSummaryCollector {
	return &verifierPackageSummaryCollector{
		path:    outputPath + verifierPackageSummarySidecarSuffix,
		records: make([]verifierPackageSummaryRecord, 0),
	}
}

func (collector *verifierPackageSummaryCollector) observe(trace metaExecutionTrace, pass string, result processResult) {
	if collector == nil {
		return
	}
	if len(collector.records) >= maxVerifierPackageSummaryRecords {
		collector.truncated = true
		return
	}
	parsed := parseVerifierPackageSummaries(result.Stdout)
	if parsed.Truncated {
		collector.truncated = true
	}
	if remaining := maxVerifierPackageSummaryTotalRows - collector.totalRows; remaining < len(parsed.Rows) {
		if remaining < 0 {
			remaining = 0
		}
		parsed.Rows = parsed.Rows[:remaining]
		parsed.Truncated = true
		collector.truncated = true
	}
	record := verifierPackageSummaryRecord{
		InvocationID:          trace.state.invocationID,
		ActionIndicatorID:     trace.action.IndicatorID,
		Activity:              trace.action.Activity,
		MetaOperation:         string(trace.action.Operation),
		Subject:               trace.action.Subject,
		OperationSequence:     trace.sequence,
		Pass:                  pass,
		CommandKind:           "verifier",
		ExitCode:              result.Observation.ExitCode,
		StdoutBytes:           result.Observation.StdoutBytes,
		RawStdoutDigest:       result.Observation.RawStdoutDigest,
		StdoutDigest:          result.Observation.StdoutDigest,
		ParseStatus:           parsed.status(),
		DiagnosticMarkers:     parsed.markers(),
		Truncated:             parsed.Truncated,
		UnrecognizedLineCount: parsed.UnrecognizedLineCount,
		Packages:              parsed.Rows,
	}
	collector.records = append(collector.records, record)
	collector.totalRows += len(parsed.Rows)
}

func (collector *verifierPackageSummaryCollector) write() error {
	if collector == nil || collector.path == "" {
		return nil
	}
	document := verifierPackageSummaryDocument{
		Schema:            verifierPackageSummarySchema,
		DiagnosticOnly:    "DIAGNOSTIC_ONLY",
		Authenticity:      "AUTHENTICITY_UNVERIFIED",
		Improvement:       "UNKNOWN",
		DiagnosticMarkers: collector.markers(),
		Truncated:         collector.truncated,
		Records:           collector.records,
	}
	payload, err := boundedVerifierPackageSummaryPayload(document)
	if err != nil {
		return err
	}
	if _, err := archivePreviousObservation(collector.path); err != nil {
		return err
	}
	return writeAtomic(collector.path, payload)
}

func (collector *verifierPackageSummaryCollector) markers() []string {
	markers := make([]string, 0, 2)
	if collector.truncated {
		markers = append(markers, "TRUNCATED_OUTPUT")
	}
	if len(collector.records) == 0 {
		markers = append(markers, "NO_VERIFIER_OBSERVATION")
	}
	return markers
}

func boundedVerifierPackageSummaryPayload(document verifierPackageSummaryDocument) ([]byte, error) {
	maxPayloadBytes := maxVerifierPackageSummaryFileBytes - 1
	payload, err := json.MarshalIndent(document, "", "  ")
	if err != nil {
		return nil, err
	}
	if len(payload) <= maxPayloadBytes {
		return append(payload, '\n'), nil
	}

	document.Truncated = true
	document.DiagnosticMarkers = appendVerifierPackageSummaryMarker(document.DiagnosticMarkers, "TRUNCATED_OUTPUT_FILE")
	for index := len(document.Records) - 1; index >= 0 && len(payload) > maxPayloadBytes; index-- {
		if len(document.Records[index].Packages) == 0 {
			continue
		}
		document.Records[index].Packages = nil
		document.Records[index].Truncated = true
		document.Records[index].ParseStatus = "TRUNCATED"
		document.Records[index].DiagnosticMarkers = appendVerifierPackageSummaryMarker(document.Records[index].DiagnosticMarkers, "TRUNCATED_OUTPUT_FILE")
		payload, err = json.MarshalIndent(document, "", "  ")
		if err != nil {
			return nil, err
		}
	}
	if len(payload) > maxPayloadBytes {
		document.Records = nil
		payload, err = json.MarshalIndent(document, "", "  ")
		if err != nil {
			return nil, err
		}
	}
	if len(payload) > maxPayloadBytes {
		return nil, fmt.Errorf("verifier package summary sidecar exceeds bounded size")
	}
	return append(payload, '\n'), nil
}

func appendVerifierPackageSummaryMarker(markers []string, marker string) []string {
	if slices.Contains(markers, marker) {
		return markers
	}
	return append(markers, marker)
}

func (state *metaExecutionTraceState) writeVerifierPackageSummary() error {
	if state == nil || state.verifierPackageSummary == nil {
		return nil
	}
	return state.verifierPackageSummary.write()
}

func (trace metaExecutionTrace) observeVerifierPackageSummary(pass, commandKind string, result processResult) {
	if commandKind != "verifier" || trace.state == nil || trace.state.verifierPackageSummary == nil {
		return
	}
	trace.state.verifierPackageSummary.observe(trace, pass, result)
}

func parseVerifierPackageSummaries(stdout []byte) verifierPackageSummaryParse {
	parsed := verifierPackageSummaryParse{Rows: make([]verifierPackageSummaryRow, 0)}
	input := stdout
	if len(input) > maxVerifierPackageSummaryStdoutBytes {
		input = input[:maxVerifierPackageSummaryStdoutBytes]
		parsed.Truncated = true
	}
	for offset := 0; offset < len(input); {
		if len(parsed.Rows) >= maxVerifierPackageSummaryRowsPerInvocation {
			parsed.Truncated = true
			break
		}
		relativeEnd := bytes.IndexByte(input[offset:], '\n')
		completeLine := relativeEnd >= 0
		var line []byte
		if completeLine {
			line = input[offset : offset+relativeEnd]
			offset += relativeEnd + 1
		} else {
			line = input[offset:]
			offset = len(input)
			if len(stdout) > len(input) {
				parsed.Truncated = true
				break
			}
		}
		if len(line) > maxVerifierPackageSummaryLineBytes {
			parsed.Truncated = true
			continue
		}
		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		row, recognized, unrecognized := parseVerifierPackageSummaryLine(line)
		if !recognized {
			parsed.UnrecognizedLineCount++
			continue
		}
		if unrecognized {
			parsed.UnrecognizedLineCount++
		}
		parsed.Rows = append(parsed.Rows, row)
	}
	return parsed
}

func parseVerifierPackageSummaryLine(line []byte) (verifierPackageSummaryRow, bool, bool) {
	if !utf8.Valid(line) {
		return verifierPackageSummaryRow{}, false, false
	}
	fields := strings.Fields(string(line))
	if len(fields) < 2 {
		return verifierPackageSummaryRow{}, false, false
	}
	status := fields[0]
	if status != "ok" && status != "FAIL" && status != "?" {
		return verifierPackageSummaryRow{}, false, false
	}
	packagePath := fields[1]
	if !validVerifierPackagePath(packagePath) {
		return verifierPackageSummaryRow{}, false, false
	}
	row := verifierPackageSummaryRow{
		Package:       packagePath,
		Status:        status,
		ElapsedStatus: "UNKNOWN",
	}
	if len(fields) == 2 {
		return row, true, false
	}
	tail := fields[2:]
	if len(tail) == 1 && tail[0] == "(cached)" {
		row.OutputMarker = verifierOutputMarkerCached
		row.OutputMarkerAuthority = verifierOutputMarkerAuthority
		return row, true, false
	}
	if len(tail) == 1 {
		if nanoseconds, ok := exactVerifierDuration(tail[0]); ok {
			row.ElapsedToken = tail[0]
			row.ElapsedStatus = "PARSED_EXACTLY"
			row.ElapsedUnit = "NANOSECONDS"
			row.ElapsedNanoseconds = &nanoseconds
			return row, true, false
		}
		if marker, ok := verifierOutputMarkerKind(tail); ok {
			row.OutputMarker = marker
			row.OutputMarkerAuthority = verifierOutputMarkerAuthority
			return row, true, false
		}
	}
	if len(tail) > 1 {
		if nanoseconds, ok := exactVerifierDuration(tail[0]); ok {
			if marker, markerOK := verifierOutputMarkerKind(tail[1:]); markerOK {
				row.ElapsedToken = tail[0]
				row.ElapsedStatus = "PARSED_EXACTLY"
				row.ElapsedUnit = "NANOSECONDS"
				row.ElapsedNanoseconds = &nanoseconds
				row.OutputMarker = marker
				row.OutputMarkerAuthority = verifierOutputMarkerAuthority
				return row, true, false
			}
		}
		if marker, ok := verifierOutputMarkerKind(tail); ok {
			row.OutputMarker = marker
			row.OutputMarkerAuthority = verifierOutputMarkerAuthority
			return row, true, false
		}
	}
	return row, true, true
}

func validVerifierPackagePath(packagePath string) bool {
	if packagePath != verifierPackageSummaryModulePrefix && !strings.HasPrefix(packagePath, verifierPackageSummaryModulePrefix+"/") {
		return false
	}
	if strings.ContainsAny(packagePath, "\\\t\r\n") || strings.Contains(packagePath, "//") || path.Clean(packagePath) != packagePath {
		return false
	}
	for segment := range strings.SplitSeq(packagePath, "/") {
		if segment == "" || segment == "." || segment == ".." {
			return false
		}
	}
	return true
}

func exactVerifierDuration(token string) (int64, bool) {
	if token == "" || token[len(token)-1] != 's' {
		return 0, false
	}
	number := token[:len(token)-1]
	if number == "" {
		return 0, false
	}
	whole := number
	fraction := ""
	if before, after, ok := strings.Cut(number, "."); ok {
		if strings.IndexByte(after, '.') >= 0 {
			return 0, false
		}
		whole = before
		fraction = after
		if whole == "" || fraction == "" || len(fraction) > 9 {
			return 0, false
		}
	}
	for _, digit := range whole {
		if digit < '0' || digit > '9' {
			return 0, false
		}
	}
	for _, digit := range fraction {
		if digit < '0' || digit > '9' {
			return 0, false
		}
	}
	seconds, err := strconv.ParseUint(whole, 10, 64)
	if err != nil || seconds > uint64(maxVerifierDurationNanoseconds)/verifierNanosecondsPerSecond {
		return 0, false
	}
	fractionValue := uint64(0)
	if fraction != "" {
		fractionValue, err = strconv.ParseUint(fraction, 10, 64)
		if err != nil {
			return 0, false
		}
		for digits := len(fraction); digits < 9; digits++ {
			fractionValue *= 10
		}
	}
	wholeNanoseconds := seconds * verifierNanosecondsPerSecond
	maxNanoseconds := uint64(maxVerifierDurationNanoseconds)
	if wholeNanoseconds > maxNanoseconds-fractionValue {
		return 0, false
	}
	return int64(wholeNanoseconds + fractionValue), true
}

func verifierOutputMarkerKind(tokens []string) (string, bool) {
	joined := strings.Join(tokens, " ")
	switch joined {
	case "[no test files]":
		return verifierOutputMarkerNoTestFiles, true
	case "[no tests to run]":
		return verifierOutputMarkerNoTestsToRun, true
	case "[build failed]":
		return verifierOutputMarkerBuildFailed, true
	case "[setup failed]":
		return verifierOutputMarkerSetupFailed, true
	default:
		return "", false
	}
}

func (parsed verifierPackageSummaryParse) status() string {
	if parsed.Truncated {
		return "TRUNCATED"
	}
	if parsed.UnrecognizedLineCount > 0 {
		return "UNRECOGNIZED_INPUT"
	}
	if len(parsed.Rows) == 0 {
		return "NO_PACKAGE_SUMMARIES"
	}
	return "COMPLETE"
}

func (parsed verifierPackageSummaryParse) markers() []string {
	markers := make([]string, 0, 2)
	if parsed.Truncated {
		markers = append(markers, "TRUNCATED_INPUT")
	}
	if parsed.UnrecognizedLineCount > 0 {
		markers = append(markers, "UNRECOGNIZED_INPUT")
	}
	if len(parsed.Rows) == 0 {
		markers = append(markers, "NO_PACKAGE_SUMMARIES")
	}
	return markers
}
