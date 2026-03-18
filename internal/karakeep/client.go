package karakeep

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/lm/karaclean/internal/engine"
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

	return engine.Bookmark{
		ID:         b.Id,
		CreatedAt:  createdAt,
		Archived:   b.Archived,
		Favourited: b.Favourited,
		Source:     source,
		Tags:       tags,
		Note:       note,
	}
}
