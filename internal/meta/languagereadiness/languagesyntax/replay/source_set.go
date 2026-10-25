package replay

import (
	"encoding/hex"
	"io/fs"

	"github.com/kimjooyoon/meta-ontology-go/internal/sourcebundle"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

// ExecuteSourceSet replays one explicit same-package case. Members retain their
// own physical spans; no concatenated document is attributed to a source file.
// Imports require package resolution and remain outside this syntax-case route.
func ExecuteSourceSet(repository fs.FS, names []string, support syntax.EntityFieldsSupport,
	implicitPorts bool) Result {
	result := Result{ObservedDecision: DecisionUnknown, SourceDigestKind: "gooo/source-bundle/v1"}
	sources, err := sourcebundle.Read(repository, names)
	if err != nil {
		return reject(result, "source-set.read: "+err.Error())
	}
	digest, err := sourcebundle.Digest(sources)
	if err != nil {
		return reject(result, err.Error())
	}
	result.SourceDigest = "sha256:" + hex.EncodeToString(digest[:])
	for _, source := range sources {
		member := ObserveSource(source.Filename, []byte(source.Content))
		result.SourceMembers = append(result.SourceMembers, member)
		result.SourceLines += member.GoooLines
	}
	result.ObservedDecision = DecisionClosed
	canonical := make([]sourcebundle.Source, 0, len(sources))
	result.ASTReplayed, result.ByteReplayed = true, true
	for _, source := range sources {
		observed, err := replaySourceSyntax(source.Filename, source.Content, support)
		if err != nil {
			return reject(result, err.Error())
		}
		result.ASTReplayed = result.ASTReplayed && observed.astReplayed
		result.ByteReplayed = result.ByteReplayed && observed.byteReplayed
		canonical = append(canonical, sourcebundle.Source{Filename: source.Filename, Content: observed.canonical})
	}
	return replaySourceSetSemantics(result, sources, canonical, support, implicitPorts)
}

func replaySourceSetSemantics(result Result, sources, canonical []sourcebundle.Source,
	support syntax.EntityFieldsSupport, implicitPorts bool) Result {
	before, err := sourcebundle.ParseWithEntityFieldsSupport(sources, support)
	if err != nil {
		return reject(result, err.Error())
	}
	after, err := sourcebundle.ParseWithEntityFieldsSupport(canonical, support)
	if err != nil {
		return reject(result, "replay: "+err.Error())
	}
	shape, err := astShape(before.File)
	if err != nil {
		return reject(result, err.Error())
	}
	result.ASTDigest, result.CanonicalDigest = digestBytes([]byte(shape)), "sha256:"+after.SourceDigest
	return executeSemanticWithImplicitActivityPorts(result, before.File, after.File, support, implicitPorts)
}
