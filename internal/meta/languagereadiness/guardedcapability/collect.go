package guardedcapability

import (
	"errors"
	"os/exec"
	"strings"
)

func Collect(root, currentHeadSHA string) (Source, error) {
	report, err := foundationReport()
	if err != nil {
		return Source{}, err
	}
	source := Source{
		CurrentHeadSHA: currentHeadSHA, WorkflowRunID: FoundationWorkflowRunID,
		ArtifactID: FoundationArtifactID, ArtifactDigest: FoundationArtifactDigest,
		ReportFileSHA: digestBytes(foundationRaw), FoundationReport: report,
	}
	source.AncestryObserved, source.FoundationAncestor = observeAncestor(root, currentHeadSHA)
	source.FoundationGuardTree, source.GuardTreesObserved =
		gitValue(root, "rev-parse", FoundationSubjectSHA+":"+guardPath)
	currentGuardTree, currentGuardObserved :=
		gitValue(root, "rev-parse", currentHeadSHA+":"+guardPath)
	source.CurrentGuardTree = currentGuardTree
	source.GuardTreesObserved = source.GuardTreesObserved && currentGuardObserved
	source.FoundationWitnessTree, source.WitnessTreesObserved =
		gitValue(root, "rev-parse", FoundationSubjectSHA+":"+witnessPath)
	currentWitnessTree, currentWitnessObserved :=
		gitValue(root, "rev-parse", currentHeadSHA+":"+witnessPath)
	source.CurrentWitnessTree = currentWitnessTree
	source.WitnessTreesObserved = source.WitnessTreesObserved && currentWitnessObserved
	return source, nil
}

func observeAncestor(root, current string) (bool, bool) {
	err := exec.Command("git", "-C", root, "merge-base", "--is-ancestor",
		FoundationSubjectSHA, current).Run()
	if err == nil {
		return true, true
	}
	exitError := &exec.ExitError{}
	if errors.As(err, &exitError) && exitError.ExitCode() == 1 {
		return true, false
	}
	return false, false
}

func gitValue(root string, arguments ...string) (string, bool) {
	arguments = append([]string{"-C", root}, arguments...)
	raw, err := exec.Command("git", arguments...).Output()
	return strings.TrimSpace(string(raw)), err == nil
}
