package http

import (
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	timeout := 5 * time.Second
	client := NewClient(timeout)

	if client == nil {
		t.Fatal("NewClient returned nil")
	}

	if client.timeout != timeout {
		t.Errorf("expected timeout %v, got %v", timeout, client.timeout)
	}

	if client.client == nil {
		t.Error("internal http client is nil")
	}
}

func TestDefaultClient(t *testing.T) {
	client := DefaultClient()

	if client == nil {
		t.Fatal("DefaultClient returned nil")
	}

	expectedTimeout := 30 * time.Second
	if client.timeout != expectedTimeout {
		t.Errorf("expected timeout %v, got %v", expectedTimeout, client.timeout)
	}
}

func TestDownloadClient_Timeout(t *testing.T) {
	client := DownloadClient()

	if client == nil {
		t.Fatal("DownloadClient returned nil")
	}

	expectedTimeout := 10 * time.Minute
	if client.timeout != expectedTimeout {
		t.Errorf("expected timeout %v, got %v", expectedTimeout, client.timeout)
	}
}

func TestGitHubRelease_FindAsset(t *testing.T) {
	release := &GitHubRelease{
		TagName: "v1.0.0",
		Assets: []struct {
			Name               string `json:"name"`
			BrowserDownloadURL string `json:"browser_download_url"`
			Size               int64  `json:"size"`
		}{
			{Name: "app-linux-amd64.tar.gz", BrowserDownloadURL: "https://example.com/linux.tar.gz", Size: 1000},
			{Name: "app-darwin-amd64.tar.gz", BrowserDownloadURL: "https://example.com/darwin.tar.gz", Size: 1100},
			{Name: "app-windows-amd64.zip", BrowserDownloadURL: "https://example.com/windows.zip", Size: 1200},
		},
	}

	// Test finding existing asset
	name, url, err := release.FindAsset("linux")
	if err != nil {
		t.Fatalf("FindAsset failed: %v", err)
	}
	if name != "app-linux-amd64.tar.gz" {
		t.Errorf("expected name 'app-linux-amd64.tar.gz', got '%s'", name)
	}
	if url != "https://example.com/linux.tar.gz" {
		t.Errorf("expected URL 'https://example.com/linux.tar.gz', got '%s'", url)
	}

	// Test finding darwin asset
	name, url, err = release.FindAsset("darwin")
	if err != nil {
		t.Fatalf("FindAsset for darwin failed: %v", err)
	}
	if name != "app-darwin-amd64.tar.gz" {
		t.Errorf("expected darwin asset name, got '%s'", name)
	}
}

func TestGitHubRelease_FindAsset_NotFound(t *testing.T) {
	release := &GitHubRelease{
		TagName: "v1.0.0",
		Assets: []struct {
			Name               string `json:"name"`
			BrowserDownloadURL string `json:"browser_download_url"`
			Size               int64  `json:"size"`
		}{
			{Name: "app-linux.tar.gz", BrowserDownloadURL: "https://example.com/linux.tar.gz", Size: 1000},
		},
	}

	_, _, err := release.FindAsset("freebsd")
	if err == nil {
		t.Error("expected error for non-existent asset")
	}
}

func TestGitHubRelease_FindAsset_Empty(t *testing.T) {
	release := &GitHubRelease{
		TagName: "v1.0.0",
		Assets:  nil,
	}

	_, _, err := release.FindAsset("linux")
	if err == nil {
		t.Error("expected error for empty assets")
	}
}
