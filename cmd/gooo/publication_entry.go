package main

import (
	"os"
	"path/filepath"
)

func samePublicationEntry(left, right string) (bool, error) {
	a, err := os.Lstat(left)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	b, err := os.Lstat(right)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if !a.Mode().IsRegular() || !b.Mode().IsRegular() || !os.SameFile(a, b) {
		return false, nil
	}
	ap, err := os.Stat(filepath.Dir(left))
	if err != nil {
		return false, err
	}
	bp, err := os.Stat(filepath.Dir(right))
	if err != nil || !os.SameFile(ap, bp) {
		return false, err
	}
	return comparePublicationNames(filepath.Dir(left), filepath.Base(left), filepath.Base(right), a)
}
