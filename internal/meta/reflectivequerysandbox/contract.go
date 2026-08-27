package reflectivequerysandbox

import (
	"fmt"
	"runtime"
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/query"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

type claimSpec struct {
	ID, Class, ProofChoice, MetaOperation, EvidenceAttempt, PriorState string
	NodeID                                                             semantic.ID
}

type operationSpec struct {
	ID      semantic.ID
	Program string
}

type sourceModel struct {
	Claims         []claimSpec
	Operations     []operationSpec
	QuerySubject   semantic.ID
	MutationTarget semantic.ID
	ReceiptTarget  semantic.ID
	MetricTarget   semantic.ID
	UnknownTarget  query.ID
}

func deriveSourceModel(ir semantic.IR) (sourceModel, error) {
	model := sourceModel{}
	for _, node := range ir.Graph.Nodes() {
		id := node.ID.String()
		switch {
		case node.Kind == semantic.Entity && strings.Contains(id, "/claim/"):
			claim, err := parseClaim(node)
			if err != nil {
				return sourceModel{}, err
			}
			model.Claims = append(model.Claims, claim)
		case node.Kind == semantic.Entity && strings.Contains(id, "/subject/query"):
			model.QuerySubject = node.ID
		case node.Kind == semantic.Entity && strings.Contains(id, "/mutation/request"):
			model.MutationTarget = node.ID
		case node.Kind == semantic.Entity && strings.Contains(id, "/metric/relation"):
			model.MetricTarget = node.ID
		case node.Kind == semantic.Entity && strings.Contains(id, "/receipt/query"):
			model.ReceiptTarget = node.ID
		case node.Kind == semantic.Activity && isSandboxOperation(node.ValueProgram):
			model.Operations = append(model.Operations, operationSpec{ID: node.ID, Program: node.ValueProgram})
		}
	}
	if len(model.Claims) == 0 {
		return sourceModel{}, fmt.Errorf("source declares no claim nodes")
	}
	if model.QuerySubject == "" || model.MutationTarget == "" || model.ReceiptTarget == "" || model.MetricTarget == "" {
		return sourceModel{}, fmt.Errorf("source is missing a formal query subject, mutation target, receipt target, or metric relation")
	}
	sort.Slice(model.Claims, func(i, j int) bool { return model.Claims[i].ID < model.Claims[j].ID })
	sort.Slice(model.Operations, func(i, j int) bool { return model.Operations[i].ID < model.Operations[j].ID })
	for _, operation := range model.Operations {
		if strings.HasPrefix(operation.Program, "reflect.query:") && strings.HasSuffix(operation.Program, ":metrics") {
			model.UnknownTarget = query.ID(operation.ID.String() + "/unknown-target")
			break
		}
	}
	if model.UnknownTarget == "" {
		return sourceModel{}, fmt.Errorf("source declares no metric query operation for unknown-target probe")
	}
	return model, nil
}

func parseClaim(node semantic.Node) (claimSpec, error) {
	const prefix = "/claim/"
	parts := strings.Split(strings.TrimPrefix(node.ID.String(), strings.Split(node.ID.String(), prefix)[0]+prefix), "/")
	if len(parts) != 6 || parts[5] == "" {
		return claimSpec{}, fmt.Errorf("claim %q must encode class/name/proof/meta-operation/evidence/prior-state", node.ID)
	}
	class, proof, prior := strings.ToUpper(parts[0]), strings.ToUpper(parts[2]), strings.ToUpper(parts[5])
	if class == "" || proof == "" || prior == "" {
		return claimSpec{}, fmt.Errorf("claim %q has an empty semantic coordinate", node.ID)
	}
	return claimSpec{
		ID:              strings.ToLower(parts[0]) + "." + parts[1],
		Class:           class,
		ProofChoice:     proof,
		MetaOperation:   parts[3],
		EvidenceAttempt: parts[4],
		PriorState:      prior,
		NodeID:          node.ID,
	}, nil
}

func isSandboxOperation(program string) bool {
	return strings.HasPrefix(program, "reflect.query:") || strings.HasPrefix(program, "reflect.attempt:") || strings.HasPrefix(program, "reflect.observation:")
}

func sourceTargets(ir semantic.IR, activity semantic.ID) []semantic.ID {
	targets := make([]semantic.ID, 0)
	for _, fact := range ir.Graph.DeterministicFacts() {
		if fact.Subject == activity && fact.Predicate == semantic.Used {
			targets = append(targets, fact.Object)
		}
	}
	sort.Slice(targets, func(i, j int) bool { return targets[i] < targets[j] })
	return targets
}

func targetForOperation(ir semantic.IR, operation operationSpec, model sourceModel) (semantic.ID, error) {
	targets := sourceTargets(ir, operation.ID)
	marker := ""
	switch operation.Program {
	case "reflect.query:structure":
		marker = "/subject/structure"
	case "reflect.query:claims":
		marker = "/claim-state/open"
	case "reflect.query:metrics":
		marker = "/metric/relation"
	case "reflect.attempt:mutation":
		marker = "/mutation/request"
	case "reflect.observation:receipt-seal":
		marker = "/receipt/query"
	}
	for _, target := range targets {
		if marker == "" || strings.Contains(target.String(), marker) {
			return target, nil
		}
	}
	return "", fmt.Errorf("operation %q has no source-backed target for marker %q", operation.Program, marker)
}

func buildContract(model sourceModel, snapshot Snapshot, attempts []Attempt, transitions []ClaimTransition) Contract {
	classes := bucketClaims(model.Claims, func(claim claimSpec) string { return claim.Class })
	proofs := bucketClaims(model.Claims, func(claim claimSpec) string { return claim.ProofChoice })
	contract := Contract{
		Schema:            Schema,
		MetricID:          MetricID,
		GoVersion:         runtime.Version(),
		Denominator:       len(model.Claims),
		Classes:           classes,
		Proofs:            proofs,
		SourceNodes:       snapshot.NodeCount,
		SourceFacts:       snapshot.FactCount,
		ClaimCount:        len(model.Claims),
		AttemptCount:      len(attempts),
		ReflectiveQueries: countAttempts(attempts, func(attempt Attempt) bool { return attempt.Operation == "query" }),
		SafeQueries: countAttempts(attempts, func(attempt Attempt) bool {
			return attempt.Operation == "query" && attempt.Decision == "PASS" && attempt.Resolution == "EXACT"
		}),
		DeniedMutations:     countAttempts(attempts, func(attempt Attempt) bool { return attempt.Operation == "mutate" && attempt.Decision == "DENIED" }),
		UnknownTargets:      countAttempts(attempts, func(attempt Attempt) bool { return attempt.Decision == "UNKNOWN" }),
		RefutedAttempts:     countAttempts(attempts, func(attempt Attempt) bool { return attempt.Decision == "REFUTED" }),
		TransitionCount:     len(transitions),
		SatisfiedIndicators: countTransitions(transitions, "DISCHARGED"),
	}
	return contract
}

func bucketClaims(claims []claimSpec, key func(claimSpec) string) []Bucket {
	totals := make(map[string]int)
	for _, claim := range claims {
		totals[key(claim)]++
	}
	names := make([]string, 0, len(totals))
	for name := range totals {
		names = append(names, name)
	}
	sort.Strings(names)
	buckets := make([]Bucket, 0, len(names))
	for _, name := range names {
		buckets = append(buckets, Bucket{Name: name, Total: totals[name]})
	}
	return buckets
}

func countAttempts(attempts []Attempt, predicate func(Attempt) bool) int {
	count := 0
	for _, attempt := range attempts {
		if predicate(attempt) {
			count++
		}
	}
	return count
}

func countTransitions(transitions []ClaimTransition, state string) int {
	count := 0
	for _, transition := range transitions {
		if transition.To == state && transition.From != state {
			count++
		}
	}
	return count
}
