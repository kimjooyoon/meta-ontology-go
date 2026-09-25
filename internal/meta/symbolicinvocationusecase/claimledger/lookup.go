package claimledger

import "strings"

func lookupAny(root map[string]any, paths []string) (string, any, bool) {
	for _, path := range paths {
		value, found := lookup(root, path)
		if found {
			return path, value, true
		}
	}
	return "", nil, false
}

func lookup(root map[string]any, path string) (any, bool) {
	var current any = root
	for part := range strings.SplitSeq(path, ".") {
		object, ok := current.(map[string]any)
		if !ok {
			return nil, false
		}
		current, ok = object[part]
		if !ok {
			return nil, false
		}
	}
	return current, true
}
