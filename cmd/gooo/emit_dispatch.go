package main

import (
	"fmt"
	"io"
)

func runExtensionCommand(args []string, stdout, stderr io.Writer) int {
	switch args[0] {
	case "emit":
		return runEmit(args[1:], stdout, stderr)
	case "certify":
		return runRetentionCertify(args[1:], OSFileReader{}, EntityFieldsCLIParser{}, stdout, stderr)
	case "consume":
		return runRetentionConsume(args[1:], OSFileReader{}, EntityFieldsCLIParser{}, stdout, stderr)
	case "authorize-discovery":
		return runContinuityAuthorize(args[1:], OSFileReader{}, stdout, stderr)
	case "certify-discovery":
		return runContinuityCertify(args[1:], OSFileReader{}, EntityFieldsCLIParser{}, stdout, stderr)
	}
	fmt.Fprintf(stderr, "gooo: command %q is not implemented yet\n", args[0])
	return exitFailure
}
