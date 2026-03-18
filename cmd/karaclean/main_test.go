package main

import (
	"testing"
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

// containsStr is a helper to avoid importing strings in test file.
func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		func() bool {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
			return false
		}())
}
