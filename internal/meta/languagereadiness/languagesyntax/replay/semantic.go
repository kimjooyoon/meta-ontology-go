package replay

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func executeSemantic(result Result, file, replayed *syntax.File) Result {
	original, err := bidir.Lower(file)
	if err != nil {
		return reject(result, "lower: "+err.Error())
	}
	replayIR, err := bidir.Lower(replayed)
	if err != nil {
		return reject(result, "replay lower: "+err.Error())
	}
	document, err := bidir.DocumentFromSyntax(file)
	if err != nil {
		return reject(result, "document: "+err.Error())
	}
	model, err := bidir.Get(document)
	if err != nil {
		return reject(result, "Get: "+err.Error())
	}
	written, err := bidir.Put(document, model)
	if err != nil {
		return reject(result, "Put: "+err.Error())
	}
	roundTrip, err := bidir.LowerDocument(written)
	if err != nil {
		return reject(result, "roundtrip lower: "+err.Error())
	}
	result.SemanticDigest = "sha256:" + original.StableHash()
	result.SemanticReplayed = bidir.EquivalentAfterRoundTrip(original, replayIR) &&
		bidir.EquivalentAfterRoundTrip(original, roundTrip)
	result.GetPut = bidir.CheckGetPut(document) == nil
	result.PutGet = bidir.CheckPutGet(document, model) == nil
	if result.ASTReplayed && result.ByteReplayed && result.SemanticReplayed && result.GetPut && result.PutGet {
		result.ObservedDecision = DecisionPass
	} else {
		result.Diagnostics = append(result.Diagnostics, "roundtrip.contract-mismatch")
	}
	return result
}
