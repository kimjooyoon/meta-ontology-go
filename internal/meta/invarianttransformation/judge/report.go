package judge

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/invarianttransformation/model"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func Indicators(summary model.Summary) []model.Indicator {
	return []model.Indicator{
		indicator("gooo.metric.bounded-transformation.source-derived-cases.v2", model.ProofFoundation, "derive-source-case-inventory", summary.SourceDerivedCases, summary.BoundedInputDomainDenominator, "="),
		indicator("gooo.metric.bounded-transformation.unique-claim-instances.v2", model.ProofCoherence, "case-qualified-claim-ledger", summary.UniqueClaimInstances, summary.CasesTotal*len(model.CanonicalValueSpecs()), "="),
		indicator("gooo.metric.bounded-transformation.accepted-transitions.v2", model.ProofCoherence, "verify-transition-digests", summary.AcceptedTransitions, summary.UniqueClaimInstances, "="),
		indicator("gooo.metric.bounded-transformation.domain-observations.v2", model.ProofFoundation, "observe-bounded-input-domain", summary.BoundedInputDomainObservations, summary.BoundedInputDomainDenominator, "="),
		indicator("gooo.metric.bounded-transformation.provisional-receipts.v2", model.ProofFoundation, "emit-provisional-receipts", summary.ProvisionalReceipts, summary.SourceDerivedCases, "="),
		indicator("gooo.metric.bounded-transformation.authorization-receipts.v2", model.ProofCoherence, "independent-authorization", summary.AuthorizationReceipts, 2, "="),
		indicator("gooo.metric.bounded-transformation.executed-effects.v2", model.ProofCoherence, "post-judgment-effect-gate", summary.ExecutedEffects, 1, "="),
		indicator("gooo.metric.bounded-transformation.independently-observed-effects.v2", model.ProofCoherence, "read-only-effect-observation", summary.IndependentlyObservedEffects, 1, "="),
		indicator("gooo.metric.bounded-transformation.unknown-effect-scopes.v2", model.ProofFoundation, "separate-transient-write-scope", summary.UnknownEffectScopes, 1, "="),
		indicator("gooo.metric.bounded-transformation.corrections.v2", model.ProofFoundation, "fixed-correction-denominator", summary.CorrectionCount, summary.CorrectionDenominator, "="),
	}
}

func indicator(id, proof, operation string, value, target int, relation string) model.Indicator {
	return model.Indicator{MetricID: id, Producer: model.ProducerID, Consumer: model.ConsumerID, MetaOperation: operation, ProofChoice: proof, Value: value, Target: target, Relation: relation, Satisfied: value == target}
}

// ValidateReport is a report consumer check. It independently judges every
// receipt and compares only against labeled validator expectations; it does
// not call producer.Build or treat deterministic replay as evidence.
func ValidateReport(report model.Report, source []byte) error {
	file, diagnostics := syntax.ParseFile(model.SourcePath, string(source))
	if diagnostics.HasErrors() {
		return fmt.Errorf("report source syntax is invalid: %s", diagnostics.Error())
	}
	semanticDigest, err := semanticSourceDigest(file)
	if err != nil {
		return err
	}
	if report.Schema != model.ReportSchema || !model.ValidHead(report.HeadSHA) || report.ExecutionID == "" || report.SourcePath != model.SourcePath ||
		report.SourceDigest != model.DigestBytes(source) || report.SemanticSourceDigest != semanticDigest || report.ContractDigest != model.ValueContractDigest() ||
		report.ValidatorContractDigest != model.ValidatorContractDigest() || report.DenominatorID != model.DenominatorID || report.DenominatorTotal != len(report.Cases) ||
		report.Digest == "" || report.Digest != model.SealReport(report).Digest {
		return fmt.Errorf("report identity or digest is invalid")
	}
	if !validRepositoryObservation(report) {
		return fmt.Errorf("repository observation is not independently bound")
	}
	ids, err := sourceCaseIDs(source)
	if err != nil || len(ids) != len(report.Cases) {
		return fmt.Errorf("report case inventory is not source-derived")
	}
	seen := map[string]bool{}
	for _, result := range report.Cases {
		if seen[result.Receipt.CaseID] {
			return fmt.Errorf("duplicate report case %q", result.Receipt.CaseID)
		}
		seen[result.Receipt.CaseID] = true
		if !contains(ids, result.Receipt.CaseID) {
			return fmt.Errorf("report case %q is not source-derived", result.Receipt.CaseID)
		}
		judgment := Judge(result.Receipt, source)
		if !judgment.Independent || !reflect.DeepEqual(judgment, result.Judgment) {
			return fmt.Errorf("case %q independent judgment mismatch", result.Receipt.CaseID)
		}
		provisionalDigest := result.Receipt.Digest
		if result.Receipt.Phase == model.ReceiptExecuted {
			provisional := result.Receipt
			provisional.Phase = model.ReceiptProvisional
			provisional.Effects = []model.Effect{}
			provisional.TempArtifactWriteAuthorized = false
			provisional.Digest = ""
			provisionalDigest = model.SealReceipt(provisional).Digest
		}
		if result.ProvisionalReceiptDigest != provisionalDigest ||
			(result.Judgment.Decision == model.DecisionAllowed && result.AuthorizationReceiptDigest != result.Receipt.AuthorizationDigest) ||
			(result.Judgment.Decision != model.DecisionAllowed && result.AuthorizationReceiptDigest != "") || result.ExecutedEffects != len(result.Receipt.Effects) {
			return fmt.Errorf("case %q receipt phase metrics are not bound", result.Receipt.CaseID)
		}
		if result.IndependentlyObservedEffects != len(result.Receipt.Effects) {
			return fmt.Errorf("case %q independently observed effect count is not bound", result.Receipt.CaseID)
		}
		expectation, ok := model.ValidatorExpectationFor(model.CanonicalContract(), result.Receipt.CaseID)
		if !ok || result.Expectation != expectation || result.Judgment.Decision != expectation.ExpectedDecision || result.Judgment.Resolution != expectation.ExpectedResolution ||
			result.Judgment.Reason != expectation.ExpectedReason || result.Judgment.Status != expectation.ExpectedStatus || len(result.Receipt.Effects) != expectation.ExpectedEffectCount || !result.Satisfied {
			return fmt.Errorf("case %q does not satisfy labeled validator expectation", result.Receipt.CaseID)
		}
	}
	if report.Decision != model.DecisionPass || report.Resolution != model.ResolutionExact || report.Reason != "ALL_BOUNDED_CASES_SATISFIED" {
		return fmt.Errorf("report top decision is not the derived bounded witness result")
	}
	if report.Summary.CorrectionCount != 12 || report.Summary.CorrectionDenominator != 12 || len(report.Indicators) != len(Indicators(report.Summary)) ||
		!reflect.DeepEqual(report.Indicators, Indicators(report.Summary)) {
		return fmt.Errorf("report metrics or correction denominator are invalid")
	}
	expectedSummary := summarizeReport(report.Cases, report.RepositoryObservation)
	if !reflect.DeepEqual(report.Summary, expectedSummary) {
		return fmt.Errorf("report summary is not derived from independently judged cases")
	}
	return nil
}

func summarizeReport(cases []model.CaseResult, observation model.RepositoryObservation) model.Summary {
	summary := model.Summary{CasesTotal: len(cases), SourceDerivedCases: len(cases), BoundedInputDomainObservations: len(cases), BoundedInputDomainDenominator: len(cases), ClaimTemplates: len(model.CanonicalValueSpecs()), CorrectionCount: 12, CorrectionDenominator: 12, RepositoryNetStatusObserved: observation.Observed, RepositoryNetStatusUnchanged: observation.State == model.RepositoryNetContentStateUnchanged, RepositoryNetContentState: observation.State, RepositoryNetSnapshotObservations: boolInt(observation.Observed), RepositoryNetSnapshotDenominator: 1, RepositoryActualOrTransientWrites: model.UnknownEffectScope, RepositoryWrites: -1, AmbientProcessAuthority: model.UnknownEffectScope}
	claimIDs := map[string]bool{}
	for _, item := range cases {
		if item.Satisfied {
			summary.CasesSatisfied++
		}
		switch item.Judgment.Decision {
		case model.DecisionAllowed:
			summary.AuthorizedCases++
		case model.DecisionRefuted:
			summary.RefutedCases++
		case model.DecisionBlocked:
			summary.OpenCases++
		}
		summary.ClaimsTotal += len(item.Receipt.Claims)
		summary.ProvisionalReceipts++
		if item.Judgment.Decision == model.DecisionAllowed {
			summary.AuthorizationReceipts++
		}
		if item.Receipt.Phase == model.ReceiptExecuted {
			summary.TempArtifactWriteAuthorized = true
		}
		summary.ExecutedEffects += len(item.Receipt.Effects)
		summary.IndependentlyObservedEffects += item.IndependentlyObservedEffects
		for _, claim := range item.Receipt.Claims {
			claimIDs[claim.ID] = true
			summary.DischargedClaims += boolInt(claim.Status == model.StatusDischarged)
			summary.RefutedClaims += boolInt(claim.Status == model.StatusRefuted)
			summary.OpenClaims += boolInt(claim.Status == model.StatusOpen)
			summary.TransitionEvents += len(claim.Transitions)
			summary.AcceptedTransitions += len(claim.Transitions)
		}
		for _, effect := range item.Receipt.Effects {
			if effect.Kind == model.EffectApproved {
				summary.ApprovedArtifactEffects++
			}
			if effect.RepositoryActualOrTransientWrites == model.UnknownEffectScope {
				summary.UnknownEffectScopes++
			}
		}
		if item.Receipt.RepositoryWritesObserved {
			if summary.RepositoryWrites < 0 {
				summary.RepositoryWrites = 0
			}
			summary.RepositoryWrites += item.Receipt.RepositoryWrites
		}
		summary.RepositoryMutationAuthorized |= boolInt(item.Receipt.RepositoryMutationAuthorized)
	}
	summary.UniqueClaimInstances = len(claimIDs)
	if summary.CasesTotal > 0 {
		summary.CoverageBPS = summary.CasesSatisfied * 10000 / summary.CasesTotal
		summary.InputDomainCoverageBPS = summary.BoundedInputDomainObservations * 10000 / summary.BoundedInputDomainDenominator
	}
	return summary
}

type repositoryEntry struct {
	Path   string `json:"path"`
	Digest string `json:"digest"`
}

func validRepositoryObservation(report model.Report) bool {
	observation := report.RepositoryObservation
	if !observation.Observed || observation.State != model.RepositoryNetContentStateUnchanged || observation.ExecutionID == "" || observation.ExecutionID != report.ExecutionID || !model.ValidDigest(observation.WitnessReportDigest) || report.Summary.RepositoryNetSnapshotDenominator != 1 || report.Summary.RepositoryNetSnapshotObservations != 1 {
		return false
	}
	before, beforeEntries, ok := validRepositorySnapshot(observation.Before, report.HeadSHA)
	if !ok {
		return false
	}
	after, afterEntries, ok := validRepositorySnapshot(observation.After, report.HeadSHA)
	if !ok || before.ExecutionID != observation.ExecutionID || after.ExecutionID != observation.ExecutionID || !reflect.DeepEqual(beforeEntries, afterEntries) {
		return false
	}
	if observation.WitnessReportDigest != unboundReportDigest(report) {
		return false
	}
	witnessReceipts := make([]string, 0, len(report.Cases))
	witnessEffects := []string{}
	witnessArtifacts := []string{}
	for _, item := range report.Cases {
		if item.Receipt.ExecutionID != report.ExecutionID || !model.ValidDigest(item.Receipt.Digest) {
			return false
		}
		witnessReceipts = append(witnessReceipts, item.Receipt.Digest)
		for _, effect := range item.Receipt.Effects {
			if effect.ExecutionID != report.ExecutionID || effect.Artifact.ExecutionID != report.ExecutionID {
				return false
			}
			witnessEffects = append(witnessEffects, effect.ExecutionReceiptDigest)
			witnessArtifacts = append(witnessArtifacts, effect.Artifact.ContentDigest)
		}
	}
	return reflect.DeepEqual(observation.WitnessReceiptDigests, witnessReceipts) && reflect.DeepEqual(observation.WitnessEffectDigests, witnessEffects) && reflect.DeepEqual(observation.WitnessArtifactDigests, witnessArtifacts)
}

func unboundReportDigest(report model.Report) string {
	unbound := report
	unbound.RepositoryObservation = model.RepositoryObservation{}
	unbound.Summary.RepositoryNetStatusObserved = false
	unbound.Summary.RepositoryNetStatusUnchanged = false
	unbound.Summary.RepositoryNetContentState = model.RepositoryNetContentStateUnknown
	unbound.Summary.RepositoryNetSnapshotObservations = 0
	unbound.Summary.RepositoryNetSnapshotDenominator = 1
	unbound.Summary.RepositoryPathAuthorization = false
	unbound.Summary.RepositoryActualOrTransientWrites = model.UnknownEffectScope
	unbound.Summary.RepositoryWrites = -1
	unbound.Summary.AmbientProcessAuthority = model.UnknownEffectScope
	unbound.Indicators = Indicators(unbound.Summary)
	return model.SealReport(unbound).Digest
}

func validRepositorySnapshot(snapshot model.RepositorySnapshot, headSHA string) (model.RepositorySnapshot, []repositoryEntry, bool) {
	if snapshot.Schema != model.RepositorySnapshotSchema || snapshot.HeadSHA != headSHA || snapshot.ExecutionID == "" || !filepath.IsAbs(snapshot.EntriesPath) || !allowedSnapshotPath(snapshot.EntriesPath) || !model.ValidDigest(snapshot.EntriesDigest) || !model.ValidDigest(snapshot.PathDigest) || snapshot.EntryCount < 0 {
		return model.RepositorySnapshot{}, nil, false
	}
	raw, err := os.ReadFile(snapshot.EntriesPath)
	if err != nil || model.DigestBytes(raw) != snapshot.EntriesDigest {
		return model.RepositorySnapshot{}, nil, false
	}
	var entries []repositoryEntry
	if json.Unmarshal(raw, &entries) != nil || len(entries) != snapshot.EntryCount {
		return model.RepositorySnapshot{}, nil, false
	}
	for _, item := range entries {
		path := filepath.FromSlash(item.Path)
		if item.Path == "" || filepath.IsAbs(path) || item.Path != filepath.ToSlash(filepath.Clean(path)) || item.Path == ".." || strings.HasPrefix(item.Path, "../") || !model.ValidDigest(item.Digest) {
			return model.RepositorySnapshot{}, nil, false
		}
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Path < entries[j].Path })
	for index := 1; index < len(entries); index++ {
		if entries[index-1].Path == entries[index].Path {
			return model.RepositorySnapshot{}, nil, false
		}
	}
	canonical, err := json.Marshal(entries)
	if err != nil || !bytes.Equal(raw, append(canonical, '\n')) || model.Digest(entries) != snapshot.PathDigest {
		return model.RepositorySnapshot{}, nil, false
	}
	return snapshot, entries, true
}

func allowedSnapshotPath(path string) bool {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return false
	}
	for _, root := range []string{snapshotTempRoot()} {
		canonicalRoot, err := filepath.EvalSymlinks(root)
		if err != nil {
			continue
		}
		canonicalRoot, err = filepath.Abs(canonicalRoot)
		if err != nil || !withinPath(canonicalRoot, absolute) {
			continue
		}
		resolved, err := filepath.EvalSymlinks(absolute)
		if err != nil {
			continue
		}
		resolved, err = filepath.Abs(resolved)
		if err == nil && withinPath(canonicalRoot, resolved) {
			return true
		}
	}
	return false
}

func snapshotTempRoot() string {
	if root := os.Getenv("RUNNER_TEMP"); root != "" {
		return root
	}
	return os.TempDir()
}

func withinPath(root, path string) bool {
	relative, err := filepath.Rel(root, path)
	return err == nil && relative != "." && relative != ".." && !filepath.IsAbs(relative) && !strings.HasPrefix(relative, ".."+string(os.PathSeparator))
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func sourceCaseIDs(source []byte) ([]string, error) {
	file, diagnostics := syntax.ParseFile(model.SourcePath, string(source))
	if diagnostics.HasErrors() {
		return nil, fmt.Errorf("source syntax: %s", diagnostics.Error())
	}
	ids := []string{}
	for _, declaration := range file.Declarations {
		activity, ok := declaration.(*syntax.ActivityDecl)
		if !ok || len(activity.Parameters) != 0 || activity.Result.Name != "Transformation" || !activity.ValueProgramPresent {
			continue
		}
		fields := stringsSplitFields(activity.ValueProgram)
		if fields["case"] != "" {
			ids = append(ids, fields["case"])
		}
	}
	return ids, nil
}

func stringsSplitFields(program string) map[string]string {
	fields := map[string]string{}
	for _, part := range splitSemicolon(program) {
		key, value, ok := cutEqual(part)
		if ok {
			fields[key] = value
		}
	}
	return fields
}

// Small source-only helpers keep report validation independent of producer.
func splitSemicolon(value string) []string         { return strings.Split(value, ";") }
func cutEqual(value string) (string, string, bool) { return strings.Cut(value, "=") }
func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
