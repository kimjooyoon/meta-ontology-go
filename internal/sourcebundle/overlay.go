package sourcebundle

import (
	"bytes"
	"io/fs"
	"path"
	"time"
)

// Overlay provides explicit file reads over caller-owned bytes. It does not
// enumerate added directory entries or establish a filesystem snapshot.
// A nil base makes reads outside the supplied members fail.
type Overlay struct {
	fs.FS
	Files map[string][]byte
}

func (overlay Overlay) Open(path string) (fs.File, error) {
	if !fs.ValidPath(path) {
		return nil, &fs.PathError{Op: "open", Path: path, Err: fs.ErrInvalid}
	}
	if raw, found := overlay.Files[path]; found {
		return &memoryFile{Reader: bytes.NewReader(raw), path: path, size: int64(len(raw))}, nil
	}
	if overlay.FS == nil {
		return nil, &fs.PathError{Op: "open", Path: path, Err: fs.ErrNotExist}
	}
	return overlay.FS.Open(path)
}

type memoryFile struct {
	*bytes.Reader
	path string
	size int64
}

func (file *memoryFile) Close() error { return nil }
func (file *memoryFile) Stat() (fs.FileInfo, error) {
	return memoryFileInfo{file.path, file.size}, nil
}

type memoryFileInfo struct {
	name string
	size int64
}

func (info memoryFileInfo) Name() string       { return path.Base(info.name) }
func (info memoryFileInfo) Size() int64        { return info.size }
func (info memoryFileInfo) Mode() fs.FileMode  { return 0444 }
func (info memoryFileInfo) ModTime() time.Time { return time.Time{} }
func (info memoryFileInfo) IsDir() bool        { return false }
func (info memoryFileInfo) Sys() any           { return nil }
