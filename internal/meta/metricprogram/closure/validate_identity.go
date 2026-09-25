package closure

import (
	"fmt"
	"strings"
)

func validateIdentity(in Input) error {
	if strings.Count(in.Repository, "/") != 1 || strings.HasPrefix(in.Repository, "/") ||
		strings.HasSuffix(in.Repository, "/") {
		return fmt.Errorf("invalid repository %q", in.Repository)
	}
	if !shaPattern.MatchString(in.SubjectSHA) {
		return fmt.Errorf("invalid subject sha %q", in.SubjectSHA)
	}
	if in.RunID < 1 || in.RunAttempt < 1 || in.Artifact.ID < 1 {
		return fmt.Errorf("run and artifact identities must be positive")
	}
	wantName := "metric-meta-program-" + in.SubjectSHA
	if in.Artifact.Name != wantName {
		return fmt.Errorf("artifact name %q does not match %q", in.Artifact.Name, wantName)
	}
	wantURL := fmt.Sprintf("https://github.com/%s/actions/runs/%d/artifacts/%d",
		in.Repository, in.RunID, in.Artifact.ID)
	if in.Artifact.URL != wantURL {
		return fmt.Errorf("artifact url %q does not match %q", in.Artifact.URL, wantURL)
	}
	return nil
}
