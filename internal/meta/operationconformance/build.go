package operationconformance

import (
	"bytes"
	"fmt"
	"go/build"
	"io"
	"os"
	"path/filepath"
)

func observeBuildSemantics(evidence SplitGoEvidence) Decision {
	if len(evidence.BuildContexts) == 0 || len(evidence.Candidates) == 0 {
		return DecisionUnknown
	}
	for _, context := range evidence.BuildContexts {
		before, err := matchFile(context, evidence.Source)
		if err != nil {
			return DecisionFail
		}
		for _, candidate := range evidence.Candidates {
			after, matchErr := matchFile(context, candidate)
			if matchErr != nil || after != before {
				return DecisionFail
			}
		}
	}
	return DecisionPass
}

func matchFile(context BuildContext, file FileEvidence) (bool, error) {
	if context.GOOS == "" || context.GOARCH == "" || file.Path == "" || len(file.Data) == 0 {
		return false, fmt.Errorf("incomplete build evidence")
	}
	value := build.Default
	value.GOOS, value.GOARCH = context.GOOS, context.GOARCH
	value.CgoEnabled, value.BuildTags = context.CgoEnabled, append([]string(nil), context.BuildTags...)
	target := filepath.Clean(filepath.FromSlash(file.Path))
	value.OpenFile = func(name string) (io.ReadCloser, error) {
		if filepath.Clean(name) != target {
			return nil, os.ErrNotExist
		}
		return io.NopCloser(bytes.NewReader(file.Data)), nil
	}
	return value.MatchFile(filepath.Dir(target), filepath.Base(target))
}
