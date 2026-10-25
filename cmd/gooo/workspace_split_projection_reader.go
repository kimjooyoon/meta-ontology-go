package main

import (
	"fmt"
	"path/filepath"
)

func recheckSourceSplitProjection(options workspaceSourceSplitOptions,
	saved, current workspaceSourceSplitReceipt, writes []atomicWrite, priorCheck func() error,
) error {
	if priorCheck != nil {
		if err := priorCheck(); err != nil {
			return err
		}
	}
	if _, err := currentWorkspaceSourceSplitProposal(options, saved); err != nil {
		return err
	}
	if _, err := verifySourceSplitOutput(current, sourceSplitWriteReader{writes: writes}); err != nil {
		return fmt.Errorf("projected split output: %w", err)
	}
	return nil
}

// sourceSplitWriteReader models replacement of directory entries. It follows
// directory aliases while keeping separate hard-link occurrences independent.
type sourceSplitWriteReader struct{ writes []atomicWrite }

func (reader sourceSplitWriteReader) ReadFile(filename string) ([]byte, error) {
	absolute, err := filepath.Abs(filename)
	if err != nil {
		return nil, err
	}
	parent, err := filepath.EvalSymlinks(filepath.Dir(absolute))
	if err != nil {
		return nil, err
	}
	address := filepath.Join(parent, filepath.Base(absolute))
	for _, write := range reader.writes {
		if write.path == address {
			return write.data, nil
		}
		shared, err := samePublicationEntry(address, write.path)
		if err != nil {
			return nil, err
		}
		if shared {
			return write.data, nil
		}
	}
	return (OSFileReader{}).ReadFile(filename)
}
