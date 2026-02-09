package app

import (
	"fmt"

	"github.com/mahirrudin/godroid/internal/cert"
	"github.com/mahirrudin/godroid/internal/frida"
	"github.com/mahirrudin/godroid/internal/magisk"
	"github.com/mahirrudin/godroid/internal/root"
	"github.com/mahirrudin/godroid/internal/ui"
)

func (h *Handler) RunRootEmulator() error {
	if err := h.EnsureDevice(); err != nil {
		return err
	}

	menu := ui.NewMenu("AVD Root Emulator")
	menu.AddWithDesc("Download & Install Magisk", "Install Magisk APK on the emulator")
	menu.AddWithDesc("Patch System Image (rootAVD)", "Root the emulator using rootAVD")
	menu.AddWithDesc("Verify Root Access", "Check if the emulator is rooted")
	menu.AddWithDesc("Configure Magisk", "Set up Magisk options")

	choice := menu.Display()
	if choice < 0 {
		return nil
	}

	magisk := root.NewMagisk(h.EmuCtrl)
	rootAVD := root.NewRootAVD(h.Config.SDKPath, h.Config.GetRootAVDDir())

	switch choice {
	case 0:
		return magisk.DownloadAndInstall()
	case 1:
		return rootAVD.InteractiveRoot()
	case 2:
		return magisk.VerifyRoot()
	case 3:
		return magisk.SetupMagisk()
	}

	return nil
}

func (h *Handler) RunConfigureEmulator() error {
	if err := h.EnsureDevice(); err != nil {
		return err
	}

	menu := ui.NewMenu("Rooted Device Configure")
	menu.AddWithDesc("Install Frida Server", "Download and install Frida server on emulator")
	menu.AddWithDesc("Start Frida Server", "Start Frida server on the emulator")
	menu.AddWithDesc("Stop Frida Server", "Stop Frida server on the emulator")
	menu.AddWithDesc("Install AlwaysTrustUserCerts", "Install Magisk module for certificate trust")
	menu.AddWithDesc("Install MagiskFrida", "Install Frida gadget injection module")
	menu.AddWithDesc("Install Zygisk SSL Unpinning", "Disable SSL pinning via Zygisk")
	menu.AddWithDesc("Configure Burp Address", fmt.Sprintf("Set Burp Suite address (current: %s)", h.Config.BurpAddress))
	menu.AddWithDesc("Configure Burp Certificate", "Install Burp Suite CA certificate")
	menu.AddWithDesc("Set Android Proxy", "Configure emulator proxy for Burp Suite")
	menu.AddWithDesc("Clear Android Proxy", "Remove proxy configuration from emulator")
	menu.AddWithDesc("Restart Emulator", "Reboot the running emulator")

	choice := menu.Display()
	if choice < 0 {
		return nil
	}

	fridaServer := frida.NewServer(h.EmuCtrl, h.Config.FridaVenvPath)
	burpCert := cert.NewBurpCert(h.EmuCtrl)
	moduleMgr := magisk.NewModuleManager(h.EmuCtrl, h.Config.FridaVenvPath)

	switch choice {
	case 0:
		return fridaServer.DownloadAndInstall()
	case 1:
		return fridaServer.StartServer()
	case 2:
		return fridaServer.StopServer()
	case 3:
		return burpCert.InstallWithMagiskModule()
	case 4:
		return moduleMgr.InstallMagiskFrida()
	case 5:
		return moduleMgr.InstallZygiskSSLUnpinning()
	case 6:
		return h.RunConfigureBurpAddress()
	case 7:
		burpCert.SetBurpAddress(h.Config.BurpAddress)
		return burpCert.DownloadAndInstall()
	case 8:
		return h.RunSetProxy()
	case 9:
		return h.RunClearProxy()
	case 10:
		return h.RunRestartEmulator()
	}

	return nil
}

func (h *Handler) RunRestartEmulator() error {
	ui.Box("Restart Emulator")
	ui.NewLine()

	if !h.EmuCtrl.IsEmulatorRunning() {
		ui.Warning("No emulator is currently running")
		return nil
	}

	ui.Info("Rebooting emulator...")
	return h.EmuCtrl.RebootEmulator()
}

func (h *Handler) RunSetProxy() error {
	ui.Box("Set Proxy")
	ui.NewLine()

	ui.Warning("This uses 'adb shell settings put global http_proxy' which may not work on all Android versions.")
	ui.Tip("If proxy doesn't work, configure it manually via Settings → Network → Proxy on the emulator.")
	ui.NewLine()

	addr := ui.GetInputDefault("Proxy address", h.Config.BurpAddress)

	var host string
	var port int
	fmt.Sscanf(addr, "%[^:]:%d", &host, &port)

	if port == 0 {
		host = "127.0.0.1"
		port = 8080
	}

	return h.EmuCtrl.SetProxy(host, port)
}

func (h *Handler) RunConfigureBurpAddress() error {
	ui.Box("Configure Burp Address")
	ui.NewLine()

	ui.Println("Current Burp address: %s", h.Config.BurpAddress)
	ui.NewLine()

	addr := ui.GetInputDefault("New Burp address (host:port)", h.Config.BurpAddress)

	if addr == "" {
		addr = "127.0.0.1:8080"
	}

	h.Config.BurpAddress = addr
	if err := h.Config.Save(); err != nil {
		ui.Warning("Failed to save config: %v", err)
	}

	ui.Success("Burp address set to: %s", h.Config.BurpAddress)
	ui.Tip("This address will be used for proxy setup and certificate download.")
	return nil
}

func (h *Handler) RunClearProxy() error {
	ui.Box("Clear Proxy")
	ui.NewLine()

	ui.Info("Removing proxy configuration...")
	return h.EmuCtrl.ClearProxy()
}
