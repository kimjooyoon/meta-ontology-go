package gosplit

import "fmt"

type BlockerKind string

const (
	RenderedPlanUnavailable BlockerKind = "RENDERED_PLAN_UNAVAILABLE"
	NoDeclarationSplit      BlockerKind = "NO_DECLARATION_SPLIT"
	DirectoryFull           BlockerKind = "DIRECTORY_CAPACITY_UNMET"
	DirectoryMixed          BlockerKind = "DIRECTORY_CONTAINS_FOLDERS"
	DirectoryUnobserved     BlockerKind = "DIRECTORY_UNOBSERVED"
)

// Blocker retains native observations without deciding the next Gooo action.
// SourceSHA256 identifies the exact bytes parsed, not a dependency closure or
// permission to publish a later plan. The original error remains inspectable.
type Blocker struct {
	Kind         BlockerKind
	Subject      string
	SourceSHA256 string
	cause        error
}

func (blocker *Blocker) Error() string { return blocker.cause.Error() }
func (blocker *Blocker) Unwrap() error { return blocker.cause }

type partitionUnavailable struct{ limit int }

func (err *partitionUnavailable) Error() string {
	return fmt.Sprintf("declarations cannot preserve source order within %d lines", err.limit)
}

func planBlocker(kind BlockerKind, plan Plan, cause error) error {
	subject := ""
	if len(plan.Parts) != 0 {
		subject = plan.Parts[0].Subject
	}
	return &Blocker{Kind: kind, Subject: subject, SourceSHA256: plan.SourceSHA256, cause: cause}
}
