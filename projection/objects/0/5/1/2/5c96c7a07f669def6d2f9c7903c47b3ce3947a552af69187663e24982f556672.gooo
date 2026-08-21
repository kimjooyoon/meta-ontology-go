package toolchain

import (
	"bufio"
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
	file, err := os.Open(filepath.Join(root, "go.mod"))
	if err != nil {
		return Requirement{}, err
	}
	defer file.Close()
	requirement := Requirement{}
	scanner := bufio.NewScanner(file)
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
	if requirement.Language == "go" {
		return Requirement{}, fmt.Errorf("go.mod has no go directive")
	}
	if requirement.Toolchain == "" {
		requirement.Toolchain = requirement.Language
	}
	return requirement, nil
}
