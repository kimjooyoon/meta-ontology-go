package selfimprovementattestation

import "strings"

func subjectMatches(subjects []Subject, name, digest string) bool {
	if len(subjects) != 1 || !strings.HasPrefix(digest, "sha256:") {
		return false
	}
	return subjects[0].Name == name && subjects[0].Digest["sha256"] == strings.TrimPrefix(digest, "sha256:")
}
