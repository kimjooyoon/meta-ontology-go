package replay

import (
	"context"
	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func executeSemantic(result Result, file, replayed *syntax.File) Result {
	return executeSemanticWithSupport(result, file, replayed, syntax.CurrentEntityFieldsSupport())
}

func executeSemanticWithSupport(result Result, file, replayed *syntax.File, support syntax.EntityFieldsSupport) Result {
	return executeSemanticWithImplicitActivityPorts(result, file, replayed, support, false)
}

func executeSemanticWithImplicitActivityPorts(result Result, file, replayed *syntax.File, support syntax.EntityFieldsSupport, allowImplicitActivityPorts bool) Result {
	lower := func(source *syntax.File) (semantic.IR, error) {
		if allowImplicitActivityPorts {
			return bidir.LowerContextWithImplicitActivityPorts(context.Background(), source, support)
		}
		return bidir.LowerContextWithEntityFieldsSupport(context.Background(), source, support)
	}
	adapt := func(source *syntax.File) (bidir.Document, error) {
		if allowImplicitActivityPorts {
			return bidir.DocumentFromSyntaxWithImplicitActivityPorts(source, support)
		}
		return bidir.DocumentFromSyntaxWithEntityFieldsSupport(source, support)
	}
	original, err := lower(file)
	if err != nil {
		return reject(result, "lower: "+err.Error())
	}
	replayIR, err := lower(replayed)
	if err != nil {
		return reject(result, "replay lower: "+err.Error())
	}
	document, err := adapt(file)
	if err != nil {
		return reject(result, "document: "+err.Error())
	}
	model, err := bidir.GetWithEntityFieldsSupport(document, support)
	if err != nil {
		return reject(result, "Get: "+err.Error())
	}
	written, err := bidir.PutWithEntityFieldsSupport(document, model, support)
	if err != nil {
		return reject(result, "Put: "+err.Error())
	}
	roundTrip, err := bidir.LowerDocumentWithEntityFieldsSupport(written, support)
	if err != nil {
		return reject(result, "roundtrip lower: "+err.Error())
	}
	result.SemanticDigest = "sha256:" + original.StableHash()
	result.SemanticReplayed = bidir.EquivalentAfterRoundTrip(original, replayIR) &&
		bidir.EquivalentAfterRoundTrip(original, roundTrip)
	result.GetPut = bidir.CheckGetPutWithEntityFieldsSupport(document, support) == nil
	result.PutGet = bidir.CheckPutGetWithEntityFieldsSupport(document, model, support) == nil
	if result.ASTReplayed && result.ByteReplayed && result.SemanticReplayed && result.GetPut && result.PutGet {
		result.ObservedDecision = DecisionPass
	} else {
		result.Diagnostics = append(result.Diagnostics, "roundtrip.contract-mismatch")
	}
	return result
}
