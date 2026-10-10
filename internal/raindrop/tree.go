package raindrop

import (
	"context"
	"fmt"
	"sort"
	"strings"
)

type CollectionPath struct {
	ID       int64  `json:"id"`
	Title    string `json:"title"`
	ParentID int64  `json:"parentId,omitempty"`
	Path     string `json:"path"`
	Depth    int    `json:"depth"`
	Count    int64  `json:"count"`
}

func (c *Client) CollectionTree(ctx context.Context) ([]CollectionPath, error) {
	roots, err := c.Collections(ctx, false)
	if err != nil {
		return nil, err
	}
	children, err := c.Collections(ctx, true)
	if err != nil {
		return nil, err
	}
	return BuildCollectionTree(append(roots, children...))
}

func BuildCollectionTree(items []Collection) ([]CollectionPath, error) {
	byID := map[int64]Collection{}
	for _, item := range items {
		byID[item.ID] = item
	}
	paths := map[int64][]string{}
	visiting := map[int64]bool{}
	var resolve func(int64) ([]string, error)
	resolve = func(id int64) ([]string, error) {
		if path, ok := paths[id]; ok {
			return path, nil
		}
		if visiting[id] {
			return nil, fmt.Errorf("collection hierarchy contains a cycle at %d", id)
		}
		item, ok := byID[id]
		if !ok {
			return []string{fmt.Sprintf("[missing parent %d]", id)}, nil
		}
		visiting[id] = true
		prefix := []string{}
		if item.Parent != nil && item.Parent.ID != 0 {
			var err error
			prefix, err = resolve(item.Parent.ID)
			if err != nil {
				return nil, err
			}
		}
		path := append(append([]string{}, prefix...), item.Title)
		paths[id] = path
		delete(visiting, id)
		return path, nil
	}
	out := []CollectionPath{}
	for id, item := range byID {
		path, err := resolve(id)
		if err != nil {
			return nil, err
		}
		row := CollectionPath{ID: id, Title: item.Title, Path: strings.Join(path, " / "), Depth: len(path) - 1, Count: item.Count}
		if item.Parent != nil {
			row.ParentID = item.Parent.ID
		}
		out = append(out, row)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Path == out[j].Path {
			return out[i].ID < out[j].ID
		}
		return out[i].Path < out[j].Path
	})
	return out, nil
}
