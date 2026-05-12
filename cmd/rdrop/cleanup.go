package main

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/KingPsychopath/raindrop-cli/internal/raindrop"
)

type cleanupTagPlan struct {
	Replace string   `json:"replace"`
	Tags    []string `json:"tags"`
	Count   int64    `json:"count"`
	Command string   `json:"command"`
}

type cleanupReportData struct {
	CollectionID int64             `json:"collectionId"`
	Search       string            `json:"search,omitempty"`
	Counts       map[string]int64  `json:"counts"`
	TopTags      []raindrop.Tag    `json:"topTags"`
	TagPlans     []cleanupTagPlan  `json:"tagPlans"`
	Commands     map[string]string `json:"commands"`
}

func cleanupReport(collectionID int64, search string, filters raindrop.Filters) cleanupReportData {
	report := cleanupReportData{
		CollectionID: collectionID,
		Search:       search,
		Counts: map[string]int64{
			"broken":     filters.Broken.Count,
			"duplicates": filters.Duplicates.Count,
			"important":  filters.Important.Count,
			"untagged":   filters.NoTag.Count,
		},
		TopTags:  topTags(filters.Tags, 20),
		Commands: cleanupCommands(collectionID, search),
	}
	report.TagPlans = tagNormalizationPlans(collectionID, filters.Tags)
	return report
}

func cleanupCommands(collectionID int64, search string) map[string]string {
	base := ""
	if collectionID != 0 {
		base = fmt.Sprintf("--collection %d ", collectionID)
	}
	if search != "" {
		base += fmt.Sprintf("--search %q ", search)
	}
	return map[string]string{
		"broken":     strings.TrimSpace("rdrop broken " + base + "--json"),
		"duplicates": strings.TrimSpace("rdrop duplicates " + base + "--json"),
		"untagged":   strings.TrimSpace("rdrop untagged " + base + "--json"),
		"filters":    strings.TrimSpace("rdrop filters " + base + "--json"),
	}
}

func topTags(tags []raindrop.Tag, limit int) []raindrop.Tag {
	out := append([]raindrop.Tag(nil), tags...)
	sort.Slice(out, func(i, j int) bool {
		if out[i].Count == out[j].Count {
			return out[i].ID < out[j].ID
		}
		return out[i].Count > out[j].Count
	})
	if len(out) > limit {
		out = out[:limit]
	}
	return out
}

func tagNormalizationPlans(collectionID int64, tags []raindrop.Tag) []cleanupTagPlan {
	type group struct {
		tags  []string
		count int64
	}
	groups := map[string]*group{}
	for _, tag := range tags {
		normalized := normalizeTag(tag.ID)
		if normalized == "" || normalized == tag.ID {
			continue
		}
		g, ok := groups[normalized]
		if !ok {
			g = &group{}
			groups[normalized] = g
		}
		g.tags = append(g.tags, tag.ID)
		g.count += tag.Count
	}

	plans := make([]cleanupTagPlan, 0, len(groups))
	for replace, group := range groups {
		sort.Strings(group.tags)
		command := fmt.Sprintf("rdrop tag merge %s %s", shellQuote(strings.Join(group.tags, ",")), shellQuote(replace))
		if collectionID != 0 {
			command += fmt.Sprintf(" --collection %d", collectionID)
		}
		plans = append(plans, cleanupTagPlan{
			Replace: replace,
			Tags:    group.tags,
			Count:   group.count,
			Command: command,
		})
	}
	sort.Slice(plans, func(i, j int) bool {
		if plans[i].Count == plans[j].Count {
			return plans[i].Replace < plans[j].Replace
		}
		return plans[i].Count > plans[j].Count
	})
	return plans
}

func normalizeTag(tag string) string {
	return strings.Join(strings.Fields(strings.ToLower(strings.TrimSpace(tag))), "-")
}

func shellQuote(value string) string {
	if value == "" {
		return "''"
	}
	if !strings.ContainsAny(value, " \t\n'\"\\$`!*?[]{}()&;|<>#~") {
		return value
	}
	return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'"
}

func printCleanupReport(w io.Writer, report cleanupReportData) {
	fmt.Fprintf(w, "cleanup report collection:%d\n", report.CollectionID)
	if report.Search != "" {
		fmt.Fprintf(w, "search: %s\n", report.Search)
	}
	fmt.Fprintf(w, "broken: %d\n", report.Counts["broken"])
	fmt.Fprintf(w, "duplicates: %d\n", report.Counts["duplicates"])
	fmt.Fprintf(w, "untagged: %d\n", report.Counts["untagged"])
	fmt.Fprintf(w, "important: %d\n", report.Counts["important"])
	fmt.Fprintln(w)
	fmt.Fprintln(w, "next commands:")
	keys := []string{"filters", "untagged", "duplicates", "broken"}
	for _, key := range keys {
		fmt.Fprintf(w, "  %s\t%s\n", key, report.Commands[key])
	}
	if len(report.TagPlans) == 0 {
		return
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "tag normalization candidates:")
	limit := len(report.TagPlans)
	if limit > 10 {
		limit = 10
	}
	for _, plan := range report.TagPlans[:limit] {
		fmt.Fprintf(w, "  %s <= %s\tcount:%d\n", plan.Replace, strings.Join(plan.Tags, ","), plan.Count)
	}
}
