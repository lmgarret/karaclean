package engine_test

import (
	"bytes"
	"context"
	"errors"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/lmgarret/karaclean/internal/engine"
)

func TestExecuteAction_ArchiveLive(t *testing.T) {
	mock := &mockAPI{}
	result := engine.ExecuteAction(context.Background(), mock, "archive", engine.Bookmark{ID: "bk-1"}, "test-rule", "", false)
	if result.Err != nil {
		t.Fatalf("unexpected error: %v", result.Err)
	}
	if len(mock.archiveBookmarkCalls) != 1 || mock.archiveBookmarkCalls[0] != "bk-1" {
		t.Errorf("expected archiveBookmarkCalls=[bk-1], got %v", mock.archiveBookmarkCalls)
	}
	if result.DryRun {
		t.Error("expected DryRun=false")
	}
}

func TestExecuteAction_UnarchiveLive(t *testing.T) {
	mock := &mockAPI{}
	result := engine.ExecuteAction(context.Background(), mock, "unarchive", engine.Bookmark{ID: "bk-1"}, "test-rule", "", false)
	if result.Err != nil {
		t.Fatalf("unexpected error: %v", result.Err)
	}
	if len(mock.unarchiveBookmarkCalls) != 1 || mock.unarchiveBookmarkCalls[0] != "bk-1" {
		t.Errorf("expected unarchiveBookmarkCalls=[bk-1], got %v", mock.unarchiveBookmarkCalls)
	}
}

func TestExecuteAction_FavouriteLive(t *testing.T) {
	mock := &mockAPI{}
	result := engine.ExecuteAction(context.Background(), mock, "favourite", engine.Bookmark{ID: "bk-1"}, "test-rule", "", false)
	if result.Err != nil {
		t.Fatalf("unexpected error: %v", result.Err)
	}
	if len(mock.favouriteCalls) != 1 || mock.favouriteCalls[0] != "bk-1" {
		t.Errorf("expected favouriteCalls=[bk-1], got %v", mock.favouriteCalls)
	}
}

func TestExecuteAction_UnfavouriteLive(t *testing.T) {
	mock := &mockAPI{}
	result := engine.ExecuteAction(context.Background(), mock, "unfavourite", engine.Bookmark{ID: "bk-1"}, "test-rule", "", false)
	if result.Err != nil {
		t.Fatalf("unexpected error: %v", result.Err)
	}
	if len(mock.unfavouriteCalls) != 1 || mock.unfavouriteCalls[0] != "bk-1" {
		t.Errorf("expected unfavouriteCalls=[bk-1], got %v", mock.unfavouriteCalls)
	}
}

func TestExecuteAction_TagLive(t *testing.T) {
	mock := &mockAPI{}
	result := engine.ExecuteAction(context.Background(), mock, "tag", engine.Bookmark{ID: "bk-1"}, "test-rule", "delete-candidate", false)
	if result.Err != nil {
		t.Fatalf("unexpected error: %v", result.Err)
	}
	if len(mock.addTagCalls) != 1 || mock.addTagCalls[0].id != "bk-1" || mock.addTagCalls[0].tag != "delete-candidate" {
		t.Errorf("expected addTagCalls=[{bk-1 delete-candidate}], got %v", mock.addTagCalls)
	}
}

func TestExecuteAction_UntagLive(t *testing.T) {
	mock := &mockAPI{}
	result := engine.ExecuteAction(context.Background(), mock, "untag", engine.Bookmark{ID: "bk-1"}, "test-rule", "stale", false)
	if result.Err != nil {
		t.Fatalf("unexpected error: %v", result.Err)
	}
	if len(mock.removeTagCalls) != 1 || mock.removeTagCalls[0].id != "bk-1" || mock.removeTagCalls[0].tag != "stale" {
		t.Errorf("expected removeTagCalls=[{bk-1 stale}], got %v", mock.removeTagCalls)
	}
}

func TestExecuteAction_TagDryRun(t *testing.T) {
	mock := &mockAPI{}
	result := engine.ExecuteAction(context.Background(), mock, "tag", engine.Bookmark{ID: "bk-1"}, "test-rule", "review", true)
	if result.Err != nil {
		t.Fatalf("unexpected error: %v", result.Err)
	}
	if len(mock.addTagCalls) != 0 {
		t.Errorf("expected no tag calls in dry-run, got %v", mock.addTagCalls)
	}
	if !result.DryRun {
		t.Error("expected DryRun=true")
	}
}

func TestExecuteAction_TagError(t *testing.T) {
	mock := &mockAPI{addTagErr: errors.New("api down")}
	result := engine.ExecuteAction(context.Background(), mock, "tag", engine.Bookmark{ID: "bk-7"}, "tag-rule", "review", false)
	if result.Err == nil {
		t.Fatal("expected error, got nil")
	}
	for _, want := range []string{"bk-7", "tag-rule", "review"} {
		if !strings.Contains(result.Err.Error(), want) {
			t.Errorf("error %q does not contain %q", result.Err.Error(), want)
		}
	}
}

func TestExecuteAction_TagDryRunLogIncludesTag(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	mock := &mockAPI{}
	_ = engine.ExecuteAction(context.Background(), mock, "tag", engine.Bookmark{ID: "bk-1"}, "rule", "delete-candidate", true)

	output := buf.String()
	for _, want := range []string{"DRY-RUN", "tag", "delete-candidate", "bk-1"} {
		if !strings.Contains(output, want) {
			t.Errorf("log output %q does not contain %q", output, want)
		}
	}
}

func TestExecuteAction_DeleteLive(t *testing.T) {
	mock := &mockAPI{}
	result := engine.ExecuteAction(context.Background(), mock, "delete", engine.Bookmark{ID: "bk-2"}, "test-rule", "", false)
	if result.Err != nil {
		t.Fatalf("unexpected error: %v", result.Err)
	}
	if len(mock.deleteBookmarkCalls) != 1 || mock.deleteBookmarkCalls[0] != "bk-2" {
		t.Errorf("expected deleteBookmarkCalls=[bk-2], got %v", mock.deleteBookmarkCalls)
	}
}

func TestExecuteAction_ArchiveDryRun(t *testing.T) {
	mock := &mockAPI{}
	result := engine.ExecuteAction(context.Background(), mock, "archive", engine.Bookmark{ID: "bk-1"}, "test-rule", "", true)
	if result.Err != nil {
		t.Fatalf("unexpected error: %v", result.Err)
	}
	if len(mock.archiveBookmarkCalls) != 0 {
		t.Errorf("expected no archive calls in dry-run, got %v", mock.archiveBookmarkCalls)
	}
	if !result.DryRun {
		t.Error("expected DryRun=true")
	}
}

func TestExecuteAction_DeleteDryRun(t *testing.T) {
	mock := &mockAPI{}
	result := engine.ExecuteAction(context.Background(), mock, "delete", engine.Bookmark{ID: "bk-2"}, "test-rule", "", true)
	if result.Err != nil {
		t.Fatalf("unexpected error: %v", result.Err)
	}
	if len(mock.deleteBookmarkCalls) != 0 {
		t.Errorf("expected no delete calls in dry-run, got %v", mock.deleteBookmarkCalls)
	}
}

func TestExecuteAction_ArchiveError(t *testing.T) {
	mock := &mockAPI{archiveBookmarkErr: errors.New("api down")}
	result := engine.ExecuteAction(context.Background(), mock, "archive", engine.Bookmark{ID: "bk-1"}, "test-rule", "", false)
	if result.Err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(result.Err.Error(), "bk-1") {
		t.Errorf("error %q does not contain bookmark ID", result.Err.Error())
	}
	if !strings.Contains(result.Err.Error(), "test-rule") {
		t.Errorf("error %q does not contain rule name", result.Err.Error())
	}
}

func TestExecuteAction_DeleteError(t *testing.T) {
	mock := &mockAPI{deleteBookmarkErr: errors.New("api down")}
	result := engine.ExecuteAction(context.Background(), mock, "delete", engine.Bookmark{ID: "bk-2"}, "test-rule", "", false)
	if result.Err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(result.Err.Error(), "bk-2") {
		t.Errorf("error %q does not contain bookmark ID", result.Err.Error())
	}
	if !strings.Contains(result.Err.Error(), "test-rule") {
		t.Errorf("error %q does not contain rule name", result.Err.Error())
	}
}

func TestExecuteAction_UnknownAction(t *testing.T) {
	mock := &mockAPI{}
	result := engine.ExecuteAction(context.Background(), mock, "unknown", engine.Bookmark{ID: "bk-1"}, "test-rule", "", false)
	if result.Err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(result.Err.Error(), "unknown action") {
		t.Errorf("error %q does not contain 'unknown action'", result.Err.Error())
	}
}

func TestExecuteAction_DryRunLogOutput(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	mock := &mockAPI{}
	bk := engine.Bookmark{ID: "bk-99", Source: "web", Tags: []string{"cleanup", "old"}, Size: 1048576}
	_ = engine.ExecuteAction(context.Background(), mock, "archive", bk, "cleanup-rule", "", true)

	output := buf.String()
	for _, want := range []string{"DRY-RUN", "archive", "bk-99", "web", "cleanup", "old", "cleanup-rule", "size=1.0 MB"} {
		if !strings.Contains(output, want) {
			t.Errorf("log output %q does not contain %q", output, want)
		}
	}
}

func TestExecuteAction_LiveLogOutput(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	mock := &mockAPI{}
	bk := engine.Bookmark{ID: "bk-50", Source: "rss", Tags: []string{"news"}, Size: 2097152}
	_ = engine.ExecuteAction(context.Background(), mock, "archive", bk, "rss-rule", "", false)

	output := buf.String()
	for _, want := range []string{"archive", "bk-50", "rss", "news", "rss-rule", "size=2.0 MB"} {
		if !strings.Contains(output, want) {
			t.Errorf("log output %q does not contain %q", output, want)
		}
	}
}

func TestHumanSize(t *testing.T) {
	tests := []struct {
		bytes int64
		want  string
	}{
		{0, "0 B"},
		{500, "500 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
	}
	for _, tt := range tests {
		got := engine.HumanSize(tt.bytes)
		if got != tt.want {
			t.Errorf("HumanSize(%d) = %q, want %q", tt.bytes, got, tt.want)
		}
	}
}

func TestBookmarkSummary_WithSize(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	mock := &mockAPI{}
	bk := engine.Bookmark{ID: "bk-1", Source: "web", Tags: []string{"test"}, Size: 2048}
	_ = engine.ExecuteAction(context.Background(), mock, "archive", bk, "rule", "", true)

	output := buf.String()
	if !strings.Contains(output, "size=2.0 KB") {
		t.Errorf("expected 'size=2.0 KB' in log output, got: %q", output)
	}
}

func TestBookmarkSummary_ZeroSize(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	mock := &mockAPI{}
	bk := engine.Bookmark{ID: "bk-1", Source: "web", Tags: []string{"test"}, Size: 0}
	_ = engine.ExecuteAction(context.Background(), mock, "archive", bk, "rule", "", true)

	output := buf.String()
	if strings.Contains(output, "size=") {
		t.Errorf("expected no 'size=' in log output for zero size, got: %q", output)
	}
}

func TestBookmarkSummary_EmptySource(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	mock := &mockAPI{}
	bk := engine.Bookmark{ID: "bk-1", Source: "", Tags: []string{"test"}}
	_ = engine.ExecuteAction(context.Background(), mock, "archive", bk, "rule", "", true)

	output := buf.String()
	if !strings.Contains(output, "(unknown)") {
		t.Errorf("expected '(unknown)' for empty source, got log: %q", output)
	}
}

func TestBookmarkSummary_EmptyTags(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	mock := &mockAPI{}
	bk := engine.Bookmark{ID: "bk-1", Source: "web", Tags: nil}
	_ = engine.ExecuteAction(context.Background(), mock, "archive", bk, "rule", "", true)

	output := buf.String()
	if !strings.Contains(output, "tags=[]") {
		t.Errorf("expected 'tags=[]' for nil tags, got log: %q", output)
	}
}
