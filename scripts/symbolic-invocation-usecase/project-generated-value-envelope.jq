def activity_valid:
  if (.activity | type) != "string" then false else (.activity | length) > 0 end;

def inputs_valid:
  if (.inputs | type) != "array" then
    false
  else
    (.inputs | length) > 0
    and all(.inputs[]; ((type == "string") and (length > 0)))
  end;

. as $input
| $input.vector.instance as $instance
| (($instance | activity_valid) and ($instance | inputs_valid)) as $valid
| {
    schema: "gooo/symbolic-invocation-value-envelope/v1",
    subject_sha: $input.subject_sha,
    language: "gooo",
    source: {
      artifact_digest: $input.source.artifact_digest,
      external_report_digest: $input.source.external_report_digest,
      validator_digest: $input.source.validator_digest,
      vector_id: $input.vector.id,
      expected: $input.vector.expected,
      proof_choice: $input.vector.proof_choice,
      meta_operation: $input.vector.meta_operation
    },
    decision: (if $valid then "READY" else "FAIL_CLOSED" end),
    resolution: (if $valid then "VALUE_EXACT" else "LOWER_RESOLUTION" end),
    reason: (if $valid
      then "SYMBOLIC_INVOCATION_VALUE_PROJECTED"
      else "SYMBOLIC_INVOCATION_VALUE_INCOMPLETE"
      end),
    invocation: (if $valid
      then {
        activity: $instance.activity,
        inputs: $instance.inputs,
        input_count: ($instance.inputs | length)
      }
      else {
        input_count: (if ($instance.inputs | type) == "array" then ($instance.inputs | length) else 0 end)
      }
      end),
    diagnostics: [
      if ($instance | activity_valid) then empty else "activity-required" end,
      if ($instance | inputs_valid) then empty else "inputs-required" end
    ],
    effects: {
      requested_invocations: (if $valid then 1 else 0 end),
      executed_invocations: 0,
      repository_writes: $input.effects.repository_writes,
      mutation_authority: $input.effects.mutation_authority
    }
  }
