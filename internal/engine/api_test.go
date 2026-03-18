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
	checkAuthErr     error
	listBookmarksRet []engine.Bookmark
	listBookmarksErr error
}

func (m *mockAPI) CheckAuth(ctx context.Context) error {
	return m.checkAuthErr
}

func (m *mockAPI) ListBookmarks(ctx context.Context) ([]engine.Bookmark, error) {
	return m.listBookmarksRet, m.listBookmarksErr
}

// Compile-time proof that mockAPI satisfies the interface.
var _ engine.KarakeepAPI = (*mockAPI)(nil)

func TestMockAPI(t *testing.T) {
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
}
