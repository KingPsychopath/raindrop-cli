package organize

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/machineyearningengineer/raindrop-cli/internal/raindrop"
)

type CreatePlan struct {
	Version   int                   `json:"version"`
	Action    string                `json:"action"`
	AccountID int64                 `json:"accountId"`
	CreatedAt string                `json:"createdAt"`
	Items     []raindrop.CreateItem `json:"items"`
	Digest    string                `json:"digest"`
}

type CreateResult struct {
	Status    string              `json:"status"`
	Requested int                 `json:"requested"`
	Items     []raindrop.Raindrop `json:"items,omitempty"`
}

func (p CreatePlan) digest() string {
	p.Digest = ""
	raw, _ := json.Marshal(p)
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

func NewCreatePlan(ctx context.Context, api *raindrop.Client, items []raindrop.CreateItem) (CreatePlan, error) {
	if err := raindrop.ValidateCreateItems(items); err != nil {
		return CreatePlan{}, err
	}
	account, err := accountID(ctx, api)
	if err != nil {
		return CreatePlan{}, err
	}
	// Copy the metadata so later changes to the caller's slices cannot alter it.
	raw, _ := json.Marshal(items)
	var copy []raindrop.CreateItem
	json.Unmarshal(raw, &copy)
	plan := CreatePlan{Version: 1, Action: "create", AccountID: account, CreatedAt: time.Now().UTC().Format(time.RFC3339), Items: copy}
	if err := validateCreateDestinations(ctx, api, copy); err != nil {
		return CreatePlan{}, err
	}
	plan.Digest = plan.digest()
	return plan, nil
}

func validateCreateDestinations(ctx context.Context, api *raindrop.Client, items []raindrop.CreateItem) error {
	required := map[int64]bool{}
	for _, item := range items {
		if item.Collection != nil && item.Collection.ID > 0 {
			required[item.Collection.ID] = true
		}
	}
	if len(required) == 0 {
		return nil
	}
	for _, children := range []bool{false, true} {
		collections, err := api.Collections(ctx, children)
		if err != nil {
			return err
		}
		for _, collection := range collections {
			delete(required, collection.ID)
		}
	}
	if len(required) > 0 {
		return fmt.Errorf("bulk creation refers to an unavailable destination collection")
	}
	return nil
}

func ApplyCreatePlan(ctx context.Context, api *raindrop.Client, plan CreatePlan, confirm bool) (CreateResult, error) {
	result := CreateResult{Status: "rejected", Requested: len(plan.Items)}
	if plan.Version != 1 || plan.Action != "create" || plan.AccountID <= 0 || plan.Digest != plan.digest() {
		return result, fmt.Errorf("invalid or changed creation plan")
	}
	if err := raindrop.ValidateCreateItems(plan.Items); err != nil {
		return result, err
	}
	account, err := accountID(ctx, api)
	if err != nil {
		return result, err
	}
	if account != plan.AccountID {
		return result, fmt.Errorf("creation plan belongs to another account")
	}
	if err := validateCreateDestinations(ctx, api, plan.Items); err != nil {
		return result, err
	}
	if !confirm {
		result.Status = "dry_run"
		return result, nil
	}
	result.Status = "outcome_unknown"
	result.Items, err = api.CreateMany(ctx, plan.Items)
	if err != nil {
		return result, fmt.Errorf("bulk create outcome is uncertain; inspect saved URLs before retrying: %w", err)
	}
	if len(result.Items) != len(plan.Items) {
		return result, fmt.Errorf("bulk create returned an incomplete result; inspect saved URLs before retrying")
	}
	seen := map[int64]bool{}
	for _, item := range result.Items {
		if item.ID <= 0 || seen[item.ID] {
			return result, fmt.Errorf("bulk create returned an invalid bookmark; inspect saved URLs before retrying")
		}
		seen[item.ID] = true
	}
	result.Status = "created"
	return result, nil
}
