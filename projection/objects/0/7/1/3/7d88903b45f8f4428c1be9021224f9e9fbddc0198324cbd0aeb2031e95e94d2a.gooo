package adapter

import (
	"fmt"
	"os"
)

func capturePath(path string) (FileObservation, error) {
	info, err := os.Lstat(path)
	if os.IsNotExist(err) {
		return FileObservation{Path: path, Kind: "missing", Lstat: LstatIdentity{}}, nil
	}
	if err != nil {
		return FileObservation{}, err
	}
	observation := FileObservation{
		Path: path, Kind: fileKind(info), Exists: true, Lstat: makeLstat(info),
	}
	if info.Mode().IsRegular() {
		data, err := os.ReadFile(path)
		if err != nil {
			return FileObservation{}, err
		}
		observation.ByteDigest = digestBytes(data)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		target, err := os.Readlink(path)
		if err != nil {
			return FileObservation{}, err
		}
		observation.ByteDigest = digestBytes([]byte(target))
	}
	if info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
		after, err := os.Lstat(path)
		if err != nil || makeLstat(after) != observation.Lstat {
			return FileObservation{}, fmt.Errorf("path changed during capture")
		}
	}
	return observation, nil
}
func fileKind(info os.FileInfo) string {
	switch {
	case info.Mode().IsRegular():
		return "file"
	case info.IsDir():
		return "directory"
	case info.Mode()&os.ModeSymlink != 0:
		return "symlink"
	default:
		return "other"
	}
}
func makeLstat(info os.FileInfo) LstatIdentity {
	return LstatIdentity{
		Exists: true, Device: statNumber(info.Sys(), "Dev"), Inode: statNumber(info.Sys(), "Ino"),
		Mode: info.Mode().String(), Size: info.Size(), ModTimeUnixNano: info.ModTime().UnixNano(),
	}
}
