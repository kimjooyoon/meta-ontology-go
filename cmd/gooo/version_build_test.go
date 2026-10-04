package main

import (
	"bytes"
	"encoding/json"
	"io"
	"runtime/debug"
	"strings"
	"testing"
)

func versionBuildFixture() *debug.BuildInfo {
	return &debug.BuildInfo{
		GoVersion: "go1.27.1", Main: debug.Module{Version: "v0.4.0-dev.0.20261004014639-4f6c7566dd43"},
		Deps:     []*debug.Module{{Path: "github.com/kimjooyoon/gooo-decision-runtime", Version: "v0.2.21-experimental"}},
		Settings: []debug.BuildSetting{{Key: "vcs.revision", Value: strings.Repeat("a", 40)}, {Key: "vcs.modified", Value: "false"}},
	}
}

func TestVersionBuildBindsActualMetadata(t *testing.T) {
	got := versionBuildMetadata(versionBuildFixture(), "go1.27.1")
	if got.Schema != "gooo/build-identity/v1" || got.GoVersion != "go1.27.1" ||
		got.CompilerSourceSHA != strings.Repeat("a", 40) || got.SourceStatus != "CLEAN_VCS" ||
		got.VCSModified != "false" || got.ModuleVersion != versionBuildFixture().Main.Version ||
		got.SDK.Version != "v0.2.21-experimental" || got.SDK.ReplacementPath != "" ||
		got.NativeGoRequired != "go1.27.1" {
		t.Fatalf("build identity = %+v", got)
	}
}

func TestVersionBuildRetainsMissingDirtyAndMalformedSource(t *testing.T) {
	for _, tc := range []struct{ name, revision, modified, status string }{
		{"missing", "", "", "UNOBSERVED_VCS"},
		{"dirty", strings.Repeat("a", 40), "true", "MODIFIED_VCS"},
		{"short", "1234", "false", "UNOBSERVED_VCS"},
		{"upper", strings.Repeat("A", 40), "false", "UNOBSERVED_VCS"},
		{"not-hex", strings.Repeat("z", 40), "false", "UNOBSERVED_VCS"},
		{"unknown-modified", strings.Repeat("a", 40), "", "UNOBSERVED_VCS"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			info := versionBuildFixture()
			info.Settings = []debug.BuildSetting{{Key: "vcs.revision", Value: tc.revision}, {Key: "vcs.modified", Value: tc.modified}}
			got := versionBuildMetadata(info, "go1.27.1")
			if got.CompilerSourceSHA != "UNBOUND_LOCAL_SOURCE" || got.SourceStatus != tc.status {
				t.Fatalf("unbound metadata = %+v", got)
			}
		})
	}
	got := versionBuildMetadata(nil, "go1.27.1")
	if got.GoVersion != "go1.27.1" || got.CompilerSourceSHA != "UNBOUND_LOCAL_SOURCE" ||
		got.ModuleVersion != "unobserved" || got.SDK.Version != "unobserved" || got.VCSModified != "unobserved" {
		t.Fatalf("missing build info = %+v", got)
	}
}

func TestVersionBuildRetainsSDKReplacement(t *testing.T) {
	info := versionBuildFixture()
	info.Deps[0].Replace = &debug.Module{Path: "../local-runtime", Version: ""}
	got := versionBuildMetadata(info, "go1.27.1")
	if got.SDK.Version != "v0.2.21-experimental" || got.SDK.ReplacementPath != "../local-runtime" ||
		got.SDK.ReplacementVersion != "unobserved" {
		t.Fatalf("SDK replacement = %+v", got.SDK)
	}
	info.Deps = nil
	if got := versionBuildMetadata(info, "go1.27.1"); got.SDK.Version != "unobserved" {
		t.Fatalf("missing SDK = %+v", got.SDK)
	}
}

func TestRunVersionBuildSupportsBothJSONFlagOrders(t *testing.T) {
	for _, args := range [][]string{{"--build", "--json"}, {"--json", "--build"}} {
		var first, second, stderr bytes.Buffer
		if code := runVersion(args, &first, &stderr); code != exitOK || stderr.Len() != 0 {
			t.Fatalf("code=%d stderr=%q", code, stderr.String())
		}
		if code := runVersion(args, &second, &stderr); code != exitOK || first.String() != second.String() {
			t.Fatalf("non-deterministic same-binary identity")
		}
		var got versionBuildInfo
		if err := json.Unmarshal(first.Bytes(), &got); err != nil {
			t.Fatal(err)
		}
		if got.Schema != "gooo/build-identity/v1" || got.NativeGoRequired != "go1.27.1" || got.GoVersion == "" {
			t.Fatalf("JSON identity=%+v", got)
		}
	}
	var text, stderr bytes.Buffer
	if code := runVersion([]string{"--build"}, &text, &stderr); code != exitOK || !strings.Contains(text.String(), "compiler source:") || !strings.Contains(text.String(), "--go-bin") {
		t.Fatalf("build text code=%d output=%q", code, text.String())
	}
}

func TestRunVersionBuildRejectsDuplicateAndUnknownFlags(t *testing.T) {
	for _, args := range [][]string{{"--build", "--build"}, {"--json", "--json"}, {"--build", "--unknown"}, {"--build", "--json", "--extra"}} {
		var output, stderr bytes.Buffer
		if code := runVersion(args, &output, &stderr); code != exitUsage || output.Len() != 0 || stderr.String() != versionUsage+"\n" {
			t.Fatalf("args=%v code=%d stderr=%q", args, code, stderr.String())
		}
	}
}

type versionClosedWriter struct{}

func (versionClosedWriter) Write([]byte) (int, error) { return 0, io.ErrClosedPipe }
func TestRunVersionBuildPreservesWriteErrors(t *testing.T) {
	for _, args := range [][]string{{"--build"}, {"--build", "--json"}} {
		if code := runVersion(args, versionClosedWriter{}, io.Discard); code != exitFailure {
			t.Fatalf("write error args=%v code=%d", args, code)
		}
	}
}
