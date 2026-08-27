package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	sandbox "github.com/kimjooyoon/meta-ontology-go/internal/meta/reflectivequerysandbox"
)

func main() {
	source := flag.String("source", "", "Gooo source")
	subject := flag.String("subject-sha", "", "exact subject commit")
	output := flag.String("output", "", "observation receipt")
	flag.Parse()
	if *source == "" || *subject == "" || *output == "" {
		fail("usage: producer -source FILE -subject-sha SHA -output FILE")
	}
	observation, err := sandbox.Observe(*source, *subject)
	if err != nil {
		fail("observe source: %v", err)
	}
	observation = sandbox.SealObservation(observation)
	data, err := json.MarshalIndent(observation, "", "  ")
	if err != nil {
		fail("encode observation: %v", err)
	}
	if err := os.WriteFile(*output, append(data, '\n'), 0o644); err != nil {
		fail("write observation: %v", err)
	}
	fmt.Printf("producer observation: %s nodes=%d facts=%d attempts=%d transitions=%d\n", observation.Schema, observation.Source.NodeCount, observation.Source.FactCount, len(observation.Attempts), len(observation.Claims))
}

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
