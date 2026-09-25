package integrationprogress

func basisPoints(numerator, denominator int) int {
	return basisPoints64(int64(numerator), int64(denominator))
}

func basisPoints64(numerator, denominator int64) int {
	if denominator == 0 {
		return 0
	}
	return int(numerator * 10000 / denominator)
}
