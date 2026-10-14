// Package organize implements reviewable, account-bound cleanup plans.
package organize

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/machineyearningengineer/raindrop-cli/internal/raindrop"
)

type Plan struct {
	Version      int                 `json:"version"`
	AccountID    int64               `json:"accountId"`
	CreatedAt    string              `json:"createdAt"`
	Action       string              `json:"action"`
	CollectionID int64               `json:"collectionId"`
	Tags         []string            `json:"tags,omitempty"`
	To           int64               `json:"to,omitempty"`
	Changes      []Change            `json:"changes,omitempty"`
	Items        []raindrop.Raindrop `json:"items"`
	Digest       string              `json:"digest"`
}

type Result struct {
	Status      string  `json:"status"`
	Applied     []int64 `json:"applied"`
	Remaining   []int64 `json:"remaining"`
	UncertainID int64   `json:"uncertainId,omitempty"`
	Error       string  `json:"error,omitempty"`
}

func accountID(ctx context.Context, api *raindrop.Client) (int64, error) {
	raw, err := api.User(ctx)
	if err != nil {
		return 0, err
	}
	var payload struct {
		User struct {
			ID int64 `json:"_id"`
		} `json:"user"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return 0, err
	}
	if payload.User.ID <= 0 {
		return 0, fmt.Errorf("account response has no valid user ID")
	}
	return payload.User.ID, nil
}

func NewPlan(ctx context.Context, api *raindrop.Client, action string, collection int64, ids []int64, tags []string, to int64) (Plan, error) {
	plan := Plan{Version: 1, CreatedAt: time.Now().UTC().Format(time.RFC3339), Action: action, CollectionID: collection, Tags: tags, To: to, Items: []raindrop.Raindrop{}}
	for _, id := range ids {
		plan.Items = append(plan.Items, raindrop.Raindrop{ID: id, Collection: raindrop.Ref{ID: collection}})
	}
	if err := plan.validate(); err != nil {
		return Plan{}, err
	}
	account, err := accountID(ctx, api)
	if err != nil {
		return Plan{}, err
	}
	plan.AccountID = account
	for i, id := range ids {
		item, err := api.Get(ctx, id)
		if err != nil {
			return Plan{}, err
		}
		if item.ID != id || item.Collection.ID != collection {
			return Plan{}, fmt.Errorf("bookmark %d is not in source collection %d", id, collection)
		}
		plan.Items[i] = item
	}
	plan.Digest = plan.digest()
	return plan, nil
}

func (p Plan) validate() error {
	if p.Version != 1 {
		return fmt.Errorf("unsupported cleanup plan version %d", p.Version)
	}
	if p.Action != "organize" && p.CollectionID != -1 && p.CollectionID != -99 && p.CollectionID <= 0 {
		return fmt.Errorf("plan needs an explicit source collection")
	}
	if len(p.Items) == 0 || len(p.Items) > 100 {
		return fmt.Errorf("plan must select 1..100 bookmarks")
	}
	seen := map[int64]bool{}
	for _, item := range p.Items {
		if item.ID <= 0 || seen[item.ID] || (p.Action != "organize" && item.Collection.ID != p.CollectionID) {
			return fmt.Errorf("invalid or duplicate bookmark selection")
		}
		seen[item.ID] = true
	}
	switch p.Action {
	case "organize":
		if p.CollectionID != 0 || len(p.Tags) != 0 || p.To != 0 || len(p.Changes) != len(p.Items) {
			return fmt.Errorf("organize plan needs one change per snapshot and no bulk fields")
		}
		for i, change := range p.Changes {
			if change.ID != p.Items[i].ID || (len(change.Tags) == 0 && change.To == nil) {
				return fmt.Errorf("invalid organize change")
			}
			if change.To != nil && (*change.To != -1 && *change.To <= 0) {
				return fmt.Errorf("invalid destination")
			}
			for _, tag := range change.Tags {
				if tag == "" {
					return fmt.Errorf("empty tag")
				}
			}
		}
	case "tag":
		if len(p.Tags) == 0 || p.To != 0 {
			return fmt.Errorf("tag plan needs tags and no destination")
		}
		for _, tag := range p.Tags {
			if tag == "" {
				return fmt.Errorf("tag cannot be empty")
			}
		}
	case "move":
		if (p.To != -1 && p.To <= 0) || p.To == p.CollectionID || len(p.Tags) != 0 {
			return fmt.Errorf("move plan needs a different positive destination or -1 and no tags")
		}
	case "delete":
		if p.To != 0 || len(p.Tags) != 0 {
			return fmt.Errorf("delete plan cannot include tags or a destination")
		}
	default:
		return fmt.Errorf("plan action must be tag, move, delete, or organize")
	}
	if p.Action != "organize" && len(p.Changes) != 0 {
		return fmt.Errorf("changes require an organize plan")
	}
	return nil
}

func (p Plan) digest() string {
	p.Digest = ""
	data, _ := json.Marshal(p)
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}

// Apply verifies every precondition before the first write. Raindrop has no
// transaction for this workflow: a later failure may leave earlier changes applied.
func Apply(ctx context.Context, api *raindrop.Client, plan Plan, yes bool) (Result, error) {
	result := Result{Status: "rejected", Applied: []int64{}, Remaining: []int64{}}
	for _, item := range plan.Items {
		result.Remaining = append(result.Remaining, item.ID)
	}
	fail := func(err error) (Result, error) { result.Error = err.Error(); return result, err }
	if err := plan.validate(); err != nil {
		return fail(err)
	}
	if plan.Digest != plan.digest() {
		return fail(fmt.Errorf("plan digest does not match; generate a fresh plan after changing it"))
	}
	account, err := accountID(ctx, api)
	if err != nil {
		return fail(err)
	}
	if account != plan.AccountID {
		return fail(fmt.Errorf("plan belongs to a different Raindrop account"))
	}
	for _, before := range plan.Items {
		current, err := api.Get(ctx, before.ID)
		if err != nil {
			return fail(err)
		}
		if !reflect.DeepEqual(current, before) {
			return fail(fmt.Errorf("bookmark %d changed since planning; generate a fresh plan", before.ID))
		}
	}
	if !yes {
		result.Status = "validated"
		return result, nil
	}
	for i, item := range plan.Items {
		switch plan.Action {
		case "organize":
			change := plan.Changes[i]
			body := map[string]any{}
			if len(change.Tags) > 0 {
				body["tags"] = appendTags(item.Tags, change.Tags)
			}
			if change.To != nil {
				body["collection"] = raindrop.Ref{ID: *change.To}
			}
			_, err = api.Update(ctx, item.ID, body)
		case "tag":
			_, err = api.Update(ctx, item.ID, map[string]any{"tags": appendTags(item.Tags, plan.Tags)})
		case "move":
			_, err = api.Update(ctx, item.ID, map[string]any{"collection": raindrop.Ref{ID: plan.To}})
		case "delete":
			_, err = api.Delete(ctx, item.ID)
		}
		if err != nil {
			result.Status = "partial"
			result.UncertainID = item.ID
			return fail(err)
		}
		result.Applied = append(result.Applied, item.ID)
		result.Remaining = result.Remaining[1:]
	}
	result.Status = "applied"
	return result, nil
}

func appendTags(before, additions []string) []string {
	tags := append([]string{}, before...)
	for _, tag := range additions {
		found := false
		for _, existing := range tags {
			if existing == tag {
				found = true
				break
			}
		}
		if !found {
			tags = append(tags, tag)
		}
	}
	return tags
}
