package lsp

import (
	"context"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/provenance"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const ExecutionPlanProvenanceSchemaPart01 = provenance.ExecutionPlanProvenanceBindingSchemaPart01

type ExecutionPlanProvenanceParamsPart01 struct {
	TextDocument     TextDocumentIdentifier                  `json:"textDocument"`
	Task             string                                  `json:"task"`
	WorkspaceDigest  string                                  `json:"workspace_digest"`
	Model            string                                  `json:"model"`
	GatewayPolicy    provenance.GatewayPolicy                `json:"gateway_policy"`
	Lifecycle        provenance.ExecutionPlanLifecyclePart01 `json:"lifecycle"`
	WorkloadIdentity *provenance.WorkloadIdentityProvenanceBinding `json:"workload_identity,omitempty"`
}

func executionPlanBindingEdgeKeyPart01(edge bidir.BindingEdge) string {
	return fmt.Sprintf("%s:%s->%s:%s", edge.SourceActivity, edge.SourcePort, edge.TargetActivity, edge.TargetPort)
}

func executionPlanTypedMetadataPart01(text string) (string, []string, []string, int) {
	file, diagnostics := syntax.ParseFile("execution-plan-provenance.gooo", text)
	if diagnostics.HasErrors() || file == nil {
		return "", nil, nil, 0
	}
	document, err := bidir.DocumentFromSyntaxWithEntityFieldsSupport(file, syntax.EntityFieldsV1Support())
	if err != nil || len(document.BindingEdges) == 0 {
		return "", nil, nil, 0
	}
	typedPlan, err := bidir.CompileTypedPlan(document)
	if err != nil {
		return "", nil, nil, 0
	}
	activityOrder := make([]string, len(typedPlan.Activities))
	for index, activity := range typedPlan.Activities {
		activityOrder[index] = string(activity)
	}
	edgeOrder := make([]string, len(typedPlan.Edges))
	for index, edge := range typedPlan.Edges {
		edgeOrder[index] = executionPlanBindingEdgeKeyPart01(edge)
	}
	return typedPlan.Digest(), activityOrder, edgeOrder, len(typedPlan.Edges)
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
	var bindingEdgeOrder []string
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
		typedPlanDigest, activityOrder, bindingEdgeOrder, runtimeBindingCount = executionPlanTypedMetadataPart01(stored.text)
		graphDigest := documentProvenanceReferenceMapDigest(references)
		if typedPlanDigest != "" {
			graphDigest = typedPlanDigest
		}
		chain = provenance.BuildSelfImprovementProvenanceChainPart01(
			documentProvenanceSymbolMapDigest(symbols),
			stored.cacheKey.sourceDigest,
			stored.result.semanticDigest,
			graphDigest,
			"",
			"",
		)
	}

	var binding provenance.ExecutionPlanProvenanceBindingPart01
	if params.WorkloadIdentity == nil {
		binding = provenance.BindExecutionPlanToProvenanceWithTypedPlanEdgesPart01(
			params.Task,
			params.WorkspaceDigest,
			params.Model,
			params.GatewayPolicy,
			params.Lifecycle,
			typedPlanDigest,
			activityOrder,
			bindingEdgeOrder,
			runtimeBindingCount,
			chain,
		)
	} else {
		binding = provenance.BindExecutionPlanToProvenanceWithTypedPlanAndWorkloadIdentityPart01(
			params.Task,
			params.WorkspaceDigest,
			params.Model,
			params.GatewayPolicy,
			params.Lifecycle,
			typedPlanDigest,
			activityOrder,
			bindingEdgeOrder,
			runtimeBindingCount,
			*params.WorkloadIdentity,
			chain,
		)
	}

	if err := binding.Validate(); err != nil {
		return featureErrorResponse(request.ID, err, ctx)
	}
	return resultResponse(request.ID, binding), nil, nil
}
