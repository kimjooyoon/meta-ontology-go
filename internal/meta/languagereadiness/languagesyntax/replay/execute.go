package replay

import (
	"io/fs"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func Execute(repository fs.FS, path, kind, expectedDiagnostic string) Result {
	return executeWithSupport(repository, path, kind, expectedDiagnostic, syntax.CurrentEntityFieldsSupport())
}

func ExecuteWithEntityFieldsSupport(repository fs.FS, path, kind, expectedDiagnostic string) Result {
	return executeWithSupportAndImplicitActivityPorts(repository, path, kind, expectedDiagnostic, syntax.EntityFieldsV1Support(), false)
}

func ExecuteWithImplicitActivityPorts(repository fs.FS, path, kind, expectedDiagnostic string) Result {
	return executeWithSupportAndImplicitActivityPorts(repository, path, kind, expectedDiagnostic, syntax.CurrentEntityFieldsSupport(), true)
}

func executeWithSupport(repository fs.FS, path, kind, expectedDiagnostic string, support syntax.EntityFieldsSupport) Result {
	return executeWithSupportAndImplicitActivityPorts(repository, path, kind, expectedDiagnostic, support, false)
}

func executeWithSupportAndImplicitActivityPorts(repository fs.FS, path, kind, expectedDiagnostic string, support syntax.EntityFieldsSupport, allowImplicitActivityPorts bool) Result {
	switch kind {
	case "VALID":
		return executeValidWithSupportAndImplicitActivityPorts(repository, path, support, allowImplicitActivityPorts)
	case "INVALID":
		return executeInvalid(repository, path, expectedDiagnostic)
	default:
		return Result{ObservedDecision: DecisionUnknown, Diagnostics: []string{"registry.unknown-kind"}}
	}
}

func executeValid(repository fs.FS, path string) Result {
	return executeValidWithSupport(repository, path, syntax.CurrentEntityFieldsSupport())
}

func executeValidWithSupport(repository fs.FS, path string, support syntax.EntityFieldsSupport) Result {
	return executeValidWithSupportAndImplicitActivityPorts(repository, path, support, false)
}

func executeValidWithSupportAndImplicitActivityPorts(repository fs.FS, path string, support syntax.EntityFieldsSupport, allowImplicitActivityPorts bool) Result {
	result := Result{ObservedDecision: DecisionClosed}
	raw, err := fs.ReadFile(repository, path)
	if err != nil {
		result.ObservedDecision, result.Diagnostics = DecisionUnknown, []string{"io.read"}
		return result
	}
	result.SourceLines, result.SourceDigest = sourceLines(raw), digestBytes(raw)
	file, diagnostics := syntax.ParseFileWithEntityFieldsSupport(path, string(raw), support)
	if diagnostics.HasErrors() {
		return reject(result, diagnostics.Error().Error())
	}
	canonical, err := syntax.FormatWithEntityFieldsSupport(file, support)
	if err != nil {
		return reject(result, "format: "+err.Error())
	}
	replayed, replayDiagnostics := syntax.ParseFileWithEntityFieldsSupport(path, canonical, support)
	if replayDiagnostics.HasErrors() {
		return reject(result, replayDiagnostics.Error().Error())
	}
	replayCanonical, err := syntax.FormatWithEntityFieldsSupport(replayed, support)
	if err != nil {
		return reject(result, "replay format: "+err.Error())
	}
	leftShape, err := astShape(file)
	if err != nil {
		return reject(result, err.Error())
	}
	rightShape, err := astShape(replayed)
	if err != nil {
		return reject(result, err.Error())
	}
	result.ASTDigest, result.CanonicalDigest = digestBytes([]byte(leftShape)), digestBytes([]byte(canonical))
	result.ASTReplayed, result.ByteReplayed = leftShape == rightShape, canonical == replayCanonical
	return executeSemanticWithImplicitActivityPorts(result, file, replayed, support, allowImplicitActivityPorts)
}

func reject(result Result, diagnostic string) Result {
	result.Diagnostics = append(result.Diagnostics, diagnostic)
	return result
}
