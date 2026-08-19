package cache

import (
	"context"
	"fmt"
	"sort"
	"testing"
	"time"
)

func measureIncrementalCache(t *testing.T, base, current []incrementalPart, expected map[int]bool, size int, mutation string) incrementalMeasurement {
	t.Helper()
	cache, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	for _, part := range base {
		if err := cache.PutPartial(part.spec, []byte(fmt.Sprintf("base:%d:%d", size, part.index))); err != nil {
			t.Fatal(err)
		}
	}
	measurement := incrementalMeasurement{started: time.Now()}
	for _, part := range current {
		part := part
		_, data, _, hit, err := cache.GetOrComputePartial(context.Background(), part.spec, func() ([]byte, error) {
			measurement.recomputations++
			return []byte(fmt.Sprintf("current:%d:%d:%s", size, part.index, mutation)), nil
		})
		if err != nil {
			t.Fatal(err)
		}
		wantHit := !expected[part.index]
		if hit != wantHit {
			t.Fatalf("%s part %d hit=%v, want %v", mutation, part.index, hit, wantHit)
		}
		if hit {
			measurement.hits++
			if string(data) != fmt.Sprintf("base:%d:%d", size, part.index) {
				t.Fatalf("%s part %d hit returned %q", mutation, part.index, data)
			}
		} else {
			measurement.misses++
			if string(data) != fmt.Sprintf("current:%d:%d:%s", size, part.index, mutation) {
				t.Fatalf("%s part %d miss returned %q", mutation, part.index, data)
			}
		}
	}
	if measurement.misses != len(expected) || measurement.recomputations != measurement.misses {
		t.Fatalf("%s measurements hits=%d misses=%d recomputations=%d expected misses=%d", mutation, measurement.hits, measurement.misses, measurement.recomputations, len(expected))
	}
	return measurement
}
func singleAffectedPart(size int) map[int]bool {
	return map[int]bool{(size / 2) * incrementalPartCount / size: true}
}
func uniqueSortedStrings(values []string) []string {
	sorted := append([]string(nil), values...)
	sort.Strings(sorted)
	unique := sorted[:0]
	for _, value := range sorted {
		if len(unique) == 0 || unique[len(unique)-1] != value {
			unique = append(unique, value)
		}
	}
	return unique
}
