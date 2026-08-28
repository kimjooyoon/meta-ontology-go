package main

func countLines(data []byte) int {
	count := 0
	for _, byteValue := range data {
		if byteValue == '\n' {
			count++
		}
	}
	if len(data) > 0 && data[len(data)-1] != '\n' {
		count++
	}
	return count
}
