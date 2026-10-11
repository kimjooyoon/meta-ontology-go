// Package referencego supplies the ordinary Go counterpart of the batch pilot.
package referencego

// Batch16 returns the number of pending jobs to process, capped at sixteen.
func Batch16(pending int64) int64 { return batch(pending, 16) }

// Batch32 changes that rule to a cap of thirty-two.
func Batch32(pending int64) int64 { return batch(pending, 32) }

func batch(pending, limit int64) int64 {
	if pending < 0 {
		return 0
	}
	if pending < limit {
		return pending
	}
	return limit
}
