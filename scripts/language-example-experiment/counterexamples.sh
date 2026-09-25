observe_counterexamples() {
	local code drift

	jq '.decision="UNKNOWN"' "$output/first.json" > "$output/unknown-top.json"
	set +e
	reduce "$golden" "$output/unknown-top.json" "$output/replay.json" "$output/profile.json" "$output/unknown-report.json"
	code=$?
	set -e
	[[ "$code" == "1" ]]
	jq -e '.decision=="FAIL_CLOSED" and .resolution=="LOWER_RESOLUTION" and .reason=="ARTIFACT_DECISION_UNKNOWN"' "$output/unknown-report.json"

	jq '.decision="FAIL_CLOSED"' "$output/first.json" > "$output/known-failure.json"
	set +e
	reduce "$golden" "$output/known-failure.json" "$output/replay.json" "$output/profile.json" "$output/known-failure-report.json"
	code=$?
	set -e
	[[ "$code" == "1" ]]
	jq -e '.resolution=="EXACT" and .reason=="ARTIFACT_DECISION_REJECTED"' "$output/known-failure-report.json"

	jq '.operation.activity="Other"' "$output/first.json" > "$output/tampered-artifact.json"
	set +e
	reduce "$golden" "$output/tampered-artifact.json" "$output/replay.json" "$output/profile.json" "$output/tampered-report.json"
	code=$?
	set -e
	[[ "$code" == "1" ]]
	jq -e '.resolution=="EXACT" and .reason=="ARTIFACT_DIGEST_INVALID"' "$output/tampered-report.json"

	drift="$output/replay-drift"
	mkdir -p "$drift"
	cp "$example/entities.gooo" "$drift/entities.gooo"
	sed 's/PayOrder/CancelOrder/' "$example/activity.gooo" > "$drift/activity.gooo"
	"$output/gooo" emit --kind operation-manifest --entry CancelOrder "$drift" > "$output/divergent-replay.json"
	set +e
	reduce "$golden" "$output/first.json" "$output/divergent-replay.json" "$output/profile.json" "$output/replay-mismatch-report.json"
	code=$?
	set -e
	[[ "$code" == "1" ]]
	jq -e '.resolution=="EXACT" and .reason=="ARTIFACT_REPLAY_MISMATCH"' "$output/replay-mismatch-report.json"

	jq '.samples[0].rss_kib=-1' "$output/profile.json" > "$output/invalid-profile.json"
	set +e
	reduce "$golden" "$output/first.json" "$output/replay.json" "$output/invalid-profile.json" "$output/invalid-profile-report.json"
	code=$?
	set -e
	[[ "$code" == "1" ]]
	jq -e '.resolution=="EXACT" and .reason=="PROFILE_SAMPLE_INVALID"' "$output/invalid-profile-report.json"

	jq '.operation.activity="Other"' "$golden" > "$output/drift-golden.json"
	set +e
	reduce "$output/drift-golden.json" "$output/first.json" "$output/replay.json" "$output/profile.json" "$output/mismatch-report.json"
	code=$?
	set -e
	[[ "$code" == "1" ]]
	jq -e '.resolution=="EXACT" and .reason=="ARTIFACT_GOLDEN_MISMATCH"' "$output/mismatch-report.json"

	jq -n '{schema:"gooo/language-example-counterexamples/v1",satisfied:6,total:6}' > "$output/counterexamples.json"
}
