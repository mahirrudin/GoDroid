// Package cert provides Burp Suite certificate installation functionality.
package cert

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mahirrudin/godroid/internal/emulator"
	"github.com/mahirrudin/godroid/internal/http"
	"github.com/mahirrudin/godroid/internal/ui"
)

const (
	// DefaultBurpAddress is the default Burp Suite proxy address.
	DefaultBurpAddress = "127.0.0.1:8080"

	// AlwaysTrustUserCertsURL is the download URL for the Magisk module.
	AlwaysTrustUserCertsURL = "https://github.com/NVISOsecurity/AlwaysTrustUserCerts/releases/download/v1.3/AlwaysTrustUserCerts_v1.3.zip"
)

// BurpCert handles Burp Suite certificate operations.
type BurpCert struct {
	emu        *emulator.Controller
	burpAddr   string
	httpClient *http.Client
}

// NewBurpCert creates a new Burp certificate manager.
func NewBurpCert(emu *emulator.Controller) *BurpCert {
	return &BurpCert{
		emu:        emu,
		burpAddr:   DefaultBurpAddress,
		httpClient: http.DefaultClient(),
	}
}

// SetBurpAddress sets the Burp Suite proxy address.
func (b *BurpCert) SetBurpAddress(addr string) {
	b.burpAddr = addr
}

// DownloadCert downloads the Burp CA certificate.
func (b *BurpCert) DownloadCert() ([]byte, error) {
	ui.Info("Downloading Burp Suite CA certificate...")
	ui.Println("Burp Proxy: http://%s", b.burpAddr)
	ui.Tip("Make sure Burp Suite is running and listening on %s", b.burpAddr)

	url := fmt.Sprintf("http://%s/cert", b.burpAddr)

	certData, err := b.httpClient.DownloadBytes(url)
	if err != nil {
		return nil, fmt.Errorf("failed to download certificate. Is Burp Suite running on %s? Error: %w", b.burpAddr, err)
	}

	ui.Success("Certificate downloaded (%d bytes)", len(certData))
	return certData, nil
}

// ConvertToSystemCert converts a DER certificate to the system format.
// Returns the filename (hash-based) and the PEM content.
func (b *BurpCert) ConvertToSystemCert(derData []byte) (string, []byte, error) {
	ui.Info("Converting certificate to system format...")

	// Parse the certificate
	cert, err := x509.ParseCertificate(derData)
	if err != nil {
		return "", nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	// Get subject hash (OpenSSL-compatible)
	hash := getSubjectHash(cert)
	filename := fmt.Sprintf("%08x.0", hash)

	// Encode as PEM
	pemBlock := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: derData,
	}
	pemData := pem.EncodeToMemory(pemBlock)

	// Add certificate info as comments (like OpenSSL)
	header := fmt.Sprintf("subject=%s\nissuer=%s\n", cert.Subject.String(), cert.Issuer.String())
	fullContent := []byte(header)
	fullContent = append(fullContent, pemData...)

	ui.Success("Certificate converted: %s", filename)
	return filename, fullContent, nil
}

// getSubjectHash calculates the subject hash for the certificate filename.
// This mimics OpenSSL's X509_subject_name_hash.
func getSubjectHash(cert *x509.Certificate) uint32 {
	// Simplified hash - in production, use proper OpenSSL-compatible hash
	// For now, use a simple hash of the subject
	data := cert.RawSubject
	var hash uint32
	for i := 0; i < len(data); i++ {
		hash = hash*31 + uint32(data[i])
	}
	return hash & 0xffffffff
}

// InstallSystemCert installs a certificate to the system store.
func (b *BurpCert) InstallSystemCert(filename string, certData []byte) error {
	ui.Info("Installing certificate to system store...")

	// Check if emulator is running
	if !b.emu.IsEmulatorRunning() {
		return fmt.Errorf("emulator is not running")
	}

	// Check for root
	if !b.emu.HasRoot() {
		return fmt.Errorf("root access required. Please root the emulator first")
	}

	// Create temp file
	tempDir, err := os.MkdirTemp("", "godroid-cert-")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	certPath := filepath.Join(tempDir, filename)
	if err := os.WriteFile(certPath, certData, 0644); err != nil {
		return fmt.Errorf("failed to write certificate: %w", err)
	}

	// Remount system as writable
	ui.Info("Remounting system partition...")
	b.emu.ADBCommand("root")
	b.emu.ADBShell("mount -o rw,remount /system")

	// Push certificate to system store
	remotePath := fmt.Sprintf("/system/etc/security/cacerts/%s", filename)
	if err := b.emu.PushFile(certPath, remotePath); err != nil {
		// Try alternative method with su
		tempRemote := fmt.Sprintf("/data/local/tmp/%s", filename)
		if err := b.emu.PushFile(certPath, tempRemote); err != nil {
			return fmt.Errorf("failed to push certificate: %w", err)
		}

		// Copy with su
		_, err = b.emu.ADBShell(fmt.Sprintf("su -c 'cp %s %s'", tempRemote, remotePath))
		if err != nil {
			return fmt.Errorf("failed to copy certificate: %w", err)
		}
	}

	// Set permissions
	b.emu.ADBShell(fmt.Sprintf("su -c 'chmod 644 %s'", remotePath))

	ui.Success("Certificate installed to %s", remotePath)
	ui.Warning("Reboot the emulator for changes to take effect")
	return nil
}

// InstallWithMagiskModule installs certificate using AlwaysTrustUserCerts Magisk module.
func (b *BurpCert) InstallWithMagiskModule() error {
	ui.Info("Installing AlwaysTrustUserCerts Magisk module...")
	ui.Println("This module automatically trusts user certificates as system certificates")

	// Check if emulator is running
	if !b.emu.IsEmulatorRunning() {
		return fmt.Errorf("emulator is not running")
	}

	// Check for root
	if !b.emu.HasRoot() {
		return fmt.Errorf("root access required")
	}

	// Download the module
	tempDir, err := os.MkdirTemp("", "godroid-atuc-")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	modulePath := filepath.Join(tempDir, "AlwaysTrustUserCerts.zip")

	ui.Info("Downloading module...")
	downloadClient := http.DownloadClient()
	err = downloadClient.Download(AlwaysTrustUserCertsURL, modulePath, func(current, total int64) {
		if total > 0 {
			bar := ui.ProgressBar(current, total, 40)
			ui.ClearLine()
			ui.Print("\r%s", bar)
		}
	})
	if err != nil {
		return fmt.Errorf("failed to download module: %w", err)
	}
	ui.NewLine()

	// Push to emulator
	remotePath := "/data/local/tmp/AlwaysTrustUserCerts.zip"
	if err := b.emu.PushFile(modulePath, remotePath); err != nil {
		return fmt.Errorf("failed to push module: %w", err)
	}

	// Install via Magisk
	ui.Info("Installing module via Magisk...")
	_, err = b.emu.ADBShell(fmt.Sprintf("su -c 'magisk --install-module %s'", remotePath))
	if err != nil {
		return fmt.Errorf("failed to install module: %w", err)
	}

	ui.Success("AlwaysTrustUserCerts module installed")
	ui.Warning("Reboot the emulator for the module to take effect")
	return nil
}

// InstallUserCert installs a certificate to the user certificate store.
func (b *BurpCert) InstallUserCert(derData []byte) error {
	ui.Info("Installing certificate to user store...")

	// Create temp file
	tempDir, err := os.MkdirTemp("", "godroid-cert-")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	certPath := filepath.Join(tempDir, "cacert.der")
	if err := os.WriteFile(certPath, derData, 0644); err != nil {
		return fmt.Errorf("failed to write certificate: %w", err)
	}

	// Push to emulator
	remotePath := "/data/local/tmp/cacert.der"
	if err := b.emu.PushFile(certPath, remotePath); err != nil {
		return fmt.Errorf("failed to push certificate: %w", err)
	}

	// Install using am command
	ui.Info("Installing via Android settings...")
	_, err = b.emu.ADBShell(fmt.Sprintf("am start -n com.android.settings/.security.credentials.InstallCAProfile -a android.intent.action.VIEW -t application/x-x509-ca-cert -d file://%s", remotePath))
	if err != nil {
		return fmt.Errorf("failed to start certificate installer: %w", err)
	}

	ui.Success("Certificate installer launched")
	ui.Tip("Complete the installation in the emulator by following the on-screen prompts")
	return nil
}

// DownloadAndInstall performs a complete certificate installation workflow.
func (b *BurpCert) DownloadAndInstall() error {
	// Download certificate
	derData, err := b.DownloadCert()
	if err != nil {
		return err
	}

	// Convert to system format
	filename, pemData, err := b.ConvertToSystemCert(derData)
	if err != nil {
		return err
	}

	// Try system installation first
	if err := b.InstallSystemCert(filename, pemData); err != nil {
		ui.Warning("System installation failed: %v", err)
		ui.Info("Falling back to user certificate installation...")

		// Try user installation
		return b.InstallUserCert(derData)
	}

	return nil
}
