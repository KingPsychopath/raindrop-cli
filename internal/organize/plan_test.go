package organize

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/KingPsychopath/raindrop-cli/internal/raindrop"
)

func TestPlanReviewAndRecovery(t *testing.T) {
	items := map[int64]raindrop.Raindrop{1: {ID: 1, Title: "One", Link: "https://one.test", Collection: raindrop.Ref{ID: 10}, Tags: []string{"existing"}, LastUpdate: "2026-10-10"}, 2: {ID: 2, Title: "Two", Collection: raindrop.Ref{ID: 10}}}
	account := int64(7)
	writes := 0
	failID := int64(0)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/user" {
			json.NewEncoder(w).Encode(map[string]any{"user": map[string]any{"_id": account}})
			return
		}
		id, _ := strconv.ParseInt(strings.TrimPrefix(r.URL.Path, "/raindrop/"), 10, 64)
		if r.Method == "GET" {
			json.NewEncoder(w).Encode(map[string]any{"item": items[id], "result": true})
			return
		}
		writes++
		if id == failID {
			w.WriteHeader(500)
			w.Write([]byte(`{"errorMessage":"uncertain upstream failure"}`))
			return
		}
		if r.Method == "PUT" {
			var body struct {
				Tags []string `json:"tags"`
			}
			json.NewDecoder(r.Body).Decode(&body)
			item := items[id]
			item.Tags = body.Tags
			items[id] = item
		}
		json.NewEncoder(w).Encode(map[string]any{"result": true, "item": items[id]})
	}))
	defer server.Close()
	api := raindrop.NewClient(server.URL, "test-token")
	plan, err := NewPlan(context.Background(), api, "tag", 10, []int64{1, 2}, []string{"new"}, 0)
	if err != nil {
		t.Fatal(err)
	}
	// A serialized review file remains valid; all validation is read-only.
	data, _ := json.Marshal(plan)
	var saved Plan
	json.Unmarshal(data, &saved)
	result, err := Apply(context.Background(), api, saved, false)
	if err != nil || result.Status != "validated" || writes != 0 {
		t.Fatalf("dry-run %v %v writes=%d", result, err, writes)
	}
	account = 8
	if _, err := Apply(context.Background(), api, saved, true); err == nil || writes != 0 {
		t.Fatal("wrong account applied")
	}
	account = 7
	changed := items[2]
	changed.Title = "Changed"
	items[2] = changed
	if _, err := Apply(context.Background(), api, saved, true); err == nil || writes != 0 {
		t.Fatal("stale plan applied before all checks completed")
	}
	changed.Title = "Two"
	items[2] = changed
	tampered := saved
	tampered.Action = "delete"
	if _, err := Apply(context.Background(), api, tampered, true); err == nil || writes != 0 {
		t.Fatal("tampered plan applied")
	}
	failID = 2
	result, err = Apply(context.Background(), api, saved, true)
	if err == nil || result.Status != "partial" || len(result.Applied) != 1 || result.UncertainID != 2 || len(result.Remaining) != 1 {
		t.Fatalf("partial result=%+v err=%v", result, err)
	}
	if strings.Join(items[1].Tags, ",") != "existing,new" {
		t.Fatal("tag append lost existing tags")
	}
	if _, err := Apply(context.Background(), api, saved, true); err == nil {
		t.Fatal("partial plan replay accepted stale first item")
	}
}

func TestInvalidPlanSelectionsNeedNoNetwork(t *testing.T) {
	api := raindrop.NewClient("http://127.0.0.1:1", "test")
	for _, p := range []struct {
		action     string
		collection int64
		ids        []int64
		tags       []string
		to         int64
	}{{"delete", 0, []int64{1}, nil, 0}, {"delete", 10, []int64{1, 1}, nil, 0}, {"tag", 10, []int64{1}, nil, 0}, {"move", 10, []int64{1}, nil, 10}} {
		if _, err := NewPlan(context.Background(), api, p.action, p.collection, p.ids, p.tags, p.to); err == nil || strings.Contains(err.Error(), "request") {
			t.Fatalf("invalid plan reached network: %v", err)
		}
	}
}
