# Real model and Executor verification

The 2026-10-10 check used macOS arm64, Python 3.12.13, Ollama 0.40.2, `qwen3:4b` (`359d7dd4bcda`), and Laya 0.4.2 from Git commit `68804629e8ccd9d616d48a40e87de9fabbeae069`. Laya's English checkpoint revision was `7b928d828b7b0e022f929d9bd2e44165aa270148`, running on MPS. The pinned Python environment is [requirements-live.txt](../integrations/laya/requirements-live.txt). These versions were tested on this host; hardware and platform support still depend on the upstream runtimes.

## Start the actual local models

From the checkout:

```sh
make build
uv venv --python 3.12 .venv-models
uv pip install --python .venv-models/bin/python -r integrations/laya/requirements-live.txt

# Terminal 1: the first start downloads the pinned checkpoint.
HF_HOME="$PWD/.artifacts/hf" \
LAYA_REVISION=7b928d828b7b0e022f929d9bd2e44165aa270148 \
LAYA_HOST=127.0.0.1 LAYA_PORT=8000 LAYA_MODELS=english \
.venv-models/bin/laya-serve
```

`LAYA_DEVICE=cpu` selects CPU explicitly. On the tested Apple Silicon host, the runtime selected MPS. Verify the actual device and revision with `curl --fail http://127.0.0.1:8000/health`. Checkpoints are downloaded model data; no Raindrop credential is required by the local model server.

Install [Ollama](https://docs.ollama.com/quickstart) from its official distribution or Homebrew. Then:

```sh
brew install ollama
# Terminal 2:
OLLAMA_HOST=127.0.0.1:11434 OLLAMA_CONTEXT_LENGTH=32768 ollama serve

# Terminal 3:
ollama pull qwen3:4b
ollama list
```

The pull downloads approximately 2.5 GB. Keep both servers running while verifying; the script does not install or start them. Ollama's [Chat Completions compatibility](https://docs.ollama.com/api/openai-compatibility) carries the model tool calls; Laya uses the System One decision contract.

## Run the whole loop without a real library

```sh
HF_HOME="$PWD/.artifacts/hf" \
LAYA_REVISION=7b928d828b7b0e022f929d9bd2e44165aa270148 \
python3 scripts/verify-models.py --command-python .venv-models/bin/python
```

This starts a disposable Raindrop REST fixture on an ephemeral loopback port and supplies a fixture token. It overrides Raindrop configuration for every subprocess. It checks real Laya over HTTP and through the local Python command adapter, then runs the real Qwen model through `rdrop agent`. The verified run completed in three model steps with two successful tool calls: classification and organization planning. The model received their actual results and returned a response. The captured snapshot plan passes dry-run, then applies only to the fixture. Existing tags and the private note survive; the tag and collection change use one update. The read-only model phase performs zero writes.

Artifacts go to the ignored `.artifacts/e2e/models` directory: actual tool results, exact plan, dry-run/apply results, fixture requests and `summary.json`. Repeat with another tool-capable model using `--model MODEL --model-url http://127.0.0.1:PORT/v1`. The check requires successful calls to `raindrop_classify` and `raindrop_organize_plan`; a plausible prose response alone fails it. A different model can fail this task despite implementing the transport.

Real inference exposed interference when two unrelated bookmarks shared one classification state. Requests now isolate each bookmark. After the fix, the API reference received `reference` (probability 0.9679, confidence 0.8546); the recipe received `cooking` (probability 0.9319, confidence 0.7306) and stayed marked for review at the default 0.8 confidence threshold. This is a two-example smoke check, not a general accuracy evaluation. Laya warns that some checkpoint temperatures for 11+ choices require clamping; evaluate your own taxonomy, especially larger category sets.

Jev was not contacted because no key was available. Its shared wire contract, malformed responses and uncertainty handling are tested with fixtures. With your key in the private environment, use `TYPESAFE_API_KEY` and run:

```sh
rdrop cleanup report --json
rdrop classify --ids REAL_ID --taxonomy taxonomy.json --json
```

Jev is a decision provider. A tool-capable model such as Qwen owns the conversational loop and can call `raindrop_classify`, which delegates to Jev or Laya.

## Use the models with your own library

```sh
rdrop cleanup report --json > cleanup-report.json
RDROP_DECISION_BASE_URL=http://127.0.0.1:8000/v1 \
RDROP_DECISION_MODEL=english \
rdrop agent --base-url http://127.0.0.1:11434/v1 --api chat --model qwen3:4b --json \
  '/no_think Review selected bookmarks, classify them using existing categories, and propose organization. Do not apply changes.' > agent-result.json
jq '.plans[0] // error("No accepted plan")' agent-result.json > reviewed-plan.json
# Inspect the actual tool results, uncertainty and exact plan before these steps.
rdrop cleanup apply reviewed-plan.json --dry-run
rdrop cleanup apply reviewed-plan.json --yes
```

The last command intentionally changes the selected bookmarks. Full contexts and plans can include private metadata; keep these files private. Omit the apply commands when only reviewing.

## Local Executor

The tested local host is `executor@2.0.0-beta.14`, reporting `apps@0.0.1-beta.69`. The published npm beta was checked directly; the website described beta.15 before that package was available. Use the host's actual framework version rather than assuming a website version has been released.

See [the integration guide](INTEGRATIONS.md#executor) for installation, source preparation, creation and deployment. The local wrapper uses `~/.local/share/rdrop-executor/keys.json`; the host keeps that file private and stores connected account credentials encrypted. OS-backed key storage is also supported by Executor; the wrapper is specifically for this file-backed local setup. Never commit that directory or copy keys into MCP configuration committed to Git.

After deployment, create a profile and issue an account connection through Executor's management tools, using the exact paths returned by `tools.search`. The secure form selects the account for that profile. Run the provider health check and read `cleanup.report`, `tags` and `collections` through the deployed profile before using writes. The adapter's two mutation tools still require exact-call host approval; deployment and account connection do not grant blanket mutation approval.

The real local deployment passed the provider account health check and read cleanup counts, tags and collections through its MCP profile. No live bookmark mutation was performed.

For reproducible CLI calls, `python3 scripts/executor-mcp.py` lists MCP tools. Save an execute request as a private JSON file:

```json
{"code":"return await tools.search({namespace:'executor',query:'appProfiles create accountConnect issue accounts check',limit:5});"}
```

```sh
python3 scripts/executor-mcp.py --tool execute --input PRIVATE_REQUEST.json
```

Inspect the returned signatures, then use the same command with the actual management calls. A call can require approval or input: show the question to the user and resume that request; never rerun a paused program or approve on their behalf. This helper uses the local host at `http://127.0.0.1:4312/mcp`; hosted setup has a separate authentication and organization boundary.
