def ready_envelope:
  .decision == "READY"
  and .resolution == "VALUE_EXACT"
  and ((.invocation.activity | type) == "string")
  and ((.invocation.activity | length) > 0)
  and ((.invocation.inputs | type) == "array")
  and ((.invocation.inputs | length) == .invocation.input_count);

def fail_closed_envelope:
  .decision == "FAIL_CLOSED"
  and .resolution == "LOWER_RESOLUTION"
  and (.diagnostics | type) == "array"
  and (.diagnostics | length) > 0;

. as $envelope
| ($envelope | ready_envelope) as $ready
| ($envelope | fail_closed_envelope) as $failed
| {
    schema: "gooo/symbolic-invocation-user-observation/v1",
    subject_sha: $envelope.subject_sha,
    language: $envelope.language,
    source: {
      envelope_digest: $envelope.digest,
      vector_id: $envelope.source.vector_id,
      artifact_digest: $envelope.source.artifact_digest
    },
    decision: (if $ready
      then "OBSERVED_READY"
      elif $failed
      then "OBSERVED_FAIL_CLOSED"
      else "FAIL_CLOSED"
      end),
    resolution: (if $ready
      then "USER_VALUE"
      elif $failed
      then "USER_GUARDRAIL"
      else "INVARIANT_ONLY"
      end),
    reason: (if $ready
      then "READY_INVOCATION_OBSERVED"
      elif $failed
      then "INCOMPLETE_INVOCATION_BLOCKED"
      else "ENVELOPE_DECISION_UNKNOWN"
      end),
    value: (if $ready
      then {
        activity: $envelope.invocation.activity,
        inputs: $envelope.invocation.inputs,
        input_count: $envelope.invocation.input_count
      }
      else {
        diagnostics: $envelope.diagnostics,
        input_count: 0
      }
      end),
    effects: {
      observed_envelopes: 1,
      executed_invocations: 0,
      repository_writes: 0,
      mutation_authority: false
    }
  }
