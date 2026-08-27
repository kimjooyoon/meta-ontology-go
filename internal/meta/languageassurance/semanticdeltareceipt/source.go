package semanticdeltareceipt

import (
	"fmt"
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/detection/semanticdelta"
)

type projectedSource struct {
	nodes  []Node
	facts  []Fact
	claims []Claim
}

type entityDecl struct {
	name string
	id   string
}

type activityDecl struct {
	name   string
	inputs []string
	output string
}

func projectSource(raw []byte) (projectedSource, error) {
	entities := map[string]entityDecl{}
	var activities []activityDecl
	packageSeen, namespaceSeen := false, false
	for lineNumber, rawLine := range strings.Split(strings.ReplaceAll(string(raw), "\r\n", "\n"), "\n") {
		line := strings.TrimSpace(rawLine)
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		switch fields[0] {
		case "package":
			if len(fields) != 2 || packageSeen {
				return projectedSource{}, fmt.Errorf("line %d: invalid package declaration", lineNumber+1)
			}
			packageSeen = true
		case "namespace":
			if len(fields) != 2 || namespaceSeen {
				return projectedSource{}, fmt.Errorf("line %d: invalid namespace declaration", lineNumber+1)
			}
			namespaceSeen = true
		case "entity":
			entity, err := parseEntity(fields)
			if err != nil {
				return projectedSource{}, fmt.Errorf("line %d: %w", lineNumber+1, err)
			}
			if _, exists := entities[entity.name]; exists {
				return projectedSource{}, fmt.Errorf("line %d: duplicate entity", lineNumber+1)
			}
			entities[entity.name] = entity
		case "activity":
			activity, err := parseActivity(line)
			if err != nil {
				return projectedSource{}, fmt.Errorf("line %d: %w", lineNumber+1, err)
			}
			activities = append(activities, activity)
		default:
			return projectedSource{}, fmt.Errorf("line %d: unsupported Gooo declaration", lineNumber+1)
		}
	}
	if !packageSeen || !namespaceSeen || len(entities) == 0 || len(activities) == 0 {
		return projectedSource{}, fmt.Errorf("incomplete Gooo source")
	}

	result := projectedSource{}
	for _, entity := range entities {
		result.nodes = append(result.nodes, Node{ID: entity.id, Kind: "ENTITY"})
	}
	for _, activity := range activities {
		activityID := activityID(activity.name)
		result.nodes = append(result.nodes, Node{ID: activityID, Kind: "ACTIVITY"})
		for index, input := range activity.inputs {
			entity, ok := entities[input]
			if !ok {
				return projectedSource{}, fmt.Errorf("activity %q references unknown input %q", activity.name, input)
			}
			fact := Fact{Subject: activityID, Predicate: "gooo:uses", Object: entity.id}
			result.facts = append(result.facts, fact)
			result.claims = append(result.claims, claimFor(activityID, "uses", index, fact))
		}
		entity, ok := entities[activity.output]
		if !ok {
			return projectedSource{}, fmt.Errorf("activity %q references unknown output %q", activity.name, activity.output)
		}
		fact := Fact{Subject: activityID, Predicate: "gooo:generates", Object: entity.id}
		result.facts = append(result.facts, fact)
		result.claims = append(result.claims, claimFor(activityID, "generates", 0, fact))
	}
	sort.Slice(result.nodes, func(i, j int) bool { return result.nodes[i].ID < result.nodes[j].ID })
	sort.Slice(result.facts, func(i, j int) bool { return factLess(result.facts[i], result.facts[j]) })
	sort.Slice(result.claims, func(i, j int) bool { return result.claims[i].ID < result.claims[j].ID })
	return result, nil
}

func parseEntity(fields []string) (entityDecl, error) {
	if len(fields) != 4 || fields[2] != "id" || len(fields[3]) < 3 || fields[3][0] != '"' || fields[3][len(fields[3])-1] != '"' {
		return entityDecl{}, fmt.Errorf("invalid entity declaration")
	}
	return entityDecl{name: fields[1], id: strings.Trim(fields[3], "\"")}, nil
}

func parseActivity(line string) (activityDecl, error) {
	body := strings.TrimSpace(strings.TrimPrefix(line, "activity"))
	open, close := strings.IndexByte(body, '('), strings.LastIndexByte(body, ')')
	if open <= 0 || close < open || !strings.HasPrefix(strings.TrimSpace(body[close+1:]), "->") {
		return activityDecl{}, fmt.Errorf("invalid activity declaration")
	}
	name := strings.TrimSpace(body[:open])
	if name == "" {
		return activityDecl{}, fmt.Errorf("activity has no name")
	}
	rawInputs := strings.TrimSpace(body[open+1 : close])
	inputs := []string{}
	if rawInputs != "" {
		for _, rawInput := range strings.Split(rawInputs, ",") {
			input := strings.TrimSpace(rawInput)
			if input == "" {
				return activityDecl{}, fmt.Errorf("activity %q has an empty input", name)
			}
			inputs = append(inputs, input)
		}
	}
	output := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(body[close+1:]), "->"))
	if output == "" || strings.ContainsAny(output, " \t") {
		return activityDecl{}, fmt.Errorf("activity %q has an invalid output", name)
	}
	return activityDecl{name: name, inputs: inputs, output: output}, nil
}

func activityID(name string) string { return "gooo://semantic-delta/activity/" + name }

func claimFor(subject, role string, index int, fact Fact) Claim {
	return Claim{ID: fmt.Sprintf("%s/claim/%s/%d", subject, role, index), Subject: fact.Subject,
		Predicate: fact.Predicate, Object: fact.Object, Status: "ASSERTED",
		Stage: "semantic-extraction", Step: "bind-signature", Reason: "DECLARATION_BOUND"}
}

func factLess(left, right Fact) bool {
	if left.Subject != right.Subject {
		return left.Subject < right.Subject
	}
	if left.Predicate != right.Predicate {
		return left.Predicate < right.Predicate
	}
	return left.Object < right.Object
}

func toSemanticSnapshot(source projectedSource) semanticdelta.Snapshot {
	nodes := make([]semanticdelta.Node, 0, len(source.nodes))
	for _, node := range source.nodes {
		nodes = append(nodes, semanticdelta.Node{ID: node.ID, Kind: node.Kind})
	}
	facts := make([]semanticdelta.Fact, 0, len(source.facts))
	for _, fact := range source.facts {
		facts = append(facts, semanticdelta.Fact{Subject: fact.Subject, Predicate: fact.Predicate, Object: fact.Object})
	}
	return semanticdelta.Snapshot{Nodes: nodes, Facts: facts}
}
