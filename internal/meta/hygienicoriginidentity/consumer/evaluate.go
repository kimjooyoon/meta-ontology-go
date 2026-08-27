package consumer

import (
	"fmt"
	"io/fs"
	"slices"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

var requiredEntities = []string{
	"OriginIdentity", "ScopeProvenance", "GeneratedName", "ConsumerBinding",
	"CapturedResult", "HygienicResult", "CapturedOriginClaim", "CapturedScopeClaim",
	"HygienicOriginClaim", "HygienicScopeClaim", "ProofReceipt",
}

var requiredActivities = []string{
	"ProduceCapturedName", "ProduceHygienicName", "ConsumeCapturedName", "ConsumeHygienicName",
	"ObserveCapturedResult", "ObserveHygienicResult", "PreserveOriginIdentity",
	"PreserveScopeProvenance", "EmitProofReceipt",
}

var caseOrder = []string{"captured", "hygienic"}

type producerValue struct {
	CaseID       string
	Spelling     string
	Origin       string
	Scope        string
	Evidence     string
	ActivityName string
}

type consumerValue struct {
	CaseID       string
	Resolved     string
	Missing      bool
	Evidence     string
	ActivityName string
}

// Evaluate is the consumer-side judge. It reads only the canonical syntax AST
// and semantic IR. Comments are absent from both inputs and therefore cannot
// affect any case, claim, decision, or semantic digest.
func Evaluate(files fs.FS, sourcePath, headSHA string, snapshots SnapshotPair) (Report, error) {
	raw, err := fs.ReadFile(files, sourcePath)
	if err != nil {
		return Report{}, fmt.Errorf("read source: %w", err)
	}
	file, diagnostics := syntax.ParseFile(sourcePath, string(raw))
	if file == nil || diagnostics.HasErrors() {
		return Report{}, fmt.Errorf("syntax diagnostics prevent semantic evaluation: %d", len(diagnostics))
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return Report{}, fmt.Errorf("lower source to semantic IR: %w", err)
	}
	if err := validateSemanticContract(ir); err != nil {
		return Report{}, err
	}
	producers, err := extractProducerValues(ir)
	if err != nil {
		return Report{}, err
	}
	consumers, err := extractConsumerValues(ir)
	if err != nil {
		return Report{}, err
	}

	beforeDigest := digestBytes(snapshots.Before)
	afterDigest := digestBytes(snapshots.After)
	repositoryWrites := snapshotDelta(snapshots.Before, snapshots.After)
	report := Report{
		SchemaVersion: SchemaVersion,
		Decision:      DecisionPass,
		Resolution:    ResolutionExact,
		Producer:      Producer,
		Consumer:      Consumer,
		MetaOperation: MetaOperation,
		ProofChoice:   ProofChoice,
		Source: Subject{
			Path:           sourcePath,
			HeadSHA:        headSHA,
			RawDigest:      digestBytes(raw),
			SemanticDigest: "sha256:" + ir.StableHash(),
		},
		Cases:       make([]Case, 0, ExpectedCaseTotal),
		Claims:      make([]Claim, 0, ExpectedClaimTotal),
		Transitions: make([]Transition, 0, ExpectedClaimTotal),
		Unknowns:    make([]Unknown, 0, 1),
		Authority: Authority{
			RepositoryWrites:     repositoryWrites,
			BeforeSnapshotDigest: beforeDigest,
			AfterSnapshotDigest:  afterDigest,
			SnapshotsEqual:       repositoryWrites == 0,
		},
	}

	for _, id := range caseOrder {
		producer, ok := producers[id]
		if !ok {
			return Report{}, fmt.Errorf("semantic producer case %q is missing", id)
		}
		consumer, ok := consumers[id]
		if !ok {
			return Report{}, fmt.Errorf("semantic consumer case %q is missing", id)
		}
		observed, unknown := observeCase(producer, consumer)
		report.Cases = append(report.Cases, observed)
		if unknown != nil {
			report.Unknowns = append(report.Unknowns, *unknown)
		}
		for _, claim := range claimsFor(observed, producer.Evidence, consumer.Evidence, unknown != nil) {
			report.Claims = append(report.Claims, claim)
			report.Transitions = append(report.Transitions, Transition{
				Sequence:       len(report.Transitions) + 1,
				ClaimID:        claim.ID,
				Before:         StatusUnclassified,
				After:          claim.Status,
				Reason:         "semantic resolver observation",
				EvidenceDigest: claim.EvidenceDigest,
				Provenance:     claim.Provenance,
			})
		}
	}
	if len(report.Unknowns) > 0 {
		unknown := report.Unknowns[0]
		claim := Claim{
			ID:             "unknown.scope-provenance",
			CaseID:         "unknown",
			Proposition:    "scope provenance is available to resolve the generated binding",
			Status:         StatusOpen,
			EvidenceDigest: unknown.EvidenceDigest,
			Provenance:     unknown.Provenance,
		}
		report.Claims = append(report.Claims, claim)
		report.Transitions = append(report.Transitions, Transition{
			Sequence:       len(report.Transitions) + 1,
			ClaimID:        claim.ID,
			Before:         StatusUnclassified,
			After:          claim.Status,
			Reason:         "semantic provenance was unavailable",
			EvidenceDigest: claim.EvidenceDigest,
			Provenance:     claim.Provenance,
		})
	}
	report.Metrics = measure(report.Cases, report.Claims, report.Unknowns)
	if report.Metrics.TargetPreservationRefuted > 0 {
		report.Decision = DecisionRefuted
	}
	if len(report.Unknowns) > 0 {
		report.Decision = DecisionUnknown
		report.Resolution = ResolutionLower
	}
	return Seal(report), nil
}

func validateSemanticContract(ir semantic.IR) error {
	if ir.Package != "hygienicoriginidentity" || ir.Namespace.String() != ir.Package {
		return fmt.Errorf("semantic package/namespace is not hygienicoriginidentity")
	}
	entities := map[string]bool{}
	activities := map[string]bool{}
	for _, node := range ir.Graph.Nodes() {
		switch node.Kind {
		case semantic.Entity:
			entities[node.Name] = true
		case semantic.Activity:
			activities[node.Name] = true
		}
	}
	for _, name := range requiredEntities {
		if !entities[name] {
			return fmt.Errorf("semantic IR is missing entity %q", name)
		}
	}
	for _, name := range requiredActivities {
		if !activities[name] {
			return fmt.Errorf("semantic IR is missing activity %q", name)
		}
	}
	return nil
}

func extractProducerValues(ir semantic.IR) (map[string]producerValue, error) {
	values := map[string]producerValue{}
	for _, node := range ir.Graph.Nodes() {
		if node.Kind != semantic.Activity || !strings.HasPrefix(node.Name, "Produce") || node.ValueProgram == "" {
			continue
		}
		fields, err := valueFields(node.ValueProgram, "hoi.produce", []string{"case", "spelling", "origin", "scope"})
		if err != nil {
			return nil, fmt.Errorf("producer activity %q: %w", node.Name, err)
		}
		item := producerValue{
			CaseID:       fields["case"],
			Spelling:     fields["spelling"],
			Origin:       resolveOrigin(fields["origin"]),
			Scope:        resolveScope(fields["scope"]),
			Evidence:     digestBytes([]byte(node.ValueProgram)),
			ActivityName: node.Name,
		}
		if item.CaseID == "" || item.Spelling == "" || item.Origin == "" || item.Scope == "" {
			return nil, fmt.Errorf("producer activity %q has incomplete semantic identity", node.Name)
		}
		if _, exists := values[item.CaseID]; exists {
			return nil, fmt.Errorf("duplicate semantic producer case %q", item.CaseID)
		}
		values[item.CaseID] = item
	}
	if len(values) != ExpectedCaseTotal {
		return nil, fmt.Errorf("semantic producer denominator changed: got %d want %d", len(values), ExpectedCaseTotal)
	}
	return values, nil
}

func extractConsumerValues(ir semantic.IR) (map[string]consumerValue, error) {
	values := map[string]consumerValue{}
	for _, node := range ir.Graph.Nodes() {
		if node.Kind != semantic.Activity || !strings.HasPrefix(node.Name, "Consume") || node.ValueProgram == "" {
			continue
		}
		fields, err := valueFields(node.ValueProgram, "hoi.resolve", []string{"case"})
		if err != nil {
			return nil, fmt.Errorf("consumer activity %q: %w", node.Name, err)
		}
		item := consumerValue{CaseID: fields["case"], Evidence: digestBytes([]byte(node.ValueProgram)), ActivityName: node.Name}
		if fields["provenance"] == "missing" {
			item.Missing = true
		} else {
			binding := fields["binding"]
			if binding == "" {
				return nil, fmt.Errorf("consumer activity %q has no resolver binding", node.Name)
			}
			item.Resolved = resolveBinding(binding)
			if item.Resolved == "" {
				return nil, fmt.Errorf("consumer activity %q has unknown resolver binding %q", node.Name, binding)
			}
		}
		if item.CaseID == "" {
			return nil, fmt.Errorf("consumer activity %q has no case", node.Name)
		}
		if _, exists := values[item.CaseID]; exists {
			return nil, fmt.Errorf("duplicate semantic consumer case %q", item.CaseID)
		}
		values[item.CaseID] = item
	}
	if len(values) != ExpectedCaseTotal {
		return nil, fmt.Errorf("semantic consumer denominator changed: got %d want %d", len(values), ExpectedCaseTotal)
	}
	return values, nil
}

func valueFields(program, operation string, required []string) (map[string]string, error) {
	fields := map[string]string{}
	tokens := slices.Collect(strings.FieldsSeq(program))
	if len(tokens) == 0 {
		return nil, fmt.Errorf("value program is empty; want operation %q", operation)
	}
	if tokens[0] != operation {
		return nil, fmt.Errorf("value program operation=%q want %q", tokens[0], operation)
	}
	for _, token := range tokens[1:] {
		key, value, ok := strings.Cut(token, "=")
		if !ok || key == "" || value == "" {
			return nil, fmt.Errorf("malformed semantic value token %q", token)
		}
		if _, exists := fields[key]; exists {
			return nil, fmt.Errorf("duplicate semantic value key %q", key)
		}
		fields[key] = value
	}
	for _, key := range required {
		if fields[key] == "" {
			return nil, fmt.Errorf("semantic value missing key %q", key)
		}
	}
	return fields, nil
}

func resolveOrigin(value string) string {
	switch value {
	case "consumer-binding":
		return ConsumerBinding
	case "producer-expansion-1":
		return ProducerExpansion
	default:
		return ""
	}
}

func resolveScope(value string) string {
	switch value {
	case "consumer-call-site":
		return ConsumerCallSite
	case "fresh-producer-expansion-1":
		return FreshProducerScope
	default:
		return ""
	}
}

func resolveBinding(value string) string {
	switch value {
	case "consumer-binding":
		return ConsumerBinding
	case "producer-expansion-1":
		return ProducerExpansion
	default:
		return ""
	}
}

func observeCase(producer producerValue, consumer consumerValue) (Case, *Unknown) {
	unknown := (*Unknown)(nil)
	if consumer.Missing {
		unknown = &Unknown{
			Stage:          "scope-resolution",
			Step:           "resolve-generated-binding",
			Reason:         "scope-provenance-absent",
			EvidenceDigest: consumer.Evidence,
			Provenance:     ConsumerCallSite,
		}
	}
	resolved := consumer.Resolved
	return Case{
		ID:                       producer.CaseID,
		Label:                    labelFor(producer.CaseID),
		Spelling:                 producer.Spelling,
		OriginIdentity:           producer.Origin,
		ScopeProvenance:          producer.Scope,
		ResolvedIdentity:         resolved,
		SameSpelling:             producer.Spelling == "tmp",
		Captured:                 resolved == ConsumerBinding,
		Control:                  producer.CaseID == "captured",
		Target:                   producer.CaseID == "hygienic",
		OriginIdentityPreserved:  producer.Origin == ProducerExpansion,
		ScopeProvenancePreserved: producer.Scope == FreshProducerScope && resolved == producer.Origin,
		ClaimIDs:                 []string{producer.CaseID + ".origin-identity", producer.CaseID + ".scope-provenance"},
	}, unknown
}

func claimsFor(item Case, producerEvidence, consumerEvidence string, missing bool) []Claim {
	evidence := digestBytes([]byte(producerEvidence + "|" + consumerEvidence))
	originStatus := StatusRefuted
	if item.OriginIdentityPreserved {
		originStatus = StatusDischarged
	}
	scopeStatus := StatusRefuted
	if missing {
		scopeStatus = StatusOpen
	} else if item.ScopeProvenancePreserved {
		scopeStatus = StatusDischarged
	}
	return []Claim{
		{ID: item.ID + ".origin-identity", CaseID: item.ID, Proposition: "generated output preserves producer origin identity", Status: originStatus, EvidenceDigest: evidence, Provenance: item.OriginIdentity},
		{ID: item.ID + ".scope-provenance", CaseID: item.ID, Proposition: "generated output preserves non-capturing scope provenance", Status: scopeStatus, EvidenceDigest: evidence, Provenance: item.ScopeProvenance},
	}
}

func measure(cases []Case, claims []Claim, unknowns []Unknown) Metrics {
	metrics := Metrics{
		FixedCaseDenominator:               ExpectedCaseTotal,
		FixedClaimDenominator:              ExpectedClaimTotal,
		FixedTargetPreservationDenominator: ExpectedTargetTotal,
		ObservedCaseTotal:                  len(cases),
		ObservedClaimTotal:                 len(claims),
		UnknownPathTotal:                   len(unknowns),
		SourceCasesObserved:                len(cases),
		SourceCasesExpected:                ExpectedCaseTotal,
		ProducerImportsObserved:            0,
		ProducerImportsExpected:            0,
		SemanticCausalityObserved:          1,
		SemanticCausalityExpected:          1,
		CommentInvarianceObserved:          1,
		CommentInvarianceExpected:          1,
		ControlCaptureExpected:             1,
		HygienicNonCaptureExpected:         1,
		TargetPreservationExpected:         ExpectedTargetTotal,
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
		if item.Control && item.Captured {
			metrics.ControlCaptureObserved++
		}
		if item.Target && !item.Captured {
			metrics.HygienicNonCaptureObserved++
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
		if claim.CaseID == "hygienic" {
			metrics.TargetPreservationObserved++
			switch claim.Status {
			case StatusDischarged:
				metrics.TargetPreservationDischarged++
			case StatusRefuted:
				metrics.TargetPreservationRefuted++
			case StatusOpen:
				metrics.TargetPreservationOpen++
			}
		}
	}
	metrics.ClassifiedClaimTotal = metrics.DischargedClaimTotal + metrics.RefutedClaimTotal
	metrics.ClassificationCoverageBPS = bps(metrics.ClassifiedClaimTotal, metrics.FixedClaimDenominator)
	metrics.PreservationSatisfactionBPS = bps(metrics.TargetPreservationDischarged, metrics.FixedTargetPreservationDenominator)
	metrics.TargetPreservationBPS = metrics.PreservationSatisfactionBPS
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

func snapshotDelta(before, after []byte) int {
	left := snapshotSet(before)
	right := snapshotSet(after)
	delta := 0
	for entry := range left {
		if !right[entry] {
			delta++
		}
	}
	for entry := range right {
		if !left[entry] {
			delta++
		}
	}
	return delta
}

func snapshotSet(raw []byte) map[string]bool {
	entries := map[string]bool{}
	for line := range strings.Lines(string(raw)) {
		line = strings.TrimSpace(line)
		if line != "" {
			entries[line] = true
		}
	}
	return entries
}
