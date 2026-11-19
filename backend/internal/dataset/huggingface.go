package dataset

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// HFFileEntry represents an entry in Hugging Face repository tree.
type HFFileEntry struct {
	Type string `json:"type"` // "file" or "directory"
	Path string `json:"path"`
	Size int64  `json:"size"`
	LFS  *struct {
		OID  string `json:"oid"`
		Size int64  `json:"size"`
	} `json:"lfs"`
}

// Client interacts with the Hugging Face dataset repository.
type Client struct {
	httpClient *http.Client
	repoID     string
}

// NewClient creates a new Hugging Face client for the specified dataset repository.
func NewClient(repoID string) *Client {
	if repoID == "" {
		repoID = "FronkonGames/steam-games-dataset"
	}
	return &Client{
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
		repoID: repoID,
	}
}

// DiscoveredParquet contains metadata of discovered Parquet dataset file.
type DiscoveredParquet struct {
	Path        string `json:"path"`
	DownloadURL string `json:"download_url"`
	Size        int64  `json:"size"`
	ExpectedSHA string `json:"expected_sha"`
}

// DiscoverParquetFiles queries the repository metadata to locate valid Parquet data files.
func (c *Client) DiscoverParquetFiles(ctx context.Context) ([]DiscoveredParquet, error) {
	// First check /tree/main/data
	urls := []string{
		fmt.Sprintf("https://huggingface.co/api/datasets/%s/tree/main/data", c.repoID),
		fmt.Sprintf("https://huggingface.co/api/datasets/%s/tree/main", c.repoID),
	}

	var found []DiscoveredParquet
	for _, apiURL := range urls {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
		if err != nil {
			continue
		}
		req.Header.Set("User-Agent", "SteamPublisherObservatory/1.0")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			continue
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			continue
		}

		var entries []HFFileEntry
		if err := json.NewDecoder(resp.Body).Decode(&entries); err != nil {
			resp.Body.Close()
			continue
		}
		resp.Body.Close()

		for _, entry := range entries {
			if entry.Type == "file" && strings.HasSuffix(strings.ToLower(entry.Path), ".parquet") {
				sha := ""
				size := entry.Size
				if entry.LFS != nil {
					sha = entry.LFS.OID
					if entry.LFS.Size > 0 {
						size = entry.LFS.Size
					}
				}
				downloadURL := fmt.Sprintf("https://huggingface.co/datasets/%s/resolve/main/%s", c.repoID, entry.Path)
				found = append(found, DiscoveredParquet{
					Path:        entry.Path,
					DownloadURL: downloadURL,
					Size:        size,
					ExpectedSHA: sha,
				})
			}
		}

		if len(found) > 0 {
			break
		}
	}

	if len(found) == 0 {
		return nil, fmt.Errorf("no Parquet files found in Hugging Face repository %s", c.repoID)
	}

	return found, nil
}

// ProgressCallback reports downloaded bytes and total bytes.
type ProgressCallback func(downloaded, total int64, percent float64)

// DownloadFile downloads a dataset file to targetPath with progress updates and SHA-256 verification.
func (c *Client) DownloadFile(ctx context.Context, downloadURL, targetPath string, onProgress ProgressCallback) (string, error) {
	tmpPath := targetPath + ".tmp"
	defer os.Remove(tmpPath)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create download request: %w", err)
	}
	req.Header.Set("User-Agent", "SteamPublisherObservatory/1.0")

	// Custom client with no timeout for large downloads (managed via context)
	downloadClient := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return nil // allow standard redirects
		},
	}

	resp, err := downloadClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to initiate download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("server returned HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	totalBytes := resp.ContentLength

	out, err := os.Create(tmpPath)
	if err != nil {
		return "", fmt.Errorf("failed to create local file: %w", err)
	}
	defer out.Close()

	hasher := sha256.New()
	writer := io.MultiWriter(out, hasher)

	buf := make([]byte, 64*1024)
	var downloaded int64
	lastReport := time.Now()

	for {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
		}

		n, rErr := resp.Body.Read(buf)
		if n > 0 {
			if _, wErr := writer.Write(buf[:n]); wErr != nil {
				return "", fmt.Errorf("write error: %w", wErr)
			}
			downloaded += int64(n)

			if onProgress != nil && (time.Since(lastReport) > 250*time.Millisecond || rErr == io.EOF) {
				pct := 0.0
				if totalBytes > 0 {
					pct = float64(downloaded) / float64(totalBytes) * 100.0
				}
				onProgress(downloaded, totalBytes, pct)
				lastReport = time.Now()
			}
		}

		if rErr != nil {
			if rErr == io.EOF {
				break
			}
			return "", fmt.Errorf("download read error: %w", rErr)
		}
	}

	out.Close()

	// Atomic rename to final path
	_ = os.MkdirAll(filepath.Dir(targetPath), 0755)
	if err := os.Rename(tmpPath, targetPath); err != nil {
		return "", fmt.Errorf("failed to finalize downloaded file: %w", err)
	}

	checksum := hex.EncodeToString(hasher.Sum(nil))
	return checksum, nil
}
