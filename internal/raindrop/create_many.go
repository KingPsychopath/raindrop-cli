package raindrop

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type CreateItem struct {
	Link        string    `json:"link"`
	Title       string    `json:"title,omitempty"`
	Tags        []string  `json:"tags,omitempty"`
	Note        string    `json:"note,omitempty"`
	Excerpt     string    `json:"excerpt,omitempty"`
	Collection  *Ref      `json:"collection,omitempty"`
	Important   *bool     `json:"important,omitempty"`
	PleaseParse *struct{} `json:"pleaseParse,omitempty"`
}

func ValidateCreateItems(items []CreateItem) error {
	if len(items) < 1 || len(items) > 100 {
		return fmt.Errorf("bulk creation requires 1..100 bookmarks")
	}
	for i, item := range items {
		parsed, err := url.Parse(item.Link)
		if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
			return fmt.Errorf("item %d requires an absolute HTTP(S) link", i+1)
		}
		if item.Collection != nil && item.Collection.ID != -1 && item.Collection.ID <= 0 {
			return fmt.Errorf("item %d destination must be -1 or a positive collection ID", i+1)
		}
		for _, tag := range item.Tags {
			if strings.TrimSpace(tag) == "" {
				return fmt.Errorf("item %d has an empty tag", i+1)
			}
		}
	}
	return nil
}

func (c *Client) CreateMany(ctx context.Context, items []CreateItem) ([]Raindrop, error) {
	if err := ValidateCreateItems(items); err != nil {
		return nil, err
	}
	var out raindropsResponse
	err := c.Do(ctx, http.MethodPost, "raindrops", nil, map[string]any{"items": items}, &out)
	return out.Items, err
}
