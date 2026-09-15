package toolchainrelease

type Target struct {
	ID            string `json:"id"`
	Runner        string `json:"runner"`
	GOOS          string `json:"goos"`
	GOARCH        string `json:"goarch"`
	ArchiveFormat string `json:"archive_format"`
}

var targetRegistry = []Target{
	{ID: "linux-amd64", Runner: "ubuntu-24.04", GOOS: "linux", GOARCH: "amd64", ArchiveFormat: "tar.gz"},
	{ID: "darwin-amd64", Runner: "macos-15-intel", GOOS: "darwin", GOARCH: "amd64", ArchiveFormat: "tar.gz"},
	{ID: "windows-amd64", Runner: "windows-2025", GOOS: "windows", GOARCH: "amd64", ArchiveFormat: "zip"},
}

func Targets() []Target {
	return append([]Target(nil), targetRegistry...)
}

func expectedTarget(id string) (Target, bool) {
	for _, target := range targetRegistry {
		if target.ID == id {
			return target, true
		}
	}
	return Target{}, false
}
