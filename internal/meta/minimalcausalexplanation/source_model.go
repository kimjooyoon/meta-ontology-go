package minimalcausalexplanation

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const (
	observedOrigin  = "OBSERVED"
	syntheticOrigin = "SYNTHETIC"
)

type sourceModel struct {
	SourcePath        string
	SourceDigest      string
	SemanticDigest    string
	Package           string
	Namespace         string
	Graph             CausalGraph
	Predicate         DecisionPredicate
	Program           MetaProgram
	Evidence          []Evidence
	Claims            []string
	PriorClaimState   string
	DecisionOutput    string
	SourceReconstruct SourceReconstruction
	EvidenceByRole    map[string]Evidence
}

type rawIndependenceObservation struct {
	Schema                     string `json:"schema"`
	ProducerPackageImportCount int    `json:"producer_package_import_count"`
	ProducerPackageImportTotal int    `json:"producer_package_import_total"`
}

type rawRepositoryObservation struct {
	Schema              string `json:"schema"`
	Status              string `json:"status"`
	WorkspaceWrites     bool   `json:"workspace_writes"`
	PromotionAuthorized bool   `json:"promotion_authorized"`
}

func reconstructSource(sourcePath string, source []byte, independence []byte) (sourceModel, error) {
	file, diagnostics := syntax.ParseFile(sourcePath, string(source))
	if file == nil || diagnostics.HasErrors() {
		return sourceModel{}, fmt.Errorf("source parse failed: %s", diagnostics.Error().Error())
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return sourceModel{}, fmt.Errorf("source lower failed: %w", err)
	}
	if err := ir.Validate(); err != nil {
		return sourceModel{}, fmt.Errorf("source semantic IR failed validation: %w", err)
	}

	model := sourceModel{
		SourcePath:        sourcePath,
		SourceDigest:      contentDigest(source),
		SemanticDigest:    "sha256:" + ir.StableHash(),
		Package:           ir.Package,
		Namespace:         ir.Namespace.String(),
		SourceReconstruct: SourceReconstruction{ASTParsed: true, IRLowered: true, SemanticDigest: "sha256:" + ir.StableHash()},
		EvidenceByRole:    make(map[string]Evidence),
	}

	producerByEntity := make(map[string]string)
	consumerByEntity := make(map[string]string)
	inputsByActivity := make(map[string][]string)
	outputsByActivity := make(map[string][]string)
	programByActivity := make(map[string]string)

	for _, node := range ir.Graph.Nodes() {
		if node.Kind != semantic.Entity {
			if node.Kind == semantic.Activity && node.ValueProgram != "" {
				programByActivity[node.ID.String()] = node.ValueProgram
			}
			continue
		}
		role := entityRole(node.ID.String())
		if role == "" {
			continue
		}
		model.Evidence = append(model.Evidence, Evidence{
			ID: node.ID.String(), Role: role, Origin: observedOrigin,
			Status: StatusUnknown, Provenance: "gooo semantic entity",
		})
	}

	for _, fact := range ir.Graph.DeterministicFacts() {
		subject, object := fact.Subject.String(), fact.Object.String()
		switch fact.Predicate {
		case semantic.Used:
			inputsByActivity[subject] = append(inputsByActivity[subject], object)
			consumerByEntity[object] = subject
		case semantic.WasGeneratedBy:
			outputsByActivity[object] = append(outputsByActivity[object], subject)
			producerByEntity[subject] = object
		}
	}

	for _, node := range ir.Graph.Nodes() {
		if _, ok := model.EvidenceByRole[entityRole(node.ID.String())]; ok {
			continue
		}
		role := entityRole(node.ID.String())
		if role == "" {
			continue
		}
		for index := range model.Evidence {
			if model.Evidence[index].ID == node.ID.String() {
				model.Evidence[index].Provenance = provenanceForEntity(node.ID.String(), producerByEntity, consumerByEntity)
				break
			}
		}
	}

	for index := range model.Evidence {
		evidence := &model.Evidence[index]
		if producer := producerByEntity[evidence.ID]; producer != "" {
			evidence.Provenance = provenanceForEntity(evidence.ID, producerByEntity, consumerByEntity)
		}
		model.EvidenceByRole[evidence.Role] = *evidence
	}

	for activityID, program := range programByActivity {
		for _, clause := range strings.Split(program, ";") {
			clause = strings.TrimSpace(clause)
			switch {
			case strings.HasPrefix(clause, "mce.operation:"):
				operation, err := parseOperation(clause, activityID, program)
				if err != nil {
					return sourceModel{}, err
				}
				model.Program.MetaOperations = append(model.Program.MetaOperations, operation)
			case strings.HasPrefix(clause, "mce.predicate:"):
				predicate, err := parsePredicate(clause)
				if err != nil {
					return sourceModel{}, err
				}
				model.Predicate = predicate
			case strings.HasPrefix(clause, "mce.claim-state:"):
				model.PriorClaimState = parseValue(clause, "mce.claim-state:")
			case strings.HasPrefix(clause, "mce.decision-output:"):
				model.DecisionOutput = parseValue(clause, "mce.decision-output:")
			case strings.HasPrefix(clause, "mce.indicators:"):
				model.Program.IndicatorDenominator = parseIntValue(clause, "mce.indicators:")
			case strings.HasPrefix(clause, "mce.program:"):
				parts := strings.Split(strings.TrimPrefix(clause, "mce.program:"), ":v1")
				endpoints := strings.Split(parts[0], "|")
				if len(endpoints) != 2 {
					return sourceModel{}, fmt.Errorf("invalid mce program value %q", clause)
				}
				model.Program.Producer, model.Program.Consumer = endpoints[0], endpoints[1]
			case strings.HasPrefix(clause, "mce.claims:"):
				model.Claims = splitValue(parseValue(clause, "mce.claims:"), "+")
			}
		}
	}
	sort.Slice(model.Program.MetaOperations, func(i, j int) bool { return model.Program.MetaOperations[i].ID < model.Program.MetaOperations[j].ID })
	if model.Program.Schema == "" {
		model.Program.Schema = SourceSchema
	}
	if model.Program.Producer == "" || model.Program.Consumer == "" {
		return sourceModel{}, fmt.Errorf("source does not declare mce program endpoints")
	}
	if model.Predicate.Value == "" || len(model.Predicate.RequiredRoles) == 0 {
		return sourceModel{}, fmt.Errorf("source does not declare an mce decision predicate")
	}
	if model.PriorClaimState == "" || model.DecisionOutput == "" || model.Program.IndicatorDenominator == 0 || len(model.Claims) == 0 {
		return sourceModel{}, fmt.Errorf("source semantic values are incomplete")
	}
	model.Predicate.DecisionOutput = model.DecisionOutput
	model.Predicate.PriorClaimState = model.PriorClaimState

	model.Graph = deriveCausalGraph(model, inputsByActivity, outputsByActivity, programByActivity)
	model.SourceReconstruct.GraphReconstructed = len(model.Graph.Nodes) > 0 && len(model.Graph.Edges) > 0
	model.SourceReconstruct.PredicateReconstructed = true
	if len(model.Evidence) == 0 || !model.SourceReconstruct.GraphReconstructed {
		return sourceModel{}, fmt.Errorf("source semantic graph has no evidence path")
	}
	if independence != nil {
		var raw rawIndependenceObservation
		if err := json.Unmarshal(independence, &raw); err != nil {
			return sourceModel{}, fmt.Errorf("independence observation: %w", err)
		}
		if raw.Schema != IndependenceSchema || raw.ProducerPackageImportCount < 0 || raw.ProducerPackageImportTotal <= 0 || raw.ProducerPackageImportCount > raw.ProducerPackageImportTotal {
			return sourceModel{}, fmt.Errorf("independence observation is invalid")
		}
		model.SourceReconstruct.ProducerPackageImportCount = raw.ProducerPackageImportCount
		model.SourceReconstruct.ProducerPackageImportTotal = raw.ProducerPackageImportTotal
	} else {
		return sourceModel{}, fmt.Errorf("independence observation is required")
	}
	return model, nil
}

func entityRole(id string) string {
	marker := "/evidence/"
	index := strings.Index(id, marker)
	if index < 0 {
		return ""
	}
	role := strings.Trim(strings.TrimPrefix(id[index+len(marker):], "/"), "/")
	if strings.Contains(role, "/") || role == "" {
		return ""
	}
	return role
}

func provenanceForEntity(id string, producer, consumer map[string]string) string {
	return "entity=" + id + ";producer=" + producer[id] + ";consumer=" + consumer[id]
}

func parseOperation(clause, activityID, program string) (MetaOperation, error) {
	value := strings.TrimPrefix(clause, "mce.operation:")
	value = strings.TrimSuffix(value, ":v1")
	parts := strings.Split(value, "|")
	if len(parts) != 4 || parts[0] == "" || parts[1] == "" || parts[2] == "" || parts[3] == "" {
		return MetaOperation{}, fmt.Errorf("invalid mce operation value %q", clause)
	}
	return MetaOperation{ID: parts[0], Activity: activityID, Producer: parts[1], Consumer: parts[2], ProofChoice: parts[3], EvidenceDigest: digestString(program)}, nil
}

func parsePredicate(clause string) (DecisionPredicate, error) {
	value := strings.TrimPrefix(clause, "mce.predicate:")
	parts := strings.SplitN(strings.TrimSuffix(value, ":v1"), ":", 2)
	if len(parts) != 2 || parts[0] != "PASS_IF" {
		return DecisionPredicate{}, fmt.Errorf("invalid mce predicate value %q", clause)
	}
	roles := splitValue(parts[1], "+")
	if len(roles) == 0 {
		return DecisionPredicate{}, fmt.Errorf("mce predicate has no required roles")
	}
	return DecisionPredicate{Value: clause, RequiredRoles: roles}, nil
}

func parseValue(clause, prefix string) string {
	value := strings.TrimPrefix(clause, prefix)
	return strings.TrimSuffix(value, ":v1")
}

func parseIntValue(clause, prefix string) int {
	value, err := strconv.Atoi(parseValue(clause, prefix))
	if err != nil {
		return 0
	}
	return value
}

func splitValue(value, separator string) []string {
	parts := strings.Split(value, separator)
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if strings.TrimSpace(part) != "" {
			result = append(result, strings.TrimSpace(part))
		}
	}
	return result
}

func deriveCausalGraph(model sourceModel, inputs, outputs map[string][]string, programs map[string]string) CausalGraph {
	graph := CausalGraph{Schema: GraphSchema, DecisionRule: model.Predicate.Value}
	for _, evidence := range model.Evidence {
		producer, consumer := "raw-observation", "unconsumed"
		for activityID, produced := range outputs {
			if containsString(produced, evidence.ID) {
				producer = activityID
			}
		}
		for activityID, consumed := range inputs {
			if containsString(consumed, evidence.ID) {
				consumer = activityID
			}
		}
		role := "DECISION_INPUT"
		if !containsString(model.Predicate.RequiredRoles, evidence.Role) {
			role = "NON_CAUSAL_LOG"
		}
		graph.Nodes = append(graph.Nodes, CausalNode{ID: evidence.ID, Role: role, Producer: producer, Consumer: consumer})
	}
	for activityID, activityInputs := range inputs {
		for _, from := range activityInputs {
			fromRole := entityRole(from)
			if fromRole == "" {
				continue
			}
			for _, to := range outputs[activityID] {
				toRole := entityRole(to)
				if toRole == "" {
					continue
				}
				relation := "PROV_USED_AND_GENERATED"
				if programs[activityID] != "" {
					relation = programs[activityID]
				}
				causal := containsString(model.Predicate.RequiredRoles, fromRole) && containsString(model.Predicate.RequiredRoles, toRole)
				graph.Edges = append(graph.Edges, CausalEdge{
					ID: digestString(from + "|" + activityID + "|" + to + "|" + relation), From: from, To: to,
					ViaActivity: activityID, Relation: relation, Causal: causal,
				})
			}
		}
	}
	sort.Slice(graph.Nodes, func(i, j int) bool { return graph.Nodes[i].ID < graph.Nodes[j].ID })
	sort.Slice(graph.Edges, func(i, j int) bool { return graph.Edges[i].ID < graph.Edges[j].ID })
	graph.Digest = digestGraph(graph)
	return graph
}

func digestGraph(graph CausalGraph) string {
	graph.Digest = ""
	digest, _ := digestValue(graph)
	return digest
}

func containsString(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}

func decodeCompilerReceipt(data []byte) (RawCompilerReceipt, error) {
	var receipt RawCompilerReceipt
	if err := json.Unmarshal(data, &receipt); err != nil {
		return RawCompilerReceipt{}, fmt.Errorf("compiler receipt: %w", err)
	}
	if receipt.Schema != CompilerReceiptSchema {
		return RawCompilerReceipt{}, fmt.Errorf("compiler receipt schema is invalid")
	}
	return receipt, nil
}

func decodeRepositoryObservation(data []byte) (RepositoryObservation, error) {
	var observation rawRepositoryObservation
	if err := json.Unmarshal(data, &observation); err != nil {
		return RepositoryObservation{}, fmt.Errorf("repository observation: %w", err)
	}
	if observation.Schema != RepositorySchema {
		return RepositoryObservation{}, fmt.Errorf("repository observation schema is invalid")
	}
	return RepositoryObservation{Schema: observation.Schema, Status: observation.Status, WorkspaceWrites: observation.WorkspaceWrites, PromotionAuthorized: observation.PromotionAuthorized}, nil
}

func observedEvidence(model sourceModel, raw RawCompilerReceipt, rawBytes []byte) []Evidence {
	digest := contentDigest(rawBytes)
	statuses := map[string]string{
		"source-parsed":           StatusUnknown,
		"semantic-ir-bound":       StatusUnknown,
		"compiler-receipt-proven": StatusUnknown,
	}
	if raw.SourcePath != "" && raw.SourceDigest != "" {
		statuses["source-parsed"] = StatusPass
	}
	if raw.SemanticFingerprint != "" && raw.CoreIRFingerprint != "" && raw.Resolution == "CORE_IR_ACTIVITY_VALUE_PROGRAM" {
		statuses["semantic-ir-bound"] = StatusPass
	}
	if raw.Decision == "VALUE_WITNESS_PROVEN" && raw.Reason == "VALUE_WITNESS_EXACT" && raw.Resolution == "CORE_IR_ACTIVITY_VALUE_PROGRAM" {
		statuses["compiler-receipt-proven"] = StatusPass
	}
	result := make([]Evidence, 0, len(model.Predicate.RequiredRoles))
	for _, role := range model.Predicate.RequiredRoles {
		evidence := model.EvidenceByRole[role]
		evidence.Origin = observedOrigin
		evidence.Status = statuses[role]
		evidence.Digest = digestString(role + "|" + digest + "|" + evidence.Status)
		evidence.Provenance = "raw compiler receipt;source=" + raw.SourcePath + ";decision=" + raw.Decision
		result = append(result, evidence)
	}
	return result
}

func syntheticNoise(model sourceModel) (Evidence, bool) {
	evidence, ok := model.EvidenceByRole["audit-noise"]
	if !ok {
		return Evidence{}, false
	}
	evidence.Origin = syntheticOrigin
	evidence.Status = StatusPass
	evidence.Digest = digestString("synthetic|audit-noise|overlong-path")
	evidence.Provenance = "synthetic overlong-path noise;not observed compiler evidence"
	return evidence, true
}

func repositoryBoundary(before, after RepositoryObservation) RepositoryBoundary {
	writes := 0
	if before.WorkspaceWrites || after.WorkspaceWrites || before.Status != "" || after.Status != "" {
		writes = 1
	}
	return RepositoryBoundary{Before: before, After: after, Writes: writes, PromotionAuthorized: before.PromotionAuthorized || after.PromotionAuthorized}
}

func digestEvidence(evidence []Evidence) string {
	digest, _ := digestValue(evidence)
	return digest
}
