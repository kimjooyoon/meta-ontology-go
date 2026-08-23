# Toolchain CLI corpus

This directory contains the versioned, non-authoring case registry for the
actual `gooo` executable. CI maps each known operation to fixed arguments; the
JSON file cannot introduce an arbitrary process or command.

The six positive paths cover text and structured identity, syntax checking,
semantic checking, and bidirectional roundtrip. The six guardrails cover
missing and unknown command boundaries. Every path is replayed byte-for-byte.
