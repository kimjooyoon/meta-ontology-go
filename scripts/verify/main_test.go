package main

import (
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/verify"
)

func TestRevisionAvailabilityFailsClosed(t *testing.T) {
	if revisionAvailable(".", "") {
		t.Fatal("empty revision was accepted")
	}
	if revisionAvailable(".", strings.Repeat("0", 40)) {
		t.Fatal("zero revision was accepted")
	}
	if revisionAvailable(".", "deadbeefdeadbeefdeadbeefdeadbeefdeadbeef") {
		t.Fatal("missing revision was accepted")
	}
}

func TestCIHEAD001EmptyScopeRevisionsFailClosed(t *testing.T) {
	if err := validateScopeRevisions("", "", strings.Repeat("a", 40)); err == nil {
		t.Fatal("empty PR revisions were accepted")
	}
}

func TestCIHEAD002InvalidScopeRevisionFailsClosed(t *testing.T) {
	if err := validateScopeRevisions("not-a-revision", strings.Repeat("a", 40), ""); err == nil {
		t.Fatal("invalid scope revision was accepted")
	}
}

func TestCIHEAD003IdenticalScopeRevisionsFailClosed(t *testing.T) {
	revision := strings.Repeat("a", 40)
	if err := validateScopeRevisions(revision, revision, ""); err == nil {
		t.Fatal("identical scope revisions were accepted")
	}
}

func TestCIHEAD004SyntheticMergeRefMismatchFailsClosed(t *testing.T) {
	actual, err := runGit(".", "rev-parse", "HEAD")
	if err != nil {
		t.Fatal(err)
	}
	actual = strings.TrimSpace(actual)
	if err := verifyPRCheckoutIdentity(".", "HEAD^", "HEAD^", actual); err == nil {
		t.Fatal("synthetic merge ref was accepted as PR scope head")
	}
}

func TestCIHEAD005CheckoutOrBaseMismatchFailsClosed(t *testing.T) {
	actual, err := runGit(".", "rev-parse", "HEAD")
	if err != nil {
		t.Fatal(err)
	}
	actual = strings.TrimSpace(actual)
	if err := verifyPRCheckoutIdentity(".", "missing-base", actual, actual); err == nil {
		t.Fatal("missing PR base was accepted")
	}
	if err := verifyPRCheckoutIdentity(".", "HEAD^", actual, strings.Repeat("b", 40)); err == nil {
		t.Fatal("mismatched checked-out PR head was accepted")
	}
}

func TestCISCOPE001UnknownAgentPushFailsClosed(t *testing.T) {
	for _, branch := range []string{"agent/syntax-current-ddaf", "agent/semantic-current-ddaf", "agent/bidir-current-ddaf", "agent/lsp-current-ddaf", "agent/"} {
		if err := checkAgentPushBranch(branch); err == nil {
			t.Errorf("unknown agent push %s was accepted", branch)
		}
	}
	if err := checkAgentPushBranch("agent/ci-generator-current7"); err != nil {
		t.Fatal(err)
	}
	if err := checkAgentPushBranch("integration"); err != nil {
		t.Fatal(err)
	}
}

func TestCISCOPE002MalformedAgentBranchFailsClosed(t *testing.T) {
	for _, branch := range []string{"agent/*", "agent/../ci-workflow", "agent//ci-workflow"} {
		if err := checkAgentPushBranch(branch); err == nil {
			t.Errorf("malformed agent branch %s was accepted", branch)
		}
	}
	if err := verify.CheckPathScopeForBranch(nil, "agent/*"); err == nil {
		t.Fatal("wildcard branch key was accepted")
	}
}

func TestCapModesRejectAmbiguousInvocation(t *testing.T) {
	if err := validateCapMode(true, true); err == nil {
		t.Fatal("caps-only and skip-caps were accepted together")
	}
	if err := validateCapMode(true, false); err != nil {
		t.Fatal(err)
	}
	if err := validateCapMode(false, true); err != nil {
		t.Fatal(err)
	}
}
