package operationprovenance

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const relationDenominator = 4

type metricSpec struct {
	ID            string
	Family        string
	PriorClaim    string
	Producer      string
	Consumer      string
	MetaOperation string
	EvidencePath  string
	DependsOn     []string
}

type scenarioSpec struct {
	ID             string
	RemoveRelation string
	Dependency     string
	Reason         string
}

type edge struct {
	From string
	To   string
	Kind string
}

type fixture struct {
	ID       string
	Mutation string
	Nodes    map[string]string
	Edges    []edge
	Metrics  []metricSpec
}

type relation struct {
	Kind string
	From string
	To   string
}

func Build(source []byte, observation WorkspaceObservation) (Receipt, error) {
	ir, err := lowerSource(source)
	if err != nil {
		return Receipt{}, err
	}
	metrics, scenarios, reconstruction, err := reconstructSemanticData(ir)
	if err != nil {
		return Receipt{}, err
	}

	receipt := Receipt{
		Schema:                  ReceiptSchema,
		Toolchain:               Toolchain,
		SourceDigest:            digestBytes(source),
		CanonicalSemanticDigest: "sha256:" + ir.StableHash(),
		SourceReconstruction:    reconstruction,
		WorkspaceObservation:    observation,
		Scenarios:               make([]ScenarioResult, 0, len(scenarios)),
	}
	for _, scenario := range scenarios {
		result, err := evaluateScenario(metrics, scenario, receipt.SourceDigest, receipt.CanonicalSemanticDigest)
		if err != nil {
			return Receipt{}, err
		}
		receipt.Scenarios = append(receipt.Scenarios, result)
	}
	return sealReceipt(receipt)
}

// BuildObserved runs the producer in an isolated repository workspace and
// binds the receipt to the observed before/after status snapshots.
func BuildObserved(source []byte, repositoryRoot string) (Receipt, error) {
	before, err := readRepositorySnapshot(repositoryRoot)
	if err != nil {
		return Receipt{}, fmt.Errorf("observe repository before producer: %w", err)
	}
	receipt, buildErr := Build(source, WorkspaceObservation{})
	after, afterErr := readRepositorySnapshot(repositoryRoot)
	if buildErr != nil {
		return Receipt{}, buildErr
	}
	if afterErr != nil {
		return Receipt{}, fmt.Errorf("observe repository after producer: %w", afterErr)
	}
	receipt.WorkspaceObservation = deriveObservation(before, after)
	return sealReceipt(receipt)
}

type repositorySnapshot struct {
	digest string
	status map[string]string
}

func readRepositorySnapshot(root string) (repositorySnapshot, error) {
	command := exec.Command("git", "-C", root, "status", "--porcelain=v1", "--untracked-files=all")
	output, err := command.Output()
	if err != nil {
		return repositorySnapshot{}, err
	}
	status := make(map[string]string)
	for _, line := range strings.Split(strings.TrimSuffix(string(output), "\n"), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		path := strings.TrimSpace(line)
		if len(path) > 3 {
			path = strings.TrimSpace(path[3:])
		}
		status[path] = line
	}
	return repositorySnapshot{digest: digestBytes(output), status: status}, nil
}

func deriveObservation(before, after repositorySnapshot) WorkspaceObservation {
	paths := make([]string, 0)
	seen := make(map[string]bool)
	for path := range before.status {
		seen[path] = true
	}
	for path := range after.status {
		seen[path] = true
	}
	for path := range seen {
		if before.status[path] != after.status[path] {
			paths = append(paths, path)
		}
	}
	sort.Strings(paths)
	writes := len(paths) > 0
	return WorkspaceObservation{
		BeforeDigest:              before.digest,
		AfterDigest:               after.digest,
		ChangedPaths:              paths,
		RepositoryWorkspaceWrites: writes,
		MutationAuthority:         writes,
	}
}

func lowerSource(source []byte) (semantic.IR, error) {
	file, diagnostics := syntax.ParseFile("main.gooo", string(source))
	if diagnostics.HasErrors() || file == nil {
		return semantic.IR{}, fmt.Errorf("Gooo source has syntax errors")
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return semantic.IR{}, fmt.Errorf("lower Gooo source: %w", err)
	}
	return ir, nil
}

func reconstructSemanticData(ir semantic.IR) ([]metricSpec, []scenarioSpec, SourceReconstruction, error) {
	metrics := make([]metricSpec, 0)
	scenarios := make([]scenarioSpec, 0)
	reconstruction := SourceReconstruction{}
	for _, node := range ir.Graph.Nodes() {
		if node.Kind != semantic.Activity || node.ValueProgram == "" {
			continue
		}
		parts := strings.Split(node.ValueProgram, "|")
		if len(parts) == 0 {
			continue
		}
		fields, err := computedFields(node.ValueProgram)
		if err != nil {
			return nil, nil, SourceReconstruction{}, fmt.Errorf("activity %s computes record: %w", node.Name, err)
		}
		switch parts[0] {
		case "metric":
			metric, err := metricFromFields(fields)
			if err != nil {
				return nil, nil, SourceReconstruction{}, fmt.Errorf("activity %s metric record: %w", node.Name, err)
			}
			metrics = append(metrics, metric)
			reconstruction.Numerator++
			reconstruction.MetricFieldsNumerator += len(fields)
		case "scenario":
			scenario, err := scenarioFromFields(fields)
			if err != nil {
				return nil, nil, SourceReconstruction{}, fmt.Errorf("activity %s scenario record: %w", node.Name, err)
			}
			scenarios = append(scenarios, scenario)
			reconstruction.Numerator++
			reconstruction.ScenarioNumerator += len(fields)
		}
	}
	if len(metrics) == 0 || len(scenarios) == 0 {
		return nil, nil, SourceReconstruction{}, fmt.Errorf("semantic model has no metric and scenario records")
	}
	sort.Slice(metrics, func(i, j int) bool { return metrics[i].ID < metrics[j].ID })
	sort.Slice(scenarios, func(i, j int) bool { return scenarios[i].ID < scenarios[j].ID })
	reconstruction.Denominator = len(metrics) + len(scenarios)
	reconstruction.MetricFieldsDenominator = len(metrics) * 8
	reconstruction.ScenarioDenominator = len(scenarios) * 4
	if reconstruction.MetricFieldsNumerator != reconstruction.MetricFieldsDenominator || reconstruction.ScenarioNumerator != reconstruction.ScenarioDenominator {
		return nil, nil, SourceReconstruction{}, fmt.Errorf("semantic record reconstruction is incomplete")
	}
	return metrics, scenarios, reconstruction, nil
}

func computedFields(value string) (map[string]string, error) {
	parts := strings.Split(value, "|")
	if len(parts) < 2 || (parts[0] != "metric" && parts[0] != "scenario") {
		return nil, fmt.Errorf("unsupported computes value %q", value)
	}
	fields := make(map[string]string, len(parts)-1)
	for _, part := range parts[1:] {
		key, raw, ok := strings.Cut(part, "=")
		if !ok || key == "" || fields[key] != "" {
			if !ok || key == "" {
				return nil, fmt.Errorf("malformed field %q", part)
			}
			return nil, fmt.Errorf("duplicate field %q", key)
		}
		fields[key] = raw
	}
	return fields, nil
}

func metricFromFields(fields map[string]string) (metricSpec, error) {
	keys := []string{"id", "family", "prior_claim", "producer", "consumer", "meta_operation", "evidence_path", "depends_on"}
	for _, key := range keys {
		if _, ok := fields[key]; !ok {
			return metricSpec{}, fmt.Errorf("missing %s", key)
		}
	}
	metric := metricSpec{
		ID:            fields["id"],
		Family:        fields["family"],
		PriorClaim:    fields["prior_claim"],
		Producer:      fields["producer"],
		Consumer:      fields["consumer"],
		MetaOperation: fields["meta_operation"],
		EvidencePath:  fields["evidence_path"],
	}
	if fields["depends_on"] != "" {
		metric.DependsOn = strings.Split(fields["depends_on"], ",")
	}
	if metric.ID == "" || metric.Family == "" || metric.PriorClaim == "" {
		return metricSpec{}, fmt.Errorf("metric identity, family, and prior claim are required")
	}
	return metric, nil
}

func scenarioFromFields(fields map[string]string) (scenarioSpec, error) {
	keys := []string{"id", "remove_relation", "dependency", "reason"}
	for _, key := range keys {
		if _, ok := fields[key]; !ok {
			return scenarioSpec{}, fmt.Errorf("missing %s", key)
		}
	}
	if fields["id"] == "" || fields["reason"] == "" {
		return scenarioSpec{}, fmt.Errorf("scenario identity and reason are required")
	}
	return scenarioSpec{ID: fields["id"], RemoveRelation: fields["remove_relation"], Dependency: fields["dependency"], Reason: fields["reason"]}, nil
}

func evaluateScenario(metrics []metricSpec, scenario scenarioSpec, sourceDigest, semanticDigest string) (ScenarioResult, error) {
	working := fixtureFromMetrics(metrics, scenario)
	if scenario.RemoveRelation != "" {
		parts := strings.SplitN(scenario.RemoveRelation, ":", 2)
		if len(parts) != 2 {
			return ScenarioResult{}, fmt.Errorf("scenario %s has malformed relation mutation", scenario.ID)
		}
		working.Edges = removeRelation(working.Edges, metrics, parts[0], parts[1])
	}
	if scenario.Dependency != "" {
		parts := strings.SplitN(scenario.Dependency, ">", 2)
		if len(parts) != 2 {
			return ScenarioResult{}, fmt.Errorf("scenario %s has malformed dependency mutation", scenario.ID)
		}
		for index := range working.Metrics {
			if working.Metrics[index].ID == parts[1] {
				working.Metrics[index].DependsOn = append(working.Metrics[index].DependsOn, parts[0])
			}
		}
	}
	return evaluateFixture(working, sourceDigest, semanticDigest), nil
}

func fixtureFromMetrics(metrics []metricSpec, scenario scenarioSpec) fixture {
	working := fixture{ID: scenario.ID, Mutation: mutationDescription(scenario), Nodes: make(map[string]string), Metrics: append([]metricSpec(nil), metrics...)}
	for _, metric := range metrics {
		working.Nodes["metric:"+metric.ID] = "metric"
		for _, value := range []struct{ id, kind string }{
			{metric.Producer, "producer"}, {metric.Consumer, "consumer"}, {metric.MetaOperation, "meta-operation"}, {metric.EvidencePath, "evidence-path"},
		} {
			if value.id != "" {
				working.Nodes[value.id] = value.kind
			}
		}
		for _, link := range relations(metric) {
			if link.From != "" && link.To != "" {
				working.Edges = append(working.Edges, edge{From: link.From, To: link.To, Kind: link.Kind})
			}
		}
	}
	return working
}

func mutationDescription(scenario scenarioSpec) string {
	return "remove_relation=" + scenario.RemoveRelation + ";dependency=" + scenario.Dependency + ";reason=" + scenario.Reason
}

func relations(metric metricSpec) []relation {
	return []relation{
		{Kind: "PRODUCES", From: metric.Producer, To: "metric:" + metric.ID},
		{Kind: "CONSUMES", From: "metric:" + metric.ID, To: metric.Consumer},
		{Kind: "OPERATES", From: metric.MetaOperation, To: "metric:" + metric.ID},
		{Kind: "EVIDENCED_BY", From: "metric:" + metric.ID, To: metric.EvidencePath},
	}
}

func removeRelation(edges []edge, metrics []metricSpec, kind, metricID string) []edge {
	var wanted relation
	for _, metric := range metrics {
		if metric.ID != metricID {
			continue
		}
		for _, link := range relations(metric) {
			if link.Kind == kind {
				wanted = link
			}
		}
	}
	filtered := make([]edge, 0, len(edges))
	for _, current := range edges {
		if current.Kind == wanted.Kind && current.From == wanted.From && current.To == wanted.To {
			continue
		}
		filtered = append(filtered, current)
	}
	return filtered
}

func evaluateFixture(fixture fixture, sourceDigest, semanticDigest string) ScenarioResult {
	nodes := fixture.Nodes
	edgeCounts := make(map[string]int)
	edgeKinds := make(map[string]int)
	for _, current := range fixture.Edges {
		if nodes[current.From] != "" && nodes[current.To] != "" {
			edgeCounts[current.From+"\x00"+current.To+"\x00"+current.Kind]++
			edgeKinds[current.Kind]++
		}
	}
	byID := make(map[string]metricSpec, len(fixture.Metrics))
	for _, metric := range fixture.Metrics {
		byID[metric.ID] = metric
	}
	memo := make(map[string]MetricResult, len(fixture.Metrics))
	visiting := make(map[string]bool)
	var evaluate func(string) MetricResult
	evaluate = func(id string) MetricResult {
		if result, ok := memo[id]; ok {
			return result
		}
		metric := byID[id]
		if visiting[id] {
			return metricResult(metric, edgeCounts, "UNKNOWN", &Issue{Stage: "DEPENDENCY", Step: "detect-cycle", Reason: "DEPENDENCY_CYCLE", Cause: "DEPENDENCY_BLOCK", BlockedBy: []string{id}}, sourceDigest, semanticDigest, fixture)
		}
		visiting[id] = true
		result := metricResult(metric, edgeCounts, "", nil, sourceDigest, semanticDigest, fixture)
		for _, dependency := range metric.DependsOn {
			upstream, ok := byID[dependency]
			if !ok {
				result = metricResult(metric, edgeCounts, "UNKNOWN", &Issue{Stage: "DEPENDENCY", Step: "resolve-upstream", Reason: "UPSTREAM_METRIC_MISSING", Cause: "DEPENDENCY_BLOCK", BlockedBy: []string{dependency}}, sourceDigest, semanticDigest, fixture)
				break
			}
			upstreamResult := evaluate(upstream.ID)
			if upstreamResult.Decision != "PASS" {
				result = metricResult(metric, edgeCounts, "UNKNOWN", &Issue{Stage: "DEPENDENCY", Step: "propagate-unknown", Reason: "UPSTREAM_" + upstreamResult.Decision, Cause: "DEPENDENCY_BLOCK", BlockedBy: []string{dependency}}, sourceDigest, semanticDigest, fixture)
				break
			}
		}
		visiting[id] = false
		memo[id] = result
		return result
	}

	results := make([]MetricResult, 0, len(fixture.Metrics))
	decisions := make(map[string]int)
	transitions := make(map[string]int)
	numerator := 0
	for _, metric := range fixture.Metrics {
		result := evaluate(metric.ID)
		results = append(results, result)
		decisions[result.Decision]++
		transitions[result.Transition.Transition]++
		numerator += result.Numerator
	}
	decision := "PASS"
	if decisions["FAIL_CLOSED"] > 0 {
		decision = "FAIL_CLOSED"
	} else if decisions["UNKNOWN"] > 0 {
		decision = "UNKNOWN"
	}
	return ScenarioResult{
		ID: fixture.ID, Mutation: fixture.Mutation,
		Graph:     GraphSummary{Nodes: len(nodes), Edges: len(fixture.Edges), EdgeKinds: edgeKinds},
		Numerator: numerator, Denominator: len(results) * relationDenominator,
		ConformanceDecision: decision, SubjectResolution: "EXACT",
		Decisions: decisions, Transitions: transitions, Metrics: results,
	}
}

func metricResult(metric metricSpec, edgeCounts map[string]int, forcedDecision string, forcedIssue *Issue, sourceDigest, semanticDigest string, fixture fixture) MetricResult {
	result := MetricResult{
		ID: metric.ID, Family: metric.Family, Claim: metric.PriorClaim,
		Denominator: relationDenominator, SubjectResolution: "EXACT", EvaluationState: "EVALUATED",
	}
	for _, link := range relations(metric) {
		count := edgeCounts[link.From+"\x00"+link.To+"\x00"+link.Kind]
		if count != 1 {
			continue
		}
		result.Numerator++
		switch link.Kind {
		case "PRODUCES":
			result.Lineage.Producer = link.From
		case "CONSUMES":
			result.Lineage.Consumer = link.To
		case "OPERATES":
			result.Lineage.MetaOperation = link.From
		case "EVIDENCED_BY":
			result.Lineage.EvidencePath = link.To
		}
	}
	if forcedDecision != "" {
		result.Decision, result.Issue = forcedDecision, forcedIssue
	} else if result.Numerator == result.Denominator {
		result.Decision = "PASS"
	} else if result.Lineage.Consumer == "" {
		result.Decision = "FAIL_CLOSED"
		result.Issue = &Issue{Stage: "LINEAGE", Step: "connect-consumer", Reason: "REQUIRED_CONSUMER_RELATION_MISSING", Cause: "DIRECT_CAUSE"}
	} else if result.Lineage.Producer == "" {
		result.Decision = "UNKNOWN"
		result.Issue = &Issue{Stage: "LINEAGE", Step: "connect-producer", Reason: "REQUIRED_PRODUCER_RELATION_MISSING", Cause: "DIRECT_CAUSE"}
	} else if result.Lineage.MetaOperation == "" {
		result.Decision = "UNKNOWN"
		result.Issue = &Issue{Stage: "LINEAGE", Step: "connect-meta-operation", Reason: "REQUIRED_META_OPERATION_RELATION_MISSING", Cause: "DIRECT_CAUSE"}
	} else {
		result.Decision = "UNKNOWN"
		result.Issue = &Issue{Stage: "LINEAGE", Step: "connect-evidence-path", Reason: "REQUIRED_EVIDENCE_RELATION_MISSING", Cause: "DIRECT_CAUSE"}
	}
	result.Transition = makeTransition(result, sourceDigest, semanticDigest, fixture)
	return result
}

func makeTransition(result MetricResult, sourceDigest, semanticDigest string, fixture fixture) ClaimTransition {
	transition := ClaimTransition{
		PriorClaim: result.Claim, ConformanceDecision: result.Decision, SubjectResolution: result.SubjectResolution,
		Stage: "CLAIM", Provenance: Provenance{SourceDigest: sourceDigest, SemanticDigest: semanticDigest, Producer: result.Lineage.Producer, Consumer: result.Lineage.Consumer, MetaOperation: result.Lineage.MetaOperation, EvidencePath: result.Lineage.EvidencePath, ScenarioMutation: fixture.Mutation},
	}
	switch result.Decision {
	case "PASS":
		transition.NextClaim, transition.Transition = "DISCHARGED", "DISCHARGED"
		transition.Step, transition.Reason = "discharge-open-claim", "PASS_DISCHARGES_OPEN_CLAIM"
	case "FAIL_CLOSED":
		transition.NextClaim, transition.Transition = "REFUTED", "REFUTED"
		transition.Step, transition.Reason = "refute-claim", "FAIL_CLOSED_IS_EXPLICIT_REFUTATION"
	default:
		transition.NextClaim, transition.Transition = "OPEN", "PRESERVED_OPEN"
		transition.Step, transition.Reason = "preserve-open-claim", "UNKNOWN_PRESERVES_OPEN_CLAIM"
	}
	evidence := struct {
		MetricID   string     `json:"metric_id"`
		Decision   string     `json:"decision"`
		Issue      *Issue     `json:"issue,omitempty"`
		Provenance Provenance `json:"provenance"`
	}{result.ID, result.Decision, result.Issue, transition.Provenance}
	payload, err := json.Marshal(evidence)
	if err != nil {
		transition.EvidenceDigest = digestText(result.ID + "|" + result.Decision + "|" + fixture.Mutation)
	} else {
		transition.EvidenceDigest = digestBytes(payload)
	}
	return transition
}
