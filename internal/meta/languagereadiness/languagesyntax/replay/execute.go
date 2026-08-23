package replay

import (
	"io/fs"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func Execute(repository fs.FS, path, kind, expectedDiagnostic string) Result {
	switch kind {
	case "VALID":
		return executeValid(repository, path)
	case "INVALID":
		return executeInvalid(repository, path, expectedDiagnostic)
	default:
		return Result{ObservedDecision: DecisionUnknown, Diagnostics: []string{"registry.unknown-kind"}}
	}
}

func executeValid(repository fs.FS, path string) Result {
	result := Result{ObservedDecision: DecisionClosed}
	raw, err := fs.ReadFile(repository, path)
	if err != nil {
		result.ObservedDecision, result.Diagnostics = DecisionUnknown, []string{"io.read"}
		return result
	}
	result.SourceLines, result.SourceDigest = sourceLines(raw), digestBytes(raw)
	file, diagnostics := syntax.ParseFile(path, string(raw))
	if diagnostics.HasErrors() {
		return reject(result, diagnostics.Error().Error())
	}
	canonical, err := syntax.Format(file)
	if err != nil {
		return reject(result, "format: "+err.Error())
	}
	replayed, replayDiagnostics := syntax.ParseFile(path, canonical)
	if replayDiagnostics.HasErrors() {
		return reject(result, replayDiagnostics.Error().Error())
	}
	replayCanonical, err := syntax.Format(replayed)
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
	return executeSemantic(result, file, replayed)
}

func reject(result Result, diagnostic string) Result {
	result.Diagnostics = append(result.Diagnostics, diagnostic)
	return result
}
