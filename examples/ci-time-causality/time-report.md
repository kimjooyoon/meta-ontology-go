# CI time causality report

decision: REFUTED (controlled corpus contains five expected refutations)
contract: ci-time-causality-v1; denominator_cells: 12; released_gooo_activities: 12; one_to_one: true
cases: total=12 CLOSED=3 UNKNOWN=4 REFUTED=5; precedence=REFUTED>UNKNOWN>CLOSED
runtime: wall_ms=1 peak_rss_kib=7436 measured_by=go-time-and-getrusage-rusage-self
input_inventory: dirs=10 files=14 physical_files=14 physical_bytes=63766 Go_files=7 Go_lines=1221 Gooo_files=1 Gooo_lines=15 root_README_included=false
authority: verification=github-actions Go=1.27 repository_writes=0 local_test_executions=0 cross_project_required_gates=0
duration rule: start and end are derived only from the same operation identity and clock domain; negative duration is REFUTED_CLOCK_ORDER; clamp-to-zero is forbidden.
aggregation: source CI and OpenTofu are separate observations; no cross-run, cross-job, or cross-provider subtraction is performed.

## Case results

- closed-same-operation-positive: CLOSED / DURATION_DERIVED / duration_ms=2500
- closed-same-operation-zero: CLOSED / DURATION_DERIVED / duration_ms=0
- closed-separate-provider-observations: CLOSED / SEPARATE_OBSERVATIONS_NOT_AGGREGATED
- unknown-missing-start: UNKNOWN / MISSING_START_TIME / unknown(stage=DURATION,step=READ_START_TIME,reason=MISSING_START_TIME,unknown_class=DIRECT_MISSING,next_operation=RESTORE_START_TIME,blocked_by=start-time)
- unknown-missing-end: UNKNOWN / MISSING_END_TIME / unknown(stage=DURATION,step=READ_END_TIME,reason=MISSING_END_TIME,unknown_class=DIRECT_MISSING,next_operation=RESTORE_END_TIME,blocked_by=end-time)
- unknown-clock-domain: UNKNOWN / CLOCK_DOMAIN_UNKNOWN / unknown(stage=DURATION,step=VALIDATE_CLOCK_DOMAIN,reason=CLOCK_DOMAIN_UNKNOWN,unknown_class=DIRECT_UNKNOWN,next_operation=DECLARE_CLOCK_DOMAIN,blocked_by=clock-domain)
- unknown-missing-artifact: UNKNOWN / ARTIFACT_MISSING / unknown(stage=EVIDENCE,step=READ_ARTIFACT,reason=ARTIFACT_MISSING,unknown_class=DIRECT_MISSING,next_operation=RESTORE_ARTIFACT,blocked_by=artifact)
- refuted-negative-duration: REFUTED / REFUTED_CLOCK_ORDER
- refuted-cross-run-subtraction: REFUTED / CROSS_RUN_JOB_PROVIDER_SUBTRACTION_FORBIDDEN
- refuted-operation-id-mismatch: REFUTED / OPERATION_ID_MISMATCH
- refuted-malformed-timezone: REFUTED / MALFORMED_TIMEZONE
- refuted-clamp-to-zero-attempt: REFUTED / CLAMP_TO_ZERO_FORBIDDEN

## Replay and output digests

replay: deterministic=true replay_count=2 first=sha256:633e1668558e97d99827399bcc9f70569f7cdb3ca436aed228cd408e4a7d02d5 second=sha256:633e1668558e97d99827399bcc9f70569f7cdb3ca436aed228cd408e4a7d02d5
time-manifest.json: sha256:67ec377f670f925dbb346bd83eee233ab53a01b05be4b56368517e2fef7554a3
operations.ndjson: sha256:c5e704fa5542a69f9e950982f4fdd492e469c900d3ace3ce0408e05113aa5707 bytes=8742
clock-domains.json: sha256:4d858df4758881a329b6e4c52d76c2d4d2e54f114419b873ba90659fc5c9988b bytes=770
duration-receipt.json: sha256:58b2236d930da8f6fd8724f766b3ca0a26a4d90de51b163194ecc8350d765a67 bytes=9519
replay-receipt.json: sha256:ee2b19165fb9d56d2ddbd3f96416ed80c4642508321cf24db36fe2b510733a71 bytes=505
time-report.md: digest is intentionally resolved by the GitHub Actions artifact API after upload (self-digest is not embedded).
retry_lineage: ci-effort attempts=1,2 append_only=true; both attempts are retained in operations.ndjson.

No score, average, percentage, or generalized speed claim is emitted.
