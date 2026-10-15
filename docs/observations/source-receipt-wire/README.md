# Reading source origins from a large execution record

The original race job116067763571 in run38626757149 used source
cd5ed4381e1800952632ddf72c4ce55ecfe63a25. Its full package inventory started431
packages:242 passed,187 skipped and2 failed at their default ten-minute package
deadline. The failing packages were `cmd/gooo` and the repository extractor.
The original [job and event summary](original-race/event-summary.json), exact
compressed log and both timeout stacks are retained here.

The CLI stack was inside `bodyexecution.exactKeys`, reparsing a6.8 MiB nested
execution receipt. Each nested object or array previously decoded another copy
of its subtree to check field spelling, after the duplicate-key scan. The
extractor stack was parsing a capacity-contract source file; its active test had
run for one second when the cumulative package deadline expired. These stacks
show active work at timeout. They do not establish a deadlock or a comparative
speedup.

The execution-record reader now checks declared field spelling and duplicate
keys in the same streaming traversal. An invocation keeps one field table for
each encountered struct type, then uses the existing typed decoder to construct
the result. Source records and parser results are never cached across calls.
Numbers, opaque record bytes, pointers, typed maps and arrays retain their wire
contracts. The32 MiB envelope limit,64-level depth limit and original source
record interpretation remain.

## New input scope73143

Three new tests ran once with the race detector. They cover exact field spelling,
escaped duplicate names, opaque nested records, byte encoding, typed containers,
nulls, large integer spelling, input bounds and trailing documents. The large
synthetic receipt has4096 rows and6,643,722 bytes. It uses two struct shape tables
and decoded in479.635458 ms in that one invocation. The complete new test package
reported2.308 seconds, including the race detector's exit delay.

[The original test output](wire73143/stdout.log), input source archive and manifest
are preserved. The timing describes this synthetic input on the local machine.
The earlier CLI/native/model/source cohorts were not locally rerun. The next
complete original CI family must establish whether both package timeouts are
resolved while all Gooo language features and counterexamples remain.
