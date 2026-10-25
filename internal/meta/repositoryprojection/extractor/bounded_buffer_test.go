package extractor

import (
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func TestBoundedBufferCopy(t *testing.T) {
	for _, limit := range []int{0, 4, 10, 12} {
		for _, readerOnly := range []bool{false, true} {
			buffer := &boundedBuffer{limit: limit}
			var source io.Reader = strings.NewReader("0123456789")
			if readerOnly {
				source = io.LimitReader(source, 10)
			}
			count, err := io.Copy(buffer, source)
			want := "0123456789"[:min(limit, 10)]
			if err != nil || count != 10 || string(buffer.Bytes()) != want || buffer.exceeded != (limit < 10) {
				t.Fatalf("limit=%d readerOnly=%t: count=%d retained=%q exceeded=%t error=%v",
					limit, readerOnly, count, buffer.Bytes(), buffer.exceeded, err)
			}
		}
	}
}

func TestBoundedBufferCopyAfterWrite(t *testing.T) {
	buffer := &boundedBuffer{limit: 4}
	_, _ = buffer.Write([]byte("01"))
	count, err := io.Copy(buffer, io.LimitReader(strings.NewReader("23456789"), 8))
	if err != nil || count != 8 || string(buffer.Bytes()) != "0123" || !buffer.exceeded {
		t.Fatalf("count=%d retained=%q exceeded=%t error=%v", count, buffer.Bytes(), buffer.exceeded, err)
	}
}

func TestBoundedBufferRetainsReadFailure(t *testing.T) {
	wantErr := errors.New("source failed after output")
	source := io.MultiReader(io.LimitReader(strings.NewReader("0123456789"), 10), boundedFailReader{wantErr})
	buffer := &boundedBuffer{limit: 4}
	count, err := io.Copy(buffer, source)
	if count != 10 || !errors.Is(err, wantErr) || string(buffer.Bytes()) != "0123" || !buffer.exceeded {
		t.Fatalf("count=%d retained=%q exceeded=%t error=%v", count, buffer.Bytes(), buffer.exceeded, err)
	}
}

type boundedFailReader struct{ err error }

func (reader boundedFailReader) Read([]byte) (int, error) { return 0, reader.err }

func TestBoundedBufferEmptyWriteAtLimit(t *testing.T) {
	for _, limit := range []int{0, 4} {
		buffer := &boundedBuffer{limit: limit}
		_, _ = buffer.Write([]byte(strings.Repeat("x", limit)))
		count, err := buffer.Write(nil)
		if count != 0 || err != nil || buffer.Len() != limit || buffer.exceeded {
			t.Fatalf("limit=%d count=%d retained=%d exceeded=%t error=%v", limit, count, buffer.Len(), buffer.exceeded, err)
		}
	}
}

func TestBoundedBufferDrainsProcessPipes(t *testing.T) {
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, executable, "-test.run=^TestBoundedBufferProcessHelper$")
	command.Env = append(os.Environ(), "GOOO_BOUNDED_BUFFER_TEST_CHILD=1")
	command.WaitDelay = time.Second
	stdout, stderr := &boundedBuffer{limit: 31}, &boundedBuffer{limit: 17}
	command.Stdout, command.Stderr = stdout, stderr
	if err := command.Run(); err != nil {
		t.Fatalf("child did not drain both pipes: %v (context: %v)", err, ctx.Err())
	}
	for _, buffer := range []*boundedBuffer{stdout, stderr} {
		if buffer.Len() != buffer.limit || !buffer.exceeded || string(buffer.Bytes()) != strings.Repeat("x", buffer.limit) {
			t.Fatalf("limit=%d retained=%d exceeded=%t", buffer.limit, buffer.Len(), buffer.exceeded)
		}
	}
}

func TestBoundedBufferProcessHelper(t *testing.T) {
	if os.Getenv("GOOO_BOUNDED_BUFFER_TEST_CHILD") != "1" {
		return
	}
	chunk := strings.Repeat("x", 64<<10)
	for range 32 {
		if _, err := io.WriteString(os.Stdout, chunk); err != nil {
			os.Exit(2)
		}
		if _, err := io.WriteString(os.Stderr, chunk); err != nil {
			os.Exit(3)
		}
	}
	os.Exit(0)
}
