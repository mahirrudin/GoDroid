# Changelog

All notable changes to GoDroid will be documented in this file.

## [Unreleased] - 2026-02-09

### Added
- **Magisk Module Installation**: New menu under "Rooted Device Configure"
  - Install MagiskFrida (version-matched with frida-tools)
  - Install Zygisk SSL Unpinning module
- **Device Status Display**: Shows connected devices count and active device in main menu
- **Device Selection Workflow**: Prompts for device selection when multiple devices connected
- **Restart Emulator**: Added to AVD Management menu
- **DisplayInline()**: New menu method that preserves banner and device status

### Fixed
- **Stop Emulator**: Fixed error when multiple devices connected
  - Now uses device serial with `-s` flag
  - Handles expected connection drop when emulator stops
- **Restart Emulator**: Fixed error with multiple devices
  - Now uses ADBCommand with device serial
- **Frida Commands**: All frida commands now target selected device
  - ListInstalledApps, ListRunningProcesses, RunScript, SpawnAndAttach, AttachToRunning

### Changed
- Menu names standardized: "Install Tools" → "Environment Setup"
- All ADB operations now prompt for device selection when multiple devices detected
- Device selection shows current selection with arrow indicator `→`

### Removed
- Duplicate MainMenu code from `internal/ui/menu.go` (using `cmd/godroid/app/mainmenu.go`)
- Removed `adbArgs` helper function - consolidated ADB calls

### Refactored
- `ADBRoot`, `InstallAPK`, `PushFile`, `PullFile` now use `ADBCommand()` (removed ~30 lines)
- Moved app-specific UI from `internal/ui/` to `cmd/godroid/app/environment.go`:
  - `EnvironmentCheckResult`, `DisplayEnvironmentResults`, `SelectArchitecture`
- `internal/ui/` now contains only reusable UI primitives (Menu, colors, prompts)

## [2.0.0] - Initial Release

### Features
- Cross-platform support (Windows, macOS, Linux)
- SDK Management with automated installation
- AVD Creation and management
- Emulator rooting with Magisk and rootAVD
- Frida integration with venv-based frida-tools
- Certificate installation for Burp Suite
- APK Analyzer with framework detection
- Built-in Frida scripts for SSL/root bypass
