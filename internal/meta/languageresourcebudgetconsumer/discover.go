package languageresourcebudgetconsumer

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func DiscoverEntry(directory string) (string, error) {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return "", fmt.Errorf("SOURCE_DIRECTORY_READ_FAILED")
	}
	activity := ""
	count := 0
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".gooo" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(directory, entry.Name()))
		if err != nil {
			return "", fmt.Errorf("SOURCE_READ_FAILED")
		}
		file, diagnostics := syntax.ParseFile(entry.Name(), string(data))
		if file == nil || diagnostics.HasErrors() {
			return "", fmt.Errorf("SOURCE_SYNTAX_INVALID")
		}
		for _, declaration := range file.Decls {
			if value, ok := declaration.(*syntax.ActivityDecl); ok {
				activity = value.Name
				count++
			}
		}
	}
	if count != 1 || activity == "" {
		return "", fmt.Errorf("SOURCE_ENTRY_CARDINALITY_INVALID")
	}
	return activity, nil
}
