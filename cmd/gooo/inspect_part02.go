package main

import (
	"encoding/json"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
	"io"
	"time"
)

func lowerInspectIRWith(file *syntax.File, timeout time.Duration, lower func(*syntax.File) (semantic.IR, error)) (semantic.IR, error) {

	if timeout <= 0 {
		return semantic.IR{}, errCommandDeadline
	}
	result := make(chan inspectLowerResult, 1)
	go func() {
		ir, err := lower(file)
		if err == nil {
			err = ir.Validate()
		}
		if err == nil {
			err = rejectCLIEntityFieldsIR(ir)
		}
		result <- inspectLowerResult{ir: ir, err: err}
	}()
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case lowered := <-result:
		return lowered.ir, lowered.err
	case <-timer.C:
		return semantic.IR{}, errCommandDeadline
	}
}
func marshalGraphDump(source []byte, ir semantic.IR) ([]byte, error) {
	dump := newGraphDump(source, ir)
	payload, err := json.Marshal(dump)
	if err != nil {
		return nil, err
	}
	if len(payload)+1 > maxGraphDumpBytes {
		return nil, errGraphDumpLimit
	}
	return append(payload, '\n'), nil
}

type writeDeadlineSetter interface {
	SetWriteDeadline(time.Time) error
}

func writeInspectOutput(output io.Writer, payload []byte, deadline time.Time) error {
	if deadlineWriter, supportsDeadline := output.(writeDeadlineSetter); supportsDeadline {
		if err := deadlineWriter.SetWriteDeadline(deadline); err == nil {
			defer deadlineWriter.SetWriteDeadline(time.Time{})
		}
	}
	for len(payload) > 0 {
		if !deadline.IsZero() && !time.Now().Before(deadline) {
			return errCommandDeadline
		}
		written, err := output.Write(payload)
		if written > 0 {
			payload = payload[written:]
		}
		if err != nil {
			return err
		}
		if written == 0 {
			return io.ErrShortWrite
		}
	}
	return nil
}
