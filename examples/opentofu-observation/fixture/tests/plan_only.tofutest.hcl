run "plan_only" {
  command = plan

  assert {
    condition     = output.observed_value == "opentofu-observation"
    error_message = "the providerless builtin-only plan did not preserve its output"
  }
}
