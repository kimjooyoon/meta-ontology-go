package replay

import (
	"io/fs"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

// EntityFieldsProfileSupport resolves exact kernel IDs. It does not select a
// newer profile from the source shape or widen the ordinary parser's default.
func EntityFieldsProfileSupport(id string) (syntax.EntityFieldsSupport, bool) {
	return syntax.EntityFieldsSupportForProfileID(id)
}

func ExecuteWithEntityFieldsProfile(repository fs.FS, path, kind, diagnostic, profileID string) Result {
	support, known := EntityFieldsProfileSupport(profileID)
	if !known {
		return Result{ObservedDecision: DecisionClosed, EntityFieldsProfile: profileID,
			Diagnostics: []string{"profile.unknown"}}
	}
	result := executeWithSupport(repository, path, kind, diagnostic, support)
	result.EntityFieldsProfile, result.ProfileDigest = support.Profile.ID, "sha256:"+support.Profile.Digest
	return result
}
