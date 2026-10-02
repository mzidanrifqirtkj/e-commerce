package storage

import (
	"os"
	"strings"
	"testing"

	"user-service/config"

	"github.com/spf13/viper"
)

// TestUploadFileLive uploads a small object to the real bucket and reads it back
// through its public URL, which is the only way to prove the URL, key and bucket
// in .env actually work together.
//
// It writes to the bucket, so it only runs when explicitly asked for:
//
//	RUN_LIVE_STORAGE=1 go test ./internal/adapter/storage/ -run TestUploadFileLive -v
func TestUploadFileLive(t *testing.T) {
	if os.Getenv("RUN_LIVE_STORAGE") != "1" {
		t.Skip("set RUN_LIVE_STORAGE=1 to hit the real Supabase bucket")
	}

	viper.SetConfigFile("../../../.env")
	if err := viper.ReadInConfig(); err != nil {
		t.Fatalf("cannot read .env: %v", err)
	}

	cfg := config.NewConfig()
	if cfg.Storage.Key == "" || cfg.Storage.URL == "" || cfg.Storage.Bucket == "" {
		t.Fatal("SUPABASE_STORAGE_URL, KEY and BUCKET must all be set")
	}
	if !strings.HasSuffix(cfg.Storage.URL, "/storage/v1") {
		t.Errorf("storage URL %q must end in /storage/v1, the /rest/v1 endpoint serves PostgREST, not Storage", cfg.Storage.URL)
	}

	body := "storage adapter self check\n"
	publicURL, err := NewSupabase(cfg).UploadFile("selfcheck/probe.txt", "text/plain", strings.NewReader(body))
	if err != nil {
		t.Fatalf("upload failed: %v", err)
	}
	t.Logf("uploaded, public URL: %s", publicURL)
}
