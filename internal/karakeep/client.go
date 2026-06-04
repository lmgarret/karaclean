package karakeep

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/lmgarret/karaclean/internal/engine"
	"github.com/oapi-codegen/oapi-codegen/v2/pkg/securityprovider"
)

// KarakeepClient wraps the oapi-codegen generated ClientWithResponses and implements engine.KarakeepAPI.
type KarakeepClient struct {
	inner *ClientWithResponses
}

// Compile-time proof that KarakeepClient satisfies the interface.
var _ engine.KarakeepAPI = (*KarakeepClient)(nil)

// NewKarakeepClient constructs a KarakeepClient with bearer auth pointed at baseURL+"/api/v1".
func NewKarakeepClient(baseURL, apiKey string) (*KarakeepClient, error) {
	bearerAuth, err := securityprovider.NewSecurityProviderBearerToken(apiKey)
	if err != nil {
		return nil, fmt.Errorf("creating auth provider: %w", err)
	}
	inner, err := NewClientWithResponses(
		baseURL+"/api/v1",
		WithRequestEditorFn(bearerAuth.Intercept),
	)
	if err != nil {
		return nil, fmt.Errorf("creating API client: %w", err)
	}
	return &KarakeepClient{inner: inner}, nil
}

// CheckAuth validates the API token by calling GET /users/me.
func (c *KarakeepClient) CheckAuth(ctx context.Context) error {
	resp, err := c.inner.GetCurrentUserWithResponse(ctx)
	if err != nil {
		return fmt.Errorf("auth check: %w", err)
	}
	switch resp.StatusCode() {
	case http.StatusOK:
		return nil
	case http.StatusUnauthorized:
		return fmt.Errorf("authentication failed: invalid API token (check KARAKEEP_API_KEY)")
	default:
		return fmt.Errorf("auth check: unexpected status %d", resp.StatusCode())
	}
}

// ListBookmarks retrieves all bookmarks using cursor-based pagination.
func (c *KarakeepClient) ListBookmarks(ctx context.Context) ([]engine.Bookmark, error) {
	var all []engine.Bookmark
	var cursor *Cursor

	for {
		limit := float32(100)
		resp, err := c.inner.ListBookmarksWithResponse(ctx, &ListBookmarksParams{
			Cursor: cursor,
			Limit:  &limit,
		})
		if err != nil {
			return nil, fmt.Errorf("listing bookmarks: %w", err)
		}
		if resp.StatusCode() != http.StatusOK {
			return nil, fmt.Errorf("listing bookmarks: unexpected status %d", resp.StatusCode())
		}

		for _, b := range resp.JSON200.Bookmarks {
			all = append(all, toEngineBookmark(b))
		}

		if resp.JSON200.NextCursor == nil {
			break
		}
		cursor = resp.JSON200.NextCursor
	}

	if all == nil {
		all = []engine.Bookmark{}
	}
	return all, nil
}

// ArchiveBookmark sets archived=true on the bookmark via PATCH /bookmarks/{id}.
func (c *KarakeepClient) ArchiveBookmark(ctx context.Context, id string) error {
	archived := true
	resp, err := c.inner.UpdateBookmarkWithResponse(ctx, id, UpdateBookmarkJSONRequestBody{
		Archived: &archived,
	})
	if err != nil {
		return fmt.Errorf("archive bookmark %s: %w", id, err)
	}
	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("archive bookmark %s: unexpected status %d", id, resp.StatusCode())
	}
	return nil
}

// UnarchiveBookmark sets archived=false on the bookmark via PATCH /bookmarks/{id}.
func (c *KarakeepClient) UnarchiveBookmark(ctx context.Context, id string) error {
	archived := false
	resp, err := c.inner.UpdateBookmarkWithResponse(ctx, id, UpdateBookmarkJSONRequestBody{
		Archived: &archived,
	})
	if err != nil {
		return fmt.Errorf("unarchive bookmark %s: %w", id, err)
	}
	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("unarchive bookmark %s: unexpected status %d", id, resp.StatusCode())
	}
	return nil
}

// FavouriteBookmark sets favourited=true on the bookmark via PATCH /bookmarks/{id}.
func (c *KarakeepClient) FavouriteBookmark(ctx context.Context, id string) error {
	return c.setFavourite(ctx, id, true)
}

// UnfavouriteBookmark sets favourited=false on the bookmark via PATCH /bookmarks/{id}.
func (c *KarakeepClient) UnfavouriteBookmark(ctx context.Context, id string) error {
	return c.setFavourite(ctx, id, false)
}

// setFavourite issues a PATCH that sets the bookmark's favourited flag.
func (c *KarakeepClient) setFavourite(ctx context.Context, id string, favourited bool) error {
	verb := "favourite"
	if !favourited {
		verb = "unfavourite"
	}
	resp, err := c.inner.UpdateBookmarkWithResponse(ctx, id, UpdateBookmarkJSONRequestBody{
		Favourited: &favourited,
	})
	if err != nil {
		return fmt.Errorf("%s bookmark %s: %w", verb, id, err)
	}
	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("%s bookmark %s: unexpected status %d", verb, id, resp.StatusCode())
	}
	return nil
}

// AddTagToBookmark attaches a tag by name via POST /bookmarks/{id}/tags.
// Karakeep creates the tag if it does not yet exist.
func (c *KarakeepClient) AddTagToBookmark(ctx context.Context, id, tagName string) error {
	var body AttachTagsToBookmarkJSONRequestBody
	body.Tags = append(body.Tags, struct {
		AttachedBy *AttachTagsToBookmarkJSONBodyTagsAttachedBy `json:"attachedBy,omitempty"`
		TagId      *string                                     `json:"tagId,omitempty"`
		TagName    *string                                     `json:"tagName,omitempty"`
	}{TagName: &tagName})
	resp, err := c.inner.AttachTagsToBookmarkWithResponse(ctx, id, body)
	if err != nil {
		return fmt.Errorf("tag bookmark %s: %w", id, err)
	}
	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("tag bookmark %s: unexpected status %d", id, resp.StatusCode())
	}
	return nil
}

// RemoveTagFromBookmark detaches a tag by name via DELETE /bookmarks/{id}/tags.
func (c *KarakeepClient) RemoveTagFromBookmark(ctx context.Context, id, tagName string) error {
	var body DetachTagsFromBookmarkJSONRequestBody
	body.Tags = append(body.Tags, struct {
		AttachedBy *DetachTagsFromBookmarkJSONBodyTagsAttachedBy `json:"attachedBy,omitempty"`
		TagId      *string                                       `json:"tagId,omitempty"`
		TagName    *string                                       `json:"tagName,omitempty"`
	}{TagName: &tagName})
	resp, err := c.inner.DetachTagsFromBookmarkWithResponse(ctx, id, body)
	if err != nil {
		return fmt.Errorf("untag bookmark %s: %w", id, err)
	}
	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("untag bookmark %s: unexpected status %d", id, resp.StatusCode())
	}
	return nil
}

// DeleteBookmark permanently removes the bookmark via DELETE /bookmarks/{id}.
func (c *KarakeepClient) DeleteBookmark(ctx context.Context, id string) error {
	resp, err := c.inner.DeleteBookmarkWithResponse(ctx, id)
	if err != nil {
		return fmt.Errorf("delete bookmark %s: %w", id, err)
	}
	if resp.StatusCode() != http.StatusOK && resp.StatusCode() != http.StatusNoContent {
		return fmt.Errorf("delete bookmark %s: unexpected status %d", id, resp.StatusCode())
	}
	return nil
}

// ListLists retrieves all lists from Karakeep.
func (c *KarakeepClient) ListLists(ctx context.Context) ([]engine.ListInfo, error) {
	resp, err := c.inner.ListListsWithResponse(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing lists: %w", err)
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("listing lists: unexpected status %d", resp.StatusCode())
	}
	lists := make([]engine.ListInfo, 0, len(resp.JSON200.Lists))
	for _, l := range resp.JSON200.Lists {
		lists = append(lists, engine.ListInfo{ID: l.Id, Name: l.Name})
	}
	return lists, nil
}

// GetListBookmarks retrieves all bookmark IDs belonging to a specific list using cursor-based pagination.
func (c *KarakeepClient) GetListBookmarks(ctx context.Context, listID string) ([]string, error) {
	var ids []string
	var cursor *Cursor
	for {
		limit := float32(100)
		resp, err := c.inner.GetListBookmarksWithResponse(ctx, listID, &GetListBookmarksParams{
			Cursor: cursor,
			Limit:  &limit,
		})
		if err != nil {
			return nil, fmt.Errorf("getting list bookmarks: %w", err)
		}
		if resp.StatusCode() != http.StatusOK {
			return nil, fmt.Errorf("getting list bookmarks: unexpected status %d", resp.StatusCode())
		}
		for _, b := range resp.JSON200.Bookmarks {
			ids = append(ids, b.Id)
		}
		if resp.JSON200.NextCursor == nil {
			break
		}
		cursor = resp.JSON200.NextCursor
	}
	if ids == nil {
		ids = []string{}
	}
	return ids, nil
}

// toEngineBookmark maps a generated Bookmark to the engine domain type.
func toEngineBookmark(b Bookmark) engine.Bookmark {
	createdAt, _ := time.Parse(time.RFC3339, b.CreatedAt)

	var source string
	if b.Source != nil {
		source = string(*b.Source)
	}

	tags := make([]string, 0, len(b.Tags))
	for _, t := range b.Tags {
		tags = append(tags, t.Name)
	}

	var note string
	if b.Note != nil {
		note = *b.Note
	}

	var size int64
	if content2, err := b.Content.AsBookmarkContent2(); err == nil && content2.Size != nil {
		size = int64(*content2.Size)
	}

	return engine.Bookmark{
		ID:         b.Id,
		CreatedAt:  createdAt,
		Archived:   b.Archived,
		Favourited: b.Favourited,
		Source:     source,
		Tags:       tags,
		Note:       note,
		Size:       size,
	}
}
