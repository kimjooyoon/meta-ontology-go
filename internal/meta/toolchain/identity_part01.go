package toolchain

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Requirement is the toolchain foundation declared by the main module.
type Requirement struct {
	Language  string
	Toolchain string
}

// ReadRequirement reads the go and optional toolchain directives without
// invoking a Go command, keeping the declared foundation independent.
func ReadRequirement(root string) (Requirement, error) {
	raw, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		return Requirement{}, err
	}
	return ParseRequirement(raw)
}

// ParseRequirement interprets the same bytes that a caller can bind to a
// recorded go.mod digest, avoiding a second read between identity and parsing.
func ParseRequirement(raw []byte) (Requirement, error) {
	requirement := Requirement{}
	scanner := bufio.NewScanner(bytes.NewReader(raw))
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) != 2 {
			continue
		}
		switch fields[0] {
		case "go":
			requirement.Language = "go" + fields[1]
		case "toolchain":
			requirement.Toolchain = fields[1]
		}
	}
	if err := scanner.Err(); err != nil {
		return Requirement{}, err
	}
	if requirement.Language == "" {
		return Requirement{}, fmt.Errorf("go.mod has no go directive")
	}
	if requirement.Toolchain == "" {
		requirement.Toolchain = requirement.Language
	}
	return requirement, nil
}
