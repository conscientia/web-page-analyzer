package config

import (
	"testing"
	"time"
)

func TestLoad_FromEnv(t *testing.T) {
	unsetAll(t)

	t.Setenv("PORT", "9090")
	t.Setenv("REQUEST_TIMEOUT", "90s")
	t.Setenv("FETCH_TIMEOUT", "15s")
	t.Setenv("LINK_CHECK_TIMEOUT", "45s")
	t.Setenv("PER_LINK_TIMEOUT", "5s")
	t.Setenv("MAX_WORKERS", "20")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Port != "9090" {
		t.Errorf("Port: want 9090, got %s", cfg.Port)
	}
	if cfg.RequestTimeout != 90*time.Second {
		t.Errorf("RequestTimeout: want 90s, got %s", cfg.RequestTimeout)
	}
	if cfg.MaxWorkers != 20 {
		t.Errorf("MaxWorkers: want 20, got %d", cfg.MaxWorkers)
	}
}

func TestValidate_MaxWorkersBelowOne(t *testing.T) {
	unsetAll(t)
	t.Setenv("MAX_WORKERS", "0")

	_, err := Load()
	if err == nil {
		t.Error("expected error when MAX_WORKERS=0, got nil")
	}
}

func TestValidate_PerLinkTimeoutExceedsLinkCheckTimeout(t *testing.T) {
	unsetAll(t)
	t.Setenv("PER_LINK_TIMEOUT", "35s")
	t.Setenv("LINK_CHECK_TIMEOUT", "30s")

	_, err := Load()
	if err == nil {
		t.Error("expected error when PER_LINK_TIMEOUT >= LINK_CHECK_TIMEOUT")
	}
}

func TestValidate_FetchTimeoutExceedsRequestTimeout(t *testing.T) {
	unsetAll(t)
	t.Setenv("FETCH_TIMEOUT", "70s")
	t.Setenv("REQUEST_TIMEOUT", "60s")

	_, err := Load()
	if err == nil {
		t.Error("expected error when FETCH_TIMEOUT >= REQUEST_TIMEOUT")
	}
}

func TestValidate_FetchPlusLinkCheckExceedsRequest(t *testing.T) {
	unsetAll(t)
	t.Setenv("REQUEST_TIMEOUT", "60s")
	t.Setenv("FETCH_TIMEOUT", "40s")
	t.Setenv("LINK_CHECK_TIMEOUT", "30s") // 40 + 30 = 70 > 60

	_, err := Load()
	if err == nil {
		t.Error("expected error when FETCH + LINK_CHECK >= REQUEST_TIMEOUT")
	}
}

func TestValidate_MalformedDuration(t *testing.T) {
	unsetAll(t)
	t.Setenv("FETCH_TIMEOUT", "bad-value")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error with malformed duration: %v", err)
	}
	// should have fallen back to default
	if cfg.FetchTimeout != defaultFetchTimeout {
		t.Errorf("expected fallback to default, got %s", cfg.FetchTimeout)
	}
}

// unsetAll clears all env vars.
func unsetAll(t *testing.T) {
	t.Helper()
	keys := []string{
		"PORT",
		"REQUEST_TIMEOUT",
		"FETCH_TIMEOUT",
		"LINK_CHECK_TIMEOUT",
		"PER_LINK_TIMEOUT",
		"MAX_WORKERS",
	}
	for _, k := range keys {
		t.Setenv(k, "")
	}
}
