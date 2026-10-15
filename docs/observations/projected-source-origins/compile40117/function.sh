run_projected_compile() {
  local stdout="$projection_work/projected-compile.stdout"
  local stderr="$projection_work/projected-compile.stderr"
  local status=0
  if go test ./... > "$stdout" 2> "$stderr"; then
    status=0
  else
    status=$?
  fi
  local stdout_bytes
  local stderr_bytes
  local stdout_digest
  local stderr_digest
  stdout_bytes="$(wc -c < "$stdout" | tr -d ' ')"
  stderr_bytes="$(wc -c < "$stderr" | tr -d ' ')"
  stdout_digest="sha256:$(sha256sum "$stdout" | cut -d' ' -f1)"
  stderr_digest="sha256:$(sha256sum "$stderr" | cut -d' ' -f1)"
  jq -n \
    --arg cwd "$repo_root" \
    --arg stdout_digest "$stdout_digest" \
    --arg stderr_digest "$stderr_digest" \
    --argjson exit_code "$status" \
    --argjson stdout_bytes "$stdout_bytes" \
    --argjson stderr_bytes "$stderr_bytes" \
    '{schema:"gooo.projected-compile-observation/v1", command:["go","test","./..."], cwd:$cwd,
      exit_code:$exit_code, stdout_bytes:$stdout_bytes, stdout_digest:$stdout_digest,
      stderr_bytes:$stderr_bytes, stderr_digest:$stderr_digest, observed:($exit_code == 0)}' \
    > "$projection_work/projected-compile.json"
  if [[ "$status" -ne 0 ]]; then
    cat "$projection_work/projected-compile.json" >&2
    cat "$stdout"
    cat "$stderr" >&2
  fi
  return "$status"
}
