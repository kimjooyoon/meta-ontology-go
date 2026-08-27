package audienceresolutionconsumer

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const (
	reportSchema     = "gooo/audience-resolution-consumer/v1"
	ledgerSchema     = "gooo/audience-resolution-ledger/v1"
	sourceKind       = "gooo"
	policyPrefix     = "gooo://audience-resolution/policy/"
	resolutionPrefix = "gooo://audience-resolution/resolution/"
	claimPrefix      = "gooo://audience-resolution/claim-state/"
	relationPrefix   = "gooo://audience-resolution/relation/"
)

type audiencePolicy struct {
	Audience    string
	Resolution  string
	Coordinates []string
}

type sourceModel struct {
	Digest      string
	Denominator int
	Audiences   []audiencePolicy
}

type spec struct {
	Producer      string
	Consumer      string
	MetaOperation string
	ProofChoice   string
	Stage         string
	Step          string
	Reason        string
}

func fixedSpecs() map[string]spec {
	return map[string]spec{
		"source.binding":               {"gooo://audience-resolution/source", "audience-resolution.projector", "bind-source-ledger", "FOUNDATION", "source", "bind", "source declaration is bound as authority"},
		"ledger.coverage":              {"audience-resolution.producer", "audience-resolution.projector", "consume-canonical-ledger", "FOUNDATION", "ledger", "coverage", "all fixed coordinates are present once"},
		"ledger.replay":                {"audience-resolution.projector", "audience-resolution.receipt", "replay-canonical-ledger", "REGRESSION", "ledger", "replay", "the same ledger replays byte-equivalently"},
		"user.coordinates":             {"audience-resolution.projector", "USER", "project-user-coordinate-set", "FOUNDATION", "projection", "user", "USER receives exactly its four coordinates"},
		"author.coordinates":           {"audience-resolution.projector", "TOOL_AUTHOR", "project-tool-author-coordinate-set", "FOUNDATION", "projection", "tool-author", "TOOL_AUTHOR receives the authoring contract"},
		"governor.coordinates":         {"audience-resolution.projector", "GOVERNOR", "project-governor-coordinate-set", "FOUNDATION", "projection", "governor", "GOVERNOR receives the full ledger"},
		"projection.nesting":           {"audience-resolution.projector", "audience-resolution.governor", "check-coordinate-nesting", "COHERENCE", "projection", "nest", "higher resolution extends the lower coordinate set"},
		"projection.shared-decision":   {"audience-resolution.receipt", "all-audiences", "lift-global-decision", "COHERENCE", "projection", "decision", "global decision is carried with local verification status"},
		"projection.resolution":        {"audience-resolution.projector", "all-audiences", "preserve-projection-resolution", "COHERENCE", "projection", "resolution", "each audience keeps its fixed resolution label"},
		"counterexample.omission":      {"audience-resolution.projector", "GOVERNOR", "execute-omitted-coordinate", "REGRESSION", "counterexample", "missing", "missing information cannot produce PASS"},
		"counterexample.contradiction": {"audience-resolution.projector", "GOVERNOR", "execute-contradictory-observation", "REGRESSION", "counterexample", "contradiction", "contradictory evidence becomes REFUTED"},
		"receipt.seal":                 {"audience-resolution.receipt", "independent.checker", "validate-receipt-digest", "REGRESSION", "receipt", "seal", "receipt is independently checkable"},
	}
}

func Check(input Input) Report {
	ledgerBytes := input.LedgerBytes
	if len(ledgerBytes) == 0 {
		ledgerBytes = inputSourceJSON(input.Ledger)
	}
	report := Report{Schema: reportSchema, RawLedgerFinalFieldsAbsent: rawLedgerHasNoFinalFields(input.Ledger, ledgerBytes), ReceiptDigestMatch: receiptDigestMatches(input.ReceiptBytes, input.Receipt.Digest)}
	model, err := reconstruct(input.SourcePath, input.Source)
	if err != nil {
		return sealReport(withIssue(report, "SOURCE_RECONSTRUCTION_UNAVAILABLE"))
	}
	report.SourceReconstruction = SourceReconstruction{ParsedAndLowered: true, DeclarationCount: model.Denominator,
		SemanticDigest: model.Digest, CanonicalIRDigest: model.Digest,
		ReceiptMatches: input.Receipt.Source.SemanticDigest == model.Digest && input.Receipt.Source.DeclarationCount == model.Denominator && input.Receipt.Source.Reconstructed}
	report.ProducerImports = auditProducerImports(input.RepoRoot)
	report.Audiences = expectedAudienceChecks(model, input.Ledger, input.Receipt)
	report.ClaimTransitionsChecked = checkTransitions(model, input.Ledger, input.Receipt)
	issues := []string{}
	if input.Ledger.Schema != ledgerSchema || input.Ledger.Subject == "" || input.Ledger.Source.Kind != sourceKind || input.Ledger.Source.Digest != rawDigest(input.Source) {
		issues = append(issues, "RAW_SOURCE_BINDING_INVALID")
	}
	if input.Ledger.Source.SemanticDigest != "" && input.Ledger.Source.SemanticDigest != model.Digest {
		issues = append(issues, "RAW_SEMANTIC_DIGEST_INVALID")
	}
	if !report.RawLedgerFinalFieldsAbsent {
		issues = append(issues, "RAW_LEDGER_CONTAINS_FINAL_FIELD")
	}
	if !report.ReceiptDigestMatch || !report.SourceReconstruction.ReceiptMatches {
		issues = append(issues, "RECEIPT_SOURCE_RECONSTRUCTION_MISMATCH")
	}
	if report.ProducerImports.Numerator != 0 {
		issues = append(issues, "CONSUMER_IMPORTS_PRODUCER")
	}
	if len(report.Audiences) != 3 || report.Audiences[2].Decision == "" {
		issues = append(issues, "AUDIENCE_PROJECTION_UNAVAILABLE")
	}
	if input.Receipt.Decision != expectedGlobal(model, input.Ledger) {
		issues = append(issues, "GLOBAL_DECISION_RECONSTRUCTION_MISMATCH")
	}
	for index, check := range report.Audiences {
		if index < len(input.Receipt.Views) && (check.Decision != input.Receipt.Views[index].LocalDecision || check.Visible != input.Receipt.Views[index].Visible || check.Required != input.Receipt.Views[index].Required) {
			issues = append(issues, "AUDIENCE_LOCAL_DECISION_MISMATCH")
		}
	}
	if len(input.Receipt.ClaimTransitions) != report.ClaimTransitionsChecked {
		issues = append(issues, "CLAIM_TRANSITION_COUNT_MISMATCH")
	}
	if len(issues) == 0 {
		report.Decision, report.Reason = "PASS", "INDEPENDENT_RAW_RECONSTRUCTION_MATCHED"
	} else {
		report.Decision, report.Reason = "REFUTED", strings.Join(issues, ";")
	}
	return sealReport(report)
}

func reconstruct(filename string, source []byte) (sourceModel, error) {
	file, diagnostics := syntax.ParseFile(filename, string(source))
	if file == nil || diagnostics.HasErrors() {
		return sourceModel{}, fmt.Errorf("parse diagnostics: %v", diagnostics.Error())
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return sourceModel{}, err
	}
	policies := map[string][]struct {
		ordinal    int
		coordinate string
	}{}
	resolutions := map[string]string{}
	claimStates := map[string]bool{}
	relationPresent := false
	for _, node := range ir.Graph.Nodes() {
		id := node.ID.String()
		if strings.HasPrefix(id, policyPrefix) {
			parts := strings.Split(strings.TrimPrefix(id, policyPrefix), "/")
			if len(parts) < 3 {
				return sourceModel{}, fmt.Errorf("incomplete policy identity")
			}
			ordinal, err := strconv.Atoi(parts[1])
			if err != nil {
				return sourceModel{}, err
			}
			policies[parts[0]] = append(policies[parts[0]], struct {
				ordinal    int
				coordinate string
			}{ordinal, strings.Join(parts[2:], "/")})
		}
		if strings.HasPrefix(id, resolutionPrefix) {
			parts := strings.Split(strings.TrimPrefix(id, resolutionPrefix), "/")
			if len(parts) < 2 {
				return sourceModel{}, fmt.Errorf("incomplete resolution identity")
			}
			resolutions[parts[0]] = strings.Join(parts[1:], "/")
		}
		if strings.HasPrefix(id, claimPrefix) {
			claimStates[strings.TrimPrefix(id, claimPrefix)] = true
		}
		if id == relationPrefix+"evidence-to-claim" {
			relationPresent = true
		}
	}
	model := sourceModel{Digest: digestBytes([]byte(ir.SemanticCanonical())), Denominator: len(ir.Graph.Nodes())}
	for _, audience := range []string{"USER", "TOOL_AUTHOR", "GOVERNOR"} {
		values := policies[audience]
		sort.Slice(values, func(i, j int) bool { return values[i].ordinal < values[j].ordinal })
		if len(values) == 0 || resolutions[audience] == "" {
			return sourceModel{}, fmt.Errorf("audience policy incomplete")
		}
		coordinates := make([]string, 0, len(values))
		for _, value := range values {
			coordinates = append(coordinates, value.coordinate)
		}
		model.Audiences = append(model.Audiences, audiencePolicy{Audience: audience, Resolution: resolutions[audience], Coordinates: coordinates})
	}
	if !nested(model.Audiences) || !claimStates["OPEN"] || !claimStates["DISCHARGED"] || !claimStates["REFUTED"] || !relationPresent {
		return sourceModel{}, fmt.Errorf("audience policy is not nested")
	}
	return model, nil
}

func expectedGlobal(model sourceModel, ledger RawLedger) string {
	records, duplicate := recordMap(ledger.Records)
	if duplicate {
		return "REFUTED"
	}
	for _, record := range ledger.Records {
		if record.Observation == "CONTRADICTORY" {
			return "REFUTED"
		}
	}
	governor := model.Audiences[2].Coordinates
	for _, coordinate := range governor {
		if !validCoordinate(records, coordinate) {
			return "UNKNOWN"
		}
	}
	if len(records) != len(governor) {
		return "UNKNOWN"
	}
	return "PASS"
}

func expectedAudienceChecks(model sourceModel, ledger RawLedger, receipt Receipt) []AudienceCheck {
	records, _ := recordMap(ledger.Records)
	checks := make([]AudienceCheck, 0, len(model.Audiences))
	required := len(model.Audiences[2].Coordinates)
	for _, audience := range model.Audiences {
		visible, satisfied := 0, 0
		contradiction := ""
		for _, coordinate := range audience.Coordinates {
			if _, ok := records[coordinate]; ok {
				visible++
			}
			if validCoordinate(records, coordinate) {
				satisfied++
			}
			if record, ok := records[coordinate]; ok && record.Observation == "CONTRADICTORY" && contradiction == "" {
				contradiction = coordinate
			}
		}
		localDecision, resolution := "PASS", "EXACT"
		claimTransition := "DISCHARGED"
		if contradiction != "" {
			localDecision, resolution, claimTransition = "REFUTED", "INVARIANT_ONLY", "REFUTED"
		} else if satisfied != len(audience.Coordinates) || len(audience.Coordinates) < required {
			localDecision, resolution, claimTransition = "UNKNOWN", "LOWER_RESOLUTION", "OPEN"
		}
		checks = append(checks, AudienceCheck{Audience: audience.Audience, Visible: visible, Required: required,
			Decision: localDecision, Resolution: resolution, ClaimTransitions: claimTransition})
	}
	_ = receipt
	return checks
}

func checkTransitions(model sourceModel, ledger RawLedger, receipt Receipt) int {
	records, _ := recordMap(ledger.Records)
	checked := 0
	index := 0
	for _, audience := range model.Audiences {
		for _, coordinate := range model.Audiences[2].Coordinates {
			if index >= len(receipt.ClaimTransitions) {
				return checked
			}
			transition := receipt.ClaimTransitions[index]
			checked++
			if transition.IndicatorID != coordinate || transition.Audience != audience.Audience || transition.Before != "OPEN" || transition.EvidenceDigest == "" {
				return checked
			}
			if !contains(audience.Coordinates, coordinate) {
				if transition.Visibility != "OMITTED" || transition.After != "OPEN" {
					return checked
				}
			} else if record, ok := records[coordinate]; !ok {
				if transition.Visibility != "OMITTED" || transition.After != "OPEN" {
					return checked
				}
			} else {
				expected := "OPEN"
				if record.Observation == "CONTRADICTORY" {
					expected = "REFUTED"
				} else if validCoordinate(records, coordinate) {
					expected = "DISCHARGED"
				}
				if transition.Visibility != "VISIBLE" || transition.After != expected {
					return checked
				}
			}
			index++
		}
	}
	return checked
}

func validCoordinate(records map[string]RawRecord, coordinate string) bool {
	record, ok := records[coordinate]
	if !ok || record.ID != coordinate || record.Coordinate != coordinate || record.PriorClaim != "OPEN" || record.Observation != "OBSERVED" {
		return false
	}
	if expected, ok := fixedSpecs()[coordinate]; ok {
		return record.Producer == expected.Producer && record.Consumer == expected.Consumer && record.MetaOperation == expected.MetaOperation && record.ProofChoice == expected.ProofChoice && record.Stage == expected.Stage && record.Step == expected.Step && record.Reason == expected.Reason
	}
	return record.Audience != "" && record.Stage != "" && record.Step != "" && record.Reason != "" && record.Producer != "" && record.Consumer != "" && record.MetaOperation != "" && record.ProofChoice != ""
}

func recordMap(records []RawRecord) (map[string]RawRecord, bool) {
	result := map[string]RawRecord{}
	duplicate := false
	for _, record := range records {
		if _, exists := result[record.Coordinate]; exists {
			duplicate = true
		}
		result[record.Coordinate] = record
	}
	return result, duplicate
}

func nested(values []audiencePolicy) bool {
	if len(values) != 3 {
		return false
	}
	for index := 1; index < len(values); index++ {
		if len(values[index].Coordinates) <= len(values[index-1].Coordinates) {
			return false
		}
		for coordinateIndex, coordinate := range values[index-1].Coordinates {
			if values[index].Coordinates[coordinateIndex] != coordinate {
				return false
			}
		}
	}
	return true
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func rawLedgerHasNoFinalFields(ledger RawLedger, raw []byte) bool {
	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		return false
	}
	forbidden := map[string]bool{"decision": true, "satisfied": true, "claim_after": true, "blocked": true, "expected_decision": true, "observed_decision": true}
	var visit func(any) bool
	visit = func(current any) bool {
		switch value := current.(type) {
		case map[string]any:
			for key, child := range value {
				if forbidden[key] || !visit(child) {
					return false
				}
			}
		case []any:
			for _, child := range value {
				if !visit(child) {
					return false
				}
			}
		}
		return true
	}
	return visit(value)
}

func inputSourceJSON(ledger RawLedger) []byte {
	value, _ := json.Marshal(ledger)
	return value
}

func auditProducerImports(root string) ImportAudit {
	directory := filepath.Join(root, "internal", "meta", "audienceresolutionconsumer")
	entries, err := os.ReadDir(directory)
	if err != nil {
		return ImportAudit{Numerator: 1, Denominator: 1, Forbidden: []string{"consumer package unavailable"}}
	}
	audit := ImportAudit{}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") {
			continue
		}
		path := filepath.Join(directory, entry.Name())
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			audit.Numerator++
			audit.Forbidden = append(audit.Forbidden, entry.Name()+":parse")
			continue
		}
		audit.Denominator += len(file.Imports)
		for _, imported := range file.Imports {
			pathValue, err := strconv.Unquote(imported.Path.Value)
			if err == nil && (pathValue == "github.com/kimjooyoon/meta-ontology-go/internal/meta/audienceresolution" || pathValue == "github.com/kimjooyoon/meta-ontology-go/cmd/audience-resolution-witness") {
				audit.Numerator++
				audit.Forbidden = append(audit.Forbidden, pathValue)
			}
		}
		if fileHasForbiddenSymbol(file) {
			audit.Numerator++
			audit.Forbidden = append(audit.Forbidden, entry.Name()+":forbidden-symbol")
		}
	}
	sort.Strings(audit.Forbidden)
	return audit
}

func fileHasForbiddenSymbol(file *ast.File) bool {
	for _, declaration := range file.Decls {
		if function, ok := declaration.(*ast.FuncDecl); ok && function.Body != nil {
			found := false
			ast.Inspect(function.Body, func(node ast.Node) bool {
				identifier, ok := node.(*ast.Ident)
				if ok && (identifier.Name == "CanonicalContract" || identifier.Name == "ValidateReceipt") {
					found = true
				}
				return !found
			})
			if found {
				return true
			}
		}
	}
	return false
}

func receiptDigestMatches(raw []byte, expected string) bool {
	if expected == "" || len(raw) == 0 {
		return false
	}
	var value map[string]any
	if json.Unmarshal(raw, &value) != nil {
		return false
	}
	delete(value, "digest")
	return digestBytes(canonicalJSON(value)) == expected
}

func rawDigest(value []byte) string { return digestBytes(value) }

func digestBytes(value []byte) string {
	sum := sha256.Sum256(value)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func canonicalJSON(value any) []byte {
	raw, _ := json.Marshal(value)
	var normalized any
	if json.Unmarshal(raw, &normalized) != nil {
		return raw
	}
	result, _ := json.Marshal(normalized)
	return result
}

func sealReport(report Report) Report {
	report.Digest = ""
	report.Digest = digestBytes(canonicalJSON(report))
	return report
}

func withIssue(report Report, reason string) Report {
	report.Decision, report.Reason = "REFUTED", reason
	return report
}
