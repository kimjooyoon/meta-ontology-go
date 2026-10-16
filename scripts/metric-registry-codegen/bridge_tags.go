package main

import (
	"fmt"
	"strconv"
	"strings"
)

// Read the complete tag declaration so an earlier valid ID cannot hide a second
// identity or malformed metadata later in the same field's source.
func nativeTagID(raw string) (string, bool, error) {
	seen := map[string]bool{}
	var id string
	for raw = strings.TrimLeft(raw, " "); raw != ""; raw = strings.TrimLeft(raw, " ") {
		colon := strings.IndexByte(raw, ':')
		if colon <= 0 {
			return "", false, fmt.Errorf("invalid tag key/value")
		}
		key := raw[:colon]
		if strings.IndexFunc(key, func(r rune) bool { return r <= ' ' || r == '"' || r == 0x7f }) >= 0 {
			return "", false, fmt.Errorf("invalid tag key/value")
		}
		if seen[key] {
			return "", false, fmt.Errorf("repeated key %q", key)
		}
		raw = raw[colon+1:]
		if raw == "" || raw[0] != '"' {
			return "", false, fmt.Errorf("invalid quoted tag value")
		}
		quoted, err := strconv.QuotedPrefix(raw)
		if err != nil {
			return "", false, fmt.Errorf("invalid quoted tag value")
		}
		value, err := strconv.Unquote(quoted)
		if err != nil {
			return "", false, fmt.Errorf("invalid quoted tag value")
		}
		raw = raw[len(quoted):]
		if raw != "" && raw[0] != ' ' {
			return "", false, fmt.Errorf("tag keys require a space")
		}
		seen[key] = true
		if key == "gooo" {
			id = value
		}
	}
	return id, seen["gooo"], nil
}
