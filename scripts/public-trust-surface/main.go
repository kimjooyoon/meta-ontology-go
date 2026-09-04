package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/publictrust"
)

func main() {
	mode := flag.String("mode", "verify", "verify or generate")
	sourcePath := flag.String("source", publictrust.CanonicalPolicyPath, "canonical .gooo policy")
	readmePath := flag.String("readme", "README.md", "README to verify")
	outputDir := flag.String("output-dir", "", "directory for generated evidence")
	flag.Parse()

	if *mode != "verify" && *mode != "generate" {
		fatal(errors.New("mode must be verify or generate"))
	}
	source, err := os.ReadFile(*sourcePath)
	if err != nil {
		fatal(fmt.Errorf("read public trust source: %w", err))
	}
	policy, generated, err := publictrust.GenerateNamed(*sourcePath, source)
	if err != nil {
		fatal(err)
	}
	if *mode == "verify" {
		if _, err := publictrust.Load(*sourcePath, source); err != nil {
			fatal(err)
		}
		committed, err := os.ReadFile(publictrust.GeneratedEvaluatorPath)
		if err != nil {
			fatal(fmt.Errorf("read committed generated evaluator: %w", err))
		}
		if !bytes.Equal(committed, generated) {
			fatal(errors.New("generated public trust evaluator is stale"))
		}
	}

	readme, err := os.ReadFile(*readmePath)
	if err != nil {
		fatal(fmt.Errorf("read README: %w", err))
	}
	readmeDrift := 0
	if !hasGeneratedBlock(string(readme), publictrust.RenderBadgeBlock(policy)) {
		readmeDrift = 1
	}
	placeholderClaims := countSecurityPlaceholders(mustRead("SECURITY.md"))
	summary := publictrust.SummaryFor(policy, readmeDrift, placeholderClaims)
	if summary.MetaRowsBound != summary.MetaRowsTotal || summary.TotalActiveBadges != publictrust.ExpectedActiveBadges || summary.UniqueClaims != summary.MetaRowsTotal || summary.UniqueTargets != summary.MetaRowsTotal || summary.UnsupportedBadges != 0 || summary.DuplicateClaims != 0 || summary.DuplicateTargets != 0 || summary.READMEGeneratedDrift != 0 || summary.SECURITYPlaceholderClaims != 0 || summary.RepositoryWrites != 0 || summary.LocalTestExecutions != 0 {
		fatal(errors.New("public trust acceptance metrics are not closed"))
	}

	manifest := publictrust.Manifest{
		Schema:           publictrust.Schema,
		Policy:           policy,
		Summary:          summary,
		SecurityEvidence: securityEvidence(policy),
	}
	if *outputDir == "" {
		return
	}
	if err := os.MkdirAll(*outputDir, 0o755); err != nil {
		fatal(fmt.Errorf("create output directory: %w", err))
	}
	writeJSON(filepath.Join(*outputDir, "public-trust-manifest.json"), manifest)
	writeJSON(filepath.Join(*outputDir, "public-trust-report.json"), summary)
	writeText(filepath.Join(*outputDir, "public-trust-report.md"), publictrust.RenderReport(policy, summary))
	writeText(filepath.Join(*outputDir, "README.badges.md"), publictrust.RenderBadgeBlock(policy))
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

func securityEvidence(policy publictrust.Policy) []string {
	links := make([]string, 0, policy.SecurityEvidenceLinks)
	for _, row := range policy.Rows {
		if strings.HasPrefix(row.Category, "Security / Supply Chain") {
			links = append(links, row.EvidenceURL)
		}
	}
	return links
}

func mustRead(path string) []byte {
	value, err := os.ReadFile(path)
	if err != nil {
		fatal(fmt.Errorf("read %s: %w", path, err))
	}
	return value
}

func writeJSON(path string, value any) {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		fatal(fmt.Errorf("encode %s: %w", path, err))
	}
	writeText(path, string(data)+"\n")
}

func writeText(path, value string) {
	if err := os.WriteFile(path, []byte(value), 0o644); err != nil {
		fatal(fmt.Errorf("write %s: %w", path, err))
	}
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
