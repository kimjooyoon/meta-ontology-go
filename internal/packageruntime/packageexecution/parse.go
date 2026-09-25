package packageexecution

import "github.com/kimjooyoon/meta-ontology-go/internal/syntax"

type parsedPackage struct {
	packageName string
	namespace   string
	source      string
	evidence    []SourceEvidence
	events      []Event
}

func parsePackage(request Request) (parsedPackage, []Diagnostic, string) {
	var declarations []syntax.Declaration
	var result parsedPackage
	for _, source := range request.Sources {
		file, diagnostics := syntax.ParseFile(source.Filename, source.Content)
		if diagnostics.HasErrors() {
			return result, syntaxDiagnostics(source.Filename, diagnostics.Errors()), "PACKAGE_SYNTAX_INVALID"
		}
		if file.Package == nil || file.Namespace == nil {
			issue := Diagnostic{Stage: "SYNTAX", Code: "PACKAGE_HEADER_MISSING", Filename: source.Filename, Message: "package and namespace headers are required"}
			return result, []Diagnostic{issue}, issue.Code
		}
		if issue := bindHeader(&result, source.Filename, file); issue != nil {
			return result, []Diagnostic{*issue}, issue.Code
		}
		declarations = append(declarations, file.Decls...)
		result.evidence = append(result.evidence, SourceEvidence{Filename: source.Filename, Digest: digestBytes([]byte(source.Content)), DeclarationCount: len(file.Decls)})
		result.events = append(result.events, Event{Sequence: len(result.events) + 1, Kind: "SOURCE_PARSED", Subject: source.Filename})
	}
	combined := &syntax.File{Package: &syntax.PackageDecl{Name: result.packageName}, Namespace: &syntax.NamespaceDecl{Name: result.namespace}, Decls: declarations, Declarations: declarations}
	formatted, err := syntax.Format(combined)
	if err != nil {
		issue := Diagnostic{Stage: "FORMAT", Code: "PACKAGE_FORMAT_INVALID", Message: err.Error()}
		return result, []Diagnostic{issue}, issue.Code
	}
	result.source = formatted
	result.events = append(result.events, Event{Sequence: len(result.events) + 1, Kind: "PACKAGE_BOUND", Subject: request.PackagePath})
	return result, nil, ""
}

func bindHeader(result *parsedPackage, filename string, file *syntax.File) *Diagnostic {
	if result.packageName == "" {
		result.packageName = file.Package.Name
		result.namespace = file.Namespace.Name
		return nil
	}
	if result.packageName != file.Package.Name || result.namespace != file.Namespace.Name {
		return diagnostic("SEMANTIC", "PACKAGE_HEADER_MISMATCH", filename, "all source files must declare the same package and namespace")
	}
	return nil
}

func syntaxDiagnostics(filename string, values syntax.Diagnostics) []Diagnostic {
	result := make([]Diagnostic, 0, len(values))
	for _, value := range values {
		result = append(result, Diagnostic{Stage: "SYNTAX", Code: string(value.Code), Filename: filename, Message: value.Message})
	}
	return result
}
