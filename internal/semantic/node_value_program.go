package semantic

import (
	"errors"
	"fmt"
	"strings"
)

func normalizeName(raw string) (string, error) {
	name := strings.Join(strings.Fields(raw), " ")
	if name == "" {
		return "", errors.New("name is empty")
	}
	return name, nil
}

func normalizeActivityValueProgram(kind Kind, raw string) (string, error) {
	if raw == "" {
		return "", nil
	}
	if raw != strings.TrimSpace(raw) {
		return "", fmt.Errorf("%w: value program has surrounding whitespace", ErrInvalidNode)
	}
	if kind != Activity {
		return "", fmt.Errorf("%w: value program requires an Activity node", ErrInvalidNode)
	}
	return raw, nil
}

func writeCanonicalNodeValueProgram(builder *strings.Builder, program string) {
	if program == "" {
		return
	}
	builder.WriteString("value-program\t")
	writeCanonicalField(builder, program)
}
