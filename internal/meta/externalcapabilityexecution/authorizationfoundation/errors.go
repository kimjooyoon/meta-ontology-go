package authorizationfoundation

import "fmt"

type resolutionError struct {
	Decision   string
	Resolution string
	Reason     string
}

func (value *resolutionError) Error() string {
	return fmt.Sprintf("%s/%s: %s", value.Decision, value.Resolution, value.Reason)
}

func unknown(reason string) error {
	return &resolutionError{Decision: "FAIL_CLOSED", Resolution: "UNKNOWN", Reason: reason}
}

func denied(reason string) error {
	return &resolutionError{Decision: "DENIED", Resolution: "EXACT", Reason: reason}
}
