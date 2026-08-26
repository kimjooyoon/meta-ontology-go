package valuecatalog

import (
	"fmt"
	"strings"
)

const (
	extensionDeclaration = "activity IncrementTwo(Integer) -> Integer"
	extensionProgramLine = extensionDeclaration + " computes \"int.add:2\""
)

func catalogSources(source []byte) ([]byte, []byte, error) {
	lines := strings.Split(string(source), "\n")
	found := -1
	for index, line := range lines {
		if line == extensionDeclaration || line == extensionProgramLine {
			if found >= 0 {
				return nil, nil, fmt.Errorf("duplicate extension declaration")
			}
			found = index
			continue
		}
		if strings.HasPrefix(line, extensionDeclaration+" computes ") {
			return nil, nil, fmt.Errorf("non-canonical extension program")
		}
	}
	if found < 0 {
		return nil, nil, fmt.Errorf("extension declaration is missing")
	}
	baseline := append([]string(nil), lines...)
	baseline[found] = extensionDeclaration
	candidate := append([]string(nil), lines...)
	candidate[found] = extensionProgramLine
	return []byte(strings.Join(baseline, "\n")), []byte(strings.Join(candidate, "\n")), nil
}

func unknownExtensionSource(candidate []byte) []byte {
	return []byte(strings.Replace(string(candidate), "int.add:2", "int.magic:2", 1))
}
