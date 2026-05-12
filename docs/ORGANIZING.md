# Organizing and Agent Workflows

`raindrop-cli` should help agents organize large libraries, but it should not make hidden destructive changes.

## Current Safe Workflow

Start with a report:

```bash
rdrop cleanup report --json > cleanup-report.json
```

The report includes:

- counts for broken, duplicate, untagged, and important raindrops
- commands to inspect each cleanup category
- tag normalization candidates, such as merging `Machine Learning` into `machine-learning`

Inspect focused sets before applying anything:

```bash
rdrop untagged --json | jq '.items[] | {id: ._id, title, link}'
rdrop duplicates --json | jq '.items[] | {id: ._id, title, link}'
rdrop broken --json | jq '.items[] | {id: ._id, title, link}'
```

Then apply explicit commands:

```bash
rdrop tag merge "Machine Learning, machine learning" machine-learning
rdrop batch tag --ids 1,2,3 --tags needs-review
rdrop batch move --ids 1,2,3 --to 12345
```

## Agent Design

Agents should follow this sequence:

1. Read `rdrop cleanup report --json`.
2. Fetch small pages of candidate items with `rdrop list`, `rdrop untagged`, `rdrop duplicates`, or `rdrop broken`.
3. Produce a written plan with exact commands.
4. Apply only the commands the user approved.
5. Re-run `rdrop cleanup report --json` and summarize the delta.

## Future Algorithm

A good cleanup algorithm should stay conservative:

- normalize obvious tag variants by trim/lowercase/space-to-hyphen
- identify near-duplicate tags without merging them automatically
- group duplicate links by canonical domain and URL
- move uncertain items to review tags or review collections instead of deleting
- prefer `--ids` over broad search mutations
- support an apply file format that can be reviewed in git before execution

The current `cleanup report` command is the first safe building block for that workflow.
