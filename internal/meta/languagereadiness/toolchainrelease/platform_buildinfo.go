package toolchainrelease

import (
	"debug/buildinfo"
	"fmt"
)

func inspectBuild(path string, input BuildInput) (BuildEvidence, error) {
	info, err := buildinfo.ReadFile(path)
	if err != nil {
		return BuildEvidence{}, err
	}
	settings := map[string]string{}
	for _, setting := range info.Settings {
		settings[setting.Key] = setting.Value
	}
	if info.GoVersion != ExpectedToolchain ||
		settings["GOOS"] != input.Target.GOOS ||
		settings["GOARCH"] != input.Target.GOARCH ||
		settings["CGO_ENABLED"] != "0" {
		return BuildEvidence{}, fmt.Errorf("TOOLCHAIN_RELEASE_BUILD_SETTINGS_MISMATCH")
	}
	if settings["vcs.revision"] != input.ExpectedHead {
		return BuildEvidence{}, fmt.Errorf("TOOLCHAIN_RELEASE_VCS_REVISION_MISMATCH")
	}
	if settings["vcs.modified"] != "false" {
		return BuildEvidence{}, fmt.Errorf("TOOLCHAIN_RELEASE_DIRTY_BUILD")
	}
	return BuildEvidence{
		VCSRevision: settings["vcs.revision"],
		VCSModified: false,
		Trimpath:    true,
		BuildVCS:    true,
		CGOEnabled:  false,
	}, nil
}
