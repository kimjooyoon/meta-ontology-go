// Package inputbinding resolves authored names against positional input ports.
package inputbinding

import (
	"fmt"
	"go/token"
	"strconv"
)

func PositionName(arity, index int) string {
	if arity == 1 {
		return "input"
	}
	return "input" + strconv.Itoa(index)
}

// Resolve retains canonical port identities while checking source aliases.
// An alias cannot borrow another position's canonical name.
func Resolve(arity int, declared []string) ([]string, error) {
	if arity < 0 || len(declared) != 0 && len(declared) != arity {
		return nil, fmt.Errorf("input names must align with declared input positions")
	}
	if len(declared) > 0 {
		present := false
		for _, name := range declared {
			present = present || name != ""
		}
		if !present {
			return nil, fmt.Errorf("input names need at least one authored name")
		}
	}
	positions := make(map[string]int, arity)
	for i := range arity {
		positions[PositionName(arity, i)] = i
	}
	result := make([]string, arity)
	seen := map[string]bool{}
	for i := range arity {
		name := PositionName(arity, i)
		if len(declared) > 0 && declared[i] != "" {
			name = declared[i]
		}
		if !token.IsIdentifier(name) || name == "_" || name == "nil" || name == "true" || name == "false" {
			return nil, fmt.Errorf("input name %q cannot bind a value", name)
		}
		if position, reserved := positions[name]; reserved && position != i {
			return nil, fmt.Errorf("input name %q belongs to position %d", name, position)
		}
		if seen[name] {
			return nil, fmt.Errorf("input name %q is repeated", name)
		}
		seen[name], result[i] = true, name
	}
	return result, nil
}
