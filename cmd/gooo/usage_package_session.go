package main

const packageSessionTopicHelp = `Keep one native artifact owner across workspace requests

Usage:
  gooo package session

Write one JSON request per line and read its response before the next request:
  {"schema":"gooo/workspace-session-request/v1","id":"first","operation":"execute","arguments":["--cases","cases.json","gooo.workspace.json"]}

Operations are execute, replay and close. Execute/replay arguments follow the
existing commands. Responses contain id, exit_code and the ordinary receipt.
Read decision and receipt.result.runtime, including the current finite counts.
Exit 0 may carry PROGRESS or input-only OBSERVED.

Save the response's receipt object before replaying it with --receipt. Every
request reloads current files and validates current input/source/tool identity.
One verified executable may be retained; each execution still runs twice.
Historical Artifact.SourceBuild is distinct from current Build and Runs.

Close after reading the last response, or finish stdin for orderly cleanup.
SIGINT/SIGTERM cancels active work. Requests are serial, with no internal queue.
Framing is JSON; omit --receipt-compression. --json and --timing are accepted.
Request lines are limited to 64 KiB, ids to 128 bytes and arguments to 48.
Malformed requests return a per-request error; a later valid request can follow.

See docs/language/workspace-sessions.md for the receipt and lifecycle contract.
`
