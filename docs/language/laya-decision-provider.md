# Laya typed decisions in Gooo

`gooo decide` sends a Gooo typed-choice request to a local Laya server when
`GOOO_LAYA_URL` points to its Jev-compatible `/v1/systemone` endpoint. Without
that setting, or when the service cannot answer, it selects the request's
declared `fallback` exactly. The JSON receipt records whether the result came
from Laya or the deterministic fallback, hashes the canonical request, and
reads the loaded checkpoint revision from Laya's local `/health` endpoint when
available.

The model receives only the request state and the declared question/options.
Its output is an advisory classification. It does not modify source, approve a
change, select CI proof, or grant repository-write authority. Keep policy and
verification decisions in Gooo's deterministic execution and proof paths.

## Install and run Laya locally

Laya requires Python 3.10 or newer. Install the server extra and start it on
loopback; the first request downloads the selected checkpoint. Use the
multilingual checkpoint for Korean inputs:

```sh
python3 -m venv .venv-laya
. .venv-laya/bin/activate
python -m pip install 'laya[serve]==0.3.21'
LAYA_HOST=127.0.0.1 LAYA_PORT=8787 LAYA_MODELS=multilingual \
  LAYA_DEVICE=cpu LAYA_PRELOAD=0 laya-serve
```

In another shell, point Gooo at the local API and execute the sample choice:

```sh
GOOO_LAYA_URL=http://127.0.0.1:8787/v1/systemone \
  go run ./cmd/gooo decide --json examples/laya-decision/request.json
```

To run without Laya, omit `GOOO_LAYA_URL`. To protect the local service with
`LAYA_API_KEY`, set the same value in `GOOO_LAYA_API_KEY` for Gooo.

## Request and receipt

The request declares the candidate IDs and descriptions, one choice question,
and a fallback ID that must be among those candidates. This makes disconnected
execution stable for the same request. The receipt includes the selected ID,
execution mode, fallback reason when applicable, provider/model/revision metadata,
probabilities and both confidence fields when returned, plus a request digest.
Do not treat model confidence as proof of correctness: evaluate the
classifications against Gooo's own accepted outcomes before making a model part
of a policy decision.

See [the local CPU evaluation](laya-local-evaluation-2026-09-29.md) for a
measured warm-request latency and process CPU/RSS baseline.
