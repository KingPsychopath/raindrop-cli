package organize

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/machineyearningengineer/raindrop-cli/internal/raindrop"
)

func TestCreationPlanReviewAccountAndExactPayload(t *testing.T) {
	account := int64(7)
	writes := 0
	items := []raindrop.CreateItem{{Link: "https://example.com/a", Tags: []string{"My exact/tag"}, Note: "private note", Collection: &raindrop.Ref{ID: 10}}, {Link: "https://example.com/b"}}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/user":
			json.NewEncoder(w).Encode(map[string]any{"user": map[string]any{"_id": account}})
		case "/collections":
			w.Write([]byte(`{"items":[{"_id":10,"title":"Reference"}]}`))
		case "/collections/childrens":
			w.Write([]byte(`{"items":[]}`))
		case "/raindrops":
			writes++
			var body struct {
				Items []raindrop.CreateItem `json:"items"`
			}
			if r.Method != "POST" || json.NewDecoder(r.Body).Decode(&body) != nil || !reflect.DeepEqual(body.Items, items) {
				t.Error("reviewed creation payload changed")
			}
			w.Write([]byte(`{"result":true,"items":[{"_id":42},{"_id":43}]}`))
		default:
			t.Errorf("unexpected route %s", r.URL.Path)
		}
	}))
	defer server.Close()
	api := raindrop.NewClient(server.URL, "fixture")
	plan, err := NewCreatePlan(context.Background(), api, items)
	if err != nil {
		t.Fatal(err)
	}
	if result, err := ApplyCreatePlan(context.Background(), api, plan, false); err != nil || result.Status != "dry_run" || writes != 0 {
		t.Fatalf("dry run wrote: %+v %v", result, err)
	}
	changed := plan
	changed.Items = append([]raindrop.CreateItem{}, plan.Items...)
	changed.Items[0].Link = "https://example.com/changed"
	if _, err := ApplyCreatePlan(context.Background(), api, changed, true); err == nil || writes != 0 {
		t.Fatal("changed plan applied")
	}
	account = 8
	if _, err := ApplyCreatePlan(context.Background(), api, plan, true); err == nil || writes != 0 {
		t.Fatal("cross-account creation applied")
	}
	account = 7
	if result, err := ApplyCreatePlan(context.Background(), api, plan, true); err != nil || result.Status != "created" || writes != 1 || len(result.Items) != 2 {
		t.Fatalf("reviewed apply failed: %+v %v writes=%d", result, err, writes)
	}
}

func TestCreationRejectsInvalidInputAndDoesNotRetryUncertainWrites(t *testing.T) {
	writes := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/user" {
			w.Write([]byte(`{"user":{"_id":7}}`))
			return
		}
		writes++
		http.Error(w, "provider failed after receiving the request", 500)
	}))
	defer server.Close()
	api := raindrop.NewClient(server.URL, "fixture")
	for _, items := range [][]raindrop.CreateItem{nil, make([]raindrop.CreateItem, 101), {{Link: "file:///private"}}, {{Link: "https://example.com", Collection: &raindrop.Ref{ID: 0}}}} {
		if _, err := NewCreatePlan(context.Background(), api, items); err == nil || writes != 0 {
			t.Fatal("invalid creation input reached API")
		}
	}
	plan, err := NewCreatePlan(context.Background(), api, []raindrop.CreateItem{{Link: "https://example.com"}})
	if err != nil {
		t.Fatal(err)
	}
	result, err := ApplyCreatePlan(context.Background(), api, plan, true)
	if err == nil || result.Status != "outcome_unknown" || writes != 1 {
		t.Fatalf("uncertain write retried or hidden: %+v %v writes=%d", result, err, writes)
	}
}
