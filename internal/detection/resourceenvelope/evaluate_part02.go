package resourceenvelope

func maxPeakRSS(samples []Sample) uint64 {
	var value uint64
	for _, sample := range samples {
		if sample.PeakRSSBytes > value {
			value = sample.PeakRSSBytes
		}
	}
	return value
}
func maxReadBytes(samples []Sample) uint64 {
	var value uint64
	for _, sample := range samples {
		if sample.ReadBytes > value {
			value = sample.ReadBytes
		}
	}
	return value
}
func maxWriteBytes(samples []Sample) uint64 {
	var value uint64
	for _, sample := range samples {
		if sample.WriteBytes > value {
			value = sample.WriteBytes
		}
	}
	return value
}
