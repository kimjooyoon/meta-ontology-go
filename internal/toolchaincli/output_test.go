package toolchaincli

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"
	"unicode/utf8"
)

func TestCaptureSeparatesEOFAndRetainedPrefix(t *testing.T) {
	cases := []struct{ name, input string }{
		{"exact", strings.Repeat("x", maxOutputBytes)},
		{"overflow-ascii", strings.Repeat("x", maxOutputBytes+1)},
		{"overflow-utf8", strings.Repeat("x", maxOutputBytes-1) + "가"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			stream, err := openInvocationStream()
			if err != nil {
				t.Fatal(err)
			}
			defer stream.close()
			stream.start()
			written := make(chan error, 1)
			go func() {
				_, err := stream.writer.Write([]byte(tc.input))
				_ = stream.writer.Close()
				written <- err
			}()
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			text, state, err := stream.finish(ctx)
			if writeErr := <-written; writeErr != nil {
				t.Fatal(writeErr)
			}
			if err != nil || !state.EOF || state.DeadlineCloseRequested ||
				state.LimitExceeded != (len(tc.input) > maxOutputBytes) ||
				state.RetainedBytes != maxOutputBytes ||
				text != tc.input[:maxOutputBytes] {
				t.Fatal("EOF and retained prefix were conflated", state, err)
			}
			if tc.name == "overflow-utf8" && utf8.ValidString(text) {
				t.Fatal("incomplete UTF-8 prefix was rewritten at the capture boundary")
			}
		})
	}
}

func TestEmptyCaptureWritePreservesLimitState(t *testing.T) {
	for _, extra := range []int{0, 1} {
		name := "exact"
		if extra != 0 {
			name = "overflow"
		}
		t.Run(name, func(t *testing.T) {
			var buffer cappedBuffer
			input := []byte(strings.Repeat("x", maxOutputBytes+extra))
			if written, err := buffer.Write(input); err != nil || written != len(input) {
				t.Fatal("initial capture did not consume its input", written, err)
			}
			if written, err := buffer.Write(nil); err != nil || written != 0 {
				t.Fatal("empty input did not remain empty", written, err)
			}
			if len(buffer.data) != maxOutputBytes || buffer.overflow != (extra != 0) ||
				buffer.String() != string(input[:maxOutputBytes]) {
				t.Fatal("empty input changed retained bytes or overflow", buffer.overflow)
			}
		})
	}
}

func TestCaptureCompletionAndCloseRequestRemainIndependent(t *testing.T) {
	cases := []struct {
		name      string
		copyErr   error
		requested bool
	}{
		{"eof", nil, false},
		{"eof-with-close-request", nil, true},
		{"read-error", os.ErrClosed, false},
		{"read-error-with-close-request", os.ErrClosed, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			stream := invocationStream{buffer: cappedBuffer{data: []byte("kept")}}
			var closeErr error
			if tc.requested {
				closeErr = context.DeadlineExceeded
			}
			text, state, err := stream.result(tc.copyErr, tc.requested, closeErr)
			if text != "kept" || state.RetainedBytes != 4 || state.LimitExceeded ||
				state.EOF != (tc.copyErr == nil) || state.DeadlineCloseRequested != tc.requested {
				t.Fatal("copy completion and close request were conflated", state, err)
			}
			if errors.Is(err, context.DeadlineExceeded) != tc.requested ||
				errors.Is(err, os.ErrClosed) != (tc.copyErr != nil) {
				t.Fatal("original copy or deadline cause disappeared", err)
			}
		})
	}
}
