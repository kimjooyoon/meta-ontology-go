package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	expansion "github.com/kimjooyoon/meta-ontology-go/internal/meta/capabilityscopedexpansion"
	independent "github.com/kimjooyoon/meta-ontology-go/internal/meta/capabilityscopedexpansion/verify"
)

type options struct {
	source, output, subject, pinnedFile, sandbox string
}

type judgeReport struct {
	Schema                     string                `json:"schema"`
	SourceDigest               string                `json:"source_digest"`
	SemanticDigest             string                `json:"semantic_digest"`
	ProviderDigest             string                `json:"provider_digest"`
	SubjectSHA                 string                `json:"subject_sha"`
	Decision                   string                `json:"decision"`
	SourceReconstruction       string                `json:"source_reconstruction"`
	SourceReconstructionPasses int                   `json:"source_reconstruction_passes"`
	SourceReconstructionTotal  int                   `json:"source_reconstruction_total"`
	ProducerImportNumerator    int                   `json:"producer_import_numerator"`
	ProducerImportDenominator  int                   `json:"producer_import_denominator"`
	Judgements                 []independent.Verdict `json:"judgements"`
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	options, err := parseOptions()
	if err != nil {
		return err
	}
	source, err := os.ReadFile(options.source)
	if err != nil {
		return err
	}
	if err := expansion.ValidateShape(source); err != nil {
		return err
	}
	if err := requireOutsideRepository(options.output); err != nil {
		return err
	}
	providerRaw, err := expansion.CaptureProvider(expansion.ProviderRequest{SubjectSHA: options.subject, PinnedFile: options.pinnedFile, SandboxRoot: options.sandbox})
	if err != nil {
		return err
	}
	if err := writeFile(filepath.Join(options.output, "provider-observations.json"), append(providerRaw, '\n')); err != nil {
		return err
	}

	suite, receipts, judgements, err := evaluateAndJudge(source, providerRaw, options.subject)
	if err != nil {
		return err
	}
	for index := range suite.Cases {
		suite.Cases[index].IndependentJudge = judgements[index].Status
		suite.Cases[index].IndependentReason = judgements[index].Reason
		if judgements[index].Status != "PASS" {
			return fmt.Errorf("independent judge rejected %s: %s", suite.Cases[index].CaseID, judgements[index].Reason)
		}
	}
	suite.IndependentJudge = "PASS"
	suite = expansion.SealSuite(suite)
	if suite.Summary.CasesPassed != suite.Summary.CasesTotal || suite.Summary.CasesTotal != expansion.FixedCaseTotal {
		return fmt.Errorf("capability-scoped expansion suite failed: %d/%d", suite.Summary.CasesPassed, suite.Summary.CasesTotal)
	}
	allowWritten, denyWritten, unknownWritten := false, false, false
	for index, receipt := range receipts {
		suite.Cases[index].ClaimStatus = claimStatus(receipt, "capability-scope-exact")
		if err := writeJSON(filepath.Join(options.output, "receipts", suite.Cases[index].CaseID+"-receipt.json"), receipt); err != nil {
			return err
		}
		if receipt.Decision == expansion.DecisionAllow && !allowWritten {
			if err := writeJSON(filepath.Join(options.output, "allow-receipt.json"), receipt); err != nil {
				return err
			}
			allowWritten = true
		}
		if receipt.Decision == expansion.DecisionDeny && !denyWritten {
			if err := writeJSON(filepath.Join(options.output, "deny-receipt.json"), receipt); err != nil {
				return err
			}
			denyWritten = true
		}
		if receipt.Decision == expansion.DecisionUnknown && !unknownWritten {
			if err := writeJSON(filepath.Join(options.output, "unknown-receipt.json"), receipt); err != nil {
				return err
			}
			unknownWritten = true
		}
	}
	judge := judgeReport{Schema: "gooo/capability-scoped-expansion-independent-judge/v2", SourceDigest: suite.SourceDigest, SemanticDigest: suite.SemanticDigest, ProviderDigest: digest(providerRaw), SubjectSHA: options.subject, Decision: "PASS", SourceReconstruction: "PASS", SourceReconstructionPasses: 1, SourceReconstructionTotal: 1, ProducerImportNumerator: 0, ProducerImportDenominator: 1, Judgements: judgements}
	if err := writeJSON(filepath.Join(options.output, "independent-judge.json"), judge); err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(options.output, "suite.json"), suite); err != nil {
		return err
	}
	if err := writeInterventions(filepath.Join(options.output, "interventions"), source, providerRaw, options.subject, suite); err != nil {
		return err
	}
	return writeSummary(filepath.Join(options.output, "summary.md"), suite)
}

func evaluateAndJudge(source, providerRaw []byte, subject string) (expansion.Suite, []expansion.Receipt, []independent.Verdict, error) {
	suite, receipts, err := expansion.EvaluateSuite(source, providerRaw, subject)
	if err != nil {
		return expansion.Suite{}, nil, nil, err
	}
	judgements := make([]independent.Verdict, 0, len(receipts))
	for index, receipt := range receipts {
		raw, err := json.MarshalIndent(receipt, "", "  ")
		if err != nil {
			return expansion.Suite{}, nil, nil, err
		}
		judgements = append(judgements, independent.Judge(source, providerRaw, append(raw, '\n')))
		suite.Cases[index].ClaimStatus = claimStatus(receipt, "capability-scope-exact")
	}
	return suite, receipts, judgements, nil
}

func writeInterventions(outputDir string, source, providerRaw []byte, subject string, base expansion.Suite) error {
	policySource := bytes.Replace(source, []byte("authorization=exact-current"), []byte("authorization=deny-all"), 1)
	if bytes.Equal(policySource, source) {
		return fmt.Errorf("policy intervention did not change a semantic value")
	}
	commentSource := append(append([]byte(nil), source...), []byte("\n// comment-only intervention\n")...)
	policy, err := interventionRecord("policy-deny-all", "semantic-policy", source, policySource, providerRaw, subject, base, false)
	if err != nil {
		return err
	}
	comment, err := interventionRecord("comment-only", "comment", source, commentSource, providerRaw, subject, base, true)
	if err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(outputDir, "policy-change.json"), policy); err != nil {
		return err
	}
	return writeJSON(filepath.Join(outputDir, "comment-only.json"), comment)
}

func interventionRecord(id, kind string, baseSource, changedSource, providerRaw []byte, subject string, base expansion.Suite, preserveDecision bool) (expansion.Intervention, error) {
	changed, _, judgements, err := evaluateAndJudge(changedSource, providerRaw, subject)
	if err != nil {
		return expansion.Intervention{}, err
	}
	for _, judgement := range judgements {
		if judgement.Status != "PASS" {
			return expansion.Intervention{}, fmt.Errorf("intervention %s failed independent judge: %s", id, judgement.Reason)
		}
	}
	baseDecision, baseClaim := allowDecisionAndClaim(base)
	changedDecision, changedClaim := allowDecisionAndClaim(changed)
	decisionPreserved := baseDecision == changedDecision
	semanticPreserved := base.SemanticDigest == changed.SemanticDigest
	if preserveDecision != decisionPreserved {
		return expansion.Intervention{}, fmt.Errorf("intervention %s decision preservation expectation failed", id)
	}
	if !preserveDecision && (decisionPreserved || baseClaim == changedClaim) {
		return expansion.Intervention{}, fmt.Errorf("intervention %s did not change authorization claim", id)
	}
	if preserveDecision && !semanticPreserved {
		return expansion.Intervention{}, fmt.Errorf("comment-only intervention changed semantic digest")
	}
	return expansion.Intervention{ID: id, Kind: kind, BaseSourceDigest: base.SourceDigest, ChangedSourceDigest: changed.SourceDigest, BaseSemanticDigest: base.SemanticDigest, ChangedSemanticDigest: changed.SemanticDigest, BaseDecision: baseDecision, ChangedDecision: changedDecision, BaseClaimStatus: baseClaim, ChangedClaimStatus: changedClaim, DecisionPreserved: decisionPreserved, SemanticDigestPreserved: semanticPreserved, IndependentJudge: "PASS"}, nil
}

func allowDecisionAndClaim(suite expansion.Suite) (string, string) {
	for _, item := range suite.Cases {
		if item.ObservedDecision == expansion.DecisionAllow || item.CaseID == "allow-current-file-time" {
			return item.ObservedDecision, item.ClaimStatus
		}
	}
	return expansion.DecisionUnknown, ""
}

func parseOptions() (options, error) {
	result := options{}
	flags := flag.NewFlagSet("capability-scoped-expansion-witness", flag.ContinueOnError)
	flags.StringVar(&result.source, "source", "", "the single Gooo source used by every case")
	flags.StringVar(&result.output, "output-dir", "", "directory outside the repository for receipts")
	flags.StringVar(&result.subject, "subject-sha", "", "exact CI subject SHA")
	flags.StringVar(&result.pinnedFile, "pinned-file", "", "CI-created file observed by the provider")
	flags.StringVar(&result.sandbox, "sandbox", "", "temporary sandbox used for before/after enforcement observations")
	if err := flags.Parse(os.Args[1:]); err != nil {
		return result, err
	}
	if result.source == "" || result.output == "" || result.subject == "" || result.pinnedFile == "" || result.sandbox == "" {
		return result, fmt.Errorf("source, output-dir, subject-sha, pinned-file, and sandbox are required")
	}
	return result, nil
}

func requireOutsideRepository(path string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	target, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	relative, err := filepath.Rel(cwd, target)
	if err != nil {
		return err
	}
	if relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return fmt.Errorf("output-dir must be outside repository: %s", path)
	}
	return nil
}

func writeJSON(path string, value any) error {
	raw, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return writeFile(path, append(raw, '\n'))
}

func writeFile(path string, raw []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, raw, 0o644)
}

func writeSummary(path string, suite expansion.Suite) error {
	content := fmt.Sprintf("# Capability-scoped expansion\n\n- source reconstruction: `%d/%d`; producer import numerator/denominator: `%d/%d`\n- source digests: raw `%s`; semantic `%s`\n- cases: `%d/%d` (allow `%d`, deny `%d`, UNKNOWN `%d`)\n- capability requests: `%d`; authorized `%d`, denied `%d`, UNKNOWN `%d`\n- CURRENT_EVIDENCE capabilities: `%d/%d`; HISTORICAL_FIXTURE declarations: `%d`\n- enforcement observations: `%d`; blocked writes `%d`, blocked mutation `%d`\n- observed effects: repository writes `%d`, mutation authority `%t`, promotion authority `%t`\n- independent judge: `%s`\n", suite.Summary.SourceReconstructionPasses, suite.Summary.SourceReconstructionTotal, suite.Summary.ProducerImportNumerator, suite.Summary.ProducerImportDenominator, suite.SourceDigest, suite.SemanticDigest, suite.Summary.CasesPassed, suite.Summary.CasesTotal, suite.Summary.AllowCases, suite.Summary.DenyCases, suite.Summary.UnknownCases, suite.Summary.CapabilityRequests, suite.Summary.CapabilityAuthorized, suite.Summary.CapabilityDenied, suite.Summary.CapabilityUnknown, suite.Summary.CurrentEvidenceCapabilities, suite.Summary.CurrentEvidenceDenominator, suite.Summary.HistoricalFixtureCapabilities, suite.Summary.EnforcementObservations, suite.Summary.BlockedWriteAttempts, suite.Summary.BlockedMutationAttempts, suite.Summary.RepositoryWrites, suite.Summary.MutationAuthority, suite.Summary.PromotionAuthority, suite.IndependentJudge)
	return writeFile(path, []byte(content))
}

func digest(raw []byte) string {
	return fmt.Sprintf("sha256:%x", sha256.Sum256(raw))
}

func claimStatus(receipt expansion.Receipt, id string) string {
	for _, claim := range receipt.Claims {
		if claim.ID == id {
			return claim.Status
		}
	}
	return ""
}
