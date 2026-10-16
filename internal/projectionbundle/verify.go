package projectionbundle

import (
	"fmt"
	"io/fs"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/detection/protectedregions"
)

// Verify reads both the compiled projection and the actual repository members.
func Verify(source, raw []byte, compiled, repository fs.FS) error {
	manifest, err := Decode(raw)
	if err != nil {
		return err
	}
	if compiled == nil || repository == nil || manifest.SourceDigest != Digest(source) {
		return fmt.Errorf("projection bundle is not bound to its source and filesystems")
	}
	for _, view := range []fs.FS{compiled, repository} {
		if err := verifyPopulation(view, manifest.Members); err != nil {
			return err
		}
	}
	for _, member := range manifest.Members {
		for _, view := range []fs.FS{compiled, repository} {
			content, err := fs.ReadFile(view, member.Path)
			if err != nil {
				return fmt.Errorf("read projection member %s: %w", member.Path, err)
			}
			if Digest(content) != member.Digest {
				return fmt.Errorf("projection member %s differs from its manifest", member.Path)
			}
			if err := verifyRegions(content, member.RegionIDs); err != nil {
				return fmt.Errorf("projection member %s ownership: %w", member.Path, err)
			}
		}
	}
	return nil
}

func verifyPopulation(view fs.FS, members []Member) error {
	prefix := strings.TrimSuffix(members[0].Path, ".generated.go")
	paths, err := fs.Glob(view, prefix+"*.generated.go")
	if err != nil || len(paths) != len(members) {
		return fmt.Errorf("projection file population differs")
	}
	expected := map[string]bool{}
	for _, member := range members {
		expected[member.Path] = true
	}
	for _, path := range paths {
		if !expected[path] {
			return fmt.Errorf("projection contains undeclared file %s", path)
		}
	}
	return nil
}

func verifyRegions(content []byte, identities []string) error {
	report := protectedregions.Validate(content)
	if err := report.Err(); err != nil {
		return err
	}
	var observed []string
	for _, region := range report.Regions {
		if region.Kind == protectedregions.Generated {
			observed = append(observed, region.ID)
		}
	}
	if len(observed) != len(identities) {
		return fmt.Errorf("projection region population differs")
	}
	for i, identity := range identities {
		if observed[i] != identity {
			return fmt.Errorf("projection region order or identity differs")
		}
	}
	return nil
}
