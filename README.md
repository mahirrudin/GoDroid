# GoDroid

**Android Security Automation Toolkit**

GoDroid is a cross-platform (Windows, macOS, Linux) command-line tool for Android mobile penetration testing. It automates the setup of a mobile security testing lab with emulator rooting, Frida server installation, and Burp Suite integration.

Inspired by [BrutDroid](https://github.com/Brut-Security/BrutDroid), rewritten in Go with enhanced cross-platform support and additional features.

## Features

- **Cross-Platform**: Works on Windows, macOS (Intel & Apple Silicon), and Linux
- **No Android Studio Required**: Uses command-line tools only
- **SDK Management**: Automated download and installation of Android SDK components
- **AVD Creation**: Interactive setup of Android Virtual Devices with multiple system images
- **Emulator Rooting**: Magisk and rootAVD integration for one-click rooting
- **Magisk Modules**: Install MagiskFrida (version-matched) and Zygisk SSL Unpinning
- **Frida Integration**: Automated Frida server installation with venv-based frida-tools
- **Certificate Installation**: Burp Suite CA certificate installation with AlwaysTrustUserCerts module
- **Multi-Device Support**: Device selection when multiple emulators/devices connected
- **APK Analyzer**: Static analysis for framework detection, SSL pinning, and root detection libraries
- **Built-in Scripts**: SSL pinning bypass, root detection bypass, and custom script support
- **Go HTTP**: Native HTTP client instead of curl for true cross-platform operation

## Prerequisites

- **Go 1.21+** (for building from source)
- **Python 3.9+** (for frida-tools)
- **Java** (for apktool - APK decompilation)
- **Hardware**: Virtualization enabled (VT-x/AMD-V for Intel/AMD, or Apple Silicon)
- **Internet**: Required for downloading SDK components, Frida, and Magisk

## Installation

### From Source

```bash
git clone https://github.com/mahirrudin/godroid.git
cd godroid
go build -o godroid ./cmd/godroid/
./godroid
```

### Cross-Compile

```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o godroid-linux ./cmd/godroid/

# macOS Intel
GOOS=darwin GOARCH=amd64 go build -o godroid-darwin ./cmd/godroid/

# macOS Apple Silicon
GOOS=darwin GOARCH=arm64 go build -o godroid-darwin-arm64 ./cmd/godroid/

# Windows
GOOS=windows GOARCH=amd64 go build -o godroid.exe ./cmd/godroid/
```

## Usage

Run GoDroid:

```bash
./godroid
```

### Main Menu Options

1. **Environment Check** - Verify all prerequisites (SDK, ADB, Python, Java)
2. **Environment Setup** - Install SDK, frida-tools, and dependencies
3. **Device Management** - Select and manage connected devices
4. **AVD Management** - Create and manage Android Virtual Devices
5. **AVD Root Emulator** - Root the device using Magisk and rootAVD
6. **Rooted Device Configuration** - Install Frida server and Burp certificate
7. **Frida Management** - Run Frida scripts and manage apps
8. **APK Management** - Analyze APKs for frameworks and security features

### Quick Start Guide

1. **First-time Setup**:
   ```
   Environment Check → Environment Setup → Install Command Line Tools
   → Install Platform Tools → Install Android Emulator
   → Install System Image (select x86_64 for Intel/AMD)
   ```

2. **Create an AVD**:
   ```
   AVD Management → Create New AVD → Select architecture → Select system image
   ```

3. **Root the Emulator**:
   ```
   AVD Root Emulator → Start Emulator → Root Emulator → Download & Install Magisk
   → Patch System Image (rootAVD)
   ```

4. **Configure for Pentesting**:
   ```
   Rooted Device Configuration → Install Frida Server
   Rooted Device Configuration → Install Burp Certificate
   Rooted Device Configuration → Install MagiskFrida (for Frida as Magisk module)
   Rooted Device Configuration → Install Zygisk SSL Unpinning
   ```

5. **Run Frida Scripts**:
   ```
   Frida Management → Run SSL Pinning Bypass → Enter package name
   ```

6. **Analyze APK**:
   ```
   APK Management → Analyze APK File → Enter path to APK
   ```

## APK Analyzer

Static analysis of Android APKs without running them:

### Detection Capabilities

| Category | Detected Items |
|----------|----------------|
| **Frameworks** | Flutter, React Native, Ionic, Cordova, Xamarin, Unity, NativeScript, Qt, Kotlin Multiplatform |
| **SSL Pinning** | OkHttp3, TrustManager, Conscrypt, Trustkit, WebViewClient, Flutter plugins, IBM WorkLight, Netty, Cronet (26 patterns) |
| **Root Detection** | RootBeer, Jail Monkey, SafetyNet, Play Integrity, Magisk, Frida, Xposed detection (10 libraries) |

### APK Analyzer Menu

1. **Apktool Status** - Check/install apktool and Java (auto-updates from GitHub)
2. **Detect Framework Only** - Quick scan without decompilation
3. **Analyze APK File** - Full analysis (requires apktool)
4. **Manage Decompiled Files** - View and delete decompiled APKs

### Requirements

- **Framework detection**: No dependencies (reads APK as ZIP)
- **SSL/Root detection**: Requires Java + apktool (auto-downloaded)

## System Images

Default system images available (API 34 recommended for rooting):

| Android Version | API Level | Architectures |
|-----------------|-----------|---------------|
| Android 11      | 30        | x86_64, arm64-v8a |
| Android 12      | 31        | x86_64, arm64-v8a |
| Android 13      | 33        | x86_64, arm64-v8a |
| Android 14      | 34        | x86_64, arm64-v8a |

**Recommended**: Android 14 (API 34) with Google APIs for best rooting compatibility.

## Configuration

Configuration is stored in `~/.godroid/config.json`:

```json
{
  "sdk_path": "/home/user/Android/Sdk",
  "frida_venv_path": "/home/user/.godroid/venv",
  "scripts_dir": "/home/user/.godroid/scripts",
  "default_avd": "GoDroid_Pixel6_API34",
  "default_device": "pixel_6",
  "default_arch": "x86_64",
  "burp_address": "127.0.0.1:8080"
}
```

## Built-in Frida Scripts

- **BYPASS-SSLPINNING.js** - SSL certificate pinning bypass
- **BYPASS-ROOT.js** - Root detection bypass
- **BYPASS-ROOT-SSLPINNING.js** - Combined SSL + root detection bypass

Custom scripts can be added via the menu or by placing `.js` files in `~/.godroid/scripts/`.

## Rooting Process

GoDroid uses the same proven rooting method as BrutDroid:

1. **Install Magisk APK** on the emulator
2. **Patch system image** using rootAVD
3. **Cold boot** the emulator
4. **Complete Magisk setup** in the app

This gives you a fully rooted emulator that passes SafetyNet for most apps.

## Magisk Modules

GoDroid can install these Magisk modules automatically:

| Module | Description |
|--------|-------------|
| **MagiskFrida** | Runs Frida server as Magisk module (version-matched with frida-tools) |
| **Zygisk SSL Unpinning** | System-wide SSL unpinning via Zygisk |
| **AlwaysTrustUserCerts** | Trust user CA certificates system-wide |

## Troubleshooting

### Emulator Not Starting
- Ensure virtualization is enabled in BIOS
- Check that AVD was created with correct architecture for your CPU
- Try cold boot: `AVD Management → Start Emulator → Yes (cold boot)`

### Rooting Fails
- Verify ANDROID_HOME environment variable is set
- Ensure you're using API 31 with Google APIs (not Play Store)
- Check rootAVD output for specific errors

### Frida Connection Issues
- Verify Frida server is running: `Rooted Device Configuration → Start Frida Server`
- Check frida-tools version matches server: `frida --version`
- Ensure emulator is rooted: `AVD Root Emulator → Verify Root Access`

### Certificate Not Trusted
- Reboot emulator after installing certificate
- For Android 7+, use AlwaysTrustUserCerts Magisk module
- Verify Burp Suite is running on correct address

## Credits

- [BrutDroid](https://github.com/Brut-Security/BrutDroid) - Original inspiration
- [Magisk](https://github.com/topjohnwu/Magisk) - Rooting solution
- [rootAVD](https://gitlab.com/newbit/rootAVD) - Emulator patching
- [Frida](https://github.com/frida/frida) - Dynamic instrumentation
- [AlwaysTrustUserCerts](https://github.com/NVISOsecurity/AlwaysTrustUserCerts) - Certificate trust module
- [Apktool](https://github.com/iBotPeaches/Apktool) - APK decompilation
- [smali-sslpin-patterns](https://github.com/AncW/smali-sslpin-patterns) - SSL pinning detection patterns
- [APK-FiD](https://github.com/AncW/APK-FiD) - Framework detection reference
- [RootBeer](https://github.com/scottyab/rootbeer) - Root detection library patterns
- [Jail Monkey](https://github.com/GantMan/jail-monkey) - React Native root detection

## License

MIT License - See [LICENSE](LICENSE) file for details.

---

**Hack smart. Stay safe. 🔐**
