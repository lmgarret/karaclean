package engine

import "context"

// KarakeepAPI defines the subset of Karakeep API operations used by the engine.
// Methods are added as phases require them (Phase 6 adds UpdateBookmark/DeleteBookmark).
type KarakeepAPI interface {
	// CheckAuth validates the API token by calling GET /users/me.
	// Returns nil on success, error on 401 or network failure.
	CheckAuth(ctx context.Context) error

	// ListBookmarks retrieves all bookmarks using cursor-based pagination.
	// Returns the complete list across all pages.
	ListBookmarks(ctx context.Context) ([]Bookmark, error)

	// ArchiveBookmark sets archived=true on the bookmark via PATCH /bookmarks/{id}.
	ArchiveBookmark(ctx context.Context, id string) error

	// DeleteBookmark permanently removes the bookmark via DELETE /bookmarks/{id}.
	DeleteBookmark(ctx context.Context, id string) error

	// ListLists retrieves all lists from Karakeep.
	ListLists(ctx context.Context) ([]ListInfo, error)

	// GetListBookmarks retrieves bookmark IDs belonging to a specific list.
	// Uses cursor-based pagination to fetch all pages.
	GetListBookmarks(ctx context.Context, listID string) ([]string, error)
}
