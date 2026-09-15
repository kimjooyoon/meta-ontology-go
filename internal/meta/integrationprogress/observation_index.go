package integrationprogress

func indexObservation(value Observation) (map[int]PullObservation, int) {
	expected := make(map[int]bool, len(PullNumbers()))
	for _, number := range PullNumbers() {
		expected[number] = true
	}
	index := make(map[int]PullObservation, len(value.PullRequests))
	conflicts := 0
	for _, pull := range value.PullRequests {
		if !expected[pull.Number] {
			conflicts++
			continue
		}
		if _, exists := index[pull.Number]; exists {
			conflicts++
			continue
		}
		index[pull.Number] = pull
	}
	if len(index) != len(expected) {
		conflicts += len(expected) - len(index)
	}
	return index, conflicts
}
