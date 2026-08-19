package bidir

import (
	"strings"
)

func relationKey(predicate Predicate, source, target ID) string {
	return string(predicate) + "\x00" + string(source) + "\x00" + string(target)
}
func cloneStringMap(values map[string]string) map[string]string {
	if len(values) == 0 {
		return nil
	}
	clone := make(map[string]string, len(values))
	for key, value := range values {
		clone[key] = value
	}
	return clone
}
func stringMapEqual(left, right map[string]string) bool {
	if len(left) != len(right) {
		return false
	}
	for key, value := range left {
		if right[key] != value {
			return false
		}
	}
	return true
}
func defaultName(id ID) string {
	value := string(id)
	if slash := strings.LastIndex(value, "/"); slash >= 0 && slash+1 < len(value) {
		return value[slash+1:]
	}
	return value
}
