package toolchainrelease

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

type branchCommitDecision struct {
	PlanSHA256 string `json:"plan_sha256"`
	BranchName string `json:"branch_name"`
	Status     string `json:"status"`
}

type DirectModuleBranchCommit struct {
	Schema      string                `json:"schema"`
	Branch      string                `json:"branch"`
	BaseCommit  string                `json:"base_commit"`
	BaseTree    string                `json:"base_tree"`
	Commit      string                `json:"commit"`
	Tree        string                `json:"tree"`
	Decision    branchCommitDecision  `json:"gooo_branch_decision"`
	Candidate   DirectModuleCandidate `json:"candidate"`
	Application string                `json:"application_status"`
	BuildStatus string                `json:"build_status"`
	Provenance  map[string]any        `json:"provenance"`
}

// CommitDirectModuleCandidate publishes all verified outputs as one local Git
// commit and atomically creates a new agent branch. The checkout is untouched.
func CommitDirectModuleCandidate(root, branch, module, proposed string) (DirectModuleBranchCommit, error) {
	if !strings.HasPrefix(branch, "agent/") || branch == "agent/" {
		return DirectModuleBranchCommit{}, fmt.Errorf("branch must be a new agent/* branch")
	}
	root, err := filepath.Abs(root)
	if err != nil {
		return DirectModuleBranchCommit{}, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	if _, err := moduleGit(ctx, root, nil, "check-ref-format", "--branch", branch); err != nil {
		return DirectModuleBranchCommit{}, fmt.Errorf("invalid branch name: %w", err)
	}
	if err := cleanModuleCheckout(ctx, root); err != nil {
		return DirectModuleBranchCommit{}, err
	}
	head, err := moduleGit(ctx, root, nil, "rev-parse", "HEAD")
	if err != nil {
		return DirectModuleBranchCommit{}, err
	}
	base := strings.TrimSpace(string(head))
	baseTreeRaw, err := moduleGit(ctx, root, nil, "rev-parse", "HEAD^{tree}")
	if err != nil {
		return DirectModuleBranchCommit{}, err
	}
	baseTree := strings.TrimSpace(string(baseTreeRaw))
	temporary, err := os.MkdirTemp("", "gooo-direct-module-branch-*")
	if err != nil {
		return DirectModuleBranchCommit{}, err
	}
	defer os.RemoveAll(temporary)
	candidatePath := filepath.Join(temporary, "candidate")
	candidate, err := ProjectDirectModuleCandidate(root, candidatePath, module, proposed)
	if err != nil {
		return DirectModuleBranchCommit{}, err
	}
	fresh, err := ObserveDirectModuleImpact(root, module, proposed)
	if err != nil || fresh.PlanSHA256 != candidate.PlanSHA256 ||
		fresh.Decision.Status != "READY_FOR_DETACHED_PROJECTION" || fresh.ArchiveHead != base {
		return DirectModuleBranchCommit{}, fmt.Errorf("candidate source or base changed before branch proposal")
	}
	if err := cleanModuleCheckout(ctx, root); err != nil {
		return DirectModuleBranchCommit{}, err
	}
	available, err := moduleBranchAvailable(ctx, root, branch)
	if err != nil {
		return DirectModuleBranchCommit{}, err
	}
	decision, err := assessModuleBranch(ctx, root, branch, fresh, candidate, available)
	if err != nil || decision.Status != "READY_FOR_BRANCH_COMMIT" {
		return DirectModuleBranchCommit{}, fmt.Errorf("Gooo did not authorize branch commit: %s: %v", decision.Status, err)
	}
	indexFile := filepath.Join(temporary, "candidate.index")
	env := moduleGitEnvironment("GIT_INDEX_FILE=" + indexFile)
	if _, err := moduleGit(ctx, root, env, "read-tree", base); err != nil {
		return DirectModuleBranchCommit{}, err
	}
	for _, file := range candidate.Files {
		if err := addCandidateBlob(ctx, root, candidatePath, env, base, file); err != nil {
			return DirectModuleBranchCommit{}, err
		}
	}
	treeRaw, err := moduleGit(ctx, root, env, "write-tree")
	if err != nil {
		return DirectModuleBranchCommit{}, err
	}
	tree := strings.TrimSpace(string(treeRaw))
	if tree == baseTree {
		return DirectModuleBranchCommit{}, fmt.Errorf("candidate did not change the Git tree")
	}
	timestampRaw, err := moduleGit(ctx, root, nil, "show", "-s", "--format=%ct", base)
	if err != nil {
		return DirectModuleBranchCommit{}, err
	}
	date := strings.TrimSpace(string(timestampRaw)) + " +0000"
	commitEnv := moduleGitEnvironment("GIT_AUTHOR_NAME=Gooo cochange",
		"GIT_AUTHOR_EMAIL=gooo-cochange@users.noreply.github.com",
		"GIT_COMMITTER_NAME=Gooo cochange",
		"GIT_COMMITTER_EMAIL=gooo-cochange@users.noreply.github.com",
		"GIT_AUTHOR_DATE="+date, "GIT_COMMITTER_DATE="+date)
	commitRaw, err := moduleGit(ctx, root, commitEnv, "commit-tree", tree, "-p", base,
		"-m", "gooo: direct module "+module+" "+proposed+" "+candidate.PlanSHA256)
	if err != nil {
		return DirectModuleBranchCommit{}, err
	}
	commit := strings.TrimSpace(string(commitRaw))
	if err := verifyCandidateCommit(ctx, root, commit, tree, candidate); err != nil {
		return DirectModuleBranchCommit{}, err
	}
	current, err := moduleGit(ctx, root, nil, "rev-parse", "HEAD")
	if err != nil || strings.TrimSpace(string(current)) != base {
		return DirectModuleBranchCommit{}, fmt.Errorf("checkout HEAD changed before branch creation")
	}
	if err := cleanModuleCheckout(ctx, root); err != nil {
		return DirectModuleBranchCommit{}, err
	}
	if _, err := moduleGit(ctx, root, nil, "update-ref", "refs/heads/"+branch, commit,
		"0000000000000000000000000000000000000000"); err != nil {
		return DirectModuleBranchCommit{}, fmt.Errorf("new branch could not be created atomically: %w", err)
	}
	graph := append([]any{}, candidate.Provenance["@graph"].([]any)...)
	activityID := "urn:gooo:direct-module-impact:branch-commit:" + candidate.PlanSHA256
	decisionID := "urn:gooo:direct-module-impact:branch-decision:" +
		candidate.PlanSHA256 + ":" + url.PathEscape(branch)
	baseID := "urn:gooo:git-commit:" + base
	used := []string{baseID, decisionID}
	for _, file := range candidate.Files {
		used = append(used, "urn:gooo:candidate-file:"+file.Path+":"+file.OutputSHA256)
	}
	graph = append(graph,
		map[string]any{"@id": decisionID, "@type": "prov:Entity",
			"gooo:status": decision.Status, "gooo:branch": branch,
			"prov:wasDerivedFrom": map[string]string{"@id": repositoryFileID(
				directModuleSourcePath, fresh.Inputs[directModuleSourcePath])}},
		map[string]any{"@id": baseID, "@type": "prov:Entity", "gooo:tree": baseTree},
		map[string]any{"@id": activityID, "@type": "prov:Activity", "prov:used": provReferences(used)},
		map[string]any{"@id": "urn:gooo:git-commit:" + commit, "@type": "prov:Entity",
			"gooo:tree": tree, "gooo:branch": branch,
			"prov:wasDerivedFrom": provReferences(used),
			"prov:wasGeneratedBy": map[string]string{"@id": activityID}})
	return DirectModuleBranchCommit{Schema: "gooo/direct-module-branch-commit/v1", Branch: branch,
		BaseCommit: base, BaseTree: baseTree, Commit: commit, Tree: tree, Decision: decision,
		Candidate: candidate, Application: "branch_committed", BuildStatus: "not_run",
		Provenance: map[string]any{"@context": candidate.Provenance["@context"], "@graph": graph}}, nil
}

func assessModuleBranch(ctx context.Context, root, branch string, impact DirectModuleImpact,
	candidate DirectModuleCandidate, available bool) (branchCommitDecision, error) {
	source, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(directModuleSourcePath)))
	if err != nil || payloadDigest(source) != impact.Inputs[directModuleSourcePath] {
		return branchCommitDecision{}, fmt.Errorf("Gooo branch decision source changed")
	}
	input, _ := json.Marshal([]any{map[string]any{
		"plan_sha256": candidate.PlanSHA256, "branch_name": branch,
		"projected_outputs": candidate.ProjectedOutputs,
		"verified_outputs":  candidate.VerifiedOutputs,
		"changed_outputs":   candidate.ChangedOutputs,
		"clean_checkout":    true, "base_bound": true,
		"branch_available": available, "agent_branch": strings.HasPrefix(branch, "agent/")}})
	value, err := bodycodegen.EvaluateFixedValue(ctx, directModuleSourcePath, source,
		"AssessBranchCommit", input)
	if err != nil || value.ActivityID != "urn:gooo:direct-module-impact:assess-branch-commit" ||
		value.SourceSHA256 != impact.Inputs[directModuleSourcePath] {
		return branchCommitDecision{}, fmt.Errorf("Gooo branch decision cannot be replayed: %v", err)
	}
	var decision branchCommitDecision
	if err := json.Unmarshal(value.Value, &decision); err != nil {
		return branchCommitDecision{}, err
	}
	if decision.PlanSHA256 != candidate.PlanSHA256 || decision.BranchName != branch {
		return branchCommitDecision{}, fmt.Errorf("Gooo branch decision changed proposal identity")
	}
	return decision, nil
}

func cleanModuleCheckout(ctx context.Context, root string) error {
	status, err := moduleGit(ctx, root, nil, "status", "--porcelain=v1", "-z", "--untracked-files=no")
	if err != nil || len(status) != 0 {
		return fmt.Errorf("tracked checkout must be clean before branch commit: %v", err)
	}
	return nil
}

func moduleBranchAvailable(ctx context.Context, root, branch string) (bool, error) {
	_, err := moduleGit(ctx, root, nil, "show-ref", "--verify", "--quiet", "refs/heads/"+branch)
	if err == nil {
		return false, nil
	}
	var exit *exec.ExitError
	if !errors.As(err, &exit) || exit.ExitCode() != 1 {
		return false, err
	}
	return true, nil
}

func addCandidateBlob(ctx context.Context, root, candidatePath string, env []string,
	base string, file DirectModuleProjectedFile) error {
	listed, err := moduleGit(ctx, root, nil, "ls-tree", "-z", base, "--", file.Path)
	parts := bytes.Split(bytes.TrimSuffix(listed, []byte{0}), []byte{0})
	if err != nil || len(parts) != 1 || !bytes.HasPrefix(parts[0], []byte("100644 blob ")) ||
		!bytes.HasSuffix(parts[0], []byte("\t"+file.Path)) {
		return fmt.Errorf("candidate output is not a regular tracked Git file: %s", file.Path)
	}
	full := filepath.Join(candidatePath, filepath.FromSlash(file.Path))
	data, err := os.ReadFile(full)
	if err != nil || payloadDigest(data) != file.OutputSHA256 {
		return fmt.Errorf("candidate output bytes changed before Git hashing: %s", file.Path)
	}
	old, err := moduleGit(ctx, root, nil, "show", base+":"+file.Path)
	if err != nil || payloadDigest(old) != file.InputSHA256 {
		return fmt.Errorf("candidate Git preimage differs: %s", file.Path)
	}
	object, err := moduleGit(ctx, root, nil, "hash-object", "-w", "--", full)
	if err != nil {
		return err
	}
	_, err = moduleGit(ctx, root, env, "update-index", "--add", "--cacheinfo", "100644",
		strings.TrimSpace(string(object)), file.Path)
	return err
}

func verifyCandidateCommit(ctx context.Context, root, commit, tree string, candidate DirectModuleCandidate) error {
	writtenTree, err := moduleGit(ctx, root, nil, "rev-parse", commit+"^{tree}")
	if err != nil || strings.TrimSpace(string(writtenTree)) != tree {
		return fmt.Errorf("candidate Git commit has a different tree")
	}
	for _, file := range candidate.Files {
		data, err := moduleGit(ctx, root, nil, "show", commit+":"+file.Path)
		if err != nil || payloadDigest(data) != file.OutputSHA256 {
			return fmt.Errorf("Git commit output differs from verified candidate: %s", file.Path)
		}
	}
	changed, err := moduleGit(ctx, root, nil, "diff-tree", "--no-commit-id", "--name-only", "-r", "-z", commit)
	if err != nil {
		return err
	}
	want := map[string]bool{}
	for _, file := range candidate.Files {
		if file.InputSHA256 != file.OutputSHA256 {
			want[file.Path] = true
		}
	}
	for _, raw := range bytes.Split(bytes.TrimSuffix(changed, []byte{0}), []byte{0}) {
		name := string(raw)
		if name == "" || !want[name] {
			return fmt.Errorf("Git commit changed an undeclared path: %s", name)
		}
		delete(want, name)
	}
	if len(want) != 0 {
		return fmt.Errorf("Git commit omitted declared changed outputs")
	}
	return nil
}

func moduleGit(ctx context.Context, root string, env []string, args ...string) ([]byte, error) {
	command := exec.CommandContext(ctx, "git", args...)
	command.Dir = root
	if env == nil {
		env = moduleGitEnvironment()
	}
	command.Env = env
	output, err := command.Output()
	if err != nil {
		var exit *exec.ExitError
		if errors.As(err, &exit) {
			return nil, fmt.Errorf("Git %s failed: %w: %s", args[0], err, strings.TrimSpace(string(exit.Stderr)))
		}
		return nil, fmt.Errorf("Git %s failed: %w", args[0], err)
	}
	return output, nil
}

func moduleGitEnvironment(extra ...string) []string {
	env := make([]string, 0, len(os.Environ())+len(extra))
	for _, item := range os.Environ() {
		if strings.HasPrefix(item, "GIT_") {
			continue
		}
		env = append(env, item)
	}
	return append(env, extra...)
}
