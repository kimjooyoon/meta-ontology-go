package feedbackstate

import "errors"

var (
	errNegativeRepositoryWrites = errors.New("FAIL_CLOSED: NEGATIVE_REPOSITORY_WRITES")
	errRepositoryWritesOverflow = errors.New("FAIL_CLOSED: REPOSITORY_WRITES_OVERFLOW")
)

func AggregateRepositoryWrites(counts ...int) (int, error) {
	for _, count := range counts {
		if count < 0 {
			return 0, errNegativeRepositoryWrites
		}
	}

	maxInt := int(^uint(0) >> 1)
	total := 0
	for _, count := range counts {
		if count > maxInt-total {
			return 0, errRepositoryWritesOverflow
		}
		total += count
	}
	return total, nil
}
