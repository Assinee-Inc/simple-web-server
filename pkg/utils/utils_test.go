package utils

import (
	"strings"
	"testing"
)

func TestGeneratePublicID_Prefix(t *testing.T) {
	prefixes := []string{"ebk_", "fil_", "usr_", "crt_", "cli_", "pur_", "txn_", "sub_"}
	for _, prefix := range prefixes {
		id := GeneratePublicID(prefix)
		if !strings.HasPrefix(id, prefix) {
			t.Errorf("expected prefix %q, got %q", prefix, id)
		}
	}
}

func TestGeneratePublicID_Length(t *testing.T) {
	prefix := "ebk_"
	id := GeneratePublicID(prefix)
	// prefix (4 chars) + 32 hex chars (UUID without hyphens) = 36
	expectedLen := len(prefix) + 32
	if len(id) != expectedLen {
		t.Errorf("expected length %d, got %d for id %q", expectedLen, len(id), id)
	}
}

func TestGeneratePublicID_NoHyphens(t *testing.T) {
	id := GeneratePublicID("ebk_")
	afterPrefix := id[4:]
	if strings.Contains(afterPrefix, "-") {
		t.Errorf("expected no hyphens after prefix, got %q", id)
	}
}

func TestGeneratePublicID_Uniqueness(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 1000; i++ {
		id := GeneratePublicID("ebk_")
		if seen[id] {
			t.Fatalf("duplicate ID generated: %q", id)
		}
		seen[id] = true
	}
}
