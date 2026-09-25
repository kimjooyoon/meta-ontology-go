package publictrust

import (
	"errors"
	"fmt"
	"go/format"
	"net/url"
	"strconv"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
	generatedpolicy "github.com/kimjooyoon/meta-ontology-go/internal/meta/publictrust/generated"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const (
	Schema                   = "gooo/public-trust-surface/v1"
	CanonicalPolicyPath      = "examples/public-trust-surface/main.gooo"
	GeneratedEvaluatorPath   = "internal/meta/publictrust/generated/evaluator.go"
	CanonicalActivity        = "DefinePublicTrustSurface"
	PolicyHead               = "public-trust-surface:v1"
	DecisionClosed           = "CLOSED"
	DecisionUnknown          = "UNKNOWN"
	DecisionRefuted          = "REFUTED"
	ExpectedMetaRows         = 16
	ExpectedActiveBadges     = 11
	ExpectedCategoryCount    = 5
	ExpectedGoVersion        = "1.27.0"
	ExpectedRepositoryWrites = 0
	ExpectedLocalTestRuns    = 0
)

type Badge struct {
	ID          string `json:"id"`
	Category    string `json:"category"`
	Claim       string `json:"claim"`
	Semantics   string `json:"semantics"`
	ImageURL    string `json:"image_url"`
	TargetURL   string `json:"target_url"`
	EvidenceURL string `json:"evidence_url"`
	State       string `json:"state"`
	Render      bool   `json:"render"`
}

type CategoryCount struct {
	Category string `json:"category"`
	Count    int    `json:"count"`
}

type Policy struct {
	Schema                 string          `json:"schema"`
	SourcePath             string          `json:"source_path"`
	SourceDigest           string          `json:"source_digest"`
	EvaluatorDigest        string          `json:"evaluator_digest"`
	ActivityCount          int             `json:"activity_count"`
	Rows                   []Badge         `json:"rows"`
	CategoryDenominators   []CategoryCount `json:"category_denominators"`
	GoVersion              string          `json:"go_version"`
	RepositoryWrites       int             `json:"repository_writes"`
	LocalTestExecutions    int             `json:"local_test_executions"`
	SecurityEvidenceLinks  int             `json:"security_evidence_links"`
	RootReadmeVisualPolicy string          `json:"root_readme_visual_policy"`
}

type Summary struct {
	Schema                    string         `json:"schema"`
	PolicySourcePath          string         `json:"policy_source_path"`
	PolicySourceDigest        string         `json:"policy_source_digest"`
	GeneratedEvaluatorDigest  string         `json:"generated_evaluator_digest"`
	CategoryCount             int            `json:"badge_category_count"`
	TotalActiveBadges         int            `json:"total_active_badges"`
	MetaRowsBound             int            `json:"meta_rows_bound"`
	MetaRowsTotal             int            `json:"meta_rows_total"`
	UniqueClaims              int            `json:"unique_claims"`
	UniqueTargets             int            `json:"unique_targets"`
	UnsupportedBadges         int            `json:"unsupported_badges"`
	DuplicateClaims           int            `json:"duplicate_claims"`
	DuplicateTargets          int            `json:"duplicate_targets"`
	READMEGeneratedDrift      int            `json:"readme_generated_drift"`
	SECURITYPlaceholderClaims int            `json:"security_placeholder_claims"`
	SecurityEvidenceLinks     int            `json:"security_evidence_links"`
	RepositoryWrites          int            `json:"ci_repository_writes"`
	LocalTestExecutions       int            `json:"local_test_executions"`
	CategoryRows              map[string]int `json:"category_rows"`
	StateRows                 map[string]int `json:"state_rows"`
}

type Manifest struct {
	Schema           string   `json:"schema"`
	Policy           Policy   `json:"policy"`
	Summary          Summary  `json:"summary"`
	SecurityEvidence []string `json:"security_evidence_links"`
}

func CompileNamed(filename string, source []byte) (Policy, error) {
	if len(source) == 0 {
		return Policy{}, errors.New("public trust policy source is empty")
	}
	file, diagnostics := syntax.ParseFile(filename, string(source))
	if diagnostics.HasErrors() {
		return Policy{}, fmt.Errorf("parse public trust policy: %w", diagnostics.Error())
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return Policy{}, fmt.Errorf("lower public trust policy: %w", err)
	}
	if ir.Package != "publictrust" || ir.Namespace.String() != "public_trust" {
		return Policy{}, fmt.Errorf("public trust policy package/namespace is %q/%q", ir.Package, ir.Namespace)
	}

	policy := Policy{Schema: Schema, SourcePath: filename}
	for _, node := range ir.Graph.Nodes() {
		if node.Kind != semantic.Activity || node.Name != CanonicalActivity {
			continue
		}
		policy.ActivityCount++
		parsed, parseErr := parseMarkers(node.ValueProgram)
		if parseErr != nil {
			return Policy{}, fmt.Errorf("activity %q: %w", CanonicalActivity, parseErr)
		}
		policy.Rows = parsed.rows
		policy.CategoryDenominators = parsed.categories
		policy.GoVersion = parsed.goVersion
		policy.RepositoryWrites = parsed.repositoryWrites
		policy.LocalTestExecutions = parsed.localTestExecutions
		policy.SecurityEvidenceLinks = parsed.securityEvidenceLinks
		policy.RootReadmeVisualPolicy = parsed.rootReadmeVisualPolicy
	}
	if policy.ActivityCount != 1 {
		return Policy{}, fmt.Errorf("public trust policy activity count = %d, want 1", policy.ActivityCount)
	}
	policy.SourceDigest = cache.HashBytes(source).String()
	policy.EvaluatorDigest = evaluatorDigest(policy.Rows)
	if err := validatePolicy(policy); err != nil {
		return Policy{}, err
	}
	return policy, nil
}

func GenerateNamed(filename string, source []byte) (Policy, []byte, error) {
	policy, err := CompileNamed(filename, source)
	if err != nil {
		return Policy{}, nil, err
	}
	generated, err := renderGenerated(filename, policy)
	if err != nil {
		return Policy{}, nil, err
	}
	return policy, generated, nil
}

func Load(filename string, source []byte) (Policy, error) {
	policy, err := CompileNamed(filename, source)
	if err != nil {
		return Policy{}, err
	}
	if policy.SourceDigest != generatedpolicy.PolicySourceDigest || policy.EvaluatorDigest != generatedpolicy.EvaluatorDigest {
		return Policy{}, errors.New("public trust policy and generated evaluator digest differ")
	}
	rows := generatedpolicy.Rows()
	if len(rows) != len(policy.Rows) {
		return Policy{}, fmt.Errorf("public trust policy and generated evaluator row counts differ: %d/%d", len(policy.Rows), len(rows))
	}
	for index, row := range policy.Rows {
		generated := rows[index]
		if row != (Badge{
			ID: generated.ID, Category: generated.Category, Claim: generated.Claim,
			Semantics: generated.Semantics, ImageURL: generated.ImageURL,
			TargetURL: generated.TargetURL, EvidenceURL: generated.EvidenceURL,
			State: generated.State, Render: generated.Render,
		}) {
			return Policy{}, fmt.Errorf("generated public trust evaluator row %d is not source-bound", index+1)
		}
	}
	return policy, nil
}

func GeneratedEvaluatorDigest() string { return generatedpolicy.EvaluatorDigest }

func SummaryFor(policy Policy, readmeDrift, securityPlaceholderClaims int) Summary {
	summary := Summary{
		Schema: Schema, PolicySourcePath: policy.SourcePath,
		PolicySourceDigest: policy.SourceDigest, GeneratedEvaluatorDigest: policy.EvaluatorDigest,
		MetaRowsBound: len(policy.Rows), MetaRowsTotal: ExpectedMetaRows,
		READMEGeneratedDrift: readmeDrift, SECURITYPlaceholderClaims: securityPlaceholderClaims,
		SecurityEvidenceLinks: policy.SecurityEvidenceLinks, RepositoryWrites: policy.RepositoryWrites,
		LocalTestExecutions: policy.LocalTestExecutions, CategoryRows: map[string]int{}, StateRows: map[string]int{},
	}
	claims := make(map[string]struct{}, len(policy.Rows))
	targets := make(map[string]struct{}, len(policy.Rows))
	for _, row := range policy.Rows {
		claims[row.Claim] = struct{}{}
		targets[row.TargetURL] = struct{}{}
		summary.CategoryRows[row.Category]++
		summary.StateRows[row.State]++
		if row.Render {
			summary.TotalActiveBadges++
			if row.State != DecisionClosed {
				summary.UnsupportedBadges++
			}
		}
	}
	summary.CategoryCount = len(summary.CategoryRows)
	summary.UniqueClaims = len(claims)
	summary.UniqueTargets = len(targets)
	summary.DuplicateClaims = len(policy.Rows) - summary.UniqueClaims
	summary.DuplicateTargets = len(policy.Rows) - summary.UniqueTargets
	return summary
}

func RenderBadgeBlock(policy Policy) string {
	var builder strings.Builder
	builder.WriteString("<!-- PUBLIC-TRUST-BADGES:BEGIN -->\n")
	builder.WriteString("### Public trust surface\n\n")
	builder.WriteString("These badges are generated from the lowered public-trust `.gooo` policy. Workflow badges report workflow results; they do not claim branch-protection or ruleset enforcement.\n\n")
	seenCategories := make(map[string]bool)
	for _, row := range policy.Rows {
		if seenCategories[row.Category] {
			continue
		}
		seenCategories[row.Category] = true
		builder.WriteString("#### ")
		builder.WriteString(row.Category)
		builder.WriteString("\n\n")
		for _, candidate := range policy.Rows {
			if candidate.Category != row.Category || !candidate.Render {
				continue
			}
			fmt.Fprintf(&builder, "[![%s](%s)](%s)\n", candidate.Claim, candidate.ImageURL, candidate.TargetURL)
		}
		builder.WriteByte('\n')
	}
	builder.WriteString("The complete row ledger, including unavailable and refuted claims, is emitted by the `Public trust surface` workflow.\n")
	builder.WriteString("<!-- PUBLIC-TRUST-BADGES:END -->\n")
	return builder.String()
}

func RenderReport(policy Policy, summary Summary) string {
	var builder strings.Builder
	fmt.Fprintf(&builder, "# Public trust surface report\n\n")
	fmt.Fprintf(&builder, "- policy: `%s`\n- source digest: `%s`\n- generated evaluator digest: `%s`\n\n", policy.SourcePath, policy.SourceDigest, policy.EvaluatorDigest)
	builder.WriteString("## Exact metrics\n\n")
	builder.WriteString("| Metric | Value |\n| --- | ---: |\n")
	fmt.Fprintf(&builder, "| badge category count | %d |\n| total active badges | %d |\n| meta rows bound / total | %d / %d |\n| unique claims | %d |\n| unique targets | %d |\n| unsupported badges | %d |\n| duplicate claims | %d |\n| duplicate targets | %d |\n| README generated drift | %d |\n| SECURITY placeholder claims | %d |\n| security evidence links | %d |\n| CI repository writes | %d |\n| local test executions | %d |\n", summary.CategoryCount, summary.TotalActiveBadges, summary.MetaRowsBound, summary.MetaRowsTotal, summary.UniqueClaims, summary.UniqueTargets, summary.UnsupportedBadges, summary.DuplicateClaims, summary.DuplicateTargets, summary.READMEGeneratedDrift, summary.SECURITYPlaceholderClaims, summary.SecurityEvidenceLinks, summary.RepositoryWrites, summary.LocalTestExecutions)
	builder.WriteString("\n## Category counts\n\n| Category | Rows |\n| --- | ---: |\n")
	for _, category := range policy.CategoryDenominators {
		fmt.Fprintf(&builder, "| %s | %d |\n", category.Category, category.Count)
	}
	builder.WriteString("\n## State counts\n\n| State | Rows |\n| --- | ---: |\n")
	for _, state := range []string{DecisionClosed, DecisionUnknown, DecisionRefuted} {
		fmt.Fprintf(&builder, "| %s | %d |\n", state, summary.StateRows[state])
	}
	builder.WriteString("\n## Security evidence links\n\n")
	for _, row := range policy.Rows {
		if strings.HasPrefix(row.Category, "Security / Supply Chain") {
			fmt.Fprintf(&builder, "- [%s](%s) — `%s`\n", row.ID, row.EvidenceURL, row.State)
		}
	}
	builder.WriteString("\nThe v19 root README change targets `dev`; the default `main` page receives it only after a later legitimate protected-main promotion.\n")
	builder.WriteString("\nThe report intentionally has no aggregate quality score. `UNKNOWN` and `REFUTED` rows remain visible in the manifest and are excluded from the rendered README badges.\n")
	return builder.String()
}

type parsedMarkers struct {
	rows                   []Badge
	categories             []CategoryCount
	goVersion              string
	repositoryWrites       int
	localTestExecutions    int
	securityEvidenceLinks  int
	rootReadmeVisualPolicy string
}

func parseMarkers(value string) (parsedMarkers, error) {
	var parsed parsedMarkers
	seenHead := false
	for part := range strings.SplitSeq(value, ";") {
		if part == "" {
			continue
		}
		switch {
		case part == PolicyHead:
			if seenHead {
				return parsedMarkers{}, errors.New("public trust policy header is duplicated")
			}
			seenHead = true
		case strings.HasPrefix(part, "meta-row-denominator="):
			if err := parseExpectedInt("meta-row-denominator", part, ExpectedMetaRows); err != nil {
				return parsedMarkers{}, err
			}
		case strings.HasPrefix(part, "active-badge-denominator="):
			if err := parseExpectedInt("active-badge-denominator", part, ExpectedActiveBadges); err != nil {
				return parsedMarkers{}, err
			}
		case strings.HasPrefix(part, "category-denominator="):
			encoded := strings.TrimPrefix(part, "category-denominator=")
			category, count, ok := strings.Cut(encoded, "|")
			if !ok || category == "" {
				return parsedMarkers{}, fmt.Errorf("category denominator %q is malformed", encoded)
			}
			parsedCount, err := strconv.Atoi(count)
			if err != nil || parsedCount < 1 {
				return parsedMarkers{}, fmt.Errorf("category denominator %q is malformed", encoded)
			}
			parsed.categories = append(parsed.categories, CategoryCount{Category: category, Count: parsedCount})
		case strings.HasPrefix(part, "workflow-go-version="):
			parsed.goVersion = strings.TrimPrefix(part, "workflow-go-version=")
		case strings.HasPrefix(part, "repository-writes="):
			value := strings.TrimPrefix(part, "repository-writes=")
			parsed.repositoryWrites, _ = strconv.Atoi(value)
		case strings.HasPrefix(part, "local-test-executions="):
			value := strings.TrimPrefix(part, "local-test-executions=")
			parsed.localTestExecutions, _ = strconv.Atoi(value)
		case strings.HasPrefix(part, "security-evidence-links="):
			value := strings.TrimPrefix(part, "security-evidence-links=")
			parsed.securityEvidenceLinks, _ = strconv.Atoi(value)
		case strings.HasPrefix(part, "root-readme-visual-sections="):
			parsed.rootReadmeVisualPolicy = strings.TrimPrefix(part, "root-readme-visual-sections=")
		case strings.HasPrefix(part, "badge-row="):
			row, err := parseBadge(strings.TrimPrefix(part, "badge-row="))
			if err != nil {
				return parsedMarkers{}, err
			}
			parsed.rows = append(parsed.rows, row)
		default:
			return parsedMarkers{}, fmt.Errorf("unknown public trust marker %q", part)
		}
	}
	if !seenHead {
		return parsedMarkers{}, errors.New("public trust policy header is missing")
	}
	return parsed, nil
}

func parseExpectedInt(name, marker string, expected int) error {
	value, err := strconv.Atoi(strings.TrimPrefix(marker, name+"="))
	if err != nil || value != expected {
		return fmt.Errorf("%s = %q, want %d", name, strings.TrimPrefix(marker, name+"="), expected)
	}
	return nil
}

func parseBadge(encoded string) (Badge, error) {
	fields := strings.Split(encoded, "|")
	if len(fields) != 9 {
		return Badge{}, fmt.Errorf("badge row has %d fields, want 9", len(fields))
	}
	row := Badge{ID: fields[0], Category: fields[1], Claim: fields[2], Semantics: fields[3], ImageURL: fields[4], TargetURL: fields[5], EvidenceURL: fields[6], State: fields[7]}
	render, err := strconv.ParseBool(fields[8])
	if err != nil {
		return Badge{}, fmt.Errorf("badge %q render flag is malformed", row.ID)
	}
	row.Render = render
	if row.ID == "" || row.Category == "" || row.Claim == "" || row.Semantics == "" {
		return Badge{}, errors.New("badge row has an empty identity or claim field")
	}
	if row.State != DecisionClosed && row.State != DecisionUnknown && row.State != DecisionRefuted {
		return Badge{}, fmt.Errorf("badge %q has unknown state %q", row.ID, row.State)
	}
	for _, value := range []string{row.ImageURL, row.TargetURL, row.EvidenceURL} {
		parsed, err := url.Parse(value)
		if err != nil || (parsed.Scheme != "https" && parsed.Scheme != "http") || parsed.Host == "" {
			return Badge{}, fmt.Errorf("badge %q has invalid URL %q", row.ID, value)
		}
	}
	return row, nil
}

func validatePolicy(policy Policy) error {
	if len(policy.Rows) != ExpectedMetaRows {
		return fmt.Errorf("public trust meta row count = %d, want %d", len(policy.Rows), ExpectedMetaRows)
	}
	if policy.GoVersion != ExpectedGoVersion || policy.RepositoryWrites != ExpectedRepositoryWrites || policy.LocalTestExecutions != ExpectedLocalTestRuns || policy.RootReadmeVisualPolicy != "preserve" {
		return errors.New("public trust policy execution or root README boundary is not canonical")
	}
	if len(policy.CategoryDenominators) != ExpectedCategoryCount {
		return fmt.Errorf("public trust category count = %d, want %d", len(policy.CategoryDenominators), ExpectedCategoryCount)
	}
	categoryRows := make(map[string]int)
	seenIDs := make(map[string]bool)
	seenCategories := make(map[string]bool)
	for _, row := range policy.Rows {
		if seenIDs[row.ID] {
			return fmt.Errorf("duplicate public trust badge ID %q", row.ID)
		}
		seenIDs[row.ID] = true
		categoryRows[row.Category]++
		if row.Render && row.State != DecisionClosed {
			return fmt.Errorf("unsupported public trust badge %q is marked renderable", row.ID)
		}
	}
	active := 0
	for _, row := range policy.Rows {
		if row.Render {
			active++
		}
	}
	if active != ExpectedActiveBadges {
		return fmt.Errorf("public trust active badge count = %d, want %d", active, ExpectedActiveBadges)
	}
	denominatorRows := 0
	for _, denominator := range policy.CategoryDenominators {
		if seenCategories[denominator.Category] {
			return fmt.Errorf("duplicate public trust category %q", denominator.Category)
		}
		seenCategories[denominator.Category] = true
		if categoryRows[denominator.Category] != denominator.Count {
			return fmt.Errorf("public trust category %q count = %d, want %d", denominator.Category, categoryRows[denominator.Category], denominator.Count)
		}
		denominatorRows += denominator.Count
	}
	if denominatorRows != ExpectedMetaRows || len(categoryRows) != ExpectedCategoryCount {
		return errors.New("public trust category denominators do not cover all meta rows")
	}
	if policy.SecurityEvidenceLinks != 8 {
		return fmt.Errorf("public trust security evidence link count = %d, want 8", policy.SecurityEvidenceLinks)
	}
	return nil
}

func evaluatorDigest(rows []Badge) string {
	var builder strings.Builder
	builder.WriteString(Schema)
	builder.WriteByte('\n')
	for _, row := range rows {
		builder.WriteString(strings.Join([]string{row.ID, row.Category, row.Claim, row.Semantics, row.ImageURL, row.TargetURL, row.EvidenceURL, row.State, strconv.FormatBool(row.Render)}, "\x00"))
		builder.WriteByte('\n')
	}
	return cache.HashBytes([]byte(builder.String())).String()
}

func renderGenerated(filename string, policy Policy) ([]byte, error) {
	var builder strings.Builder
	fmt.Fprintln(&builder, "// Code generated by the public trust surface generator; DO NOT EDIT.")
	fmt.Fprintln(&builder)
	fmt.Fprintln(&builder, "package publictrustgenerated")
	fmt.Fprintln(&builder)
	fmt.Fprintln(&builder, "const (")
	fmt.Fprintf(&builder, "\tSchema = %q\n", Schema)
	fmt.Fprintf(&builder, "\tPolicySourcePath = %q\n", filename)
	fmt.Fprintf(&builder, "\tPolicySourceDigest = %q\n", policy.SourceDigest)
	fmt.Fprintf(&builder, "\tEvaluatorDigest = %q\n", policy.EvaluatorDigest)
	fmt.Fprintln(&builder, ")")
	fmt.Fprintln(&builder)
	fmt.Fprintln(&builder, "type Badge struct {")
	fmt.Fprintln(&builder, "\tID, Category, Claim, Semantics, ImageURL, TargetURL, EvidenceURL, State string")
	fmt.Fprintln(&builder, "\tRender bool")
	fmt.Fprintln(&builder, "}")
	fmt.Fprintln(&builder)
	fmt.Fprintln(&builder, "var rows = [...]Badge{")
	for _, row := range policy.Rows {
		fmt.Fprintf(&builder, "\t{ID: %q, Category: %q, Claim: %q, Semantics: %q, ImageURL: %q, TargetURL: %q, EvidenceURL: %q, State: %q, Render: %t},\n", row.ID, row.Category, row.Claim, row.Semantics, row.ImageURL, row.TargetURL, row.EvidenceURL, row.State, row.Render)
	}
	fmt.Fprintln(&builder, "}")
	fmt.Fprintln(&builder)
	fmt.Fprintln(&builder, "func Rows() []Badge { return append([]Badge(nil), rows[:]...) }")
	return format.Source([]byte(builder.String()))
}
