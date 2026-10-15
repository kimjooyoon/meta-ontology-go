package syntaxregistration

import (
	"io/fs"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesyntax"
)

func InspectInventory(repository fs.FS) (languagesyntax.SourceInventory, error) {
	raw, err := fs.ReadFile(repository, corpusPath)
	if err != nil {
		return languagesyntax.SourceInventory{}, err
	}
	return languagesyntax.InspectSourceInventory(repository, raw)
}
