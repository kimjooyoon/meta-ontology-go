// Package linecapscontract defines a small, standard-library-only oracle
// contract for line-cap experiments. It measures source and AST facts without
// claiming that the future gooo-hosted verifier is implemented.
package linecapscontract

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"sort"
)

const (
	// Schema is the versioned evidence envelope consumed by later adapters.
	Schema = "gooo/linecaps-evidence/v1"
	// Producer is stable across implementation refactors of this oracle.
	Producer = "gooo://detector/linecaps/ast"
)

// Decision is the result of the local line-cap criterion.
type Decision string

const (
	Pass     Decision = "pass"
	Fail     Decision = "fail"
	Deferred Decision = "deferred"
)

// Rule identifies a measurable policy outcome.
type Rule string

const (
	RuleFileLines     Rule = "file-lines"
	RuleFunctionLines Rule = "function-lines"
	RuleParseFile     Rule = "parse-file"
)

// Limits are inclusive. Values exactly at a limit pass the local criterion.
type Limits struct {
	MaxFileLines     int `json:"max_file_lines"`
	MaxFunctionLines int `json:"max_function_lines"`
}

// DefaultLimits is the current DAMP/DRY policy under experiment.
func DefaultLimits() Limits {
	return Limits{MaxFileLines: 300, MaxFunctionLines: 75}
}

func (l Limits) valid() error {
	if l.MaxFileLines <= 0 || l.MaxFunctionLines <= 0 {
		return fmt.Errorf("linecaps limits must be positive")
	}
	return nil
}

// Case declares a falsifiable experiment. DeferredReason represents a later
// semantic or host-parity criterion and never turns a deferred result into a
// pass.
type Case struct {
	ID             string   `json:"id"`
	Hypothesis     string   `json:"hypothesis"`
	Fixture        string   `json:"fixture"`
	Limits         Limits   `json:"limits"`
	Expected       Decision `json:"expected,omitempty"`
	DeferredReason string   `json:"deferred_reason,omitempty"`
}

func (c Case) valid() error {
	if c.ID == "" || c.Hypothesis == "" || c.Fixture == "" {
		return fmt.Errorf("linecaps case requires id, hypothesis, and fixture")
	}
	if err := c.Limits.valid(); err != nil {
		return err
	}
	if c.Expected != "" && c.Expected != Pass && c.Expected != Fail && c.Expected != Deferred {
		return fmt.Errorf("linecaps case %q has invalid expected decision %q", c.ID, c.Expected)
	}
	return nil
}

// FunctionSpan is a source-backed AST measurement, including both endpoints.
type FunctionSpan struct {
	Name      string `json:"name"`
	StartLine int    `json:"start_line"`
	EndLine   int    `json:"end_line"`
	Lines     int    `json:"lines"`
}

// Measurement contains facts that AST, IR, codegen, and CI adapters can reuse.
type Measurement struct {
	SourceDigest string         `json:"source_digest"`
	FileLines    int            `json:"file_lines"`
	Parseable    bool           `json:"parseable"`
	ParseError   string         `json:"parse_error,omitempty"`
	Functions    []FunctionSpan `json:"functions"`
}

// Finding records why a local criterion failed.
type Finding struct {
	Rule      Rule   `json:"rule"`
	Name      string `json:"name,omitempty"`
	StartLine int    `json:"start_line,omitempty"`
	EndLine   int    `json:"end_line,omitempty"`
	Actual    int    `json:"actual,omitempty"`
	Limit     int    `json:"limit,omitempty"`
	Detail    string `json:"detail,omitempty"`
}

// Evidence is the deterministic output envelope for one experiment case.
// OutcomeMatches is the only comparison against the case expectation; it is
// not a self-hosting or promotion claim.
type Evidence struct {
	Schema         string      `json:"schema"`
	Producer       string      `json:"producer"`
	CaseID         string      `json:"case_id"`
	Fixture        string      `json:"fixture"`
	Hypothesis     string      `json:"hypothesis"`
	Limits         Limits      `json:"limits"`
	Decision       Decision    `json:"decision"`
	Expected       Decision    `json:"expected,omitempty"`
	OutcomeMatches bool        `json:"outcome_matches"`
	DeferredReason string      `json:"deferred_reason,omitempty"`
	Measurement    Measurement `json:"measurement"`
	Findings       []Finding   `json:"findings"`
}

// Measure parses one source buffer and returns facts even when parsing fails.
// A parse error is evidence for a negative case, not an implementation success.
func Measure(path string, source []byte) (Measurement, error) {
	digest := sha256.Sum256(source)
	measurement := Measurement{
		SourceDigest: hex.EncodeToString(digest[:]),
		FileLines:    lineCount(source),
		Functions:    make([]FunctionSpan, 0),
	}
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, source, parser.ParseComments)
	if err != nil {
		measurement.ParseError = err.Error()
		return measurement, nil
	}
	measurement.Parseable = true
	ast.Inspect(file, func(node ast.Node) bool {
		span, ok := functionSpan(fset, node)
		if ok {
			measurement.Functions = append(measurement.Functions, span)
		}
		return true
	})
	sort.Slice(measurement.Functions, func(i, j int) bool {
		left, right := measurement.Functions[i], measurement.Functions[j]
		if left.StartLine != right.StartLine {
			return left.StartLine < right.StartLine
		}
		return left.EndLine < right.EndLine
	})
	return measurement, nil
}

// EvaluateCase applies only the declared line-cap criterion. A valid result
// with DeferredReason is explicitly deferred after local checks pass.
func EvaluateCase(c Case, path string, source []byte) (Evidence, error) {
	if err := c.valid(); err != nil {
		return Evidence{}, err
	}
	measurement, err := Measure(path, source)
	if err != nil {
		return Evidence{}, err
	}
	findings := findingsFor(measurement, c.Limits)
	decision := Pass
	if len(findings) > 0 {
		decision = Fail
	} else if c.DeferredReason != "" {
		decision = Deferred
	}
	matches := c.Expected == "" || c.Expected == decision
	return Evidence{
		Schema: Schema, Producer: Producer, CaseID: c.ID, Fixture: c.Fixture,
		Hypothesis: c.Hypothesis, Limits: c.Limits, Decision: decision,
		Expected: c.Expected, OutcomeMatches: matches,
		DeferredReason: c.DeferredReason, Measurement: measurement, Findings: findings,
	}, nil
}

// EvaluateCases sorts cases by stable ID and returns deterministic evidence.
// Sources are keyed by Case.Fixture so callers can load fixtures independently.
func EvaluateCases(cases []Case, sources map[string][]byte) ([]Evidence, error) {
	ordered := append([]Case(nil), cases...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].ID < ordered[j].ID })
	evidence := make([]Evidence, 0, len(ordered))
	for i, experiment := range ordered {
		if i > 0 && ordered[i-1].ID == experiment.ID {
			return nil, fmt.Errorf("duplicate linecaps case %q", experiment.ID)
		}
		source, ok := sources[experiment.Fixture]
		if !ok {
			return nil, fmt.Errorf("missing linecaps fixture %q", experiment.Fixture)
		}
		result, err := EvaluateCase(experiment, experiment.Fixture, source)
		if err != nil {
			return nil, err
		}
		evidence = append(evidence, result)
	}
	return evidence, nil
}

// JSON serializes one evidence envelope with stable field and array ordering.
func (e Evidence) JSON() ([]byte, error) {
	data, err := json.MarshalIndent(e, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}

func lineCount(source []byte) int {
	if len(source) == 0 {
		return 0
	}
	lines := bytes.Count(source, []byte{'\n'})
	if source[len(source)-1] != '\n' {
		lines++
	}
	return lines
}

func functionSpan(fset *token.FileSet, node ast.Node) (FunctionSpan, bool) {
	name := ""
	switch function := node.(type) {
	case *ast.FuncDecl:
		name = function.Name.Name
		if function.Recv != nil {
			name = "method " + name
		}
	case *ast.FuncLit:
		name = "function literal"
	default:
		return FunctionSpan{}, false
	}
	start := fset.Position(node.Pos()).Line
	end := fset.Position(node.End()).Line
	return FunctionSpan{Name: name, StartLine: start, EndLine: end, Lines: end - start + 1}, true
}

func findingsFor(measurement Measurement, limits Limits) []Finding {
	findings := make([]Finding, 0)
	if measurement.FileLines > limits.MaxFileLines {
		findings = append(findings, Finding{Rule: RuleFileLines, Actual: measurement.FileLines, Limit: limits.MaxFileLines})
	}
	if !measurement.Parseable {
		findings = append(findings, Finding{Rule: RuleParseFile, Detail: measurement.ParseError})
		return findings
	}
	for _, function := range measurement.Functions {
		if function.Lines > limits.MaxFunctionLines {
			findings = append(findings, Finding{Rule: RuleFunctionLines, Name: function.Name, StartLine: function.StartLine, EndLine: function.EndLine, Actual: function.Lines, Limit: limits.MaxFunctionLines})
		}
	}
	return findings
}
