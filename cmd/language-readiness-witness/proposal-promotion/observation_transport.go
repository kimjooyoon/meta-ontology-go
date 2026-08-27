package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
)

type proposalObservationTransport struct {
	base  http.RoundTripper
	store *proposalObservationStore
}

type proposalObservedTransportFailure struct{ Reason string }

func (failure *proposalObservedTransportFailure) Error() string { return failure.Reason }

func (transport *proposalObservationTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	if transport.store != nil && transport.store.replay {
		observed, err := transport.store.next(request.Method, request.URL.String())
		if err != nil {
			return nil, err
		}
		if observed.Failure != "" {
			return nil, &proposalObservedTransportFailure{Reason: observed.Failure}
		}
		response := &http.Response{
			StatusCode: observed.StatusCode,
			Status:     fmt.Sprintf("%d", observed.StatusCode),
			Header:     make(http.Header),
			Body:       io.NopCloser(bytes.NewReader(observed.Body)),
			Request:    request,
		}
		if observed.Link != "" {
			response.Header.Set("Link", observed.Link)
		}
		return response, nil
	}
	base := transport.base
	if base == nil {
		base = http.DefaultTransport
	}
	response, err := base.RoundTrip(request)
	if err != nil {
		transport.store.record(proposalObservedResponse{Kind: request.Method, URL: request.URL.String(), Failure: "TRANSPORT_ERROR"})
		return nil, err
	}
	body, readErr := io.ReadAll(io.LimitReader(response.Body, proposalObservationMaxBytes+1))
	_ = response.Body.Close()
	if readErr != nil || len(body) > proposalObservationMaxBytes {
		transport.store.record(proposalObservedResponse{Kind: request.Method, URL: request.URL.String(), StatusCode: response.StatusCode, Failure: "BODY_READ_ERROR"})
		if readErr == nil {
			readErr = fmt.Errorf("response exceeds fixed observation bound")
		}
		return nil, readErr
	}
	transport.store.record(proposalObservedResponse{
		Kind: request.Method, URL: request.URL.String(), StatusCode: response.StatusCode,
		Body: body, Link: response.Header.Get("Link"),
	})
	response.Body = io.NopCloser(bytes.NewReader(body))
	return response, nil
}

const proposalObservationMaxBytes = 16 << 20
