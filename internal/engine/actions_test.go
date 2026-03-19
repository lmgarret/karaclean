package engine_test

import (
	"bytes"
	"context"
	"errors"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/lm/karaclean/internal/engine"
)

func TestExecuteAction_ArchiveLive(t *testing.T) {
	mock := &mockAPI{}
	result := engine.ExecuteAction(context.Background(), mock, "archive", engine.Bookmark{ID: "bk-1"}, "test-rule", false)
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

func TestExecuteAction_DeleteLive(t *testing.T) {
	mock := &mockAPI{}
	result := engine.ExecuteAction(context.Background(), mock, "delete", engine.Bookmark{ID: "bk-2"}, "test-rule", false)
	if result.Err != nil {
		t.Fatalf("unexpected error: %v", result.Err)
	}
	if len(mock.deleteBookmarkCalls) != 1 || mock.deleteBookmarkCalls[0] != "bk-2" {
		t.Errorf("expected deleteBookmarkCalls=[bk-2], got %v", mock.deleteBookmarkCalls)
	}
}

func TestExecuteAction_ArchiveDryRun(t *testing.T) {
	mock := &mockAPI{}
	result := engine.ExecuteAction(context.Background(), mock, "archive", engine.Bookmark{ID: "bk-1"}, "test-rule", true)
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
	result := engine.ExecuteAction(context.Background(), mock, "delete", engine.Bookmark{ID: "bk-2"}, "test-rule", true)
	if result.Err != nil {
		t.Fatalf("unexpected error: %v", result.Err)
	}
	if len(mock.deleteBookmarkCalls) != 0 {
		t.Errorf("expected no delete calls in dry-run, got %v", mock.deleteBookmarkCalls)
	}
}

func TestExecuteAction_ArchiveError(t *testing.T) {
	mock := &mockAPI{archiveBookmarkErr: errors.New("api down")}
	result := engine.ExecuteAction(context.Background(), mock, "archive", engine.Bookmark{ID: "bk-1"}, "test-rule", false)
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
	result := engine.ExecuteAction(context.Background(), mock, "delete", engine.Bookmark{ID: "bk-2"}, "test-rule", false)
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
	result := engine.ExecuteAction(context.Background(), mock, "unknown", engine.Bookmark{ID: "bk-1"}, "test-rule", false)
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
	bk := engine.Bookmark{ID: "bk-99", Source: "web", Tags: []string{"cleanup", "old"}}
	_ = engine.ExecuteAction(context.Background(), mock, "archive", bk, "cleanup-rule", true)

	output := buf.String()
	for _, want := range []string{"DRY-RUN", "archive", "bk-99", "web", "cleanup", "old", "cleanup-rule"} {
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
	bk := engine.Bookmark{ID: "bk-50", Source: "rss", Tags: []string{"news"}}
	_ = engine.ExecuteAction(context.Background(), mock, "archive", bk, "rss-rule", false)

	output := buf.String()
	for _, want := range []string{"archive", "bk-50", "rss", "news", "rss-rule"} {
		if !strings.Contains(output, want) {
			t.Errorf("log output %q does not contain %q", output, want)
		}
	}
}

func TestBookmarkSummary_EmptySource(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	mock := &mockAPI{}
	bk := engine.Bookmark{ID: "bk-1", Source: "", Tags: []string{"test"}}
	_ = engine.ExecuteAction(context.Background(), mock, "archive", bk, "rule", true)

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
	_ = engine.ExecuteAction(context.Background(), mock, "archive", bk, "rule", true)

	output := buf.String()
	if !strings.Contains(output, "tags=[]") {
		t.Errorf("expected 'tags=[]' for nil tags, got log: %q", output)
	}
}
