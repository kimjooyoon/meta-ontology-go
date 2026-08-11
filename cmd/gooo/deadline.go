package main

import (
	"errors"
	"time"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const commandDeadline = 5 * time.Second

var errCommandDeadline = errors.New("command deadline exceeded")

type readResult struct {
	source []byte
	err    error
}

type parseResult struct {
	file        *syntax.File
	diagnostics syntax.Diagnostics
}

func parseWithDeadline(parser SourceParser, filename, source string, timeout time.Duration) (*syntax.File, syntax.Diagnostics, error) {
	if timeout <= 0 {
		return nil, nil, errCommandDeadline
	}
	result := make(chan parseResult, 1)
	go func() {
		file, diagnostics := parser.ParseFile(filename, source)
		result <- parseResult{file: file, diagnostics: diagnostics}
	}()
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case parsed := <-result:
		return parsed.file, parsed.diagnostics, nil
	case <-timer.C:
		return nil, nil, errCommandDeadline
	}
}

func remainingDeadline(deadline time.Time) time.Duration {
	return time.Until(deadline)
}

func readSourceWithDeadline(reader SourceReader, filename string, timeout time.Duration) ([]byte, error) {
	if timeout <= 0 {
		return nil, errCommandDeadline
	}
	result := make(chan readResult, 1)
	go func() {
		source, err := readSource(reader, filename)
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
