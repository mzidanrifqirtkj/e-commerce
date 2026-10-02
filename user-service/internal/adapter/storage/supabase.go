package storage

import (
	"io"
	"strings"
	"user-service/config"

	"github.com/labstack/gommon/log"
	storage_go "github.com/supabase-community/storage-go"
)

type SupabaseInterface interface {
	UploadFile(filePath string, contentType string, file io.Reader) (string, error)
}

type supabaseStruct struct {
	cfg *config.Config
}

// UploadFile stores the file at filePath inside the configured bucket and
// returns the public URL the frontend can show. filePath is the path inside the
// bucket, e.g. "avatars/<user-id>.png".
func (s *supabaseStruct) UploadFile(filePath string, contentType string, file io.Reader) (string, error) {
	client := storage_go.NewClient(s.cfg.Storage.URL, strings.TrimSpace(s.cfg.Storage.Key), map[string]string{"Content-Type": contentType})

	_, err := client.UploadFile(s.cfg.Storage.Bucket, filePath, file)
	if err != nil {
		log.Errorf("Error uploading file: %v", err)
		return "", err
	}

	result := client.GetPublicUrl(s.cfg.Storage.Bucket, filePath)
	return result.SignedURL, nil
}

func NewSupabase(cfg *config.Config) SupabaseInterface {
	// The client wraps its own http.Client with a connection pool, so it is built
	// once here instead of on every upload. The trailing slash is trimmed because
	// the library concatenates the base URL with "/object/...".
	return &supabaseStruct{
		cfg: cfg,
	}
}
