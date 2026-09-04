package semantic

import (
	"fmt"
	"strings"
)

// Policy is the semantic IR for a first-class meta-programming policy. The
// child node slices retain source order because transition order is the
// precedence order used by a policy evaluator.
type Policy struct {
	ID          ID
	Name        string
	States      []PolicyState
	Transitions []PolicyTransition
	Cases       []PolicyCase
	Span        Span
}

type PolicyState struct {
	Name string
	Span Span
}

type PolicyTransition struct {
	From string
	To   string
	Span Span
}

type PolicyCase struct {
	Name       string
	Evidence   []PolicyEvidence
	Resolution PolicyResolution
	Span       Span
}

type PolicyEvidence struct {
	Name  string
	Value string
	Span  Span
}

type PolicyResolution struct {
	Decision       string
	Stage          string
	Step           int
	Reason         string
	DecisionStage  string
	DecisionStep   int
	DecisionReason string
	UnknownClass   string
	NextOperation  string
	BlockedBy      []string
	Role           string
	MetaOperation  string
	ProofChoice    string
	Claim          string
	Span           Span
}

func (p Policy) Normalized() (Policy, error) {
	id, err := ParseIdentity(p.ID.String())
	if err != nil {
		return Policy{}, fmt.Errorf("%w: policy id: %v", ErrInvalidNode, err)
	}
	name := strings.TrimSpace(p.Name)
	if name == "" {
		return Policy{}, fmt.Errorf("%w: policy name is empty", ErrInvalidNode)
	}
	span := p.Span.Normalized()
	if err := span.Validate(); err != nil {
		return Policy{}, fmt.Errorf("%w: policy span: %v", ErrInvalidNode, err)
	}
	out := Policy{ID: id, Name: name, Span: span,
		States:      append([]PolicyState(nil), p.States...),
		Transitions: append([]PolicyTransition(nil), p.Transitions...),
		Cases:       make([]PolicyCase, len(p.Cases))}
	for index, current := range p.Cases {
		current.Name = strings.TrimSpace(current.Name)
		if current.Name == "" {
			return Policy{}, fmt.Errorf("%w: policy case %d has an empty name", ErrInvalidNode, index)
		}
		if len(current.Evidence) == 0 {
			return Policy{}, fmt.Errorf("%w: policy case %q has no evidence", ErrInvalidNode, current.Name)
		}
		current.Evidence = append([]PolicyEvidence(nil), current.Evidence...)
		for evidenceIndex, evidence := range current.Evidence {
			evidence.Name = strings.TrimSpace(evidence.Name)
			evidence.Value = strings.TrimSpace(evidence.Value)
			if evidence.Name == "" || evidence.Value == "" {
				return Policy{}, fmt.Errorf("%w: policy case %q evidence %d is incomplete", ErrInvalidNode, current.Name, evidenceIndex)
			}
			current.Evidence[evidenceIndex] = evidence
		}
		resolution := current.Resolution
		resolution.Decision = strings.TrimSpace(resolution.Decision)
		resolution.Stage = strings.TrimSpace(resolution.Stage)
		resolution.Reason = strings.TrimSpace(resolution.Reason)
		resolution.DecisionStage = strings.TrimSpace(resolution.DecisionStage)
		resolution.DecisionReason = strings.TrimSpace(resolution.DecisionReason)
		resolution.UnknownClass = strings.TrimSpace(resolution.UnknownClass)
		resolution.NextOperation = strings.TrimSpace(resolution.NextOperation)
		resolution.Role = strings.TrimSpace(resolution.Role)
		resolution.MetaOperation = strings.TrimSpace(resolution.MetaOperation)
		resolution.ProofChoice = strings.TrimSpace(resolution.ProofChoice)
		resolution.Claim = strings.TrimSpace(resolution.Claim)
		resolution.BlockedBy = append([]string(nil), resolution.BlockedBy...)
		for blockedIndex, blocked := range resolution.BlockedBy {
			resolution.BlockedBy[blockedIndex] = strings.TrimSpace(blocked)
			if resolution.BlockedBy[blockedIndex] == "" {
				return Policy{}, fmt.Errorf("%w: policy case %q has an empty blocked_by entry", ErrInvalidNode, current.Name)
			}
		}
		if resolution.Decision == "" || resolution.Stage == "" || resolution.Step <= 0 || resolution.Reason == "" || resolution.DecisionStage == "" || resolution.DecisionStep <= 0 || resolution.DecisionReason == "" || resolution.Role == "" || resolution.MetaOperation == "" || resolution.ProofChoice == "" || resolution.Claim == "" {
			return Policy{}, fmt.Errorf("%w: policy case %q has incomplete resolution", ErrInvalidNode, current.Name)
		}
		current.Resolution = resolution
		out.Cases[index] = current
	}
	if err := out.Validate(); err != nil {
		return Policy{}, err
	}
	return out, nil
}

func (p Policy) Validate() error {
	if _, err := ParseIdentity(p.ID.String()); err != nil {
		return fmt.Errorf("%w: policy id: %v", ErrInvalidNode, err)
	}
	if strings.TrimSpace(p.Name) == "" {
		return fmt.Errorf("%w: policy name is empty", ErrInvalidNode)
	}
	states := make(map[string]struct{}, len(p.States))
	for _, state := range p.States {
		name := strings.TrimSpace(state.Name)
		if name == "" {
			return fmt.Errorf("%w: policy state name is empty", ErrInvalidNode)
		}
		if _, exists := states[name]; exists {
			return fmt.Errorf("%w: duplicate policy state %q", ErrInvalidNode, name)
		}
		states[name] = struct{}{}
	}
	transitions := make(map[string]struct{}, len(p.Transitions))
	for _, transition := range p.Transitions {
		if _, ok := states[transition.From]; !ok {
			return fmt.Errorf("%w: transition source %q is not declared", ErrInvalidNode, transition.From)
		}
		if _, ok := states[transition.To]; !ok {
			return fmt.Errorf("%w: transition target %q is not declared", ErrInvalidNode, transition.To)
		}
		key := transition.From + "\x00" + transition.To
		if _, exists := transitions[key]; exists {
			return fmt.Errorf("%w: duplicate policy transition %q -> %q", ErrInvalidNode, transition.From, transition.To)
		}
		transitions[key] = struct{}{}
	}
	cases := make(map[string]struct{}, len(p.Cases))
	for _, current := range p.Cases {
		if _, exists := states[current.Name]; !exists {
			return fmt.Errorf("%w: case %q is not a declared state", ErrInvalidNode, current.Name)
		}
		if _, exists := states[current.Resolution.Decision]; !exists {
			return fmt.Errorf("%w: case %q resolves to undeclared state %q", ErrInvalidNode, current.Name, current.Resolution.Decision)
		}
		if _, exists := transitions[current.Name+"\x00"+current.Resolution.Decision]; !exists {
			return fmt.Errorf("%w: case %q has no matching transition", ErrInvalidNode, current.Name)
		}
		if _, exists := cases[current.Name]; exists {
			return fmt.Errorf("%w: duplicate policy case %q", ErrInvalidNode, current.Name)
		}
		cases[current.Name] = struct{}{}
	}
	return nil
}

func (p Policy) Canonical() string {
	if normalized, err := p.Normalized(); err == nil {
		p = normalized
	}
	var b strings.Builder
	b.WriteString("policy\t")
	writeCanonicalField(&b, p.ID.String())
	writeCanonicalField(&b, p.Name)
	writeCanonicalSpan(&b, p.Span)
	for _, state := range p.States {
		b.WriteString("policy-state\t")
		writeCanonicalField(&b, state.Name)
		writeCanonicalSpan(&b, state.Span)
	}
	for _, transition := range p.Transitions {
		b.WriteString("policy-transition\t")
		writeCanonicalField(&b, transition.From)
		writeCanonicalField(&b, transition.To)
		writeCanonicalSpan(&b, transition.Span)
	}
	for _, current := range p.Cases {
		b.WriteString("policy-case\t")
		writeCanonicalField(&b, current.Name)
		writeCanonicalSpan(&b, current.Span)
		for _, evidence := range current.Evidence {
			b.WriteString("policy-evidence\t")
			writeCanonicalField(&b, evidence.Name)
			writeCanonicalField(&b, evidence.Value)
			writeCanonicalSpan(&b, evidence.Span)
		}
		b.WriteString("policy-resolution\t")
		writeResolutionCanonical(&b, current.Resolution, true)
	}
	return b.String()
}

func (p Policy) SemanticCanonical() string {
	if normalized, err := p.Normalized(); err == nil {
		p = normalized
	}
	var b strings.Builder
	b.WriteString("policy\t")
	writeCanonicalField(&b, p.ID.String())
	for _, state := range p.States {
		b.WriteString("policy-state\t")
		writeCanonicalField(&b, state.Name)
	}
	for _, transition := range p.Transitions {
		b.WriteString("policy-transition\t")
		writeCanonicalField(&b, transition.From)
		writeCanonicalField(&b, transition.To)
	}
	for _, current := range p.Cases {
		b.WriteString("policy-case\t")
		writeCanonicalField(&b, current.Name)
		for _, evidence := range current.Evidence {
			b.WriteString("policy-evidence\t")
			writeCanonicalField(&b, evidence.Name)
			writeCanonicalField(&b, evidence.Value)
		}
		b.WriteString("policy-resolution\t")
		writeResolutionCanonical(&b, current.Resolution, false)
	}
	return b.String()
}

func writeResolutionCanonical(builder *strings.Builder, resolution PolicyResolution, withSpan bool) {
	writeCanonicalField(builder, resolution.Decision)
	writeCanonicalField(builder, resolution.Stage)
	writeCanonicalField(builder, fmt.Sprint(resolution.Step))
	writeCanonicalField(builder, resolution.Reason)
	writeCanonicalField(builder, resolution.DecisionStage)
	writeCanonicalField(builder, fmt.Sprint(resolution.DecisionStep))
	writeCanonicalField(builder, resolution.DecisionReason)
	writeCanonicalField(builder, resolution.UnknownClass)
	writeCanonicalField(builder, resolution.NextOperation)
	for _, blocked := range resolution.BlockedBy {
		writeCanonicalField(builder, blocked)
	}
	writeCanonicalField(builder, resolution.Role)
	writeCanonicalField(builder, resolution.MetaOperation)
	writeCanonicalField(builder, resolution.ProofChoice)
	writeCanonicalField(builder, resolution.Claim)
	if withSpan {
		writeCanonicalSpan(builder, resolution.Span)
	}
	builder.WriteByte('\n')
}

// PolicyNodeCounts reports the exact first-class grammar node counts in a
// deterministic map for CI evidence and tooling.
func (p Policy) PolicyNodeCounts() map[string]int {
	counts := map[string]int{
		"policy": 1, "state": len(p.States), "transition": len(p.Transitions),
		"case": len(p.Cases), "evidence": 0, "resolution": 0,
	}
	for _, current := range p.Cases {
		counts["evidence"] += len(current.Evidence)
		counts["resolution"]++
	}
	return counts
}

func validatePolicies(policies []Policy) error {
	seen := make(map[ID]struct{}, len(policies))
	for _, policy := range policies {
		normalized, err := policy.Normalized()
		if err != nil {
			return err
		}
		if _, exists := seen[normalized.ID]; exists {
			return fmt.Errorf("%w: duplicate policy %s", ErrInvalidNode, normalized.ID)
		}
		seen[normalized.ID] = struct{}{}
	}
	return nil
}
