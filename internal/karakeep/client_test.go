package karakeep_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/lm/karaclean/internal/karakeep"
)

func TestCheckAuth_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/users/me" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte("Unauthorized"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"id": "user1", "email": "a@b.com"})
	}))
	defer srv.Close()

	client, err := karakeep.NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if err := client.CheckAuth(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCheckAuth_Unauthorized(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("Unauthorized"))
	}))
	defer srv.Close()

	client, err := karakeep.NewClient(srv.URL, "bad-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	err = client.CheckAuth(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	const want = "authentication failed: invalid API token (check KARAKEEP_API_KEY)"
	if !strings.Contains(err.Error(), want) {
		t.Errorf("error %q does not contain %q", err.Error(), want)
	}
}

func TestCheckAuth_NetworkError(t *testing.T) {
	client, err := karakeep.NewClient("http://127.0.0.1:1", "any-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	err = client.CheckAuth(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "auth check") {
		t.Errorf("error %q does not contain 'auth check'", err.Error())
	}
}

func TestCheckAuth_UnexpectedStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	client, err := karakeep.NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	err = client.CheckAuth(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "unexpected status 500") {
		t.Errorf("error %q does not contain 'unexpected status 500'", err.Error())
	}
}

// bookmarksResponse builds a minimal JSON response matching the generated PaginatedBookmarks type.
func bookmarksResponse(bookmarks []map[string]any, nextCursor *string) map[string]any {
	resp := map[string]any{
		"bookmarks": bookmarks,
	}
	if nextCursor != nil {
		resp["nextCursor"] = *nextCursor
	} else {
		resp["nextCursor"] = nil
	}
	return resp
}

func sampleBookmark(id string) map[string]any {
	return map[string]any{
		"id":            id,
		"createdAt":     "2024-01-15T10:30:00.000Z",
		"archived":      false,
		"favourited":    false,
		"taggingStatus": "success",
		"note":          "test note",
		"type":          "link",
		"content": map[string]any{
			"type":  "link",
			"url":   "https://example.com",
			"title": "Example",
		},
		"tags": []map[string]any{
			{"id": "tag1", "name": "go"},
			{"id": "tag2", "name": "testing"},
		},
	}
}

func TestListBookmarks_SinglePage(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/bookmarks" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(bookmarksResponse(
			[]map[string]any{sampleBookmark("b1"), sampleBookmark("b2")},
			nil,
		))
	}))
	defer srv.Close()

	client, err := karakeep.NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	results, err := client.ListBookmarks(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 bookmarks, got %d", len(results))
	}
	if results[0].ID != "b1" {
		t.Errorf("expected ID b1, got %s", results[0].ID)
	}
	if len(results[0].Tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(results[0].Tags))
	}
	if results[0].Note != "test note" {
		t.Errorf("expected note 'test note', got %q", results[0].Note)
	}
}

func TestListBookmarks_Pagination(t *testing.T) {
	reqCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqCount++
		w.Header().Set("Content-Type", "application/json")
		switch reqCount {
		case 1:
			cursor := "page2cursor"
			_ = json.NewEncoder(w).Encode(bookmarksResponse(
				[]map[string]any{sampleBookmark("b1"), sampleBookmark("b2")},
				&cursor,
			))
		default:
			if r.URL.Query().Get("cursor") != "page2cursor" {
				t.Errorf("expected cursor=page2cursor, got %q", r.URL.Query().Get("cursor"))
			}
			_ = json.NewEncoder(w).Encode(bookmarksResponse(
				[]map[string]any{sampleBookmark("b3")},
				nil,
			))
		}
	}))
	defer srv.Close()

	client, err := karakeep.NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	results, err := client.ListBookmarks(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 3 {
		t.Errorf("expected 3 bookmarks across 2 pages, got %d", len(results))
	}
}

func TestListBookmarks_Empty(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(bookmarksResponse(nil, nil))
	}))
	defer srv.Close()

	client, err := karakeep.NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	results, err := client.ListBookmarks(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if results == nil {
		t.Error("expected empty slice, got nil")
	}
	if len(results) != 0 {
		t.Errorf("expected 0 bookmarks, got %d", len(results))
	}
}

func TestListBookmarks_ErrorStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	client, err := karakeep.NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	_, err = client.ListBookmarks(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
