package organize

import (
	"context"
	"encoding/json"
	"github.com/KingPsychopath/raindrop-cli/internal/raindrop"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func boolean(v bool) *bool { return &v }
func TestContextProposalContractAndCombinedApply(t *testing.T) {
	item := raindrop.Raindrop{ID: 42, Title: "Docs", Collection: raindrop.Ref{ID: -1}, Tags: []string{"My own tag"}, Note: "private-note"}
	writes := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/user":
			w.Write([]byte(`{"user":{"_id":7}}`))
		case "/collections":
			w.Write([]byte(`{"items":[{"_id":10,"title":"Work"}]}`))
		case "/collections/childrens":
			w.Write([]byte(`{"items":[{"_id":11,"title":"Docs","parent":{"$id":10}}]}`))
		case "/tags/0":
			w.Write([]byte(`{"items":[{"_id":"My own tag","count":3},{"_id":"reference","count":2}]}`))
		case "/raindrop/42":
			if r.Method == "PUT" {
				writes++
				var body struct {
					Tags       []string     `json:"tags"`
					Collection raindrop.Ref `json:"collection"`
				}
				json.NewDecoder(r.Body).Decode(&body)
				item.Tags = body.Tags
				item.Collection = body.Collection
			}
			json.NewEncoder(w).Encode(map[string]any{"item": item, "result": true})
		default:
			t.Errorf("unexpected request %s", r.URL.Path)
		}
	}))
	defer server.Close()
	api := raindrop.NewClient(server.URL, "fixture")
	c, err := NewContext(context.Background(), api, []int64{42}, Scope{ExplicitIDs: true}, Policy{})
	if err != nil {
		t.Fatal(err)
	}
	if c.Collections[0].Path != "Work" || c.Collections[1].Path != "Work / Docs" {
		t.Fatal("missing hierarchy")
	}
	raw, _ := json.Marshal(c.ModelView())
	if strings.Contains(string(raw), "private-note") || strings.Contains(string(raw), "snapshots") {
		t.Fatal("model view exposed notes or snapshots")
	}
	destination := int64(11)
	proposal := Proposal{ContextDigest: c.Digest, Suggestions: []Suggestion{{Change: Change{ID: 42, Tags: []string{"reference", "My own tag"}, To: &destination}, Reason: "Technical reference", NeedsReview: boolean(false)}}}
	out, err := PlanProposal(c, proposal)
	if err != nil || out.Plan == nil {
		t.Fatalf("%+v %v", out, err)
	}
	result, err := Apply(context.Background(), api, *out.Plan, true)
	if err != nil || result.Status != "applied" || writes != 1 || item.Collection.ID != 11 || strings.Join(item.Tags, ",") != "My own tag,reference" || item.Note != "private-note" {
		t.Fatalf("lost metadata or split write: %+v %+v %v", item, result, err)
	}
	proposal.Suggestions[0].NeedsReview = boolean(true)
	out, err = PlanProposal(c, proposal)
	if err != nil || out.Plan != nil || len(out.NeedsReview) != 1 {
		t.Fatal("uncertain suggestion became an apply plan")
	}
	for _, change := range []Suggestion{
		{Change: Change{ID: 999}, Reason: "x", NeedsReview: boolean(false)},
		{Change: Change{ID: 42, Tags: []string{"Invented"}}, Reason: "x", NeedsReview: boolean(false)},
		{Change: Change{ID: 42, To: new(int64)}, Reason: "x", NeedsReview: boolean(false)},
		{Change: Change{ID: 42, Tags: []string{"reference"}}, Reason: "x"},
	} {
		if _, err := PlanProposal(c, Proposal{ContextDigest: c.Digest, Suggestions: []Suggestion{change}}); err == nil {
			t.Fatalf("invalid proposal accepted %+v", change)
		}
	}
	modified := c
	modified.Policy.AllowNewTags = true
	if _, err := PlanProposal(modified, proposal); err == nil {
		t.Fatal("tampered context accepted")
	}
}
