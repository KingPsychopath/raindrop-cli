package organize

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/machineyearningengineer/raindrop-cli/internal/raindrop"
)

// Policy adds meanings to existing names. New names must be explicitly enabled
// and declared in TagRules; models cannot create arbitrary names.
type Policy struct {
	Instructions    string            `json:"instructions,omitempty"`
	TagRules        map[string]string `json:"tagRules,omitempty"`
	CollectionRules map[string]string `json:"collectionRules,omitempty"`
	AllowNewTags    bool              `json:"allowNewTags"`
}
type Scope struct {
	CollectionID int64  `json:"collectionId"`
	Search       string `json:"search,omitempty"`
	Page         int    `json:"page"`
	PerPage      int    `json:"perPage"`
	ExplicitIDs  bool   `json:"explicitIds"`
}
type Bookmark struct {
	ID           int64    `json:"id"`
	Title        string   `json:"title"`
	Link         string   `json:"link"`
	Excerpt      string   `json:"excerpt"`
	Tags         []string `json:"tags"`
	CollectionID int64    `json:"collectionId"`
}
type LibraryContext struct {
	Version     int                       `json:"version"`
	AccountID   int64                     `json:"accountId"`
	CreatedAt   string                    `json:"createdAt"`
	Scope       Scope                     `json:"scope"`
	Policy      Policy                    `json:"policy"`
	Tags        []raindrop.Tag            `json:"tags"`
	Collections []raindrop.CollectionPath `json:"collections"`
	Bookmarks   []Bookmark                `json:"bookmarks"`
	Snapshots   []raindrop.Raindrop       `json:"snapshots,omitempty"`
	Digest      string                    `json:"digest"`
}

func (c LibraryContext) digest() string {
	c.Digest = ""
	raw, _ := json.Marshal(c)
	sum := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(sum[:])
}

// ModelView omits private notes and full apply preconditions. Retain the local
// context file (or server cache) to build a plan from the model's response.
func (c LibraryContext) ModelView() LibraryContext { c.Snapshots = nil; return c }
func NewContext(ctx context.Context, api *raindrop.Client, ids []int64, scope Scope, policy Policy) (LibraryContext, error) {
	c := LibraryContext{Version: 1, CreatedAt: time.Now().UTC().Format(time.RFC3339), Scope: scope, Policy: policy, Bookmarks: []Bookmark{}, Snapshots: []raindrop.Raindrop{}}
	if len(ids) > 50 {
		return c, fmt.Errorf("context selects at most 50 bookmarks")
	}
	seen := map[int64]bool{}
	for _, id := range ids {
		if id <= 0 || seen[id] {
			return c, fmt.Errorf("context IDs must be positive and unique")
		}
		seen[id] = true
	}
	var err error
	c.AccountID, err = accountID(ctx, api)
	if err != nil {
		return c, err
	}
	c.Collections, err = api.CollectionTree(ctx)
	if err != nil {
		return c, err
	}
	c.Collections = append(c.Collections, raindrop.CollectionPath{ID: -1, Title: "Unsorted", Path: "Unsorted"})
	c.Tags, err = api.Tags(ctx, 0)
	if err != nil {
		return c, err
	}
	known := map[string]bool{}
	for _, tag := range c.Tags {
		known[tag.ID] = true
	}
	for _, id := range ids {
		item, err := api.Get(ctx, id)
		if err != nil {
			return c, err
		}
		if item.ID != id {
			return c, fmt.Errorf("unexpected bookmark ID")
		}
		c.Snapshots = append(c.Snapshots, item)
		c.Bookmarks = append(c.Bookmarks, Bookmark{ID: item.ID, Title: item.Title, Link: item.Link, Excerpt: item.Excerpt, Tags: item.Tags, CollectionID: item.Collection.ID})
		for _, tag := range item.Tags {
			if !known[tag] {
				c.Tags = append(c.Tags, raindrop.Tag{ID: tag})
				known[tag] = true
			}
		}
	}
	sort.Slice(c.Tags, func(i, j int) bool { return c.Tags[i].ID < c.Tags[j].ID })
	for tag, rule := range policy.TagRules {
		if tag == "" || rule == "" {
			return c, fmt.Errorf("tag rules require names and descriptions")
		}
		if !known[tag] && !policy.AllowNewTags {
			return c, fmt.Errorf("policy names new tag %q; explicitly allow new tags first", tag)
		}
	}
	c.Digest = c.digest()
	return c, nil
}

type Change struct {
	ID   int64    `json:"id"`
	Tags []string `json:"tags,omitempty"`
	To   *int64   `json:"collectionId,omitempty"`
}
type Suggestion struct {
	Change
	Reason      string `json:"reason"`
	NeedsReview *bool  `json:"needsReview"`
}
type Proposal struct {
	ContextDigest string       `json:"contextDigest"`
	Suggestions   []Suggestion `json:"suggestions"`
}
type PlannedProposal struct {
	ContextDigest string       `json:"contextDigest"`
	Plan          *Plan        `json:"plan,omitempty"`
	NeedsReview   []Suggestion `json:"needsReview"`
	Unchanged     []int64      `json:"unchanged"`
	Unmentioned   []int64      `json:"unmentioned"`
}

// PlanProposal validates all suggestions before constructing any apply file.
// Validation is structural, not a guarantee that a model's judgement is correct.
func PlanProposal(c LibraryContext, proposal Proposal) (PlannedProposal, error) {
	out := PlannedProposal{ContextDigest: c.Digest, NeedsReview: []Suggestion{}, Unchanged: []int64{}, Unmentioned: []int64{}}
	if c.Version != 1 || c.AccountID <= 0 || c.Digest != c.digest() || proposal.ContextDigest != c.Digest {
		return out, fmt.Errorf("context version, account, or digest mismatch; use the original context file")
	}
	if len(c.Snapshots) != len(c.Bookmarks) || len(c.Snapshots) > 50 {
		return out, fmt.Errorf("context needs its original snapshots")
	}
	items := map[int64]raindrop.Raindrop{}
	for i, item := range c.Snapshots {
		if item.ID <= 0 || c.Bookmarks[i].ID != item.ID {
			return out, fmt.Errorf("invalid context snapshot")
		}
		if _, ok := items[item.ID]; ok {
			return out, fmt.Errorf("duplicate context ID")
		}
		items[item.ID] = item
	}
	tags := map[string]bool{}
	for _, tag := range c.Tags {
		tags[tag.ID] = true
	}
	if c.Policy.AllowNewTags {
		for tag := range c.Policy.TagRules {
			tags[tag] = true
		}
	}
	collections := map[int64]bool{-1: true}
	for _, row := range c.Collections {
		if row.ID > 0 {
			collections[row.ID] = true
		}
	}
	plan := Plan{Version: 1, AccountID: c.AccountID, CreatedAt: time.Now().UTC().Format(time.RFC3339), Action: "organize", Items: []raindrop.Raindrop{}, Changes: []Change{}}
	seen := map[int64]bool{}
	for _, suggestion := range proposal.Suggestions {
		item, ok := items[suggestion.ID]
		if !ok || seen[suggestion.ID] {
			return out, fmt.Errorf("unknown or duplicate proposal ID %d", suggestion.ID)
		}
		seen[suggestion.ID] = true
		selected := map[string]bool{}
		additions := []string{}
		for _, tag := range suggestion.Tags {
			if !tags[tag] || selected[tag] {
				return out, fmt.Errorf("unknown or duplicate proposed tag %q", tag)
			}
			selected[tag] = true
			exists := false
			for _, before := range item.Tags {
				if before == tag {
					exists = true
					break
				}
			}
			if !exists {
				additions = append(additions, tag)
			}
		}
		if suggestion.To != nil && !collections[*suggestion.To] {
			return out, fmt.Errorf("unknown or forbidden destination %d", *suggestion.To)
		}
		if suggestion.Reason == "" {
			return out, fmt.Errorf("each suggestion needs a reason")
		}
		if suggestion.NeedsReview == nil {
			return out, fmt.Errorf("needsReview must be explicitly true or false")
		}
		if *suggestion.NeedsReview {
			out.NeedsReview = append(out.NeedsReview, suggestion)
			continue
		}
		change := Change{ID: item.ID, Tags: additions, To: suggestion.To}
		if change.To != nil && *change.To == item.Collection.ID {
			change.To = nil
		}
		if len(change.Tags) == 0 && change.To == nil {
			out.Unchanged = append(out.Unchanged, item.ID)
			continue
		}
		plan.Items = append(plan.Items, item)
		plan.Changes = append(plan.Changes, change)
	}
	for _, item := range c.Snapshots {
		if !seen[item.ID] {
			out.Unmentioned = append(out.Unmentioned, item.ID)
		}
	}
	if len(plan.Items) > 0 {
		if err := plan.validate(); err != nil {
			return out, err
		}
		plan.Digest = plan.digest()
		out.Plan = &plan
	}
	return out, nil
}
