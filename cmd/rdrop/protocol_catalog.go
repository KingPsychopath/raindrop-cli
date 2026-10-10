package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// One catalog drives MCP schemas, CLI-to-tool conversion and discovery. The
// normal CLI handlers remain the owners of API behavior and validation.
type parameter struct {
	Name        string
	Flag        string
	Type        string
	Required    bool
	Description string
}

type toolSpec struct {
	Command     string      `json:"command"`
	Description string      `json:"description"`
	Write       bool        `json:"write"`
	Destructive bool        `json:"destructive"`
	JSONFlag    bool        `json:"-"`
	Prefix      []string    `json:"-"`
	Parameters  []parameter `json:"-"`
}

func param(name, flag, kind string, required bool) parameter {
	return parameter{name, flag, kind, required, name}
}
func str(name, flag string, required bool) parameter       { return param(name, flag, "string", required) }
func num(name, flag string, required bool) parameter       { return param(name, flag, "integer", required) }
func boolean(name, flag string) parameter                  { return param(name, flag, "boolean", false) }
func csv(name, flag, kind string, required bool) parameter { return param(name, flag, kind, required) }
func body() parameter                                      { return param("body", "", "object", true) }

var scopeParameters = []parameter{num("collection", "collection", false), str("search", "search", false)}
var listParameters = []parameter{num("collection", "collection", false), str("search", "search", false), num("page", "page", false), num("perPage", "per-page", false), str("sort", "sort", false), boolean("nested", "nested")}

func spec(command, description string, write, destructive, jsonFlag bool, parameters ...parameter) toolSpec {
	return toolSpec{Command: command, Description: description, Write: write, Destructive: destructive, JSONFlag: jsonFlag, Parameters: parameters}
}

func toolCatalog() []toolSpec {
	catalog := []toolSpec{
		spec("me", "Read the authenticated account.", false, false, true),
		spec("stats", "Read system collection counts and account metadata.", false, false, true),
		spec("collections", "List root collections, or children when requested.", false, false, true, boolean("children", "children")),
		spec("collections tree", "Read all collections with parent IDs and breadcrumb paths.", false, false, true),
		spec("get", "Read one bookmark.", false, false, true, num("id", "", true)),
		spec("tags", "List tags and counts in a collection (0 means all).", false, false, true, num("collection", "collection", false)),
		spec("filters", "Read cleanup counts, tags and types.", false, false, true, scopeParameters...),
		spec("cleanup report", "Start organizing here. Read library health and review candidates; no changes.", false, false, true, scopeParameters...),
		spec("highlights", "Read highlights with pagination.", false, false, false, num("collection", "collection", false), num("page", "page", false), num("perPage", "per-page", false)),
		spec("suggest", "Ask Raindrop for suggestions for a URL; suggestions need review.", false, false, true, str("url", "", true)),
		spec("covers", "Search featured collection icons.", false, false, false, str("search", "", false)),
		spec("import parse-url", "Read parsed metadata for a URL.", false, false, false, str("url", "", true)),
		spec("import exists", "Check which URLs are already saved.", false, false, false, csv("urls", "", "strings", true)),
		spec("sharing list", "Read collection collaborators.", false, false, false, num("collection", "", true)),
		spec("backups", "List account backups.", false, false, false),
		spec("backup create", "Create an account backup after explicit review.", true, false, false),
		spec("add", "Create a bookmark; review the URL and destination.", true, false, true, str("url", "", true), str("title", "title", false), csv("tags", "tags", "strings", false), num("collection", "collection", false), str("note", "note", false), str("excerpt", "excerpt", false), boolean("parse", "parse")),
		spec("update", "Update one bookmark with an explicit JSON object.", true, true, false, num("id", "", true), body()),
		spec("delete", "Delete one bookmark. Items already in Trash may be permanently deleted.", true, true, false, num("id", "", true)),
		spec("user update", "Update authenticated user settings.", true, true, false, body()),
		spec("collection create", "Create a collection.", true, false, false, str("title", "", true), num("parent", "parent", false)),
		spec("collection update", "Update collection settings.", true, true, false, num("id", "", true), body()),
		spec("collection delete", "Delete a collection and move its bookmarks to Trash.", true, true, false, num("id", "", true)),
		spec("collection sort", "Sort collections (title, -title, or -count).", true, true, false, str("sort", "", true)),
		spec("collection expand", "Expand or collapse collections.", true, true, false, param("expanded", "", "boolean", true)),
		spec("collection merge", "Merge reviewed collection IDs into a destination.", true, true, false, num("to", "to", true), csv("ids", "ids", "integers", true)),
		spec("collection clean", "Delete empty collections after reviewing the collection tree.", true, true, false),
		spec("collection empty-trash", "Permanently empty all Trash after explicit review.", true, true, false),
		spec("tag rename", "Rename a tag throughout the selected collection.", true, true, false, str("from", "", true), str("to", "", true), num("collection", "collection", false)),
		spec("tag merge", "Merge tag names throughout a collection; cannot be scoped by search.", true, true, false, csv("from", "", "strings", true), str("to", "", true), num("collection", "collection", false)),
		spec("tag delete", "Remove tags throughout a collection.", true, true, false, csv("tags", "", "strings", true), num("collection", "collection", false)),
		spec("highlight add", "Add a highlight to a bookmark.", true, false, false, num("id", "", true), str("text", "text", true), str("color", "color", false), str("note", "note", false)),
		spec("highlight update", "Update a highlight.", true, true, false, num("id", "", true), str("highlight", "", true), str("text", "text", false), str("color", "color", false), str("note", "note", false)),
		spec("highlight delete", "Remove one highlight.", true, true, false, num("id", "", true), str("highlight", "", true)),
		spec("sharing invite", "Invite collaborators to a collection.", true, true, false, num("collection", "", true), csv("emails", "emails", "strings", true), str("role", "role", false)),
		spec("sharing role", "Change a collaborator's role.", true, true, false, num("collection", "", true), num("user", "", true), str("role", "role", true)),
		spec("sharing remove", "Remove a collaborator.", true, true, false, num("collection", "", true), num("user", "", true)),
		spec("sharing unshare", "Disable sharing on a collection.", true, true, false, num("collection", "", true)),
		spec("sharing join", "Join a collection using its invitation token.", true, true, false, num("collection", "", true), str("token", "token", true)),
	}
	for _, command := range []string{"list", "untagged", "duplicates", "broken"} {
		catalog = append(catalog, spec(command, "List bookmarks with bounded pagination and Raindrop search syntax.", false, false, true, listParameters...))
	}
	for _, action := range []string{"tag", "move", "delete"} {
		params := []parameter{num("collection", "collection", true), csv("ids", "ids", "integers", true)}
		if action == "tag" {
			params = append(params, csv("tags", "tags", "strings", true))
		}
		if action == "move" {
			params = append(params, num("to", "to", true))
		}
		entry := spec("batch "+action+" preview", "Create a read-only snapshot plan for explicit IDs in a source collection. Save and review the exact plan before cleanup apply.", false, false, action == "delete", params...)
		entry.Prefix = []string{"--dry-run"}
		catalog = append(catalog, entry)
	}
	return catalog
}

func (s toolSpec) name() string { return "raindrop_" + strings.ReplaceAll(s.Command, " ", "_") }

func (s toolSpec) schema() map[string]any {
	props := map[string]any{}
	required := []string{}
	for _, p := range s.Parameters {
		prop := map[string]any{"type": p.Type, "description": p.Description}
		switch p.Type {
		case "strings":
			prop["type"] = "array"
			prop["items"] = map[string]any{"type": "string", "minLength": 1, "pattern": "^[^,]+$"}
			prop["minItems"] = 1
		case "integers":
			prop["type"] = "array"
			prop["items"] = map[string]any{"type": "integer", "minimum": 1}
			prop["minItems"] = 1
			prop["maxItems"] = 100
			prop["uniqueItems"] = true
		case "string":
			if p.Required {
				prop["minLength"] = 1
			}
		case "integer":
			if p.Name == "page" {
				prop["minimum"] = 0
			} else if p.Name == "perPage" {
				prop["minimum"] = 1
				prop["maximum"] = 50
			} else if p.Name == "id" || p.Name == "user" {
				prop["minimum"] = 1
			}
		}
		if p.Name == "role" {
			prop["enum"] = []string{"viewer", "member"}
		}
		if s.Command == "collection sort" {
			prop["enum"] = []string{"title", "-title", "-count"}
		}
		props[p.Name] = prop
		if p.Required {
			required = append(required, p.Name)
		}
	}
	if s.Write {
		props["confirm"] = map[string]any{"type": "boolean", "const": true, "description": "Explicit intent to apply this reviewed operation; the host must obtain human authorization."}
		required = append(required, "confirm")
	}
	return map[string]any{"type": "object", "properties": props, "required": required, "additionalProperties": false}
}

func (s toolSpec) args(input map[string]json.RawMessage) ([]string, error) {
	command := strings.TrimSuffix(s.Command, " preview")
	if command == "collections tree" {
		command = "collections"
	}
	args := strings.Fields(command)
	args = append(args, s.Prefix...)
	if s.JSONFlag {
		args = append(args, "--json")
	}
	if s.Command == "collections tree" {
		args = append(args, "--tree")
	}
	if s.Command == "collection empty-trash" {
		args = append(args, "--yes")
	}
	positional := []string{}
	hasFlags := s.JSONFlag
	for _, p := range s.Parameters {
		if p.Flag != "" {
			hasFlags = true
		}
		raw, ok := input[p.Name]
		if !ok {
			continue
		}
		var value string
		switch p.Type {
		case "string":
			if err := json.Unmarshal(raw, &value); err != nil {
				return nil, err
			}
		case "integer":
			var n int64
			if err := json.Unmarshal(raw, &n); err != nil {
				return nil, err
			}
			value = strconv.FormatInt(n, 10)
		case "boolean":
			var b bool
			if err := json.Unmarshal(raw, &b); err != nil {
				return nil, err
			}
			value = strconv.FormatBool(b)
		case "strings":
			var values []string
			if err := json.Unmarshal(raw, &values); err != nil {
				return nil, err
			}
			value = strings.Join(values, ",")
		case "integers":
			var values []int64
			if err := json.Unmarshal(raw, &values); err != nil {
				return nil, err
			}
			value = int64CSV(values)
		case "object":
			var object map[string]any
			if err := json.Unmarshal(raw, &object); err != nil || object == nil {
				return nil, fmt.Errorf("body must be an object")
			}
			value = string(raw)
		default:
			return nil, fmt.Errorf("unsupported parameter type %s", p.Type)
		}
		if p.Flag != "" {
			args = append(args, "--"+p.Flag+"="+value)
		} else {
			positional = append(positional, value)
		}
	}
	if hasFlags && len(positional) > 0 {
		args = append(args, "--")
	}
	return append(args, positional...), nil
}
