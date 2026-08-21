package main

type cliQueryValidationError struct {
	code    string
	message string
}

func (err cliQueryValidationError) Error() string {
	return err.code + ": " + err.message
}
