package main

import (
	"fmt"
	"sort"
	"strings"
	"unicode"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func discoverTools(tools []*mcp.Tool, query, name string, limit int) ([]*mcp.Tool, error) {
	if limit < 0 || limit > 100 {
		return nil, fmt.Errorf("--limit must be between 0 and 100")
	}
	if query != "" && name != "" {
		return nil, fmt.Errorf("use either --search or --name")
	}
	if name != "" {
		for _, tool := range tools {
			if tool.Name == name {
				return []*mcp.Tool{tool}, nil
			}
		}
		return nil, fmt.Errorf("tool %q is unavailable in this catalog", name)
	}
	if query == "" {
		if limit > 0 && len(tools) > limit {
			return tools[:limit], nil
		}
		return tools, nil
	}
	terms := strings.FieldsFunc(strings.ToLower(query), func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsDigit(r) })
	stop := map[string]bool{"a": true, "an": true, "the": true, "to": true, "with": true, "for": true, "my": true, "and": true, "in": true, "of": true, "using": true}
	type match struct {
		tool  *mcp.Tool
		score int
	}
	matches := []match{}
	for _, tool := range tools {
		name, description := strings.ToLower(tool.Name), strings.ToLower(tool.Description)
		score := 0
		for _, term := range terms {
			if stop[term] {
				continue
			}
			if strings.Contains(name, term) {
				score += 3
			} else if strings.Contains(description, term) {
				score++
			}
		}
		if score > 0 {
			matches = append(matches, match{tool, score})
		}
	}
	sort.SliceStable(matches, func(i, j int) bool {
		if matches[i].score != matches[j].score {
			return matches[i].score > matches[j].score
		}
		return matches[i].tool.Name < matches[j].tool.Name
	})
	if limit == 0 {
		limit = 5
	}
	result := []*mcp.Tool{}
	for _, item := range matches {
		if len(result) == limit {
			break
		}
		result = append(result, item.tool)
	}
	return result, nil
}
