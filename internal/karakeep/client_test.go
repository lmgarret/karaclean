package karakeep_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/lmgarret/karaclean/internal/karakeep"
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

	client, err := karakeep.NewKarakeepClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewKarakeepClient: %v", err)
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

	client, err := karakeep.NewKarakeepClient(srv.URL, "bad-token")
	if err != nil {
		t.Fatalf("NewKarakeepClient: %v", err)
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
	client, err := karakeep.NewKarakeepClient("http://127.0.0.1:1", "any-token")
	if err != nil {
		t.Fatalf("NewKarakeepClient: %v", err)
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

	client, err := karakeep.NewKarakeepClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewKarakeepClient: %v", err)
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

	client, err := karakeep.NewKarakeepClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewKarakeepClient: %v", err)
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

	client, err := karakeep.NewKarakeepClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewKarakeepClient: %v", err)
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

	client, err := karakeep.NewKarakeepClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewKarakeepClient: %v", err)
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

	client, err := karakeep.NewKarakeepClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewKarakeepClient: %v", err)
	}
	_, err = client.ListBookmarks(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestArchiveBookmark_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/bookmarks/bk-123" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), `"archived":true`) {
			t.Errorf("body does not contain archived:true: %s", string(body))
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(sampleBookmark("bk-123"))
	}))
	defer srv.Close()

	client, err := karakeep.NewKarakeepClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewKarakeepClient: %v", err)
	}
	if err := client.ArchiveBookmark(context.Background(), "bk-123"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestArchiveBookmark_ErrorStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	client, err := karakeep.NewKarakeepClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewKarakeepClient: %v", err)
	}
	err = client.ArchiveBookmark(context.Background(), "bk-123")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "archive bookmark") {
		t.Errorf("error %q does not contain 'archive bookmark'", err.Error())
	}
	if !strings.Contains(err.Error(), "unexpected status 500") {
		t.Errorf("error %q does not contain 'unexpected status 500'", err.Error())
	}
}

func TestUnarchiveBookmark_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/bookmarks/bk-123" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), `"archived":false`) {
			t.Errorf("body does not contain archived:false: %s", string(body))
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(sampleBookmark("bk-123"))
	}))
	defer srv.Close()

	client, err := karakeep.NewKarakeepClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewKarakeepClient: %v", err)
	}
	if err := client.UnarchiveBookmark(context.Background(), "bk-123"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUnarchiveBookmark_ErrorStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	client, err := karakeep.NewKarakeepClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewKarakeepClient: %v", err)
	}
	err = client.UnarchiveBookmark(context.Background(), "bk-123")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "unarchive bookmark") {
		t.Errorf("error %q does not contain 'unarchive bookmark'", err.Error())
	}
}

func TestFavouriteBookmark_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), `"favourited":true`) {
			t.Errorf("body does not contain favourited:true: %s", string(body))
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(sampleBookmark("bk-1"))
	}))
	defer srv.Close()

	client, err := karakeep.NewKarakeepClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewKarakeepClient: %v", err)
	}
	if err := client.FavouriteBookmark(context.Background(), "bk-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUnfavouriteBookmark_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), `"favourited":false`) {
			t.Errorf("body does not contain favourited:false: %s", string(body))
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(sampleBookmark("bk-1"))
	}))
	defer srv.Close()

	client, err := karakeep.NewKarakeepClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewKarakeepClient: %v", err)
	}
	if err := client.UnfavouriteBookmark(context.Background(), "bk-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUnfavouriteBookmark_ErrorStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	client, err := karakeep.NewKarakeepClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewKarakeepClient: %v", err)
	}
	err = client.UnfavouriteBookmark(context.Background(), "bk-1")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "unfavourite bookmark") {
		t.Errorf("error %q does not contain 'unfavourite bookmark'", err.Error())
	}
}

func TestAddTagToBookmark_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/bookmarks/bk-1/tags" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), `"tagName":"delete-candidate"`) {
			t.Errorf("body does not contain tagName: %s", string(body))
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"attached": []string{"tag-1"}})
	}))
	defer srv.Close()

	client, err := karakeep.NewKarakeepClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewKarakeepClient: %v", err)
	}
	if err := client.AddTagToBookmark(context.Background(), "bk-1", "delete-candidate"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAddTagToBookmark_ErrorStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	client, err := karakeep.NewKarakeepClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewKarakeepClient: %v", err)
	}
	err = client.AddTagToBookmark(context.Background(), "bk-1", "x")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "tag bookmark") {
		t.Errorf("error %q does not contain 'tag bookmark'", err.Error())
	}
}

func TestRemoveTagFromBookmark_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/bookmarks/bk-1/tags" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), `"tagName":"stale"`) {
			t.Errorf("body does not contain tagName: %s", string(body))
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"detached": []string{"tag-1"}})
	}))
	defer srv.Close()

	client, err := karakeep.NewKarakeepClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewKarakeepClient: %v", err)
	}
	if err := client.RemoveTagFromBookmark(context.Background(), "bk-1", "stale"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRemoveTagFromBookmark_ErrorStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	client, err := karakeep.NewKarakeepClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewKarakeepClient: %v", err)
	}
	err = client.RemoveTagFromBookmark(context.Background(), "bk-1", "stale")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "untag bookmark") {
		t.Errorf("error %q does not contain 'untag bookmark'", err.Error())
	}
}

func TestDeleteBookmark_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/bookmarks/bk-456" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	client, err := karakeep.NewKarakeepClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewKarakeepClient: %v", err)
	}
	if err := client.DeleteBookmark(context.Background(), "bk-456"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteBookmark_Success204(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/bookmarks/bk-789" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	client, err := karakeep.NewKarakeepClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewKarakeepClient: %v", err)
	}
	if err := client.DeleteBookmark(context.Background(), "bk-789"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// listsResponse builds a minimal JSON response matching the ListLists response type.
func listsResponse(lists []map[string]any) map[string]any {
	return map[string]any{
		"lists": lists,
	}
}

func sampleList(id, name string) map[string]any {
	return map[string]any{
		"id":               id,
		"name":             name,
		"icon":             "list",
		"public":           false,
		"hasCollaborators": false,
		"userRole":         "owner",
	}
}

func TestListLists_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/lists" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(listsResponse([]map[string]any{
			sampleList("list-1", "Read Later"),
			sampleList("list-2", "RSS Feeds"),
		}))
	}))
	defer srv.Close()

	client, err := karakeep.NewKarakeepClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewKarakeepClient: %v", err)
	}
	lists, err := client.ListLists(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(lists) != 2 {
		t.Fatalf("expected 2 lists, got %d", len(lists))
	}
	if lists[0].ID != "list-1" || lists[0].Name != "Read Later" {
		t.Errorf("unexpected first list: %+v", lists[0])
	}
	if lists[1].ID != "list-2" || lists[1].Name != "RSS Feeds" {
		t.Errorf("unexpected second list: %+v", lists[1])
	}
}

func TestListLists_ErrorStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	client, err := karakeep.NewKarakeepClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewKarakeepClient: %v", err)
	}
	_, err = client.ListLists(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "listing lists") {
		t.Errorf("error %q does not contain 'listing lists'", err.Error())
	}
}

func TestListLists_NetworkError(t *testing.T) {
	client, err := karakeep.NewKarakeepClient("http://127.0.0.1:1", "any-token")
	if err != nil {
		t.Fatalf("NewKarakeepClient: %v", err)
	}
	_, err = client.ListLists(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "listing lists") {
		t.Errorf("error %q does not contain 'listing lists'", err.Error())
	}
}

func TestGetListBookmarks_SinglePage(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/lists/list-1/bookmarks" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(bookmarksResponse(
			[]map[string]any{sampleBookmark("b1"), sampleBookmark("b2")},
			nil,
		))
	}))
	defer srv.Close()

	client, err := karakeep.NewKarakeepClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewKarakeepClient: %v", err)
	}
	ids, err := client.GetListBookmarks(context.Background(), "list-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ids) != 2 {
		t.Fatalf("expected 2 IDs, got %d", len(ids))
	}
	if ids[0] != "b1" || ids[1] != "b2" {
		t.Errorf("unexpected IDs: %v", ids)
	}
}

func TestGetListBookmarks_Pagination(t *testing.T) {
	reqCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqCount++
		w.Header().Set("Content-Type", "application/json")
		switch reqCount {
		case 1:
			cursor := "page2"
			_ = json.NewEncoder(w).Encode(bookmarksResponse(
				[]map[string]any{sampleBookmark("b1"), sampleBookmark("b2")},
				&cursor,
			))
		default:
			if r.URL.Query().Get("cursor") != "page2" {
				t.Errorf("expected cursor=page2, got %q", r.URL.Query().Get("cursor"))
			}
			_ = json.NewEncoder(w).Encode(bookmarksResponse(
				[]map[string]any{sampleBookmark("b3")},
				nil,
			))
		}
	}))
	defer srv.Close()

	client, err := karakeep.NewKarakeepClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewKarakeepClient: %v", err)
	}
	ids, err := client.GetListBookmarks(context.Background(), "list-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ids) != 3 {
		t.Fatalf("expected 3 IDs across 2 pages, got %d", len(ids))
	}
}

func TestGetListBookmarks_Empty(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(bookmarksResponse(nil, nil))
	}))
	defer srv.Close()

	client, err := karakeep.NewKarakeepClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewKarakeepClient: %v", err)
	}
	ids, err := client.GetListBookmarks(context.Background(), "list-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ids == nil {
		t.Error("expected empty slice, got nil")
	}
	if len(ids) != 0 {
		t.Errorf("expected 0 IDs, got %d", len(ids))
	}
}

func TestGetListBookmarks_ErrorStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	client, err := karakeep.NewKarakeepClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewKarakeepClient: %v", err)
	}
	_, err = client.GetListBookmarks(context.Background(), "list-1")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "getting list bookmarks") {
		t.Errorf("error %q does not contain 'getting list bookmarks'", err.Error())
	}
}

func TestDeleteBookmark_ErrorStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	client, err := karakeep.NewKarakeepClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewKarakeepClient: %v", err)
	}
	err = client.DeleteBookmark(context.Background(), "bk-456")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "delete bookmark") {
		t.Errorf("error %q does not contain 'delete bookmark'", err.Error())
	}
	if !strings.Contains(err.Error(), "unexpected status 500") {
		t.Errorf("error %q does not contain 'unexpected status 500'", err.Error())
	}
}
