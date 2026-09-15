package artifactemit

type symbolicReaderChecks struct {
	Schema             bool
	Subject            bool
	Metric             bool
	Decision           bool
	Resolution         bool
	InternalDigest     bool
	UpstreamDigests    bool
	UnknownBranches    bool
	UniqueIndicatorID  bool
	UserPresent        bool
	ToolPresent        bool
	GovernorPresent    bool
	UserCountBound     bool
	ToolCountBound     bool
	GovernorCountBound bool
	UserNested         bool
	ToolNested         bool
	ReaderResolutions  bool
}

func (checks symbolicReaderChecks) passed() bool {
	return checks.Schema && checks.Subject && checks.Metric && checks.Decision &&
		checks.Resolution && checks.InternalDigest && checks.UpstreamDigests &&
		checks.UnknownBranches && checks.UniqueIndicatorID && checks.UserPresent &&
		checks.ToolPresent && checks.GovernorPresent && checks.UserCountBound &&
		checks.ToolCountBound && checks.GovernorCountBound && checks.UserNested &&
		checks.ToolNested && checks.ReaderResolutions
}
