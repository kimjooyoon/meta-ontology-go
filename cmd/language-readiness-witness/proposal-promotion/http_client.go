package main

import (
	"net/http"
	"net/url"
	"strings"
	"time"
)

type proposalRedirectFailure struct{}

func (proposalRedirectFailure) Error() string           { return "PROPOSAL_REDIRECT_ORIGIN_MISMATCH" }
func (proposalRedirectFailure) RedirectOriginMismatch() {}

func newProposalHTTPClient(apiURL string, store *proposalObservationStore) (*http.Client, error) {
	base, err := url.Parse(strings.TrimRight(apiURL, "/"))
	if err != nil || base.Scheme == "" || base.Host == "" || base.User != nil || base.Fragment != "" {
		return nil, &proposalRedirectFailure{}
	}
	client := &http.Client{
		Timeout:   30 * time.Second,
		Transport: &proposalObservationTransport{base: http.DefaultTransport, store: store},
	}
	client.CheckRedirect = func(request *http.Request, _ []*http.Request) error {
		if request == nil || request.URL == nil || request.URL.User != nil || request.URL.Fragment != "" ||
			!strings.EqualFold(request.URL.Scheme, base.Scheme) || !strings.EqualFold(request.URL.Host, base.Host) {
			return proposalRedirectFailure{}
		}
		return nil
	}
	return client, nil
}
