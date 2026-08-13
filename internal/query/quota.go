package query

// queryWorkQuota bounds graph-edge inspection during one query operation.
// A zero limit keeps the historical unbounded direct Go API; versioned
// envelopes always provide a positive limit.
type queryWorkQuota struct {
	remaining int
}

func newQueryWorkQuota(limit int) *queryWorkQuota {
	if limit <= 0 {
		return &queryWorkQuota{remaining: -1}
	}
	return &queryWorkQuota{remaining: limit}
}

func (quota *queryWorkQuota) take() bool {
	if quota.remaining < 0 {
		return true
	}
	if quota.remaining == 0 {
		return false
	}
	quota.remaining--
	return true
}

func (quota *queryWorkQuota) exhausted() bool {
	return quota.remaining == 0
}
