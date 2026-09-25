package main

import (
	"time"
)

func readPreviousWithDeadline(reader SourceReader, filename string, timeout time.Duration) ([]byte, error) {
	if timeout <= 0 {
		return nil, errCommandDeadline
	}
	result := make(chan readResult, 1)
	go func() {
		source, err := reader.ReadFile(filename)
		if err == nil && int64(len(source)) > maxOutputBytes {
			err = outputLimitError(maxOutputBytes)
		}
		result <- readResult{source: source, err: err}
	}()
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case read := <-result:
		return read.source, read.err
	case <-timer.C:
		return nil, errCommandDeadline
	}
}
