# Gooo trace debugger experiment

This experiment projects the deterministic source-execution event stream up to a selected event.

```sh
gooo debug --json --entry PayOrder --break-event ACTIVITY_INVOKED examples/billing/main.gooo
```

The receipt can prove that a fixed event was reached and expose the trace prefix. It does not claim
interactive control, variable inspection, or time travel.
