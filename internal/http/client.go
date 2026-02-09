// Package http provides HTTP client utilities for downloading files and making requests.
package http

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Client wraps the standard HTTP client with additional functionality.
type Client struct {
	client  *http.Client
	timeout time.Duration
}

// NewClient creates a new HTTP client with the specified timeout.
func NewClient(timeout time.Duration) *Client {
	return &Client{
		client: &http.Client{
			Timeout: timeout,
		},
		timeout: timeout,
	}
}

// DefaultClient returns a client with default settings.
func DefaultClient() *Client {
	return NewClient(30 * time.Second)
}

// DownloadClient returns a client optimized for downloads (longer timeout).
func DownloadClient() *Client {
	return NewClient(10 * time.Minute)
}

// ProgressCallback is called during download with current and total bytes.
type ProgressCallback func(current, total int64)

// Download downloads a file from a URL to the specified destination.
func (c *Client) Download(url, destPath string, progress ProgressCallback) error {
	// Create destination directory if needed
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Create request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set user agent
	req.Header.Set("User-Agent", "GoDroid/1.0")

	// Make request
	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status: %s", resp.Status)
	}

	// Create destination file
	out, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	// Copy with progress if callback provided
	if progress != nil {
		return c.copyWithProgress(out, resp.Body, resp.ContentLength, progress)
	}

	_, err = io.Copy(out, resp.Body)
	return err
}

// copyWithProgress copies data while reporting progress.
func (c *Client) copyWithProgress(dst io.Writer, src io.Reader, total int64, progress ProgressCallback) error {
	buf := make([]byte, 32*1024)
	var written int64

	for {
		nr, readErr := src.Read(buf)
		if nr > 0 {
			nw, writeErr := dst.Write(buf[0:nr])
			if nw > 0 {
				written += int64(nw)
				progress(written, total)
			}
			if writeErr != nil {
				return writeErr
			}
			if nr != nw {
				return io.ErrShortWrite
			}
		}
		if readErr != nil {
			if readErr == io.EOF {
				break
			}
			return readErr
		}
	}
	return nil
}

// DownloadBytes downloads a URL and returns the content as bytes.
func (c *Client) DownloadBytes(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", "GoDroid/1.0")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status: %s", resp.Status)
	}

	return io.ReadAll(resp.Body)
}

// GetJSON fetches a URL and unmarshals the JSON response.
func (c *Client) GetJSON(url string, result interface{}) error {
	data, err := c.DownloadBytes(url)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, result)
}

// GitHubRelease represents a GitHub release.
type GitHubRelease struct {
	TagName    string `json:"tag_name"`
	Name       string `json:"name"`
	Draft      bool   `json:"draft"`
	Prerelease bool   `json:"prerelease"`
	Assets     []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
		Size               int64  `json:"size"`
	} `json:"assets"`
}

// GetLatestGitHubRelease fetches the latest release information for a repository.
func (c *Client) GetLatestGitHubRelease(repo string) (*GitHubRelease, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo)

	var release GitHubRelease
	if err := c.GetJSON(url, &release); err != nil {
		return nil, fmt.Errorf("failed to fetch release info: %w", err)
	}

	return &release, nil
}

// GetGitHubReleaseByTag fetches a specific release by tag.
func (c *Client) GetGitHubReleaseByTag(repo, tag string) (*GitHubRelease, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/tags/%s", repo, tag)

	var release GitHubRelease
	if err := c.GetJSON(url, &release); err != nil {
		return nil, fmt.Errorf("failed to fetch release info: %w", err)
	}

	return &release, nil
}

// FindAsset finds an asset in a release by partial name match.
func (r *GitHubRelease) FindAsset(namePart string) (string, string, error) {
	for _, asset := range r.Assets {
		if strings.Contains(asset.Name, namePart) {
			return asset.Name, asset.BrowserDownloadURL, nil
		}
	}
	return "", "", fmt.Errorf("asset containing '%s' not found", namePart)
}

// Download is a convenience function using the default client.
func Download(url, destPath string, progress ProgressCallback) error {
	return DownloadClient().Download(url, destPath, progress)
}

// DownloadBytes is a convenience function using the default client.
func DownloadBytes(url string) ([]byte, error) {
	return DefaultClient().DownloadBytes(url)
}

// GetJSON is a convenience function using the default client.
func GetJSON(url string, result interface{}) error {
	return DefaultClient().GetJSON(url, result)
}

// GetLatestGitHubRelease is a convenience function using the default client.
func GetLatestGitHubRelease(repo string) (*GitHubRelease, error) {
	return DefaultClient().GetLatestGitHubRelease(repo)
}
