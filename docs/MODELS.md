# Models and provider contracts

Use an external agent over MCP when it already owns the model loop. Use `rdrop agent` when you want this executable to run that loop. Use `classify` for small fixed-choice decisions. Use context/proposal files when your model or rule engine has a different interface. These paths share the library vocabulary and reviewed-plan boundary.

## Direct tool-calling loop

```sh
rdrop agent --model your-model --json 'Review and propose organization.'
rdrop agent --base-url http://127.0.0.1:1234/v1 --api chat --model local-model --json 'Inspect my tags.'
```

| Setting | Behavior |
| --- | --- |
| `--model`, `RDROP_MODEL` | Required tool-capable model ID |
| `--base-url`, `RDROP_MODEL_BASE_URL` | API root; default `https://api.openai.com/v1` |
| `RDROP_MODEL_API_KEY` | Bearer credential; `OPENAI_API_KEY` fallback |
| `--api responses` | Default: Responses tools and function-call outputs; `store:false`; replays encrypted reasoning output |
| `--api chat` | Chat Completions function tools and tool messages for compatible gateways |
| `--allow-write` | Authorizes model-selected mutations; otherwise only reads and planning tools |
| `--max-steps`, `--max-tool-calls` | Default 12 model requests and 32 total calls |
| `--max-output-tokens` | Default 4096 per model request |
| `--timeout` | Default 10 minutes overall; individual HTTP request limit 120 seconds |
| `--json` | Result, status, step/call counts, independent tool results and exact generated plans |

The loop reads a cleanup report and bounded context first. It validates every tool name, input schema and call ID; checks a whole model batch before executing calls; runs calls sequentially; and returns tool failures to the model for correction. Unknown tools, malformed calls, incomplete provider output and exhausted limits stop execution. Oversized tool output is replaced by a bounded error; exact generated plans are retained locally in the JSON result.

The caller chooses the right wire protocol and a model that supports it. Not every OpenAI-compatible service implements Responses, tool calls or reasoning replay. OpenAI's current tool-calling models may require Responses; consult the [official function-calling guide](https://developers.openai.com/api/docs/guides/function-calling). This client is a bounded synchronous loop, without streaming, persisted sessions or unattended scheduling.

`--allow-write` is explicit authorization to expose account, sharing and destructive tools to the model. For a review-first workflow, leave it off, inspect `.plans`, and run `cleanup apply` yourself. A model-generated `confirm:true` is not evidence of a person approving that particular call.

## Typed decision providers

`classify` sends ID, title, link, excerpt and existing tags for 1..25 exact bookmarks. It excludes private notes. Taxonomy is a JSON object of category names and descriptions; `needs_review` is reserved as an abstention category.

```json
{"reference":"Technical documentation", "read-later":"Long-form reading"}
```

The shared System One request is `{model,state,questions}`. Each bookmark gets a `choice` question keyed by its ID; criteria contain your category descriptions plus `needs_review`. HTTP providers receive `POST <base-url>/systemone`. Command providers receive the same JSON on stdin and must return exactly one JSON response on stdout:

```json
{"model":"local-model","answers":{"101":{"type":"choice","choice":"reference","confidence":0.9,"probabilities":{"reference":0.85,"read-later":0.1,"needs_review":0.05},"abstention":false}}}
```

Answers must cover every question exactly, use known labels, supply finite confidence and a full probability distribution between 0 and 1 summing to approximately 1 (tolerance 0.02). `needsReview` is true for abstention, `needs_review`, confidence below `--min-confidence` or chosen probability below `--min-probability`; both thresholds default to 0.8. Provider scores and calibration differ. Laya's `confidence` is its own score; the optional `x_jev_confidence` is not used by this client. Evaluate reviewed examples before trusting a threshold. Results never apply tags by themselves.

### Jev

The default provider is [TypeSafe System One](https://docs.typesafe.ai/introduction/quickstart), API root `https://api.typesafe.ai/v1`, model `jev-latest`. Set `TYPESAFE_API_KEY`; optional `TYPESAFE_BASE_URL` and `TYPESAFE_MODEL` remain supported. [Jev introduction](https://typesafe.ai/blog/introducing-system-one-models-and-jev).

### Local Laya over HTTP

[Laya](https://github.com/NandhaKishorM/laya) provides a System One-compatible service. Install and manage its runtime separately, selecting the checkpoint and hardware appropriate to your data:

```sh
python -m venv .venv
. .venv/bin/activate
python -m pip install 'laya[serve]'
LAYA_HOST=127.0.0.1 LAYA_PORT=8000 LAYA_MODELS=english laya-serve
```

In another shell:

```sh
RDROP_DECISION_PROVIDER=systemone \
RDROP_DECISION_BASE_URL=http://127.0.0.1:8000/v1 \
RDROP_DECISION_MODEL=english \
rdrop classify --ids 101,102 --taxonomy taxonomy.json --json
```

Check the chosen Laya version's serving options and port. Initial use can download checkpoints. For offline use, configure local checkpoints using Laya's [local setup guide](https://github.com/NandhaKishorM/laya/blob/main/LOCAL_SETUP.md). Upstream benchmark results do not establish accuracy or latency for your bookmark taxonomy.

### Local command adapter

The supplied Python adapter calls `Router.predict` and normalizes typed answers. Install Laya in your own virtual environment, then point to that Python executable and the adapter's absolute path:

```sh
export RDROP_DECISION_PROVIDER=command
export RDROP_DECISION_COMMAND="$PWD/.venv/bin/python"
export RDROP_DECISION_ARGS='["/absolute/path/raindrop-cli/integrations/laya/classify.py"]'
export RDROP_DECISION_MODEL=english
rdrop classify --ids 101,102 --taxonomy taxonomy.json --json
```

Any trusted executable implementing the same contract can replace it, including another local inference runtime. Arguments are a JSON array, executed directly without a shell. Models and MCP tool callers cannot pick the executable. Standard API key variables and `RAINDROP_TOKEN` are removed from its inherited environment; other environment variables and normal local filesystem permissions remain available. Treat the configured process as trusted local code. Timeout is 120 seconds, stdout limit 8 MiB, stderr limit 64 KiB; diagnostics never become JSON output.

### Other HTTP providers

Configure `RDROP_DECISION_BASE_URL`, `RDROP_DECISION_API_KEY` and `RDROP_DECISION_MODEL`. Those take precedence over `TYPESAFE_*`. Remote endpoints require HTTPS and a key; loopback HTTP may run without one. Services with another schema need a small adapter that normalizes their response. Fixed-choice classification chooses one category; multi-tag/move suggestions use the [provider-neutral proposal contract](ORGANIZING.md#provider-neutral-proposals).

## Data and evidence

Direct models receive library counts, tag names, collection paths and selected bookmark metadata. Additional read tools, such as `get`, highlights and snapshot planning, can expose private content when the model calls them. Full apply plans returned to the model include private notes in their snapshots. Classification omits notes; MCP context keeps its raw snapshots local. Saved full contexts and plans contain private metadata and must be stored accordingly.

Fixture tests verify transport, schema rejection, tool round trips, budgets, abstention and planning. They do not evaluate model quality or verify a live service/account. A useful evaluation set contains real reviewed tags/moves, ambiguous items that should abstain, existing spelling variants, long excerpts and held-out examples. Compare each provider's mistakes and coverage before changing thresholds or enabling writes.
