package lsp

import (
	"context"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/provenance"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const ExecutionPlanProvenanceSchemaPart01 = provenance.ExecutionPlanProvenanceBindingSchemaPart01

type ExecutionPlanProvenanceParamsPart01 struct {
	TextDocument    TextDocumentIdentifier                  `json:"textDocument"`
	Task            string                                  `json:"task"`
	WorkspaceDigest string                                  `json:"workspace_digest"`
	Model           string                                  `json:"model"`
	GatewayPolicy   provenance.GatewayPolicy                `json:"gateway_policy"`
	Lifecycle       provenance.ExecutionPlanLifecyclePart01 `json:"lifecycle"`
}

func executionPlanTypedMetadataPart01(text string) (string, []string, int) {
	file, diagnostics := syntax.ParseFile("execution-plan-provenance.gooo", text)
	if diagnostics.HasErrors() || file == nil {
		return "", nil, 0
	}
	document, err := bidir.DocumentFromSyntaxWithEntityFieldsSupport(file, syntax.EntityFieldsV1Support())
	if err != nil || len(document.BindingEdges) == 0 {
		return "", nil, 0
	}
	typedPlan, err := bidir.CompileTypedPlan(document)
	if err != nil {
		return "", nil, 0
	}
	activityOrder := make([]string, len(typedPlan.Activities))
	for index, activity := range typedPlan.Activities {
		activityOrder[index] = string(activity)
	}
	return typedPlan.Digest(), activityOrder, len(typedPlan.Edges)
}

func (server *Server) executionPlanProvenanceRequest(
	ctx context.Context,
	request requestEnvelope,
) (*responseEnvelope, [][]byte, error) {
	var params ExecutionPlanProvenanceParamsPart01
	if decodeParams(request.Params, &params) != nil || params.TextDocument.URI == "" {
		return responseOrNil(request.ID, invalidParams, "Invalid execution-plan provenance parameters"), nil, nil
	}
	if err := server.refresh(ctx, params.TextDocument.URI); err != nil {
		return featureErrorResponse(request.ID, err, ctx)
	}

	var chain provenance.SelfImprovementProvenanceChainPart01
	var typedPlanDigest string
	var activityOrder []string
	var runtimeBindingCount int
	server.mu.RLock()
	stored, exists := server.documents[params.TextDocument.URI]
	if exists {
		key := stored.cacheKey
		result := cloneParseResult(stored.result)
		stored = &document{
			version:  stored.version,
			text:     stored.text,
			cacheKey: key,
			result:   result,
		}
	}
	server.mu.RUnlock()

	if exists && stored.cacheKey.sourceDigest != "" {
		symbols := documentProvenanceSymbols(stored.result)
		references := documentProvenanceReferences(stored.result)
		chain = provenance.BuildSelfImprovementProvenanceChainPart01(
			documentProvenanceSymbolMapDigest(symbols),
			stored.cacheKey.sourceDigest,
			stored.result.semanticDigest,
			documentProvenanceReferenceMapDigest(references),
			"",
			"",
		)
		typedPlanDigest, activityOrder, runtimeBindingCount = executionPlanTypedMetadataPart01(stored.text)
	}

	binding := provenance.BindExecutionPlanToProvenanceWithTypedPlanPart01(
		params.Task,
		params.WorkspaceDigest,
		params.Model,
		params.GatewayPolicy,
		params.Lifecycle,
		typedPlanDigest,
		activityOrder,
		runtimeBindingCount,
		chain,
	)
	if err := binding.Validate(); err != nil {
		return featureErrorResponse(request.ID, err, ctx)
	}
	return resultResponse(request.ID, binding), nil, nil
}

