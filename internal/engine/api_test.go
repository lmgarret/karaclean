package engine_test

import (
	"context"
	"errors"
	"testing"

	"github.com/lm/karaclean/internal/engine"
)

// mockAPI is a test double that implements engine.KarakeepAPI.
// Its existence proves the interface is mockable without importing the karakeep package.
type mockAPI struct {
	checkAuthErr       error
	listBookmarksRet   []engine.Bookmark
	listBookmarksErr   error
	archiveBookmarkErr error
	deleteBookmarkErr  error

	archiveBookmarkCalls []string
	deleteBookmarkCalls  []string

	listListsRet         []engine.ListInfo
	listListsErr         error
	getListBookmarksRet  map[string][]string
	getListBookmarksErr  error
}

func (m *mockAPI) CheckAuth(ctx context.Context) error {
	return m.checkAuthErr
}

func (m *mockAPI) ListBookmarks(ctx context.Context) ([]engine.Bookmark, error) {
	return m.listBookmarksRet, m.listBookmarksErr
}

func (m *mockAPI) ArchiveBookmark(ctx context.Context, id string) error {
	m.archiveBookmarkCalls = append(m.archiveBookmarkCalls, id)
	return m.archiveBookmarkErr
}

func (m *mockAPI) DeleteBookmark(ctx context.Context, id string) error {
	m.deleteBookmarkCalls = append(m.deleteBookmarkCalls, id)
	return m.deleteBookmarkErr
}

func (m *mockAPI) ListLists(ctx context.Context) ([]engine.ListInfo, error) {
	return m.listListsRet, m.listListsErr
}

func (m *mockAPI) GetListBookmarks(ctx context.Context, listID string) ([]string, error) {
	if m.getListBookmarksRet != nil {
		return m.getListBookmarksRet[listID], m.getListBookmarksErr
	}
	return nil, m.getListBookmarksErr
}

// Compile-time proof that mockAPI satisfies the interface.
var _ engine.KarakeepAPI = (*mockAPI)(nil)

func TestMockAPI(t *testing.T) { //nolint:gocyclo // subtest fan-out, each case is trivial
	t.Run("CheckAuth returns configured error", func(t *testing.T) {
		want := errors.New("auth failed")
		api := &mockAPI{checkAuthErr: want}
		if err := api.CheckAuth(context.Background()); err != want {
			t.Errorf("got %v, want %v", err, want)
		}
	})

	t.Run("CheckAuth returns nil on success", func(t *testing.T) {
		api := &mockAPI{}
		if err := api.CheckAuth(context.Background()); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("ListBookmarks returns configured bookmarks", func(t *testing.T) {
		want := []engine.Bookmark{{ID: "b1"}, {ID: "b2"}}
		api := &mockAPI{listBookmarksRet: want}
		got, err := api.ListBookmarks(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != len(want) {
			t.Errorf("got %d bookmarks, want %d", len(got), len(want))
		}
	})

	t.Run("ListBookmarks returns configured error", func(t *testing.T) {
		want := errors.New("list failed")
		api := &mockAPI{listBookmarksErr: want}
		_, err := api.ListBookmarks(context.Background())
		if err != want {
			t.Errorf("got %v, want %v", err, want)
		}
	})

	t.Run("ArchiveBookmark returns nil on success", func(t *testing.T) {
		api := &mockAPI{}
		if err := api.ArchiveBookmark(context.Background(), "bk-1"); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if len(api.archiveBookmarkCalls) != 1 || api.archiveBookmarkCalls[0] != "bk-1" {
			t.Errorf("expected archiveBookmarkCalls=[bk-1], got %v", api.archiveBookmarkCalls)
		}
	})

	t.Run("ArchiveBookmark returns configured error", func(t *testing.T) {
		want := errors.New("archive failed")
		api := &mockAPI{archiveBookmarkErr: want}
		if err := api.ArchiveBookmark(context.Background(), "bk-1"); err != want {
			t.Errorf("got %v, want %v", err, want)
		}
	})

	t.Run("DeleteBookmark returns nil on success", func(t *testing.T) {
		api := &mockAPI{}
		if err := api.DeleteBookmark(context.Background(), "bk-2"); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if len(api.deleteBookmarkCalls) != 1 || api.deleteBookmarkCalls[0] != "bk-2" {
			t.Errorf("expected deleteBookmarkCalls=[bk-2], got %v", api.deleteBookmarkCalls)
		}
	})

	t.Run("DeleteBookmark returns configured error", func(t *testing.T) {
		want := errors.New("delete failed")
		api := &mockAPI{deleteBookmarkErr: want}
		if err := api.DeleteBookmark(context.Background(), "bk-2"); err != want {
			t.Errorf("got %v, want %v", err, want)
		}
	})

	t.Run("ListLists returns configured lists", func(t *testing.T) {
		want := []engine.ListInfo{
			{ID: "list-1", Name: "Read Later"},
			{ID: "list-2", Name: "Favorites"},
		}
		api := &mockAPI{listListsRet: want}
		got, err := api.ListLists(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != len(want) {
			t.Fatalf("got %d lists, want %d", len(got), len(want))
		}
		for i := range want {
			if got[i].ID != want[i].ID || got[i].Name != want[i].Name {
				t.Errorf("list[%d] = %+v, want %+v", i, got[i], want[i])
			}
		}
	})

	t.Run("ListLists returns configured error", func(t *testing.T) {
		want := errors.New("list lists failed")
		api := &mockAPI{listListsErr: want}
		_, err := api.ListLists(context.Background())
		if err != want {
			t.Errorf("got %v, want %v", err, want)
		}
	})

	t.Run("GetListBookmarks returns configured IDs", func(t *testing.T) {
		api := &mockAPI{
			getListBookmarksRet: map[string][]string{
				"list-1": {"bk-1", "bk-2"},
				"list-2": {"bk-3"},
			},
		}
		got, err := api.GetListBookmarks(context.Background(), "list-1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != 2 || got[0] != "bk-1" || got[1] != "bk-2" {
			t.Errorf("got %v, want [bk-1 bk-2]", got)
		}

		got2, err := api.GetListBookmarks(context.Background(), "list-2")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got2) != 1 || got2[0] != "bk-3" {
			t.Errorf("got %v, want [bk-3]", got2)
		}
	})

	t.Run("GetListBookmarks returns configured error", func(t *testing.T) {
		want := errors.New("get list bookmarks failed")
		api := &mockAPI{getListBookmarksErr: want}
		_, err := api.GetListBookmarks(context.Background(), "list-1")
		if err != want {
			t.Errorf("got %v, want %v", err, want)
		}
	})
}
