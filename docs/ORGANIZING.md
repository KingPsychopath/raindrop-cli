# Organizing with existing library context

The user's tags and collection paths are the starting vocabulary. A model proposes changes; `rdrop` validates names, scope and snapshots; the caller reviews the plan and applies it intentionally. Neither a valid schema nor a high confidence score proves a good categorization.

## First inspect

```sh
rdrop cleanup report --json > cleanup-report.json
rdrop collections --tree --json
rdrop tags --json
rdrop untagged --collection=-1 --per-page 25 --json
```

Reports show Raindrop's counts and candidate tag normalization commands. Search-scoped reports omit collection-wide tag normalization commands. Do not automatically execute commands from a report; decide the intended naming rules first.

## Context and naming rules

```sh
rdrop context --ids 101,102 --policy policy.json > library-context.json
jq 'del(.snapshots)' library-context.json > model-context.json
```

Optional `policy.json`:

```json
{
  "instructions": "Preserve my case and spelling. Multiple tags are welcome. Use collections for projects, tags for subjects.",
  "tagRules": {"reference": "Reusable technical documentation"},
  "collectionRules": {"42": "Resources for the active website project"},
  "allowNewTags": false
}
```

A tag rule may describe an existing name. To introduce a name deliberately, declare it in `tagRules` and set `allowNewTags:true` before exporting context. Models cannot introduce arbitrary new names. Case, spacing and spelling remain exact; aliases and tag merges require a separate reviewed decision.

The context has version, account ID, timestamp, scope, policy, global tags, collection paths, selected bookmark metadata, snapshots and a SHA-256 digest. Default scope is one page of 25, at most 50. Exact IDs cannot be mixed with search/page/collection filters. Collection paths include nested parents and Unsorted; moves into Trash are excluded from proposals.

Full snapshots contain private notes. Keep the original local file for later plan construction; give the model only `del(.snapshots)` or a `--model-view` export. A separate export creates a separate digest, so preserve the original model view and digest when planning. MCP keeps full snapshots in a cache of 16 contexts per process and returns only the model view from `raindrop_context`. A later planning tool returns a full apply plan, including snapshots and private notes; that tool result is visible to the connected model/host. Restarting or evicting that cache requires a fresh context.

## Provider-neutral proposals

Any responsive HTTP API, local process, rules engine or tool-calling model can return the [versioned proposal schema](schemas/proposal.schema.json):

```json
{
  "contextDigest": "sha256:copy-the-exact-context-digest",
  "suggestions": [
    {"id":101,"tags":["reference","go"],"collectionId":42,"reason":"Go reference for this project","needsReview":false},
    {"id":102,"tags":["read-later"],"reason":"Metadata is ambiguous","needsReview":true}
  ]
}
```

The example names and destination must actually exist in that context or its approved naming policy. `tags` append; omitted `collectionId` leaves the collection alone. `needsReview` and `reason` are required. No delete, tag replacement, collection creation or tag merge is inferred through this contract.

```sh
rdrop organize plan --context library-context.json --proposal proposal.json > organization.json
jq '.plan // error("No accepted changes")' organization.json > reviewed-plan.json
```

All suggestions are validated before returning any plan. Unknown or duplicate IDs, tag names and destinations reject the proposal. Entries marked `needsReview:true` are withheld. Unchanged and unmentioned IDs are reported explicitly. Missing suggestions do not imply that a model inspected the rest of the library. Review both the plan and withheld entries; edit the proposal and regenerate rather than editing a digested plan.

For MCP, call `raindrop_context`, then `raindrop_organize_plan` with that digest. For direct models, `agent --json` captures exact plans under `.plans`; do not reconstruct an apply file from the model's narrative response.

## Apply and recover

```sh
rdrop cleanup apply reviewed-plan.json --dry-run
rdrop cleanup apply reviewed-plan.json --yes
```

Apply validates the file version, digest, account and every snapshot before the first write. Tags and a move on the same bookmark are one update. Existing tags and unrelated fields remain intact. Manual tag/move/delete plans are also available through `cleanup plan` and ID-based batch `--dry-run`.

There is no Raindrop transaction or conditional version update. A concurrent edit after checking can race a write. Failure after some writes returns `status:partial`, applied IDs, remaining IDs and an `uncertainId`: the failed request may have reached the server. Inspect affected bookmarks, export a fresh context and plan only remaining work. The CLI does not retry mutations automatically. A replay commonly fails stale checks, but digest validation is not a substitute for inspecting uncertain outcomes.

A digest detects accidental or unreviewed changes to a saved file; it is not an authorization signature. A caller who can construct a new file can construct its digest. Host permissions and deliberate review remain necessary.

## Direct bulk commands

```sh
rdrop batch tag --collection=-1 --ids 101,102 --tags reference --dry-run > reviewed-plan.json
rdrop cleanup plan --action move --collection=-1 --ids 101,102 --to 42 > reviewed-plan.json
rdrop cleanup plan --action delete --collection=42 --ids 101,102 > reviewed-plan.json
```

Without `--dry-run`, existing direct batch tag/move commands immediately update the explicit source collection. Search-based batch deletion is a dynamic query; its preview is a count and apply command, not a snapshot plan. Matches can change before execution. Prefer exact IDs and saved plans for agent organizing. Deleting from Trash and emptying Trash may be permanent.

TSV sanitizes terminal control characters; JSON preserves original metadata. Treat titles, excerpts, links, notes, API responses and model output as data, never shell commands or policy instructions.
