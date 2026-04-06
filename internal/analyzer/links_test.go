package analyzer

import (
	"testing"
)

func TestPrepareLinks(t *testing.T) {
	raw := []Link{
		{URL: "https://example.com/a", IsInternal: true},
		{URL: "https://example.com/b", IsInternal: true},
		{URL: "https://example.com/a", IsInternal: true}, // duplicate
		{URL: "https://external.com", IsInternal: false},
		{URL: "https://external.com", IsInternal: false}, // duplicate external
	}

	stats, unique := prepareLinks(raw)

	// raw counts include duplicates
	if stats.Internal != 3 {
		t.Errorf("Internal = %d, want 3", stats.Internal)
	}
	if stats.External != 2 {
		t.Errorf("External = %d, want 2", stats.External)
	}

	// unique links deduplicated
	if len(unique) != 3 {
		t.Errorf("unique links = %d, want 3", len(unique))
	}
}

func TestPrepareLinks_Empty(t *testing.T) {
	stats, unique := prepareLinks([]Link{})

	if stats.Internal != 0 || stats.External != 0 {
		t.Errorf("expected zero stats for empty input, got %+v", stats)
	}
	if len(unique) != 0 {
		t.Errorf("expected empty unique links, got %d", len(unique))
	}
}

func TestPrepareLinks_AllInternal(t *testing.T) {
	raw := []Link{
		{URL: "https://example.com/a", IsInternal: true},
		{URL: "https://example.com/b", IsInternal: true},
		{URL: "https://example.com/c", IsInternal: true},
	}

	stats, unique := prepareLinks(raw)

	if stats.Internal != 3 {
		t.Errorf("Internal = %d, want 3", stats.Internal)
	}
	if stats.External != 0 {
		t.Errorf("External = %d, want 0", stats.External)
	}
	if len(unique) != 3 {
		t.Errorf("unique = %d, want 3", len(unique))
	}
}

func TestPrepareLinks_AllDuplicates(t *testing.T) {
	raw := []Link{
		{URL: "https://example.com/a", IsInternal: true},
		{URL: "https://example.com/a", IsInternal: true},
		{URL: "https://example.com/a", IsInternal: true},
	}

	stats, unique := prepareLinks(raw)

	// raw count includes all duplicates
	if stats.Internal != 3 {
		t.Errorf("Internal = %d, want 3", stats.Internal)
	}

	// but unique should only have one
	if len(unique) != 1 {
		t.Errorf("unique = %d, want 1", len(unique))
	}
}
