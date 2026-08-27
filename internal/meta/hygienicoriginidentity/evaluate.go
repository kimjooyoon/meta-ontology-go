package hygienicoriginidentity

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"strings"
)

var requiredEntities = []string{
	"OriginIdentity", "ScopeProvenance", "GeneratedName", "ConsumerBinding",
	"CapturedResult", "HygienicResult", "CapturedOriginClaim", "CapturedScopeClaim",
	"HygienicOriginClaim", "HygienicScopeClaim", "ProofReceipt",
}

var requiredActivities = []string{
	"ProduceSameSpelling", "ConsumeName", "ObserveCapturedResult",
	"ObserveHygienicResult", "PreserveOriginIdentity", "PreserveScopeProvenance",
	"EmitProofReceipt",
}

var caseOrder = []string{"captured", "hygienic"}

// Evaluate is an independent, read-only judge. It intentionally does not use
// the Gooo parser: only the source bytes, the closed experiment annotations,
// and this fixed identity oracle determine the receipt.
func Evaluate(files fs.FS, sourcePath, headSHA string) (Report, error) {
	raw, err := fs.ReadFile(files, sourcePath)
	if err != nil {
		return Report{}, fmt.Errorf("read source: %w", err)
	}
	document, err := parseSource(raw)
	if err != nil {
		return Report{}, err
	}
	report := Report{
		SchemaVersion: SchemaVersion,
		Decision:      DecisionPass,
		Resolution:    ResolutionExact,
		Producer:      Producer,
		Consumer:      Consumer,
		MetaOperation: MetaOperation,
		ProofChoice:   ProofChoice,
		Source: Subject{
			Path:    sourcePath,
			HeadSHA: headSHA,
			Digest:  digestBytes(raw),
		},
		Cases:     make([]Case, 0, ExpectedCaseTotal),
		Claims:    make([]Claim, 0, ExpectedClaimTotal),
		Unknowns:  append([]Unknown(nil), document.Unknowns...),
		Authority: Authority{},
	}
	for _, id := range caseOrder {
		item := document.Cases[id]
		observed := observeCase(item)
		report.Cases = append(report.Cases, observed)
		report.Claims = append(report.Claims, claimsFor(observed)...)
	}
	report.Metrics = measure(report.Cases, report.Claims, report.Unknowns)
	if len(report.Unknowns) > 0 {
		report.Decision = DecisionUnknown
		report.Resolution = ResolutionLower
	}
	return seal(report), nil
}

func parseSource(raw []byte) (sourceDocument, error) {
	document := sourceDocument{
		Entities:   map[string]bool{},
		Activities: map[string]bool{},
		Cases:      map[string]sourceCase{},
		Unknowns:   []Unknown{},
	}
	scanner := bufio.NewScanner(strings.NewReader(string(raw)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		switch {
		case strings.HasPrefix(line, "package "):
			document.Package = strings.TrimSpace(strings.TrimPrefix(line, "package "))
		case strings.HasPrefix(line, "namespace "):
			document.Namespace = strings.TrimSpace(strings.TrimPrefix(line, "namespace "))
		case strings.HasPrefix(line, "entity "):
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				document.Entities[fields[1]] = true
			}
		case strings.HasPrefix(line, "activity "):
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				name := fields[1]
				if index := strings.IndexByte(name, '('); index >= 0 {
					name = name[:index]
				}
				document.Activities[name] = true
			}
		case strings.HasPrefix(line, "# experiment.case "):
			item, err := parseCaseAnnotation(strings.TrimPrefix(line, "# experiment.case "))
			if err != nil {
				return sourceDocument{}, err
			}
			if _, exists := document.Cases[item.ID]; exists {
				return sourceDocument{}, fmt.Errorf("duplicate experiment case %q", item.ID)
			}
			document.Cases[item.ID] = item
		case strings.HasPrefix(line, "# experiment.unknown "):
			unknown, err := parseUnknownAnnotation(strings.TrimPrefix(line, "# experiment.unknown "))
			if err != nil {
				return sourceDocument{}, err
			}
			document.Unknowns = append(document.Unknowns, unknown)
		}
	}
	if err := scanner.Err(); err != nil {
		return sourceDocument{}, fmt.Errorf("scan source: %w", err)
	}
	if document.Package != "hygienicoriginidentity" || document.Namespace != document.Package {
		return sourceDocument{}, fmt.Errorf("source package/namespace is not hygienicoriginidentity")
	}
	for _, name := range requiredEntities {
		if !document.Entities[name] {
			return sourceDocument{}, fmt.Errorf("source is missing entity %q", name)
		}
	}
	for _, name := range requiredActivities {
		if !document.Activities[name] {
			return sourceDocument{}, fmt.Errorf("source is missing activity %q", name)
		}
	}
	for _, id := range caseOrder {
		if _, ok := document.Cases[id]; !ok {
			return sourceDocument{}, fmt.Errorf("source is missing experiment case %q", id)
		}
	}
	if len(document.Cases) != ExpectedCaseTotal {
		return sourceDocument{}, fmt.Errorf("source case denominator changed: got %d want %d", len(document.Cases), ExpectedCaseTotal)
	}
	return document, nil
}

func parseCaseAnnotation(value string) (sourceCase, error) {
	fields, err := keyValues(value, []string{"id", "spelling", "origin", "scope", "resolves", "expected"})
	if err != nil {
		return sourceCase{}, fmt.Errorf("invalid experiment case: %w", err)
	}
	return sourceCase{ID: fields["id"], Spelling: fields["spelling"], Origin: fields["origin"], Scope: fields["scope"], Resolves: fields["resolves"], Expected: fields["expected"]}, nil
}

func parseUnknownAnnotation(value string) (Unknown, error) {
	fields, err := keyValues(value, []string{"stage", "step", "reason"})
	if err != nil {
		return Unknown{}, fmt.Errorf("invalid experiment unknown: %w", err)
	}
	return Unknown{Stage: fields["stage"], Step: fields["step"], Reason: fields["reason"]}, nil
}

func keyValues(value string, required []string) (map[string]string, error) {
	fields := map[string]string{}
	for _, token := range strings.Fields(value) {
		pair := strings.SplitN(token, "=", 2)
		if len(pair) != 2 || pair[0] == "" || pair[1] == "" {
			return nil, fmt.Errorf("malformed key/value %q", token)
		}
		if _, exists := fields[pair[0]]; exists {
			return nil, fmt.Errorf("duplicate key %q", pair[0])
		}
		fields[pair[0]] = pair[1]
	}
	for _, key := range required {
		if fields[key] == "" {
			return nil, fmt.Errorf("missing key %q", key)
		}
	}
	return fields, nil
}

func observeCase(item sourceCase) Case {
	origin, scope, resolved := identityValues(item)
	originPreserved := origin == ProducerExpansion
	scopePreserved := scope == FreshProducerScope && resolved == origin
	captured := resolved == ConsumerBinding
	return Case{
		ID:                       item.ID,
		Label:                    labelFor(item.ID),
		Spelling:                 item.Spelling,
		OriginIdentity:           origin,
		ScopeProvenance:          scope,
		ResolvedIdentity:         resolved,
		SameSpelling:             item.Spelling == "tmp",
		Captured:                 captured,
		OriginIdentityPreserved:  originPreserved,
		ScopeProvenancePreserved: scopePreserved,
		ClaimIDs:                 []string{item.ID + ".origin-identity", item.ID + ".scope-provenance"},
	}
}

func identityValues(item sourceCase) (string, string, string) {
	identity := func(value string) string {
		switch value {
		case CapturedOrigin:
			return ConsumerBinding
		case HygienicOrigin:
			return ProducerExpansion
		default:
			return value
		}
	}
	scope := func(value string) string {
		switch value {
		case CapturedScope:
			return ConsumerCallSite
		case HygienicScope:
			return FreshProducerScope
		default:
			return value
		}
	}
	return identity(item.Origin), scope(item.Scope), identity(item.Resolves)
}

func claimsFor(item Case) []Claim {
	originStatus := StatusRefuted
	if item.OriginIdentityPreserved {
		originStatus = StatusDischarged
	}
	scopeStatus := StatusRefuted
	if item.ScopeProvenancePreserved {
		scopeStatus = StatusDischarged
	}
	return []Claim{
		{ID: item.ID + ".origin-identity", CaseID: item.ID, Proposition: "generated output preserves producer origin identity", Status: originStatus},
		{ID: item.ID + ".scope-provenance", CaseID: item.ID, Proposition: "generated output preserves non-capturing scope provenance", Status: scopeStatus},
	}
}

func measure(cases []Case, claims []Claim, unknowns []Unknown) Metrics {
	metrics := Metrics{
		FixedCaseDenominator:  ExpectedCaseTotal,
		FixedClaimDenominator: ExpectedClaimTotal,
		ObservedCaseTotal:     len(cases),
		ObservedClaimTotal:    len(claims),
		UnknownPathTotal:      len(unknowns),
	}
	for _, item := range cases {
		if item.SameSpelling {
			metrics.SameSpellingCaseTotal++
		}
		if item.Captured {
			metrics.CapturedCaseTotal++
		} else {
			metrics.NonCapturedCaseTotal++
		}
	}
	for _, claim := range claims {
		switch claim.Status {
		case StatusDischarged:
			metrics.DischargedClaimTotal++
		case StatusRefuted:
			metrics.RefutedClaimTotal++
		case StatusOpen:
			metrics.OpenClaimTotal++
		}
	}
	metrics.ClassifiedClaimTotal = metrics.DischargedClaimTotal + metrics.RefutedClaimTotal
	metrics.ClassificationCoverageBPS = bps(metrics.ClassifiedClaimTotal, metrics.FixedClaimDenominator)
	metrics.PreservationSatisfactionBPS = bps(metrics.DischargedClaimTotal, metrics.FixedClaimDenominator)
	return metrics
}

func labelFor(id string) string {
	if id == "captured" {
		return "same spelling captured by consumer"
	}
	return "same spelling hygienic non-capture"
}

func bps(value, total int) int {
	if total == 0 {
		return 0
	}
	return value * 10000 / total
}

func seal(report Report) Report {
	report.ReceiptDigest = ""
	report.ReceiptDigest = digestJSON(report)
	return report
}

func digestBytes(value []byte) string {
	sum := sha256.Sum256(value)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func digestJSON(value any) string {
	encoded, _ := json.Marshal(value)
	return digestBytes(encoded)
}
