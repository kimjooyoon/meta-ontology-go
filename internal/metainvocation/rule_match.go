package metainvocation

import (
	"path"
	"strings"
)

func ruleMatches(operation, file string) bool {
	switch operation {
	case operationDocsRule:
		return strings.HasPrefix(file, "docs/")
	case operationGoRule:
		return path.Ext(file) == ".go"
	case operationYAMLRule:
		extension := path.Ext(file)
		return extension == ".yaml" || extension == ".yml"
	default:
		return false
	}
}
