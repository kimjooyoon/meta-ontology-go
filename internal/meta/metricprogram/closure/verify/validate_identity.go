package verify

import (
	"fmt"
	"strings"
)

func validateIdentity(in Input) error {
	if strings.Count(in.Repository, "/") != 1 || !shaPattern.MatchString(in.SubjectSHA) {
		return fmt.Errorf("invalid repository or subject identity")
	}
	if in.RunID < 1 || in.RunAttempt < 1 || in.Artifact.ID < 1 {
		return fmt.Errorf("run and artifact identities must be positive")
	}
	if in.Artifact.Name != "metric-meta-program-"+in.SubjectSHA {
		return fmt.Errorf("artifact name is not bound to subject")
	}
	wantURL := fmt.Sprintf("https://github.com/%s/actions/runs/%d/artifacts/%d",
		in.Repository, in.RunID, in.Artifact.ID)
	if in.Artifact.URL != wantURL {
		return fmt.Errorf("artifact url is not bound to run")
	}
	return nil
}
