package sourcebundle

import (
	"errors"
	"fmt"
	"io/fs"
	"strings"
	"testing"
	"testing/fstest"
)

type observedFS struct {
	fs.FS
	opens, closes, bytes int
	failRead, failClose  string
}

type observedFile struct {
	fs.File
	owner *observedFS
	name  string
}

func (f *observedFS) Open(name string) (fs.File, error) {
	f.opens++
	file, err := f.FS.Open(name)
	if err != nil {
		return nil, err
	}
	return &observedFile{file, f, name}, nil
}

var errRead = errors.New("injected read failure")
var errClose = errors.New("injected close failure")

func (f *observedFile) Read(p []byte) (int, error) {
	if f.name == f.owner.failRead {
		return 0, errRead
	}
	n, err := f.File.Read(p)
	f.owner.bytes += n
	return n, err
}

func (f *observedFile) Close() error {
	f.owner.closes++
	err := f.File.Close()
	if f.name == f.owner.failClose {
		return errClose
	}
	return err
}

func TestReadValidatesEveryNameBeforeOpeningFiles(t *testing.T) {
	for _, bad := range []string{"../b.gooo", "/b.gooo", "a/../b.gooo", "A.gooo", "a\\b.gooo", "b.go"} {
		files := &observedFS{FS: fstest.MapFS{"a.gooo": {Data: []byte("first")}}}
		if got, err := Read(files, []string{"a.gooo", bad}); err == nil || got != nil || files.opens != 0 {
			t.Fatalf("invalid/duplicate name reached I/O: name=%q got=%v opens=%d err=%v", bad, got, files.opens, err)
		}
	}
}

func TestReadFailureReturnsNoEarlierMembersAndClosesHandles(t *testing.T) {
	for _, failure := range []string{"open", "read", "close"} {
		t.Run(failure, func(t *testing.T) {
			files := &observedFS{FS: fstest.MapFS{
				"a.gooo": {Data: []byte("first")}, "b.gooo": {Data: []byte("second")},
			}}
			want, wantClosed := fs.ErrNotExist, 1
			switch failure {
			case "open":
				delete(files.FS.(fstest.MapFS), "b.gooo")
			case "read":
				files.failRead, want, wantClosed = "b.gooo", errRead, 2
			case "close":
				files.failClose, want, wantClosed = "b.gooo", errClose, 2
			}
			got, err := Read(files, []string{"a.gooo", "b.gooo"})
			if got != nil || !errors.Is(err, want) || files.closes != wantClosed {
				t.Fatalf("partial source or lost I/O error/handle: got=%v err=%v closes=%d", got, err, files.closes)
			}
		})
	}
}

func TestReadBoundsActualBytesWithoutTrustingFileSize(t *testing.T) {
	for _, aggregate := range []bool{false, true} {
		files, names := fstest.MapFS{}, []string{}
		count, size, budget := 1, MaxSourceBytes+2, MaxSourceBytes
		if aggregate {
			count, size, budget = MaxTotalBytes/MaxSourceBytes+1, MaxSourceBytes, MaxTotalBytes
		}
		data := []byte(strings.Repeat("x", size))
		for i := range count {
			name := fmt.Sprintf("member-%d.gooo", i)
			files[name] = &fstest.MapFile{Data: data}
			names = append(names, name)
		}
		observed := &observedFS{FS: files}
		got, err := Read(observed, names)
		if got != nil || err == nil || observed.bytes != budget+1 || observed.closes != count {
			t.Fatalf("unbounded or partial read: aggregate=%t bytes=%d closes=%d err=%v", aggregate,
				observed.bytes, observed.closes, err)
		}
	}
}
