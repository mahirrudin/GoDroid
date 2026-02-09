package app

import (
	"fmt"
	"os"

	"github.com/mahirrudin/godroid/internal/frida"
	"github.com/mahirrudin/godroid/internal/ui"
)

func (h *Handler) RunFridaTools() error {
	if err := h.EnsureDevice(); err != nil {
		return err
	}

	menu := ui.NewMenu("Frida Tools")
	menu.AddWithDesc("List Installed Apps", "Show all apps on the emulator")
	menu.AddWithDesc("Run SSL Pinning Bypass", "Bypass SSL certificate pinning")
	menu.AddWithDesc("Run Root Detection Bypass", "Bypass root/emulator detection")
	menu.AddWithDesc("Run Custom Script", "Execute a custom Frida script")
	menu.AddWithDesc("Add Custom Script", "Add a new Frida script")
	menu.AddWithDesc("List Scripts", "Show available Frida scripts")

	choice := menu.Display()
	if choice < 0 {
		return nil
	}

	scriptMgr := frida.NewScriptManager(h.Config.ScriptsDir, h.Config.FridaVenvPath, h.EmuCtrl)

	switch choice {
	case 0:
		return h.RunListApps(scriptMgr)
	case 1:
		return h.RunFridaScript(scriptMgr, "BYPASS-SSLPINNING.js")
	case 2:
		return h.RunFridaScript(scriptMgr, "BYPASS-ROOT.js")
	case 3:
		return h.RunCustomScript(scriptMgr)
	case 4:
		return h.RunAddCustomScript(scriptMgr)
	case 5:
		return h.RunListScripts(scriptMgr)
	}

	return nil
}

func (h *Handler) RunListApps(scriptMgr *frida.ScriptManager) error {
	apps, err := scriptMgr.ListInstalledApps()
	if err != nil {
		return err
	}

	ui.Box("Installed Applications")
	ui.NewLine()

	for _, app := range apps {
		status := ""
		if app.PID > 0 {
			status = fmt.Sprintf(" (PID: %d)", app.PID)
		}
		ui.Println("  %s%s", app.Identifier, status)
		if app.Name != "" {
			ui.ColorPrintln(ui.Dim, "    %s", app.Name)
		}
	}

	return nil
}

func (h *Handler) RunFridaScript(scriptMgr *frida.ScriptManager, scriptName string) error {
	scriptPath := fmt.Sprintf("%s/%s", h.Config.ScriptsDir, scriptName)

	// Check if script exists
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		ui.Warning("Script %s not found. Creating default script...", scriptName)
		if err := h.CreateDefaultScripts(); err != nil {
			return err
		}
	}

	// Get target package
	pkg := ui.GetInput("Enter package name (e.g., com.example.app)")
	if pkg == "" {
		return fmt.Errorf("package name required")
	}

	return scriptMgr.RunScript(scriptPath, pkg)
}

func (h *Handler) RunCustomScript(scriptMgr *frida.ScriptManager) error {
	scripts := scriptMgr.ListScripts()
	if len(scripts) == 0 {
		return fmt.Errorf("no scripts available")
	}

	ui.Println("Available scripts:")
	for i, s := range scripts {
		custom := ""
		if s.IsCustom {
			custom = " (custom)"
		}
		ui.Println("  %d. %s%s", i+1, s.Name, custom)
	}
	ui.NewLine()

	choice := ui.GetInput("Select script")
	var idx int
	fmt.Sscanf(choice, "%d", &idx)
	if idx < 1 || idx > len(scripts) {
		return fmt.Errorf("invalid selection")
	}

	pkg := ui.GetInput("Enter package name")
	if pkg == "" {
		return fmt.Errorf("package name required")
	}

	return scriptMgr.RunScript(scripts[idx-1].Path, pkg)
}

func (h *Handler) RunAddCustomScript(scriptMgr *frida.ScriptManager) error {
	ui.Box("Add Custom Frida Script")
	ui.NewLine()

	name := ui.GetInput("Script name (without .js)")
	if name == "" {
		return fmt.Errorf("name required")
	}

	ui.Println("Enter script code (press Enter twice to finish):")
	ui.NewLine()

	// Read multi-line input
	var lines []string
	var emptyCount int
	for {
		var line string
		fmt.Scanln(&line)
		if line == "" {
			emptyCount++
			if emptyCount >= 2 {
				break
			}
		} else {
			emptyCount = 0
		}
		lines = append(lines, line)
	}

	code := ""
	for _, l := range lines {
		code += l + "\n"
	}

	return scriptMgr.AddCustomScript(name, code)
}

func (h *Handler) RunListScripts(scriptMgr *frida.ScriptManager) error {
	scripts := scriptMgr.ListScripts()

	ui.Box("Available Frida Scripts")
	ui.NewLine()

	if len(scripts) == 0 {
		ui.Warning("No scripts found")
		return nil
	}

	for _, s := range scripts {
		custom := ""
		if s.IsCustom {
			custom = " [Custom]"
		}
		ui.Println("  • %s%s", s.Name, custom)
		ui.ColorPrintln(ui.Dim, "    %s", s.Path)
	}

	return nil
}

// InitDefaultScripts creates the scripts directory and default scripts if they don't exist.
func (h *Handler) InitDefaultScripts() {
	// Create scripts directory
	if err := os.MkdirAll(h.Config.ScriptsDir, 0755); err != nil {
		return // Silently fail on startup
	}

	scripts := map[string]string{
		"BYPASS-SSLPINNING.js":      sslBypassScript,
		"BYPASS-ROOT.js":            rootBypassScript,
		"BYPASS-ROOT-SSLPINNING.js": combinedBypassScript,
	}

	for name, content := range scripts {
		path := fmt.Sprintf("%s/%s", h.Config.ScriptsDir, name)
		// Only create if doesn't exist
		if _, err := os.Stat(path); os.IsNotExist(err) {
			os.WriteFile(path, []byte(content), 0644)
		}
	}
}

func (h *Handler) CreateDefaultScripts() error {
	if err := os.MkdirAll(h.Config.ScriptsDir, 0755); err != nil {
		return err
	}

	scripts := map[string]string{
		"BYPASS-SSLPINNING.js":      sslBypassScript,
		"BYPASS-ROOT.js":            rootBypassScript,
		"BYPASS-ROOT-SSLPINNING.js": combinedBypassScript,
	}

	for name, content := range scripts {
		path := fmt.Sprintf("%s/%s", h.Config.ScriptsDir, name)
		// Overwrite existing scripts
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return err
		}
	}

	ui.Success("Default scripts created in %s", h.Config.ScriptsDir)
	return nil
}

// Default Frida scripts
const sslBypassScript = `// SSL Pinning Bypass Script
Java.perform(function() {
    console.log("[*] SSL Pinning Bypass loaded");
    
    // TrustManager bypass
    var TrustManager = Java.registerClass({
        name: 'com.godroid.TrustManager',
        implements: [Java.use('javax.net.ssl.X509TrustManager')],
        methods: {
            checkClientTrusted: function(chain, authType) {},
            checkServerTrusted: function(chain, authType) {},
            getAcceptedIssuers: function() { return []; }
        }
    });
    
    // SSLContext bypass
    var SSLContext = Java.use('javax.net.ssl.SSLContext');
    SSLContext.init.overload('[Ljavax.net.ssl.KeyManager;', '[Ljavax.net.ssl.TrustManager;', 'java.security.SecureRandom').implementation = function(km, tm, sr) {
        console.log("[+] SSLContext.init() bypassed");
        this.init(km, [TrustManager.$new()], sr);
    };
    
    console.log("[+] SSL Pinning bypass active");
});
`

const rootBypassScript = `// Root Detection Bypass Script
Java.perform(function() {
    console.log("[*] Root Detection Bypass loaded");
    
    // File.exists bypass
    var File = Java.use('java.io.File');
    File.exists.implementation = function() {
        var path = this.getAbsolutePath();
        var rootPaths = ['su', 'Superuser', 'magisk', 'busybox'];
        for (var i = 0; i < rootPaths.length; i++) {
            if (path.indexOf(rootPaths[i]) !== -1) {
                console.log("[+] Blocked root check: " + path);
                return false;
            }
        }
        return this.exists();
    };
    
    // Build.TAGS bypass
    var Build = Java.use('android.os.Build');
    Build.TAGS.value = "release-keys";
    
    console.log("[+] Root detection bypass active");
});
`

const combinedBypassScript = `// Combined SSL Pinning + Root Detection Bypass
Java.perform(function() {
    console.log("[*] Combined Bypass - SSL Pinning + Root Detection loaded");
    
    // SSL Pinning Bypass
    try {
        var TrustManager = Java.registerClass({
            name: 'com.godroid.PintoorTrustManager',
            implements: [Java.use('javax.net.ssl.X509TrustManager')],
            methods: {
                checkClientTrusted: function(chain, authType) {},
                checkServerTrusted: function(chain, authType) {},
                getAcceptedIssuers: function() { return []; }
            }
        });
        
        var SSLContext = Java.use('javax.net.ssl.SSLContext');
        SSLContext.init.overload('[Ljavax.net.ssl.KeyManager;', '[Ljavax.net.ssl.TrustManager;', 'java.security.SecureRandom').implementation = function(km, tm, sr) {
            console.log("[+] SSL Pinning bypassed");
            this.init(km, [TrustManager.$new()], sr);
        };
    } catch(e) {
        console.log("[-] SSL bypass error: " + e);
    }
    
    // Root Detection Bypass
    try {
        var File = Java.use('java.io.File');
        File.exists.implementation = function() {
            var path = this.getAbsolutePath();
            if (path.indexOf('su') !== -1 || path.indexOf('magisk') !== -1) {
                console.log("[+] Root path hidden: " + path);
                return false;
            }
            return this.exists();
        };
    } catch(e) {
        console.log("[-] Root bypass error: " + e);
    }
    
    console.log("[+] PintooR bypass active");
});
`
