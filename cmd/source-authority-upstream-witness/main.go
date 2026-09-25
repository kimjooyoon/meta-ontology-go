package main

import (
	"context"
	"fmt"
	"os"
)

func main() {
	if err := run(context.Background(), os.Args[1:], newHTTPFetcher()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
