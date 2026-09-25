package impactcoverage

import (
	"fmt"
	"sort"
)

func addLen(total uint64, pair sourcePair) (uint64, error) {
	count := uint64(0)
	if pair.before != nil {
		count = uint64(len(pair.before.Bindings))
	}
	if pair.after != nil {
		var err error
		count, err = checkedAdd(count, uint64(len(pair.after.Bindings)))
		if err != nil {
			return 0, err
		}
	}
	return checkedAdd(total, count)
}
func changed(pair sourcePair, beforeIDs, afterIDs []string) bool {
	if pair.before == nil || pair.after == nil {
		return true
	}
	return pair.before.BlobDigest != pair.after.BlobDigest ||
		!equalStrings(beforeIDs, afterIDs)
}
func unionIDs(left, right []string) []string {
	values := append(append([]string{}, left...), right...)
	return sortedUnique(values)
}
func sortedUnique(values []string) []string {
	result := append([]string{}, values...)
	sort.Strings(result)
	if len(result) < 2 {
		return result
	}
	write := 1
	for _, value := range result[1:] {
		if value != result[write-1] {
			result[write] = value
			write++
		}
	}
	return result[:write]
}
func equalStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}
func checkedAdd(left, right uint64) (uint64, error) {
	if ^uint64(0)-left < right {
		return 0, fmt.Errorf("deterministic work unit overflow")
	}
	return left + right, nil
}
