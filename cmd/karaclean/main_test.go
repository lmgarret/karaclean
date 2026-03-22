package main

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/lm/karaclean/internal/engine"
)

func TestRequireEnv(t *testing.T) {
	t.Run("returns value when env var is set", func(t *testing.T) {
		t.Setenv("TEST_REQUIREENV_VAR", "myvalue")
		got, err := requireEnv("TEST_REQUIREENV_VAR")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "myvalue" {
			t.Errorf("got %q, want %q", got, "myvalue")
		}
	})

	t.Run("returns error when env var is missing", func(t *testing.T) {
		t.Setenv("TEST_REQUIREENV_MISSING", "")
		_, err := requireEnv("TEST_REQUIREENV_MISSING")
		if err == nil {
			t.Fatal("expected error for missing env var, got nil")
		}
		if !containsStr(err.Error(), "TEST_REQUIREENV_MISSING") {
			t.Errorf("error %q does not name the missing variable", err.Error())
		}
	})

	t.Run("covers KARAKEEP_URL missing case (CONF-03a)", func(t *testing.T) {
		t.Setenv("KARAKEEP_URL", "")
		_, err := requireEnv("KARAKEEP_URL")
		if err == nil {
			t.Fatal("expected error for missing KARAKEEP_URL, got nil")
		}
	})

	t.Run("covers KARAKEEP_API_KEY missing case (CONF-03b)", func(t *testing.T) {
		t.Setenv("KARAKEEP_API_KEY", "")
		_, err := requireEnv("KARAKEEP_API_KEY")
		if err == nil {
			t.Fatal("expected error for missing KARAKEEP_API_KEY, got nil")
		}
	})
}

func TestResolveDryRun(t *testing.T) {
	tests := []struct {
		name      string
		flagSet   bool
		flagVal   bool
		envVal    string
		configVal bool
		want      bool
	}{
		{"flag true wins over all", true, true, "", false, true},
		{"flag true wins over env false", true, true, "false", false, true},
		{"flag false wins over env true", true, false, "true", true, false},
		{"env true wins over config false", false, false, "true", false, true},
		{"env 1 is truthy", false, false, "1", false, true},
		{"env false overrides config true", false, false, "false", true, false},
		{"config true as fallback", false, false, "", true, true},
		{"default is false", false, false, "", false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveDryRun(tt.flagSet, tt.flagVal, tt.envVal, tt.configVal)
			if got != tt.want {
				t.Errorf("resolveDryRun(%v, %v, %q, %v) = %v, want %v",
					tt.flagSet, tt.flagVal, tt.envVal, tt.configVal, got, tt.want)
			}
		})
	}
}

// containsStr is a helper to avoid importing strings in test file.
func containsStr(s, substr string) bool {
	return strings.Contains(s, substr)
}

// testAPI is a minimal mock implementing only what validateListNames needs.
type testAPI struct {
	listListsRet []engine.ListInfo
	listListsErr error
}

func (t *testAPI) CheckAuth(ctx context.Context) error                                    { return nil }
func (t *testAPI) ListBookmarks(ctx context.Context) ([]engine.Bookmark, error)           { return nil, nil }
func (t *testAPI) ArchiveBookmark(ctx context.Context, id string) error                   { return nil }
func (t *testAPI) DeleteBookmark(ctx context.Context, id string) error                    { return nil }
func (t *testAPI) ListLists(ctx context.Context) ([]engine.ListInfo, error)               { return t.listListsRet, t.listListsErr }
func (t *testAPI) GetListBookmarks(ctx context.Context, listID string) ([]string, error)  { return nil, nil }

func TestValidateListNames(t *testing.T) {
	api := &testAPI{
		listListsRet: []engine.ListInfo{
			{ID: "1", Name: "Read Later"},
			{ID: "2", Name: "RSS Feeds"},
		},
	}
	err := validateListNames(context.Background(), api, []string{"Read Later", "RSS Feeds"})
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
}

func TestValidateListNames_Missing(t *testing.T) {
	api := &testAPI{
		listListsRet: []engine.ListInfo{
			{ID: "1", Name: "Read Later"},
		},
	}
	err := validateListNames(context.Background(), api, []string{"Read Later", "No Such List", "Also Missing"})
	if err == nil {
		t.Fatal("expected error for missing list names")
	}
	// Verify ALL missing names are reported (D-13)
	msg := err.Error()
	if !strings.Contains(msg, "No Such List") || !strings.Contains(msg, "Also Missing") {
		t.Errorf("error should list all missing names, got: %s", msg)
	}
}

func TestValidateListNames_APIError(t *testing.T) {
	api := &testAPI{
		listListsErr: fmt.Errorf("connection refused"),
	}
	err := validateListNames(context.Background(), api, []string{"Read Later"})
	if err == nil {
		t.Fatal("expected error when ListLists fails")
	}
	if !strings.Contains(err.Error(), "validating list names") {
		t.Errorf("error %q does not contain 'validating list names'", err.Error())
	}
}
