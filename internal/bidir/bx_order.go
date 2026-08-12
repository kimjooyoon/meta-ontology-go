package bidir

import (
	"fmt"
	"sort"
	"strings"
)

func orderedSequences(model Model) ([]string, []string) {
	return orderedPortSequence(model), orderedRelationSequence(model)
}

func orderedPortSequence(model Model) []string {
	activities := make([]Node, 0)
	for _, node := range model.Nodes {
		if node.Kind == ActivityKind {
			activities = append(activities, node)
		}
	}
	sort.Slice(activities, func(i, j int) bool { return activities[i].ID < activities[j].ID })
	sequence := make([]string, 0)
	for _, activity := range activities {
		inputs, outputs := activityPorts(model, activity.ID)
		sort.SliceStable(inputs, func(i, j int) bool { return referenceSourceOrderLess(inputs[i], inputs[j]) })
		sort.SliceStable(outputs, func(i, j int) bool { return referenceSourceOrderLess(outputs[i], outputs[j]) })
		appendPortSequence(&sequence, activity.ID, "input", inputs)
		appendPortSequence(&sequence, activity.ID, "output", outputs)
	}
	return sequence
}

func activityPorts(model Model, activity ID) ([]Reference, []Reference) {
	var inputs, outputs []Reference
	for _, relation := range model.Relations {
		if relation.Kind == PredicateUsed && relation.Source == activity {
			if node, exists := model.node(relation.Target); exists {
				inputs = append(inputs, relationReference(node, relation))
			}
		}
		if relation.Kind == PredicateWasGeneratedBy && relation.Target == activity {
			if node, exists := model.node(relation.Source); exists {
				outputs = append(outputs, relationReference(node, relation))
			}
		}
	}
	return inputs, outputs
}

func relationReference(node Node, relation Relation) Reference {
	return Reference{ID: node.ID, Name: node.Name, Namespace: node.Namespace, Span: relation.Span}
}

func appendPortSequence(sequence *[]string, activity ID, direction string, references []Reference) {
	for index, reference := range references {
		*sequence = append(*sequence, fmt.Sprintf("%s|%s|%d|%s|%s", direction, activity, index, reference.ID, spanText(reference.Span)))
	}
}

func orderedRelationSequence(model Model) []string {
	relations := append([]Relation(nil), model.Relations...)
	sort.SliceStable(relations, func(i, j int) bool { return relationSourceOrderLess(relations[i], relations[j]) })
	sequence := make([]string, len(relations))
	for index, relation := range relations {
		sequence[index] = fmt.Sprintf("%s|%s|%s|%s|%s", relation.Kind, relation.Source, relation.Target, relation.ID, spanText(relation.Span))
	}
	return sequence
}

func relationSourceOrderLess(left, right Relation) bool {
	leftValid, rightValid := left.Span.Valid(), right.Span.Valid()
	if leftValid != rightValid {
		return leftValid
	}
	if leftValid && !sameSpan(left.Span, right.Span) {
		return spanLess(left.Span, right.Span)
	}
	return relationLess(left, right)
}

func sameSpan(left, right SourceSpan) bool {
	return left == right
}

func spanText(span SourceSpan) string {
	var builder strings.Builder
	writeSpan(&builder, span)
	return builder.String()
}

func sequenceHash(sequence []string) string {
	var builder strings.Builder
	for _, value := range sequence {
		writePart(&builder, value)
	}
	return digest(builder.String())
}
