package main

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/originlenspolicy"
	"github.com/kimjooyoon/meta-ontology-go/internal/viewlens"
)

const bodyViewHelp = `Inspect canonical source origins with an advisory presentation

` + bodyViewUsage + `

  gooo body-view --capabilities
  gooo body-view --activity Difference examples/named-inputs/main.gooo
  gooo body-view --activity Difference --actor kimjooyoon --audience person \
    --prefer inputs --format text examples/named-inputs/main.gooo

JSON retains the complete canonical source report. Projections use references
into that report. Hints never edit source/IR, select codegen paths or change the
canonical status. Invalid hint JSON, unknown view and unavailable hint policy
fall back with diagnostic codes. Source program errors remain source errors.

The host may preload a small hint via "gooo view-lens load", then pass the hint
object as --hint JSON. Rendering performs no preference I/O, lock, queue, clock,
network or model calls. No hints are loaded automatically by compiler commands.
This removes storage waiting; rendering still has explicit CPU/output cost.

The response provides a record_event and a next operation. After the response,
a host may persist it with "gooo view-lens record --event JSON". Record uses a
nonblocking lock, refuses stale snapshots and preserves each accepted event in
a losslessly compressed archive. Profiles retain counts and 32 recent choices;
the event archive retains older records. Storage is explicit and advisory.

--prefer pins a preference; --prefer auto clears it. --view affects one request.
Otherwise the Gooo policy chooses the most selected view, with an audience
default on ties. Counts describe requests, not reading or inferred intent.
A host should choose a distinct opaque --actor for each user or agent.
`

const viewLensHelp = `Detached local storage for advisory origin views

  gooo view-lens load --actor kimjooyoon --audience person [--state-dir directory]
  gooo view-lens record --event '<record_event JSON>' [--state-dir directory]
  gooo view-lens read --kind event|profile --digest sha256:... [--state-dir directory]

Load returns a hint object for body-view --hint. Record accepts the emitted event,
checks the current profile origin and archives its canonical JSON bytes before
updating the profile. A stale or busy profile returns state=UNAVAILABLE with an
issue; persisted=false. Retrying the last accepted event is idempotent.
Storage failure never changes compiler results. Hosts own scheduling and retries;
the CLI does not claim background writes survive process exit.
Read restores raw_bytes_base64 with its raw_digest; it preserves historical
event and profile bytes without granting execution authority.
`

func writeBodyViewCapabilities(stdout, stderr io.Writer) int {
	value := map[string]any{"schema": "gooo/body-view-capabilities/v1",
		"output_schema": "gooo/body-view/v1", "default_format": "json", "default_audience": "agent",
		"default_actor": "agent", "domain": viewlens.Domain, "authority": "advisory",
		"hint_byte_limit": 4096, "automatic_storage": false, "critical_path_storage_wait": false,
		"views": []map[string]string{{"alias": "full", "id": viewlens.Full},
			{"alias": "inputs", "id": viewlens.NamedInputs}},
		"preference_actions": []string{"prefer", "clear_preference", "load", "record", "read"},
		"canonical_pointer":  "#/canonical", "policy": originlenspolicy.Origin()}
	if err := json.NewEncoder(stdout).Encode(value); err != nil {
		fmt.Fprintln(stderr, err)
		return exitFailure
	}
	return exitOK
}
