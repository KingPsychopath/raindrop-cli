package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/KingPsychopath/raindrop-cli/internal/raindrop"
)

func PrintJSON(w io.Writer, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(w, string(data))
	return err
}

func PrintRaindrops(w io.Writer, items []raindrop.Raindrop) {
	for _, item := range items {
		tags := strings.Join(item.Tags, ",")
		if tags == "" {
			tags = "-"
		}
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\n", item.ID, item.Title, item.Link, item.Domain, tags)
	}
}

func PrintCollections(w io.Writer, items []raindrop.Collection) {
	for _, item := range items {
		parent := "-"
		if item.Parent != nil {
			parent = fmt.Sprint(item.Parent.ID)
		}
		fmt.Fprintf(w, "%d\t%s\t%d\tparent:%s\n", item.ID, item.Title, item.Count, parent)
	}
}

func PrintTags(w io.Writer, items []raindrop.Tag) {
	for _, item := range items {
		fmt.Fprintf(w, "%s\t%d\n", item.ID, item.Count)
	}
}
