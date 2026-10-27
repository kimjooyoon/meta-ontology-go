package languagesyntax

import "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesyntax/replay"

func validCaseSourceEvidence(item CaseResult, source Source) bool {
	names := caseSourcesFor(expectedRegistry(), item.Definition.ID)
	evidence := item.Evidence
	if len(names) == 0 {
		return evidence.SourceDigestKind == "" && len(evidence.SourceMembers) == 0
	}
	if evidence.ObservedDecision == replay.DecisionUnknown && len(evidence.SourceMembers) == 0 {
		return evidence.SourceDigest == "" && evidence.SourceLines == 0 &&
			(evidence.SourceDigestKind == "" || evidence.SourceDigestKind == "gooo/source-bundle/v1")
	}
	if evidence.SourceDigestKind != "gooo/source-bundle/v1" || len(evidence.SourceMembers) != len(names) {
		return false
	}
	lines := 0
	for i, member := range evidence.SourceMembers {
		if member.Path != names[i] || member.GoooLines < 0 || !validDigest(member.SourceDigest) {
			return false
		}
		if source.ObservationKnown && !observedCaseMember(source.GoooFiles, member) {
			return false
		}
		lines += member.GoooLines
	}
	return evidence.SourceLines == lines
}

func observedCaseMember(files []replay.FileObservation, member replay.FileObservation) bool {
	for _, file := range files {
		if file.Path == member.Path {
			return file == member
		}
	}
	return false
}
