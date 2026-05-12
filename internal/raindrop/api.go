package raindrop

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

type ListOptions struct {
	CollectionID int64
	Search       string
	Sort         string
	Page         int
	PerPage      int
	Nested       bool
}

func (c *Client) Raw(ctx context.Context, method, path string, query url.Values, body any) (json.RawMessage, error) {
	var raw json.RawMessage
	err := c.Do(ctx, method, path, query, body, &raw)
	return raw, err
}

func (c *Client) RawBytes(ctx context.Context, method, path string, query url.Values, body any) ([]byte, string, error) {
	return c.Bytes(ctx, method, path, query, body)
}

func (c *Client) User(ctx context.Context) (json.RawMessage, error) {
	return c.Raw(ctx, http.MethodGet, "user", nil, nil)
}

func (c *Client) UpdateUser(ctx context.Context, body map[string]any) (json.RawMessage, error) {
	return c.Raw(ctx, http.MethodPut, "user", nil, body)
}

func (c *Client) Stats(ctx context.Context) (UserStats, error) {
	var out UserStats
	err := c.Do(ctx, http.MethodGet, "user/stats", nil, nil, &out)
	return out, err
}

func (c *Client) Collections(ctx context.Context, children bool) ([]Collection, error) {
	path := "collections"
	if children {
		path = "collections/childrens"
	}
	var out collectionsResponse
	err := c.Do(ctx, http.MethodGet, path, nil, nil, &out)
	return out.Items, err
}

func (c *Client) CreateCollection(ctx context.Context, title string, parentID int64) (json.RawMessage, error) {
	body := map[string]any{"title": title}
	if parentID != 0 {
		body["parent"] = Ref{ID: parentID}
	}
	return c.Raw(ctx, http.MethodPost, "collection", nil, body)
}

func (c *Client) UpdateCollection(ctx context.Context, id int64, body map[string]any) (json.RawMessage, error) {
	return c.Raw(ctx, http.MethodPut, fmt.Sprintf("collection/%d", id), nil, body)
}

func (c *Client) DeleteCollection(ctx context.Context, id int64) (json.RawMessage, error) {
	return c.Raw(ctx, http.MethodDelete, fmt.Sprintf("collection/%d", id), nil, nil)
}

func (c *Client) SortCollections(ctx context.Context, sort string) (json.RawMessage, error) {
	return c.Raw(ctx, http.MethodPut, "collections", nil, map[string]any{"sort": sort})
}

func (c *Client) ExpandCollections(ctx context.Context, expanded bool) (json.RawMessage, error) {
	return c.Raw(ctx, http.MethodPut, "collections", nil, map[string]any{"expanded": expanded})
}

func (c *Client) MergeCollections(ctx context.Context, to int64, ids []int64) (json.RawMessage, error) {
	return c.Raw(ctx, http.MethodPut, "collections/merge", nil, map[string]any{"to": to, "ids": ids})
}

func (c *Client) CleanCollections(ctx context.Context) (json.RawMessage, error) {
	return c.Raw(ctx, http.MethodPut, "collections/clean", nil, nil)
}

func (c *Client) CollectionCovers(ctx context.Context, search string) (json.RawMessage, error) {
	path := "collections/covers"
	if search != "" {
		path = fmt.Sprintf("collections/covers/%s", url.PathEscape(search))
	}
	return c.Raw(ctx, http.MethodGet, path, nil, nil)
}

func (c *Client) Tags(ctx context.Context, collectionID int64) ([]Tag, error) {
	var out tagsResponse
	err := c.Do(ctx, http.MethodGet, fmt.Sprintf("tags/%d", collectionID), nil, nil, &out)
	return out.Items, err
}

func (c *Client) RenameTags(ctx context.Context, collectionID int64, tags []string, replace string) (json.RawMessage, error) {
	return c.Raw(ctx, http.MethodPut, fmt.Sprintf("tags/%d", collectionID), nil, map[string]any{
		"tags":    tags,
		"replace": replace,
	})
}

func (c *Client) DeleteTags(ctx context.Context, collectionID int64, tags []string) (json.RawMessage, error) {
	return c.Raw(ctx, http.MethodDelete, fmt.Sprintf("tags/%d", collectionID), nil, map[string]any{"tags": tags})
}

func (c *Client) Filters(ctx context.Context, collectionID int64, search string) (Filters, error) {
	q := url.Values{}
	if search != "" {
		q.Set("search", search)
	}
	var out Filters
	err := c.Do(ctx, http.MethodGet, fmt.Sprintf("filters/%d", collectionID), q, nil, &out)
	return out, err
}

func (c *Client) List(ctx context.Context, opts ListOptions) ([]Raindrop, int64, error) {
	q := url.Values{}
	if opts.Search != "" {
		q.Set("search", opts.Search)
	}
	if opts.Sort != "" {
		q.Set("sort", opts.Sort)
	}
	if opts.Page > 0 {
		q.Set("page", strconv.Itoa(opts.Page))
	}
	if opts.PerPage > 0 {
		q.Set("perpage", strconv.Itoa(opts.PerPage))
	}
	if opts.Nested {
		q.Set("nested", "true")
	}
	var out raindropsResponse
	err := c.Do(ctx, http.MethodGet, fmt.Sprintf("raindrops/%d", opts.CollectionID), q, nil, &out)
	return out.Items, out.Count, err
}

func (c *Client) Get(ctx context.Context, id int64) (Raindrop, error) {
	var out raindropResponse
	err := c.Do(ctx, http.MethodGet, fmt.Sprintf("raindrop/%d", id), nil, nil, &out)
	return out.Item, err
}

func (c *Client) Create(ctx context.Context, body map[string]any) (Raindrop, error) {
	var out raindropResponse
	err := c.Do(ctx, http.MethodPost, "raindrop", nil, body, &out)
	return out.Item, err
}

func (c *Client) Update(ctx context.Context, id int64, body map[string]any) (Raindrop, error) {
	var out raindropResponse
	err := c.Do(ctx, http.MethodPut, fmt.Sprintf("raindrop/%d", id), nil, body, &out)
	return out.Item, err
}

func (c *Client) Delete(ctx context.Context, id int64) (json.RawMessage, error) {
	return c.Raw(ctx, http.MethodDelete, fmt.Sprintf("raindrop/%d", id), nil, nil)
}

func (c *Client) BatchUpdate(ctx context.Context, collectionID int64, body map[string]any) (json.RawMessage, error) {
	return c.Raw(ctx, http.MethodPut, fmt.Sprintf("raindrops/%d", collectionID), nil, body)
}

func (c *Client) BatchDelete(ctx context.Context, collectionID int64, search string, ids []int64) (json.RawMessage, error) {
	q := url.Values{}
	if search != "" {
		q.Set("search", search)
	}
	body := map[string]any{}
	if len(ids) > 0 {
		body["ids"] = ids
	}
	return c.Raw(ctx, http.MethodDelete, fmt.Sprintf("raindrops/%d", collectionID), q, body)
}

func (c *Client) Suggest(ctx context.Context, link string) (json.RawMessage, error) {
	return c.Raw(ctx, http.MethodPost, "raindrop/suggest", nil, map[string]any{"link": link})
}

func (c *Client) Cache(ctx context.Context, id int64) ([]byte, string, error) {
	return c.RawBytes(ctx, http.MethodGet, fmt.Sprintf("raindrop/%d/cache", id), nil, nil)
}

func (c *Client) Highlights(ctx context.Context, collectionID int64, page, perPage int) (json.RawMessage, error) {
	q := url.Values{}
	if page > 0 {
		q.Set("page", strconv.Itoa(page))
	}
	if perPage > 0 {
		q.Set("perpage", strconv.Itoa(perPage))
	}
	path := "highlights"
	if collectionID != 0 {
		path = fmt.Sprintf("highlights/%d", collectionID)
	}
	return c.Raw(ctx, http.MethodGet, path, q, nil)
}

func (c *Client) ParseURL(ctx context.Context, link string) (json.RawMessage, error) {
	q := url.Values{"url": []string{link}}
	return c.Raw(ctx, http.MethodGet, "import/url/parse", q, nil)
}

func (c *Client) URLsExist(ctx context.Context, urls []string) (json.RawMessage, error) {
	return c.Raw(ctx, http.MethodPost, "import/url/exists", nil, map[string]any{"urls": urls})
}

func (c *Client) Export(ctx context.Context, collectionID int64, format, search, sort string) ([]byte, string, error) {
	q := url.Values{}
	if search != "" {
		q.Set("search", search)
	}
	if sort != "" {
		q.Set("sort", sort)
	}
	return c.RawBytes(ctx, http.MethodGet, fmt.Sprintf("raindrops/%d/export.%s", collectionID, format), q, nil)
}

func (c *Client) Sharing(ctx context.Context, collectionID int64) (json.RawMessage, error) {
	return c.Raw(ctx, http.MethodGet, fmt.Sprintf("collection/%d/sharing", collectionID), nil, nil)
}

func (c *Client) Share(ctx context.Context, collectionID int64, emails []string, role string) (json.RawMessage, error) {
	return c.Raw(ctx, http.MethodPost, fmt.Sprintf("collection/%d/sharing", collectionID), nil, map[string]any{
		"emails": emails,
		"role":   role,
	})
}

func (c *Client) Unshare(ctx context.Context, collectionID int64) (json.RawMessage, error) {
	return c.Raw(ctx, http.MethodDelete, fmt.Sprintf("collection/%d/sharing", collectionID), nil, nil)
}

func (c *Client) UpdateCollaborator(ctx context.Context, collectionID, userID int64, role string) (json.RawMessage, error) {
	return c.Raw(ctx, http.MethodPut, fmt.Sprintf("collection/%d/sharing/%d", collectionID, userID), nil, map[string]any{"role": role})
}

func (c *Client) RemoveCollaborator(ctx context.Context, collectionID, userID int64) (json.RawMessage, error) {
	return c.Raw(ctx, http.MethodDelete, fmt.Sprintf("collection/%d/sharing/%d", collectionID, userID), nil, nil)
}

func (c *Client) JoinCollection(ctx context.Context, collectionID int64, token string) (json.RawMessage, error) {
	return c.Raw(ctx, http.MethodPost, fmt.Sprintf("collection/%d/join", collectionID), nil, map[string]any{"token": token})
}

func (c *Client) Backups(ctx context.Context) (json.RawMessage, error) {
	return c.Raw(ctx, http.MethodGet, "backups", nil, nil)
}

func (c *Client) CreateBackup(ctx context.Context) ([]byte, string, error) {
	return c.RawBytes(ctx, http.MethodGet, "backup", nil, nil)
}

func (c *Client) DownloadBackup(ctx context.Context, id, format string) ([]byte, string, error) {
	return c.RawBytes(ctx, http.MethodGet, fmt.Sprintf("backup/%s.%s", url.PathEscape(id), format), nil, nil)
}

func (c *Client) UploadRaindropFile(ctx context.Context, filePath string, collectionID int64) (json.RawMessage, error) {
	fields := map[string]string{}
	if collectionID != 0 {
		fields["collectionId"] = strconv.FormatInt(collectionID, 10)
	}
	return c.Multipart(ctx, http.MethodPut, "raindrop/file", nil, "file", filePath, fields)
}

func (c *Client) UploadRaindropCover(ctx context.Context, id int64, filePath string) (json.RawMessage, error) {
	return c.Multipart(ctx, http.MethodPut, fmt.Sprintf("raindrop/%d/cover", id), nil, "cover", filePath, nil)
}

func (c *Client) UploadCollectionCover(ctx context.Context, id int64, filePath string) (json.RawMessage, error) {
	return c.Multipart(ctx, http.MethodPut, fmt.Sprintf("collection/%d/cover", id), nil, "cover", filePath, nil)
}

func (c *Client) ImportFile(ctx context.Context, filePath string) (json.RawMessage, error) {
	return c.Multipart(ctx, http.MethodPost, "import/file", nil, "import", filePath, nil)
}
