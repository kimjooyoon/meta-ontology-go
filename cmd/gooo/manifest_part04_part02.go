package main

import (
	"fmt"
	"strconv"
	"strings"
)

func markerID(line, prefix string) (string, error) {
	rest := strings.TrimSpace(strings.TrimPrefix(line, prefix))
	if !strings.HasPrefix(rest, "id=") {
		return "", fmt.Errorf("marker %q has no id", prefix)
	}
	value := strings.TrimSpace(strings.TrimPrefix(rest, "id="))
	if len(value) < 2 || value[0] != '"' {
		return "", fmt.Errorf("marker %q has invalid id", prefix)
	}
	end := 1
	for end < len(value) {
		if value[end] == '"' && value[end-1] != '\\' {
			break
		}
		end++
	}
	if end >= len(value) {
		return "", fmt.Errorf("marker %q has unterminated id", prefix)
	}
	id, err := strconv.Unquote(value[:end+1])
	if err != nil || id == "" {
		return "", fmt.Errorf("marker %q has invalid id", prefix)
	}
	return id, nil
}
