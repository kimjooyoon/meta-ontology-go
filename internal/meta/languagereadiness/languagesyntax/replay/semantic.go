package replay

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func executeSemantic(result Result, file, replayed *syntax.File) Result {
	return executeSemanticWithSupport(result, file, replayed, syntax.CurrentEntityFieldsSupport())
}

func executeSemanticWithSupport(result Result, file, replayed *syntax.File, support syntax.EntityFieldsSupport) Result {
	original, err := bidir.LowerContextWithEntityFieldsSupport(nil, file, support)
	if err != nil {
		return reject(result, "lower: "+err.Error())
	}
	replayIR, err := bidir.LowerContextWithEntityFieldsSupport(nil, replayed, support)
	if err != nil {
		return reject(result, "replay lower: "+err.Error())
	}
	document, err := bidir.DocumentFromSyntaxWithEntityFieldsSupport(file, support)
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
