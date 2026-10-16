package unresolvedpolicy

import (
	"fmt"
	"io/fs"
	"path"

	"github.com/kimjooyoon/meta-ontology-go/internal/projectionbundle"
)

// VerifySource binds the diagnostic policy to the actual repository snapshot.
func VerifySource(repository fs.FS) error {
	if repository == nil {
		return fmt.Errorf("operation origin policy repository is missing")
	}
	if err := VerifyProjection(); err != nil {
		return err
	}
	origin := Origin()
	for _, member := range []struct{ path, digest string }{
		{origin.SourcePath, origin.SourceDigest}, {origin.ProjectionPath, origin.ProjectionDigest},
	} {
		raw, err := fs.ReadFile(repository, member.path)
		if err != nil {
			return fmt.Errorf("read operation origin policy %s: %w", member.path, err)
		}
		if projectionbundle.Digest(raw) != member.digest {
			return fmt.Errorf("operation origin policy member %s differs from compiled origin", member.path)
		}
	}
	actual, err := fs.Sub(repository, path.Dir(origin.SourcePath))
	if err != nil {
		return err
	}
	return projectionbundle.Verify(source, manifest, members, actual)
}
