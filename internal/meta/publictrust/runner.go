package publictrust

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type RunOptions struct {
	Mode       string
	SourcePath string
	ReadmePath string
	OutputDir  string
}

func Execute(options RunOptions) error {
	options = defaultRunOptions(options)
	policy, err := loadPolicy(options)
	if err != nil {
		return err
	}
	summary, err := loadRunSummary(options, policy)
	if err != nil {
		return err
	}
	if err := validateSummary(summary); err != nil {
		return err
	}
	return writeOutputs(options.OutputDir, policy, summary)
}

func defaultRunOptions(options RunOptions) RunOptions {
	if options.Mode == "" {
		options.Mode = "verify"
	}
	if options.SourcePath == "" {
		options.SourcePath = CanonicalPolicyPath
	}
	if options.ReadmePath == "" {
		options.ReadmePath = "README.md"
	}
	return options
}

func loadPolicy(options RunOptions) (Policy, error) {
	if options.Mode != "verify" && options.Mode != "generate" {
		return Policy{}, errors.New("mode must be verify or generate")
	}
	source, err := os.ReadFile(options.SourcePath)
	if err != nil {
		return Policy{}, fmt.Errorf("read public trust source: %w", err)
	}
	policy, generated, err := GenerateNamed(options.SourcePath, source)
	if err != nil {
		return Policy{}, err
	}
	if options.Mode != "verify" {
		return policy, nil
	}
	if _, err := Load(options.SourcePath, source); err != nil {
		return Policy{}, err
	}
	committed, err := os.ReadFile(GeneratedEvaluatorPath)
	if err != nil {
		return Policy{}, fmt.Errorf("read committed generated evaluator: %w", err)
	}
	if !bytes.Equal(committed, generated) {
		return Policy{}, errors.New("generated public trust evaluator is stale")
	}
	return policy, nil
}

func loadRunSummary(options RunOptions, policy Policy) (Summary, error) {
	readme, err := os.ReadFile(options.ReadmePath)
	if err != nil {
		return Summary{}, fmt.Errorf("read README: %w", err)
	}
	readmeDrift := 0
	if !hasGeneratedBlock(string(readme), RenderBadgeBlock(policy)) {
		readmeDrift = 1
	}
	security, err := os.ReadFile("SECURITY.md")
	if err != nil {
		return Summary{}, fmt.Errorf("read SECURITY.md: %w", err)
	}
	return SummaryFor(policy, readmeDrift, countSecurityPlaceholders(security)), nil
}

func writeOutputs(outputDir string, policy Policy, summary Summary) error {
	if outputDir == "" {
		return nil
	}
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}
	manifest := Manifest{Schema: Schema, Policy: policy, Summary: summary, SecurityEvidence: securityEvidence(policy)}
	if err := writeJSON(filepath.Join(outputDir, "public-trust-manifest.json"), manifest); err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(outputDir, "public-trust-report.json"), summary); err != nil {
		return err
	}
	if err := writeText(filepath.Join(outputDir, "public-trust-report.md"), RenderReport(policy, summary)); err != nil {
		return err
	}
	return writeText(filepath.Join(outputDir, "README.badges.md"), RenderBadgeBlock(policy))
}

func validateSummary(summary Summary) error {
	if summary.MetaRowsBound != summary.MetaRowsTotal || summary.TotalActiveBadges != ExpectedActiveBadges || summary.UniqueClaims != summary.MetaRowsTotal || summary.UniqueTargets != summary.MetaRowsTotal || summary.UnsupportedBadges != 0 || summary.DuplicateClaims != 0 || summary.DuplicateTargets != 0 || summary.READMEGeneratedDrift != 0 || summary.SECURITYPlaceholderClaims != 0 || summary.RepositoryWrites != 0 || summary.LocalTestExecutions != 0 {
		return errors.New("public trust acceptance metrics are not closed")
	}
	return nil
}

func hasGeneratedBlock(readme, expected string) bool {
	begin := "<!-- PUBLIC-TRUST-BADGES:BEGIN -->"
	end := "<!-- PUBLIC-TRUST-BADGES:END -->"
	start := strings.Index(readme, begin)
	if start < 0 {
		return false
	}
	endOffset := strings.Index(readme[start:], end)
	if endOffset < 0 {
		return false
	}
	actual := readme[start : start+endOffset+len(end)]
	return actual == strings.TrimSuffix(expected, "\n")
}

func countSecurityPlaceholders(source []byte) int {
	text := strings.ToLower(string(source))
	placeholders := []string{"use this section", "tell them where", "check |", "white_check_mark", "< 4.0"}
	count := 0
	for _, placeholder := range placeholders {
		if strings.Contains(text, placeholder) {
			count++
		}
	}
	return count
}

func securityEvidence(policy Policy) []string {
	links := make([]string, 0, policy.SecurityEvidenceLinks)
	for _, row := range policy.Rows {
		if strings.HasPrefix(row.Category, "Security / Supply Chain") {
			links = append(links, row.EvidenceURL)
		}
	}
	return links
}

func writeJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("encode %s: %w", path, err)
	}
	return writeText(path, string(data)+"\n")
}

func writeText(path, value string) error {
	if err := os.WriteFile(path, []byte(value), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}
