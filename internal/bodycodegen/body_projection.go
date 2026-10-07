package bodycodegen

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
)

type bodyProjection struct {
	file              *ast.File
	fset              *token.FileSet
	function          *ast.FuncDecl
	constructs, units int
	calls             []pureCallFunction
}

func renderParameters(packageName, activityName, activityID string, parameters []InputParameter,
	outputType, body, route string, records ...RecordType) ([]byte, int, int, int, int, error) {
	return renderParametersWithCalls(packageName, activityName, activityID, parameters, outputType, body, route, nil, records)
}

func renderParametersWithCalls(packageName, activityName, activityID string, parameters []InputParameter,
	outputType, body, route string, calls []pureCallFunction, records []RecordType) ([]byte, int, int, int, int, error) {
	p, err := prepareBodyProjectionWithCalls(packageName, activityName, parameters, outputType, body, calls, records)
	if err != nil {
		return nil, 0, 0, 0, 0, err
	}
	lowered, units, err := p.lower(packageName, activityName, outputType, route, parameterNames(parameters), records)
	if err != nil {
		return nil, 0, 0, 0, 0, err
	}
	raw, err := p.emit(packageName, activityID, records)
	return raw, p.constructs, lowered, p.units, units, err
}

func prepareBodyProjection(packageName, activityName string, parameters []InputParameter,
	outputType, body string, records []RecordType) (bodyProjection, error) {
	return prepareBodyProjectionWithCalls(packageName, activityName, parameters, outputType, body, nil, records)
}

func prepareBodyProjectionWithCalls(packageName, activityName string, parameters []InputParameter,
	outputType, body string, calls []pureCallFunction, records []RecordType) (bodyProjection, error) {
	p := bodyProjection{fset: token.NewFileSet(), calls: calls}
	wrapped := fmt.Sprintf("package %s\n%sfunc %s(%s) %s {\n%s\n}\n", packageName,
		RecordDeclarations(records, false), activityName, parameterDeclaration(parameters), outputType, body)
	for _, call := range calls {
		wrapped += call.declaration()
	}
	var err error
	p.file, err = parser.ParseFile(p.fset, "body.goo", wrapped, parser.AllErrors)
	if err != nil {
		return p, fmt.Errorf("parse computes body: %w", err)
	}
	if len(p.file.Decls) != len(records)+1+len(calls) || len(p.file.Imports) != 0 {
		return p, fmt.Errorf("computes must contain only the declared function body")
	}
	var ok bool
	p.function, ok = findFunction(p.file, activityName)
	if !ok {
		return p, fmt.Errorf("activity body did not produce a function")
	}
	readonly := parameterNames(parameters)
	// An absent else falls through to the next statement. Whole-function
	// termination and Go type checking still validate every current path.
	p.constructs, err = validateBlockInputs(p.function.Body, readonly, readonly, true, records...)
	if err != nil {
		return p, err
	}
	if !blockTerminates(p.function.Body) {
		return p, fmt.Errorf("activity %q body must return on every control-flow path", activityName)
	}
	p.units = semanticUnitCount(p.function.Body)
	if err := p.validatePureCalls(records); err != nil {
		return p, err
	}
	return p, validateRecordLiterals(p.function.Body, records)
}

func (p bodyProjection) lower(packageName, activityName, outputType, route string,
	readonly map[string]bool, records []RecordType) (int, int, error) {
	if route == guardReturnRoute {
		if !lowerGuardReturn(p.function.Body) {
			return 0, 0, fmt.Errorf("activity %q does not match the guard-return route shape", activityName)
		}
	} else if route == mergeResultRoute {
		if !lowerMergeResult(p.function.Body, outputType) {
			return 0, 0, fmt.Errorf("activity %q does not match the merge-result route shape", activityName)
		}
	} else if route != preserveRoute {
		return 0, 0, fmt.Errorf("unknown body-codegen route %q", route)
	}
	constructs, err := validateBlockInputs(p.function.Body, readonly, readonly, true, records...)
	if err != nil {
		return 0, 0, err
	}
	if !blockTerminates(p.function.Body) {
		return 0, 0, fmt.Errorf("lowered activity %q body does not return on every control-flow path", activityName)
	}
	inferred := normalizeIntegerLocalInitializers(packageName, p.file, p.fset)
	units := semanticUnitCount(p.function.Body) - inferred
	for _, call := range p.calls {
		function, _ := findFunction(p.file, call.identity.Name)
		count, _ := validateBlockInputs(function.Body, parameterNames(call.parameters), parameterNames(call.parameters), true, records...)
		constructs += count
		units += semanticUnitCount(function.Body)
	}
	if err := typecheck(packageName, p.file, p.fset); err != nil {
		return 0, 0, fmt.Errorf("typecheck generated activity: %w", err)
	}
	return constructs, units, typecheckAndLowerRecords(packageName, p.file, p.fset, records)
}

func (p bodyProjection) emit(packageName, activityID string, records []RecordType) ([]byte, error) {
	var out bytes.Buffer
	fmt.Fprintf(&out, "package %s\n\n", packageName)
	out.WriteString(RecordDeclarations(records, true))
	fmt.Fprintf(&out, "//gooo:generated:start id=%q kind=\"activity\"\n", activityID)
	if err := format.Node(&out, p.fset, p.function); err != nil {
		return nil, fmt.Errorf("format generated activity: %w", err)
	}
	out.WriteByte('\n')
	fmt.Fprintf(&out, "//gooo:generated:end id=%q kind=\"activity\"\n", activityID)
	for _, call := range p.calls {
		function, _ := findFunction(p.file, call.identity.Name)
		fmt.Fprintf(&out, "//gooo:generated:start id=%q kind=\"activity\"\n", call.identity.ActivityID)
		if err := format.Node(&out, p.fset, function); err != nil {
			return nil, err
		}
		fmt.Fprintf(&out, "\n//gooo:generated:end id=%q kind=\"activity\"\n", call.identity.ActivityID)
	}
	raw, err := format.Source(out.Bytes())
	if err != nil {
		return nil, fmt.Errorf("format generated source: %w", err)
	}
	return raw, nil
}
