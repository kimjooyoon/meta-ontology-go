package main

import (
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
	source, output, subject string
}

type judgeReport struct {
	Schema       string                `json:"schema"`
	SourceDigest string                `json:"source_digest"`
	SubjectSHA   string                `json:"subject_sha"`
	Decision     string                `json:"decision"`
	Judgements   []independent.Verdict `json:"judgements"`
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
	suite, receipts := expansion.EvaluateSuite(source, options.subject)
	judgements := make([]independent.Verdict, 0, len(receipts))
	for index, receipt := range receipts {
		raw, err := json.MarshalIndent(receipt, "", "  ")
		if err != nil {
			return err
		}
		raw = append(raw, '\n')
		judgement := independent.Judge(source, raw)
		judgements = append(judgements, judgement)
		result := &suite.Cases[index]
		result.IndependentJudge = judgement.Status
		result.IndependentReason = judgement.Reason
		if judgement.Status != "PASS" || judgement.Decision != result.ObservedDecision || judgement.Resolution != result.ObservedResolution {
			return fmt.Errorf("independent judge rejected %s: %s", result.CaseID, judgement.Reason)
		}
		name := result.CaseID + "-receipt.json"
		if err := writeJSON(filepath.Join(options.output, "receipts", name), receipt); err != nil {
			return err
		}
		if result.CaseID == "allow-exact" {
			if err := writeJSON(filepath.Join(options.output, "allow-receipt.json"), receipt); err != nil {
				return err
			}
		}
		if result.CaseID == "deny-undeclared-network" {
			if err := writeJSON(filepath.Join(options.output, "deny-receipt.json"), receipt); err != nil {
				return err
			}
		}
		if result.CaseID == "unknown-missing-evidence" {
			if err := writeJSON(filepath.Join(options.output, "unknown-receipt.json"), receipt); err != nil {
				return err
			}
		}
	}
	suite.IndependentJudge = "PASS"
	suite = expansion.SealSuite(suite)
	if suite.Summary.CasesPassed != expansion.FixedCaseTotal || len(judgements) != expansion.FixedCaseTotal {
		return fmt.Errorf("capability-scoped expansion suite failed: %d/%d", suite.Summary.CasesPassed, expansion.FixedCaseTotal)
	}
	if err := writeJSON(filepath.Join(options.output, "suite.json"), suite); err != nil {
		return err
	}
	judge := judgeReport{Schema: "gooo/capability-scoped-expansion-independent-judge/v1",
		SourceDigest: suite.SourceDigest, SubjectSHA: options.subject, Decision: "PASS", Judgements: judgements}
	if err := writeJSON(filepath.Join(options.output, "independent-judge.json"), judge); err != nil {
		return err
	}
	return writeSummary(filepath.Join(options.output, "summary.md"), suite)
}

func parseOptions() (options, error) {
	result := options{}
	flags := flag.NewFlagSet("capability-scoped-expansion-witness", flag.ContinueOnError)
	flags.StringVar(&result.source, "source", "", "the single Gooo source used by every case")
	flags.StringVar(&result.output, "output-dir", "", "directory outside the repository for receipts")
	flags.StringVar(&result.subject, "subject-sha", "", "exact CI subject SHA")
	if err := flags.Parse(os.Args[1:]); err != nil {
		return result, err
	}
	if result.source == "" || result.output == "" || result.subject == "" {
		return result, fmt.Errorf("source, output-dir, and subject-sha are required")
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
	content := fmt.Sprintf("# Capability-scoped expansion\n\n- source digest: `%s`\n- Gooo source comparisons: one source, allow and deny receipts\n- cases: `%d/%d` (allow `%d`, deny `%d`, UNKNOWN `%d`)\n- capability permissions: requested `%d`, authorized `%d`, denied `%d`, UNKNOWN `%d`\n- blocked effect attempts: repository writes `%d`, mutation `%d`\n- observed effects: repository writes `%d`, mutation authority `%t`, promotion authority `%t`\n- independent judge: `%s`\n", suite.SourceDigest, suite.Summary.CasesPassed, suite.Summary.CasesTotal, suite.Summary.AllowCases, suite.Summary.DenyCases, suite.Summary.UnknownCases, suite.Summary.CapabilityRequests, suite.Summary.CapabilityAuthorized, suite.Summary.CapabilityDenied, suite.Summary.CapabilityUnknown, suite.Summary.BlockedWriteAttempts, suite.Summary.BlockedMutationAttempts, suite.Summary.RepositoryWrites, suite.Summary.MutationAuthority, suite.Summary.PromotionAuthority, suite.IndependentJudge)
	return writeFile(path, []byte(content))
}
