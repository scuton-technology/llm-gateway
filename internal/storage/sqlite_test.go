package storage

import (
	"path/filepath"
	"testing"
	"time"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()

	store, err := New(filepath.Join(t.TempDir(), "gateway.db"))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	t.Cleanup(func() {
		_ = store.Close()
	})
	return store
}

func TestProviderSettingsAreEncryptedAtRest(t *testing.T) {
	store := newTestStore(t)

	if err := store.SaveProviderSetting(ProviderSetting{
		Provider:  "openai",
		APIKey:    "sk-test-secret",
		IsEnabled: true,
	}); err != nil {
		t.Fatalf("SaveProviderSetting() error = %v", err)
	}

	var rawValue string
	if err := store.db.QueryRow(`SELECT api_key FROM provider_settings WHERE provider = ?`, "openai").Scan(&rawValue); err != nil {
		t.Fatalf("raw query error = %v", err)
	}
	if rawValue == "sk-test-secret" || rawValue == "" {
		t.Fatalf("api_key stored without encryption: %q", rawValue)
	}

	setting, err := store.GetProviderSetting("openai")
	if err != nil {
		t.Fatalf("GetProviderSetting() error = %v", err)
	}
	if setting.APIKey != "sk-test-secret" {
		t.Fatalf("GetProviderSetting().APIKey = %q, want original value", setting.APIKey)
	}
}

func TestSessionsAreStoredHashed(t *testing.T) {
	store := newTestStore(t)

	token, err := store.CreateSession(time.Hour)
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}

	var storedToken string
	if err := store.db.QueryRow(`SELECT token FROM sessions LIMIT 1`).Scan(&storedToken); err != nil {
		t.Fatalf("session query error = %v", err)
	}
	if storedToken == token {
		t.Fatalf("session token stored in plaintext")
	}
	if !store.ValidateSession(token) {
		t.Fatalf("ValidateSession() returned false for a valid token")
	}
}
