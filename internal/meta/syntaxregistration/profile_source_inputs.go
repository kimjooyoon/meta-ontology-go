package syntaxregistration

import (
	"bytes"
	"fmt"
	"io/fs"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesyntax"
	"github.com/kimjooyoon/meta-ontology-go/internal/sourcebundle"
)

// Pin every member even though migration writes only the case's primary source.
// Generate rereads the same inputs; replay uses these bytes rather than a second
// unbound read of the mutable input filesystem.
func readProfileFamilyInputs(repository fs.FS, inputs map[string][]byte, definition languagesyntax.CaseDefinition) error {
	var registry languagesyntax.Registry
	if err := decodeStrict(inputs[corpusPath], &registry); err != nil {
		return err
	}
	names, err := languagesyntax.CaseSourceMembers(registry, definition)
	if err != nil || len(names) == 0 {
		return err
	}
	sources, err := sourcebundle.Read(repository, names)
	if err != nil {
		return err
	}
	for _, source := range sources {
		raw := []byte(source.Content)
		if prior, found := inputs[source.Filename]; found && !bytes.Equal(prior, raw) {
			return fmt.Errorf("profile source family changed during input observation")
		}
		inputs[source.Filename] = raw
	}
	return nil
}
