package syntaxregistration

import (
	"bytes"
	"io/fs"
	"path"
	"time"
)

type profileOverlay struct {
	fs.FS
	members map[string][]byte
}

func (overlay profileOverlay) Open(path string) (fs.File, error) {
	if !fs.ValidPath(path) {
		return nil, &fs.PathError{Op: "open", Path: path, Err: fs.ErrInvalid}
	}
	if raw, found := overlay.members[path]; found {
		return &profileMemoryFile{Reader: bytes.NewReader(raw), path: path, size: int64(len(raw))}, nil
	}
	if overlay.FS == nil {
		return nil, &fs.PathError{Op: "open", Path: path, Err: fs.ErrNotExist}
	}
	return overlay.FS.Open(path)
}

type profileMemoryFile struct {
	*bytes.Reader
	path string
	size int64
}

func (file *profileMemoryFile) Close() error { return nil }
func (file *profileMemoryFile) Stat() (fs.FileInfo, error) {
	return profileFileInfo{file.path, file.size}, nil
}

type profileFileInfo struct {
	name string
	size int64
}

func (info profileFileInfo) Name() string       { return path.Base(info.name) }
func (info profileFileInfo) Size() int64        { return info.size }
func (info profileFileInfo) Mode() fs.FileMode  { return 0444 }
func (info profileFileInfo) ModTime() time.Time { return time.Time{} }
func (info profileFileInfo) IsDir() bool        { return false }
func (info profileFileInfo) Sys() any           { return nil }
