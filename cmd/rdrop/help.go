package main

var helpTopics = map[string]string{
	"doctor": `
Usage:
  rdrop doctor

Checks that RAINDROP_TOKEN is configured and accepted by the Raindrop API.
`,
	"me": `
Usage:
  rdrop me [--json]
  rdrop user update '{"fullName":"Name"}'

Reads or updates authenticated user details.
`,
	"user": `
Usage:
  rdrop user update '{"fullName":"Name"}'

Updates authenticated user settings through Raindrop's user endpoint.
`,
	"stats": `
Usage:
  rdrop stats [--json]

Shows system collection counts and account metadata.
`,
	"collections": `
Usage:
  rdrop collections [--children] [--json]

Lists root collections. Use --children for nested collections.
`,
	"collection": `
Usage:
  rdrop collection create [--parent id] <title>
  rdrop collection update <id> '{"title":"New name"}'
  rdrop collection delete <id>
  rdrop collection sort title|-title|-count
  rdrop collection expand true|false
  rdrop collection merge --to id --ids id,id
  rdrop collection clean
  rdrop collection empty-trash

Manages collections. Deleting a collection moves contained raindrops to Trash.
`,
	"covers": `
Usage:
  rdrop covers [search-text]

Lists featured collection cover icons, or searches cover icons by text.
`,
	"cover": `
Usage:
  rdrop cover raindrop <raindrop-id> <image-path>
  rdrop cover collection <collection-id> <image-path>

Uploads PNG, GIF, or JPEG cover images.
`,
	"export": `
Usage:
  rdrop export --format csv|html|zip [--collection id] [--search query] [--sort -created] > bookmarks.csv

Exports raindrops to stdout so normal shell redirection and pipes work.
`,
	"import": `
Usage:
  rdrop import parse-url <url>
  rdrop import exists <url,url>
  rdrop import file <bookmarks.html>

Parses URLs, checks whether URLs are saved, or parses a browser/Pocket/Instapaper bookmark HTML file.
`,
	"upload": `
Usage:
  rdrop upload file [--collection id] <path>

Uploads a local file as a raindrop.
`,
	"tags": `
Usage:
  rdrop tags [--collection id] [--json]

Lists tags and counts. Default output is tab-separated.
`,
	"tag": `
Usage:
  rdrop tag rename [--collection id] <old> <new>
  rdrop tag merge [--collection id] <from,from2> <to>
  rdrop tag delete [--collection id] <tag,tag2>

Renames, merges, or deletes tags, optionally inside one collection.
`,
	"filters": `
Usage:
  rdrop filters [--collection id] [--search query] [--json]

Shows counts for broken links, duplicates, untagged items, favorites, tags, and types.
`,
	"cleanup": `
Usage:
  rdrop cleanup report [--collection id] [--search query] [--json]

Builds a safe cleanup report without mutating anything. The report highlights broken links, duplicates, untagged items, and tag normalization candidates.
`,
	"list": `
Usage:
  rdrop list [--collection id] [--search query] [--page n] [--per-page n] [--sort -created] [--nested] [--json]
  rdrop search [list flags] <query>
  rdrop untagged [list flags]
  rdrop duplicates [list flags]
  rdrop broken [list flags]

Lists raindrops. Human output is tab-separated; --json is best for agents and jq.
`,
	"search": `
Usage:
  rdrop search [--collection id] [--page n] [--per-page n] [--sort score] [--nested] [--json] <query>

Searches raindrops using Raindrop search syntax.
`,
	"get": `
Usage:
  rdrop get [--json] <id>

Gets one raindrop by ID.
`,
	"add": `
Usage:
  rdrop add [--title text] [--tags a,b] [--collection id] [--note text] [--excerpt text] [--parse] <url>

Creates a link raindrop. Flags may appear before or after the URL.
`,
	"update": `
Usage:
  rdrop update <id> '{"tags":["go"],"important":true}'

Updates one raindrop with a JSON object.
`,
	"delete": `
Usage:
  rdrop delete <id>

Moves one raindrop to Trash. Deleting from Trash can remove permanently.
`,
	"batch": `
Usage:
  rdrop batch tag --ids 1,2 --tags go,docs [--collection id]
  rdrop batch move --ids 1,2 --to collectionID [--collection id]
  rdrop batch delete --ids 1,2 [--collection id]
  rdrop batch delete --search "notag:true" [--collection id]

Bulk operations. Prefer targeted IDs or a carefully tested search.
`,
	"highlights": `
Usage:
  rdrop highlights [--collection id] [--page n] [--per-page n]

Lists highlights globally or for one collection.
`,
	"highlight": `
Usage:
  rdrop highlight add <raindrop-id> --text text [--color yellow] [--note text]
  rdrop highlight update <raindrop-id> <highlight-id> [--text text] [--color yellow] [--note text]
  rdrop highlight delete <raindrop-id> <highlight-id>

Adds, updates, or removes highlights on a raindrop.
`,
	"cache": `
Usage:
  rdrop cache <raindrop-id>

Writes the permanent-copy response to stdout. Permanent copy requires Raindrop Pro.
`,
	"sharing": `
Usage:
  rdrop sharing list <collection-id>
  rdrop sharing invite <collection-id> --emails a@b.com,c@d.com [--role viewer|member]
  rdrop sharing role <collection-id> <user-id> --role viewer|member
  rdrop sharing remove <collection-id> <user-id>
  rdrop sharing unshare <collection-id>
  rdrop sharing join <collection-id> --token token

Manages collection sharing and collaborators.
`,
	"backups": `
Usage:
  rdrop backups
  rdrop backup create
  rdrop backup download <backup-id> --format csv|html > backup.csv

Lists, creates, or downloads Raindrop backups.
`,
	"backup": `
Usage:
  rdrop backup create
  rdrop backup download <backup-id> --format csv|html > backup.csv

Creates or downloads backups.
`,
	"suggest": `
Usage:
  rdrop suggest <url> [--json]

Asks Raindrop to suggest collections and tags for a URL.
`,
	"raw": `
Usage:
  rdrop raw METHOD PATH [json-body]

Calls a JSON REST endpoint directly. Use first-class commands for multipart upload endpoints.
`,
}
