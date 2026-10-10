package cli

import (
	"bytes"
	"encoding/json"
	"github.com/KingPsychopath/raindrop-cli/internal/raindrop"
	"strings"
	"testing"
)

func TestTSVControlsCannotCreateExtraRows(t *testing.T) {
	var out bytes.Buffer
	item := raindrop.Raindrop{ID: 1, Title: "Title\n\x1b[31m\tmore", Tags: []string{"a\nb"}, Link: "https://fixture.test"}
	PrintRaindrops(&out, []raindrop.Raindrop{item})
	if strings.Count(out.String(), "\n") != 1 || strings.Count(out.String(), "\t") != 4 || strings.Contains(out.String(), "\x1b") {
		t.Fatal(out.String())
	}
	raw, _ := json.Marshal(item)
	var decoded raindrop.Raindrop
	json.Unmarshal(raw, &decoded)
	if decoded.Title != item.Title {
		t.Fatal("JSON metadata changed")
	}
	if _, err := JSONObject("null"); err == nil {
		t.Fatal("null object accepted")
	}
}
