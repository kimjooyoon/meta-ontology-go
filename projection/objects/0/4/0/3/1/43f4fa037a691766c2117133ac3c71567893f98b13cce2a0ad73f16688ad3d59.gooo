package semanticbinding

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strconv"
	"strings"
)

func writeField(builder *strings.Builder, value string) {
	builder.WriteString(strconv.Itoa(len(value)))
	builder.WriteByte(':')
	builder.WriteString(value)
	builder.WriteByte('\n')
}
func digestString(value string) string {
	digest := sha256.Sum256([]byte(value))
	return hex.EncodeToString(digest[:])
}
func sortRecords(bindings []Binding, obligations []Obligation) {
	sort.SliceStable(bindings, func(left, right int) bool {
		return compareBinding(bindings[left], bindings[right]) < 0
	})
	sort.SliceStable(obligations, func(left, right int) bool {
		return compareObligation(obligations[left], obligations[right]) < 0
	})
}
func compareBinding(left, right Binding) int {
	if value := strings.Compare(left.ID, right.ID); value != 0 {
		return value
	}
	if value := strings.Compare(string(left.Role), string(right.Role)); value != 0 {
		return value
	}
	return 0
}
func compareObligation(left, right Obligation) int {
	if value := strings.Compare(left.ID, right.ID); value != 0 {
		return value
	}
	if value := strings.Compare(left.Subject, right.Subject); value != 0 {
		return value
	}
	return strings.Compare(left.Pressure, right.Pressure)
}
