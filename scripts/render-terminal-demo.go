package main

import (
	"fmt"
	"html"
	"io"
	"os"
)

type textLine struct {
	X     int
	Y     int
	Fill  string
	Size  int
	Value string
}

func main() {
	if err := render(os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func render(w io.Writer) error {
	lines := []textLine{
		{132, 54, "#9FB0C3", 15, "rdrop - Shell commands, MCP tools and model workflows"},
		{48, 112, "#5EEAD4", 18, "$"},
		{72, 112, "#E6EDF3", 18, "rdrop mcp tools | jq length"},
		{72, 146, "#A7F3D0", 18, "26 read-only tools; writes require --allow-write"},
		{48, 198, "#5EEAD4", 18, "$"},
		{72, 198, "#E6EDF3", 18, "rdrop collections --tree"},
		{72, 232, "#C9D1D9", 18, "Reference / Engineering        184"},
		{72, 260, "#C9D1D9", 18, "Reading / Essays               126"},
		{72, 288, "#C9D1D9", 18, "Unsorted                        93"},
		{48, 340, "#5EEAD4", 18, "$"},
		{72, 340, "#E6EDF3", 18, "rdrop cleanup apply reviewed-plan.json --dry-run"},
		{72, 374, "#C9D1D9", 18, "validated: account, digest and every bookmark snapshot"},
		{72, 402, "#FDE68A", 18, "next: review the exact plan, then explicitly apply"},
	}

	fmt.Fprintln(w, `<svg width="960" height="460" viewBox="0 0 960 460" fill="none" xmlns="http://www.w3.org/2000/svg" role="img" aria-labelledby="title desc">`)
	fmt.Fprintln(w, `  <title id="title">raindrop-cli terminal demo</title>`)
	fmt.Fprintln(w, `  <desc id="desc">A terminal screenshot-style demo showing rdrop commands for discovering MCP tools, browsing collection paths, and validating a reviewed plan.</desc>`)
	fmt.Fprintln(w, `  <rect width="960" height="460" rx="18" fill="#101820"/>`)
	fmt.Fprintln(w, `  <rect x="24" y="24" width="912" height="412" rx="14" fill="#0B1117" stroke="#26313D"/>`)
	fmt.Fprintln(w, `  <rect x="24" y="24" width="912" height="48" rx="14" fill="#17212B"/>`)
	fmt.Fprintln(w, `  <circle cx="52" cy="48" r="7" fill="#FF5F57"/>`)
	fmt.Fprintln(w, `  <circle cx="76" cy="48" r="7" fill="#FFBD2E"/>`)
	fmt.Fprintln(w, `  <circle cx="100" cy="48" r="7" fill="#28C840"/>`)
	for _, line := range lines {
		fmt.Fprintf(w, `  <text x="%d" y="%d" fill="%s" font-family="Menlo, Consolas, monospace" font-size="%d">%s</text>`+"\n",
			line.X, line.Y, line.Fill, line.Size, html.EscapeString(line.Value))
	}
	fmt.Fprintln(w, `</svg>`)
	return nil
}
