// Package bodycodegen lowers a deliberately small, pure activity-body language
// from .gooo source into a deterministic Go projection.
package bodycodegen

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/scanner"
	"go/token"
	"go/types"
	"maps"
	"math"
	"strings"
	"time"

	"github.com/kimjooyoon/meta-ontology-go/internal/decisionroute"
)

const (
	schema              = "gooo/body-codegen-report/v3"
	preserveRoute       = "preserve"
	guardReturnRoute    = "guard-return"
	mergeResultRoute    = "merge-result"
	routeDecisionBudget = 3 * time.Second
	seededSelectionV1   = "sha256_seeded_weighted_choice/v1"
)

// Result is an immutable report and generated source projection.
type Result struct {
	Report     Report `json:"report"`
	Source     string `json:"source"`
	GoooSource string `json:"gooo_source,omitempty"`
}

// Report describes what was lowered and how the output was checked.
type Report struct {
	Schema                 string                  `json:"schema"`
	Decision               string                  `json:"decision"`
	Activity               string                  `json:"activity"`
	ActivityID             string                  `json:"activity_id"`
	InputType              string                  `json:"input_type"`
	InputParameters        []InputParameter        `json:"input_parameters,omitempty"`
	RecordTypes            []RecordType            `json:"record_types,omitempty"`
	OutputType             string                  `json:"output_type"`
	PlanSHA256             string                  `json:"plan_sha256"`
	CompilerSourceSHA      string                  `json:"compiler_source_sha"`
	SourceDigest           string                  `json:"source_digest"`
	ProgramDigest          string                  `json:"program_digest"`
	GeneratedDigest        string                  `json:"generated_digest"`
	ReplayDigest           string                  `json:"replay_digest"`
	Route                  string                  `json:"route"`
	RouteDecision          decisionroute.Receipt   `json:"route_decision"`
	RouteSelection         RouteSelectionReceipt   `json:"route_selection"`
	RouteDecisionLatencyMS float64                 `json:"route_decision_latency_ms"`
	BodyFill               *IRBodyFillReceipt      `json:"body_fill,omitempty"`
	BodySearch             *IRBodySearchReceipt    `json:"body_search,omitempty"`
	BodyPaths              *BodyPathReceipt        `json:"body_paths,omitempty"`
	RecordAssembly         *RecordAssemblyReceipt  `json:"record_assembly,omitempty"`
	CandidateRoutes        []string                `json:"candidate_routes"`
	EquivalenceRule        string                  `json:"equivalence_rule"`
	SourceConstructs       int                     `json:"source_constructs"`
	LoweredConstructs      int                     `json:"lowered_constructs"`
	SourceSemanticUnits    int                     `json:"source_semantic_units"`
	LoweredSemanticUnits   int                     `json:"lowered_semantic_units"`
	CompletenessPercent    float64                 `json:"completeness_percent"`
	RouteEquivalence       RouteEquivalenceReceipt `json:"route_equivalence"`
	TypecheckPassed        bool                    `json:"typecheck_passed"`
	DeterministicReplay    bool                    `json:"deterministic_replay"`
	RepositoryWrites       int                     `json:"repository_writes"`
	UnsupportedConstructs  string                  `json:"unsupported_constructs"`
	CompletenessReceipt    *CompletenessReceipt    `json:"completeness_receipt,omitempty"`
}

// RouteSelectionReceipt makes an optional weighted route draw replayable without
// exposing the caller's seed. The weights are the normalized distribution used
// for this draw; the Gooo emitter still validates the selected route.
type RouteSelectionReceipt struct {
	Method        string             `json:"method"`
	SeedSHA256    string             `json:"seed_sha256,omitempty"`
	DrawHex       string             `json:"draw_hex,omitempty"`
	Weights       map[string]float64 `json:"weights,omitempty"`
	ProposedRoute string             `json:"proposed_route"`
	FinalRoute    string             `json:"final_route"`
}

// Generate compiles one .gooo activity body into a marked Go source region.
// The accepted body subset is local declarations/assignments, conditionals,
// and returns over scalar or declared record values. Calls and effects fail closed.
func Generate(filename string, source []byte, activityName string) (Result, error) {
	return GenerateWithPlanner(context.Background(), filename, source, activityName, "", "")
}

// GenerateWithPlanner asks Laya to choose only among source-shape-preserving
// lowering routes that are valid for this activity. With no usable provider,
// decisionroute selects the declared deterministic fallback.
func GenerateWithPlanner(ctx context.Context, filename string, source []byte, activityName, endpoint, apiKey string) (Result, error) {
	return GenerateWithPlannerAndSampleSeed(ctx, filename, source, activityName, endpoint, apiKey, "")
}

// GenerateWithPlannerAndSampleSeed behaves like GenerateWithPlanner unless a
// non-empty sampleSeed is supplied. In that case, Laya probabilities (when
// available) or an equal prior over eligible routes drive a replayable draw.
func GenerateWithPlannerAndSampleSeed(ctx context.Context, filename string, source []byte, activityName, endpoint, apiKey, sampleSeed string) (Result, error) {
	if sampleSeed != "" && strings.TrimSpace(sampleSeed) == "" {
		return Result{}, fmt.Errorf("route sample seed must contain a non-whitespace character")
	}
	prepared, err := prepareActivityBody(filename, source, activityName)
	if err != nil {
		return Result{}, err
	}
	choice, err := chooseBodyRoute(ctx, prepared, source, endpoint, apiKey, sampleSeed)
	if err != nil {
		return Result{}, err
	}
	return prepared.selectedResult(source, choice)
}

func goTypeForEntity(entity string) (string, bool) {
	switch entity {
	case "Integer":
		return "int64", true
	case "Boolean":
		return "bool", true
	case "Text":
		return "string", true
	default:
		return "", false
	}
}

type generatedRoute struct {
	report Report
	source []byte
}

func generateRoute(packageName, activityName, activityID, inputType, outputType, body, route string) (generatedRoute, error) {
	return generateRouteParameters(packageName, activityName, activityID, []InputParameter{{Name: "input", Type: inputType}}, outputType, body, route)
}

func generateRouteParameters(packageName, activityName, activityID string, parameters []InputParameter, outputType, body, route string, records ...RecordType) (generatedRoute, error) {
	generated, sourceConstructs, loweredConstructs, sourceUnits, loweredUnits, err := renderParameters(packageName, activityName, activityID, parameters, outputType, body, route, records...)
	if err != nil {
		return generatedRoute{}, err
	}
	equivalenceRule := routeEquivalenceRule(route)
	equivalence, err := routeEquivalenceParameters(packageName, activityName, parameters, outputType, body, generated, equivalenceRule, records...)
	if err != nil {
		return generatedRoute{}, err
	}
	if !equivalence.Equivalent {
		return generatedRoute{}, fmt.Errorf("route %q failed its canonical semantic equivalence receipt", route)
	}
	completeness := 0.0
	if sourceUnits > 0 {
		covered := min(loweredUnits, sourceUnits)
		completeness = float64(covered) * 100 / float64(sourceUnits)
	}
	var inputParameters []InputParameter
	if len(parameters) > 1 {
		inputParameters = append([]InputParameter(nil), parameters...)
	}
	return generatedRoute{source: generated, report: Report{
		Schema: schema, Decision: "PASS", Activity: activityName, ActivityID: activityID,
		InputType: parameterTypeLabel(parameters), InputParameters: inputParameters, OutputType: outputType, RecordTypes: records,
		Route: route, EquivalenceRule: equivalenceRule,
		SourceConstructs: sourceConstructs, LoweredConstructs: loweredConstructs,
		SourceSemanticUnits: sourceUnits, LoweredSemanticUnits: loweredUnits,
		CompletenessPercent: completeness, RouteEquivalence: equivalence,
		TypecheckPassed: true,
	}}, nil
}

func routeEquivalenceRule(route string) string {
	switch route {
	case guardReturnRoute:
		return "if-return-else-return-to-guard-return-v1"
	case mergeResultRoute:
		return "if-return-else-return-to-explicit-result-join-v1"
	default:
		return "source-shape-preserving-v1"
	}
}

func render(packageName, activityName, activityID, inputType, outputType, body, route string) ([]byte, int, int, int, int, error) {
	return renderParameters(packageName, activityName, activityID, []InputParameter{{Name: "input", Type: inputType}}, outputType, body, route)
}

func candidateRoutes(packageName, activityName, body string) ([]decisionroute.Option, string, error) {
	options := []decisionroute.Option{{
		ID:          preserveRoute,
		Description: "Preserve the explicit source control-flow shape while lowering each accepted statement and expression directly.",
	}}
	wrapped := fmt.Sprintf("package %s\nfunc %s(input int64) int64 {\n%s\n}\n", packageName, activityName, body)
	file, err := parser.ParseFile(token.NewFileSet(), "body.goo", wrapped, parser.AllErrors|parser.ParseComments)
	if err != nil {
		return nil, "", fmt.Errorf("inspect codegen route candidates: %w", err)
	}
	function, ok := findFunction(file, activityName)
	if !ok {
		return nil, "", fmt.Errorf("activity body did not produce a function")
	}
	shape := "general-pure-body"
	if len(file.Comments) == 0 && isGuardReturnShape(function.Body) {
		shape = "single-if-else-with-one-return-per-branch"
		options = append(options, decisionroute.Option{
			ID:          guardReturnRoute,
			Description: "For a single pure if/else whose branches each return once, emit the true-branch guard followed by the false-branch return; this removes one nesting block while preserving both outcomes.",
		})
		options = append(options, decisionroute.Option{
			ID:          mergeResultRoute,
			Description: "For a single pure if/else whose branches each return once, assign both branch values to one typed result local and return it after the conditional; this makes the control-flow join explicit.",
		})
	}
	return options, shape, nil
}

func sampleRoute(seed, requestDigest string, options []decisionroute.Option, supplied map[string]float64, fallback string) (RouteSelectionReceipt, error) {
	if strings.TrimSpace(seed) == "" || requestDigest == "" || len(options) == 0 {
		return RouteSelectionReceipt{}, fmt.Errorf("seed, request digest, and eligible routes are required")
	}
	weights, err := normalizedRouteWeights(options, supplied, fallback)
	if err != nil {
		return RouteSelectionReceipt{}, err
	}
	material := []byte(seededSelectionV1 + "\x00" + seed + "\x00" + requestDigest)
	for _, option := range options {
		material = append(material, 0)
		material = append(material, option.ID...)
	}
	drawDigest := sha256.Sum256(material)
	drawValue := binary.BigEndian.Uint64(drawDigest[:8])
	draw := float64(drawValue>>11) / float64(uint64(1)<<53)

	var selected string
	cumulative := 0.0
	for _, option := range options {
		weight := weights[option.ID]
		if weight == 0 {
			continue
		}
		selected = option.ID
		cumulative += weight
		if draw < cumulative {
			break
		}
	}
	if selected == "" {
		return RouteSelectionReceipt{}, fmt.Errorf("eligible route distribution has no positive weight")
	}
	return RouteSelectionReceipt{
		Method: seededSelectionV1, SeedSHA256: digest([]byte(seed)),
		DrawHex: hex.EncodeToString(drawDigest[:8]), Weights: weights,
		ProposedRoute: selected, FinalRoute: selected,
	}, nil
}

func normalizedRouteWeights(options []decisionroute.Option, supplied map[string]float64, fallback string) (map[string]float64, error) {
	eligible := make(map[string]bool, len(options))
	for _, option := range options {
		eligible[option.ID] = true
	}
	for route := range supplied {
		if !eligible[route] {
			return nil, fmt.Errorf("route probability names ineligible route %q", route)
		}
	}

	weights := make(map[string]float64, len(options))
	if len(supplied) == 0 {
		uniform := 1.0 / float64(len(options))
		for _, option := range options {
			weights[option.ID] = uniform
		}
		return weights, nil
	}
	total := 0.0
	for _, option := range options {
		weight := supplied[option.ID]
		if math.IsNaN(weight) || math.IsInf(weight, 0) || weight < 0 {
			return nil, fmt.Errorf("route %q has an invalid probability", option.ID)
		}
		weights[option.ID] = weight
		total += weight
	}
	if total == 0 {
		if !eligible[fallback] {
			return nil, fmt.Errorf("zero route probabilities have no eligible fallback")
		}
		for _, option := range options {
			weights[option.ID] = 0
		}
		weights[fallback] = 1
		return weights, nil
	}
	for _, option := range options {
		weights[option.ID] /= total
	}
	return weights, nil
}

func isGuardReturnShape(body *ast.BlockStmt) bool {
	if body == nil || len(body.List) != 1 {
		return false
	}
	conditional, ok := body.List[0].(*ast.IfStmt)
	if !ok || conditional.Init != nil || conditional.Else == nil || len(conditional.Body.List) != 1 {
		return false
	}
	if _, ok := conditional.Body.List[0].(*ast.ReturnStmt); !ok {
		return false
	}
	otherwise, ok := conditional.Else.(*ast.BlockStmt)
	if !ok || len(otherwise.List) != 1 {
		return false
	}
	_, ok = otherwise.List[0].(*ast.ReturnStmt)
	return ok
}

func lowerGuardReturn(body *ast.BlockStmt) bool {
	if !isGuardReturnShape(body) {
		return false
	}
	conditional := body.List[0].(*ast.IfStmt)
	otherwise := conditional.Else.(*ast.BlockStmt)
	fallback := otherwise.List[0]
	conditional.Else = nil
	body.List = []ast.Stmt{conditional, fallback}
	return true
}

func lowerMergeResult(body *ast.BlockStmt, outputType string) bool {
	if !isGuardReturnShape(body) {
		return false
	}
	conditional := body.List[0].(*ast.IfStmt)
	otherwise := conditional.Else.(*ast.BlockStmt)
	thenReturn := conditional.Body.List[0].(*ast.ReturnStmt)
	elseReturn := otherwise.List[0].(*ast.ReturnStmt)
	resultName := "_goooResult"
	result := ast.NewIdent(resultName)
	declaration := &ast.DeclStmt{Decl: &ast.GenDecl{
		Tok: token.VAR,
		Specs: []ast.Spec{&ast.ValueSpec{
			Names: []*ast.Ident{ast.NewIdent(resultName)},
			Type:  ast.NewIdent(outputType),
		}},
	}}
	merge := &ast.IfStmt{
		Cond: conditional.Cond,
		Body: &ast.BlockStmt{List: []ast.Stmt{&ast.AssignStmt{
			Lhs: []ast.Expr{ast.NewIdent(resultName)}, Tok: token.ASSIGN, Rhs: thenReturn.Results,
		}}},
		Else: &ast.BlockStmt{List: []ast.Stmt{&ast.AssignStmt{
			Lhs: []ast.Expr{ast.NewIdent(resultName)}, Tok: token.ASSIGN, Rhs: elseReturn.Results,
		}}},
	}
	body.List = []ast.Stmt{declaration, merge, &ast.ReturnStmt{Results: []ast.Expr{result}}}
	return true
}

func semanticUnitCount(node ast.Node) int {
	count := 0
	ast.Inspect(node, func(current ast.Node) bool {
		switch value := current.(type) {
		case ast.Stmt:
			if _, container := value.(*ast.BlockStmt); !container {
				count++
			}
		case ast.Expr:
			count++
		}
		return true
	})
	return count
}

func findFunction(file *ast.File, name string) (*ast.FuncDecl, bool) {
	for _, declaration := range file.Decls {
		if function, ok := declaration.(*ast.FuncDecl); ok && function.Name.Name == name {
			return function, true
		}
	}
	return nil, false
}

func validateBlock(block *ast.BlockStmt, inputName string, inherited map[string]bool, allowGuardReturn bool) (int, error) {
	return validateBlockInputs(block, map[string]bool{inputName: true}, inherited, allowGuardReturn)
}

func cloneNames(names map[string]bool) map[string]bool {
	clone := make(map[string]bool, len(names))
	maps.Copy(clone, names)
	return clone
}

func validateExpression(expression ast.Expr) error {
	switch value := expression.(type) {
	case *ast.Ident, *ast.BasicLit:
		return nil
	case *ast.ParenExpr:
		return validateExpression(value.X)
	case *ast.SelectorExpr:
		return validateExpression(value.X)
	case *ast.CompositeLit:
		return validateRecordExpression(value)
	case *ast.UnaryExpr:
		if value.Op != token.SUB && value.Op != token.NOT && value.Op != token.AND {
			return fmt.Errorf("unsupported unary operator %s", value.Op)
		}
		return validateExpression(value.X)
	case *ast.BinaryExpr:
		switch value.Op {
		case token.ADD, token.SUB, token.MUL, token.EQL, token.NEQ, token.LSS, token.LEQ, token.GTR, token.GEQ, token.LAND, token.LOR:
		default:
			return fmt.Errorf("unsupported binary operator %s", value.Op)
		}
		if err := validateExpression(value.X); err != nil {
			return err
		}
		return validateExpression(value.Y)
	default:
		return fmt.Errorf("unsupported activity expression %T", expression)
	}
}

func blockTerminates(block *ast.BlockStmt) bool {
	if block == nil || len(block.List) == 0 {
		return false
	}
	switch last := block.List[len(block.List)-1].(type) {
	case *ast.ReturnStmt:
		return true
	case *ast.IfStmt:
		if last.Else == nil || !blockTerminates(last.Body) {
			return false
		}
		switch otherwise := last.Else.(type) {
		case *ast.BlockStmt:
			return blockTerminates(otherwise)
		case *ast.IfStmt:
			return blockTerminates(&ast.BlockStmt{List: []ast.Stmt{otherwise}})
		}
	}
	return false
}

func typecheck(packageName string, file *ast.File, fset *token.FileSet) error {
	configuration := types.Config{Importer: importer.Default()}
	_, err := configuration.Check(packageName, fset, []*ast.File{file}, nil)
	return err
}

// normalizeIntegerLocalInitializers gives inferred integer locals the DSL's
// Integer representation. Go otherwise defaults an integer literal to int,
// which is not interchangeable with the compiler's int64 Integer type.
// Boolean and text locals retain Go's normal inference.
func normalizeIntegerLocalInitializers(packageName string, file *ast.File, fset *token.FileSet) int {
	information := &types.Info{Types: make(map[ast.Expr]types.TypeAndValue)}
	configuration := types.Config{Importer: importer.Default(), Error: func(error) {}}
	// The initial check can report the int/int64 mismatch this normalization
	// addresses. go/types still records initializer types for declarations it
	// can analyze; the final check below reports all remaining errors.
	_, _ = configuration.Check(packageName, fset, []*ast.File{file}, information)

	inferred := 0
	ast.Inspect(file, func(node ast.Node) bool {
		spec, ok := node.(*ast.ValueSpec)
		if !ok || spec.Type != nil || len(spec.Names) != 1 || len(spec.Values) != 1 {
			return true
		}
		value, ok := information.Types[spec.Values[0]]
		if !ok || value.Type == nil {
			return true
		}
		basic, ok := value.Type.Underlying().(*types.Basic)
		if !ok || (basic.Kind() != types.Int && basic.Kind() != types.UntypedInt) {
			return true
		}
		spec.Type = ast.NewIdent("int64")
		inferred++
		return true
	})
	return inferred
}

func rewriteLetDeclarations(body string) (string, error) {
	fset := token.NewFileSet()
	file := fset.AddFile("computes", fset.Base(), len(body))
	var sourceScanner scanner.Scanner
	sourceScanner.Init(file, []byte(body), nil, scanner.ScanComments)
	var offsets []int
	for {
		position, kind, literal := sourceScanner.Scan()
		if kind == token.EOF {
			break
		}
		if kind == token.IDENT && literal == "let" {
			offsets = append(offsets, file.Offset(position))
		}
	}
	if len(offsets) == 0 {
		return body, nil
	}
	var output strings.Builder
	last := 0
	for _, offset := range offsets {
		if offset < last || offset+3 > len(body) {
			return "", fmt.Errorf("invalid let token in computes body")
		}
		output.WriteString(body[last:offset])
		output.WriteString("var")
		last = offset + 3
	}
	output.WriteString(body[last:])
	return output.String(), nil
}

func digest(value []byte) string { return fmt.Sprintf("sha256:%x", sha256.Sum256(value)) }
