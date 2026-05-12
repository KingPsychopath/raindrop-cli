# Organizing Large Raindrop Libraries

`raindrop-cli` is designed to help organize large bookmark libraries without making hidden destructive changes.

The safest workflow is:

1. Generate a report.
2. Inspect a small candidate set.
3. Write down exact commands.
4. Apply explicit changes.
5. Re-run the report and compare the result.

## Start With A Report

```bash
rdrop cleanup report
rdrop cleanup report --json > cleanup-report.json
```

The report includes:

- counts for broken, duplicate, untagged, and important bookmarks
- commands to inspect each cleanup category
- top tags by count
- tag normalization candidates, such as merging `Machine Learning` into `machine-learning`

The report is read-only.

## Inspect Before Mutating

Use focused list commands before applying a plan:

```bash
rdrop untagged --json | jq '.items[] | {id: ._id, title, link}'
rdrop duplicates --json | jq '.items[] | {id: ._id, title, link}'
rdrop broken --json | jq '.items[] | {id: ._id, title, link}'
```

Limit scope with a collection or search query:

```bash
rdrop cleanup report --collection 123 --json
rdrop untagged --collection 123 --json
rdrop list --search "created:>2025-01-01 notag:true" --json
```

## Apply Explicit Changes

Prefer ID-based commands for bulk mutations:

```bash
rdrop tag merge "Machine Learning,machine learning" machine-learning
rdrop batch tag --ids 101,102,103 --tags needs-review
rdrop batch move --ids 101,102,103 --to 98765
```

Bulk tag and move require `--ids` so a broad search cannot accidentally rewrite a large library.

## Delete Workflow

Plan deletes first:

```bash
rdrop batch delete --ids 101,102 --dry-run
rdrop batch delete --search "notag:true" --dry-run --json
```

Apply only after the plan looks right:

```bash
rdrop batch delete --ids 101,102 --yes
rdrop batch delete --search "notag:true" --yes
```

Emptying Trash is intentionally explicit:

```bash
rdrop collection empty-trash --yes
```

## Suggested Agent Workflow

Agents should follow this sequence:

1. Read `rdrop cleanup report --json`.
2. Fetch small pages of candidate items with `rdrop list`, `rdrop untagged`, `rdrop duplicates`, or `rdrop broken`.
3. Produce a written plan with exact commands.
4. Use `--dry-run` for delete plans.
5. Apply only commands the user approved.
6. Re-run `rdrop cleanup report --json`.
7. Summarize the delta.

## Conservative Cleanup Rules

Good cleanup behavior should stay boring:

- normalize obvious tag variants by trim/lowercase/space-to-hyphen
- identify near-duplicate tags without merging them automatically
- group duplicate links by canonical domain and URL
- move uncertain items to review tags or review collections instead of deleting
- prefer `--ids` over broad search mutations
- keep apply plans reviewable in plain text or JSON

`cleanup report` is the first safe building block for that workflow.
