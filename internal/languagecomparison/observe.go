package languagecomparison

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
	"runtime"
	"strings"
	"time"

	"github.com/kimjooyoon/meta-ontology-go/internal/sourceexecution"
)

const comparisonScope = "DECLARATION_SIGNATURE_PARSE_AND_RESOLUTION"

func ObserveRuntime(request Request) Receipt {
	receipt := Receipt{
		Schema: ReceiptSchema, ContractID: ContractID, SubjectSHA: request.SubjectSHA,
		ExecutableDigest: request.ExecutableDigest, Decision: "FAIL_CLOSED", Resolution: "EXACT",
		Reason: "COMPARISON_REQUEST_INVALID", Scope: comparisonScope,
		GoooFilename: request.GoooFilename, GoFilename: request.GoFilename, Entry: request.Entry,
		GoooSourceDigest: digestBytes([]byte(request.GoooSource)), GoSourceDigest: digestBytes([]byte(request.GoSource)),
		Runner: Runner{GoVersion: runtime.Version(), OS: runtime.GOOS, Arch: runtime.GOARCH,
			CPUs: runtime.NumCPU(), Label: request.RunnerLabel},
		Samples: []Sample{}, Effects: Effects{}, NotClaimed: defaultNonClaims(),
	}
	if !validSubject(request.SubjectSHA) || !validDigest(request.ExecutableDigest) ||
		strings.TrimSpace(request.RunnerLabel) == "" || strings.TrimSpace(request.GoooFilename) == "" ||
		strings.TrimSpace(request.GoFilename) == "" || strings.TrimSpace(request.Entry) == "" ||
		request.GoooSource == "" || request.GoSource == "" || request.Samples != SamplesPerLanguage {
		return seal(receipt)
	}
	goooRequest := sourceexecution.Request{
		Filename: request.GoooFilename, Source: request.GoooSource, Entry: request.Entry,
	}
	if gooo := sourceexecution.Execute(goooRequest); gooo.Decision != "PASS" {
		receipt.Reason = "GOOO_" + gooo.Reason
		return seal(receipt)
	}
	goEntry, err := resolveGoEntry(request.GoFilename, request.GoSource, request.Entry)
	if err != nil {
		receipt.Reason = "GO_BASELINE_" + err.Error()
		return seal(receipt)
	}
	goooWarm := sourceexecution.Execute(goooRequest)
	goWarm, err := resolveGoEntry(request.GoFilename, request.GoSource, request.Entry)
	if goooWarm.Decision != "PASS" || err != nil ||
		!reflect.DeepEqual(goooWarm.Entry, goEntry) || !reflect.DeepEqual(goWarm, goEntry) {
		receipt.Reason = "COMPARISON_WARMUP_DIVERGED"
		return seal(receipt)
	}

	for sequence := 1; sequence <= request.Samples; sequence++ {
		sample := Sample{Sequence: sequence, GoooDecision: "PASS", GoDecision: "PASS"}
		var gooo sourceexecution.Receipt
		var baseline sourceexecution.Entry
		var goErr error
		if sequence%2 == 1 {
			sample.FirstMeasured = "gooo"
			gooo, sample.Gooo = measure(func() sourceexecution.Receipt { return sourceexecution.Execute(goooRequest) })
			baselineResult, measured := measure(func() goResult {
				entry, resolveErr := resolveGoEntry(request.GoFilename, request.GoSource, request.Entry)
				return goResult{entry: entry, err: resolveErr}
			})
			baseline, sample.Go, goErr = baselineResult.entry, measured, baselineResult.err
		} else {
			sample.FirstMeasured = "go"
			baselineResult, measured := measure(func() goResult {
				entry, resolveErr := resolveGoEntry(request.GoFilename, request.GoSource, request.Entry)
				return goResult{entry: entry, err: resolveErr}
			})
			baseline, sample.Go, goErr = baselineResult.entry, measured, baselineResult.err
			gooo, sample.Gooo = measure(func() sourceexecution.Receipt { return sourceexecution.Execute(goooRequest) })
		}
		if gooo.Decision != "PASS" {
			sample.GoooDecision = "FAIL_CLOSED"
		}
		if goErr != nil {
			sample.GoDecision = "FAIL_CLOSED"
		}
		sample.GoooOutputDigest = digestValue(gooo.Entry)
		sample.GoOutputDigest = digestValue(baseline)
		receipt.Samples = append(receipt.Samples, sample)
		if sample.GoooDecision != "PASS" || sample.GoDecision != "PASS" || sample.GoooOutputDigest != sample.GoOutputDigest {
			receipt.Reason = "COMPARISON_OUTPUT_MISMATCH"
			return seal(receipt)
		}
	}
	receipt.Summary = summarize(request.Samples, receipt.Samples)
	if receipt.Summary.EquivalentOutputSamples != request.Samples || receipt.Summary.GoooOutputDigestVariants != 1 ||
		receipt.Summary.GoOutputDigestVariants != 1 || receipt.Summary.Gooo.WallMedianNanoseconds <= 0 ||
		receipt.Summary.Go.WallMedianNanoseconds <= 0 || receipt.Summary.Gooo.TotalAllocMedianBytes == 0 ||
		receipt.Summary.Go.TotalAllocMedianBytes == 0 {
		receipt.Reason = "COMPARISON_MEASUREMENT_UNKNOWN"
		return seal(receipt)
	}
	receipt.Decision, receipt.Resolution, receipt.Reason = "PASS", "RUNNER_SCOPED",
		"EQUIVALENT_DECLARATION_SIGNATURE_OBSERVED"
	return seal(receipt)
}

type goResult struct {
	entry sourceexecution.Entry
	err   error
}

func measure[T any](run func() T) (T, Measurement) {
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	started := time.Now()
	result := run()
	measurement := Measurement{WallNanoseconds: time.Since(started).Nanoseconds()}
	runtime.ReadMemStats(&after)
	if after.TotalAlloc >= before.TotalAlloc {
		measurement.TotalAllocBytes = after.TotalAlloc - before.TotalAlloc
	}
	return result, measurement
}

func resolveGoEntry(filename, source, requested string) (sourceexecution.Entry, error) {
	file, err := parser.ParseFile(token.NewFileSet(), filename, source, parser.ParseComments|parser.AllErrors)
	if err != nil {
		return sourceexecution.Entry{}, fmt.Errorf("SOURCE_INVALID")
	}
	if file.Name == nil || file.Name.Name == "" {
		return sourceexecution.Entry{}, fmt.Errorf("PACKAGE_UNKNOWN")
	}
	namespace := directive(file.Doc, "gooo:namespace")
	if namespace == "" {
		return sourceexecution.Entry{}, fmt.Errorf("NAMESPACE_UNKNOWN")
	}
	typeIDs := map[string]string{}
	for _, declaration := range file.Decls {
		group, ok := declaration.(*ast.GenDecl)
		if !ok || group.Tok != token.TYPE {
			continue
		}
		for _, spec := range group.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok || typeSpec.Name == nil {
				continue
			}
			doc := typeSpec.Doc
			if doc == nil && len(group.Specs) == 1 {
				doc = group.Doc
			}
			id := directive(doc, "gooo:id")
			if id == "" || typeIDs[typeSpec.Name.Name] != "" {
				return sourceexecution.Entry{}, fmt.Errorf("TYPE_IDENTITY_UNKNOWN")
			}
			typeIDs[typeSpec.Name.Name] = id
		}
	}
	var function *ast.FuncDecl
	for _, declaration := range file.Decls {
		candidate, ok := declaration.(*ast.FuncDecl)
		if ok && candidate.Recv == nil && candidate.Name != nil && candidate.Name.Name == requested {
			if function != nil {
				return sourceexecution.Entry{}, fmt.Errorf("ENTRY_AMBIGUOUS")
			}
			function = candidate
		}
	}
	if function == nil || function.Type == nil || function.Type.Params == nil || function.Type.Results == nil ||
		len(function.Type.Results.List) != 1 {
		return sourceexecution.Entry{}, fmt.Errorf("ENTRY_UNKNOWN")
	}
	entry := sourceexecution.Entry{Package: file.Name.Name, Namespace: namespace, Activity: requested,
		Inputs: []sourceexecution.Binding{}}
	for _, field := range function.Type.Params.List {
		name, ok := field.Type.(*ast.Ident)
		if !ok || len(field.Names) == 0 {
			return sourceexecution.Entry{}, fmt.Errorf("INPUT_SIGNATURE_UNKNOWN")
		}
		id := typeIDs[name.Name]
		if id == "" {
			return sourceexecution.Entry{}, fmt.Errorf("INPUT_IDENTITY_UNKNOWN")
		}
		for range field.Names {
			entry.Inputs = append(entry.Inputs, sourceexecution.Binding{Name: name.Name, ID: id})
		}
	}
	output, ok := function.Type.Results.List[0].Type.(*ast.Ident)
	if !ok || len(function.Type.Results.List[0].Names) > 0 {
		return sourceexecution.Entry{}, fmt.Errorf("OUTPUT_SIGNATURE_UNKNOWN")
	}
	outputID := typeIDs[output.Name]
	if outputID == "" {
		return sourceexecution.Entry{}, fmt.Errorf("OUTPUT_IDENTITY_UNKNOWN")
	}
	entry.Output = sourceexecution.Binding{Name: output.Name, ID: outputID}
	return entry, nil
}

func directive(group *ast.CommentGroup, name string) string {
	if group == nil {
		return ""
	}
	prefix := name + " "
	for _, line := range strings.Split(group.Text(), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, prefix) {
			return strings.TrimSpace(strings.TrimPrefix(line, prefix))
		}
	}
	return ""
}
