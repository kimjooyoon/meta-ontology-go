#!/usr/bin/env bash
set -euo pipefail

if [[ "$#" -ne 6 ]]; then
  echo "usage: select-eligibility.sh exact-sha platform-result aggregate-result fallback downloaded output" >&2
  exit 2
fi

exact_sha="$1"
platform_result="$2"
aggregate_result="$3"
fallback="$4"
downloaded="$5"
output="$6"
artifact_ref="gooo-release-eligibility-$exact_sha/eligibility.json"

mkdir -p "$(dirname "$fallback")" "$(dirname "$output")"
if [[ "$platform_result" == "success" && "$aggregate_result" == "success" && -f "$downloaded" ]]; then
  cp "$downloaded" "$output"
  exit 0
fi

if [[ "$platform_result" != "success" || "$aggregate_result" != "success" ]]; then
  reason="RELEASE_READINESS_UPSTREAM_NOT_SUCCESS"
  unknown_class="UPSTREAM_JOB_NON_SUCCESS"
  blocked_by=$(jq -nc     --arg platform "$platform_result"     --arg aggregate "$aggregate_result" '
      [
        (if $platform != "success" then "platform:" + $platform else empty end),
        (if $aggregate != "success" then "aggregate:" + $aggregate else empty end)
      ]')
else
  reason="RELEASE_READINESS_ELIGIBILITY_EVIDENCE_UNAVAILABLE"
  unknown_class="ELIGIBILITY_ARTIFACT_UNAVAILABLE"
  blocked_by=$(jq -nc --arg artifact "$artifact_ref" '[$artifact]')
fi

jq -n   --arg sha "$exact_sha"   --arg platform "$platform_result"   --arg aggregate "$aggregate_result"   --arg reason "$reason"   --arg unknown_class "$unknown_class"   --arg next_operation "RERUN_GOOO_RELEASE_READINESS"   --argjson blocked_by "$blocked_by"   '{
    schema:"gooo/release-eligibility/v1",
    head_sha:$sha,
    decision:"FAIL_CLOSED",
    resolution:"OPERATION_CLASS",
    reason:$reason,
    next_operation:$next_operation,
    mutation_allowed:false,
    upstream:{platform:$platform,aggregate:$aggregate},
    claim:{
      id:"release://gooo/experimental-candidate-evidence",
      entity:"GoooExperimentalReleaseCandidateEvidence",
      status:"ACTIVE",
      state:"UNKNOWN",
      resolution:"OPERATION_CLASS",
      stage:"CI",
      step:"AGGREGATE_RELEASE_EVIDENCE",
      reason:$reason,
      unknown_class:$unknown_class,
      next_operation:$next_operation,
      blocked_by:$blocked_by
    },
    summary:{total_work:7,closed:0,unknown:7,refuted:0,repository_writes:null}
  }' > "$fallback"
cp "$fallback" "$output"
