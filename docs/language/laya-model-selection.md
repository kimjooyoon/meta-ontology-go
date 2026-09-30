# Laya model selection

Typed decision requests may select a Laya model with the optional
`provider_model` string:

```json
{
  "schema": "gooo/typed-decision-request/v1",
  "state": "account data",
  "question": {
    "id": "decision",
    "instructions": "Choose a result.",
    "options": [
      {"id": "allow", "description": "Allow the action"},
      {"id": "deny", "description": "Reject the action"}
    ]
  },
  "fallback": "deny",
  "provider_model": "english"
}
```

The accepted explicit values are `english`, `multilingual`, and
`typed-decisions`. An omitted or empty value selects the provider's automatic
default. `decisionroute.ValidateProviderModel` checks a selection before the
plan phase chooses a provider or source body; unsupported values are rejected.

An explicit value is bound into the canonical typed-request SHA-256. Omitted
and empty values are omitted from canonical JSON, preserving the historical
digest. For Laya, the value is sent as the top-level `model` property in the
request JSON. It is not sent as a language hint. Receipts record the selection
as `requested_provider_model`.

When no provider endpoint is configured, resolution does not contact a
provider. It selects the request's declared fallback and records
`NOT_CONFIGURED`, along with any requested model. With an explicit model, a
response is accepted only when `routing.model` exactly matches the request.
A missing or different routing model selects the declared fallback with
`PROVIDER_RESULT_INVALID`.
