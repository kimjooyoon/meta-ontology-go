package languageartifactoracle

import (
	"fmt"
	"strconv"
	"strings"
)

func parseHeader(line, keyword string) (string, error) {
	fields := strings.Fields(line)
	if len(fields) != 2 || fields[0] != keyword || fields[1] == "" {
		return "", fmt.Errorf("unsupported %s declaration", keyword)
	}
	return fields[1], nil
}

func parseEntity(line string) (artifactBinding, error) {
	fields := strings.Fields(line)
	if len(fields) != 4 || fields[0] != "entity" || fields[2] != "id" {
		return artifactBinding{}, fmt.Errorf("unsupported entity declaration")
	}
	id, err := strconv.Unquote(fields[3])
	if err != nil || fields[1] == "" || id == "" {
		return artifactBinding{}, fmt.Errorf("invalid entity declaration")
	}
	return artifactBinding{Name: fields[1], ID: id}, nil
}
