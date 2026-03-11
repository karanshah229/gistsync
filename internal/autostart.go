package internal

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/karanshah229/gistsync/internal/storage"
)

const (
	macPlistName    = "com.karanshah229.gistsync.plist"
	linuxUnitName   = "gistsync.service"
	windowsShortcut = "gistsync.lnk"
)

// PathProvider abstracts the resolution of autostart file paths
type PathProvider interface {
	GetMacPlistPath() (string, error)
	GetLinuxUnitPath() (string, error)
	GetWindowsShortcutPath() (string, error)
}

type defaultPathProvider struct{}

func (p defaultPathProvider) GetMacPlistPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "Library", "LaunchAgents", macPlistName), nil
}

func (p defaultPathProvider) GetLinuxUnitPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "systemd", "user", linuxUnitName), nil
}

func (p defaultPathProvider) GetWindowsShortcutPath() (string, error) {
	appData := os.Getenv("APPDATA")
	if appData == "" {
		return "", fmt.Errorf("APPDATA environment variable not set")
	}
	return filepath.Join(appData, "Microsoft", "Windows", "Start Menu", "Programs", "Startup", windowsShortcut), nil
}

var currentPathProvider PathProvider = defaultPathProvider{}

// InstallAutostart installs the autostart mechanism for the current OS
func InstallAutostart() error {
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	configDir, err := storage.GetConfigDir()
	if err != nil {
		return fmt.Errorf("failed to get config directory: %w", err)
	}
	logPath := filepath.Join(configDir, "gistsync.log")

	switch runtime.GOOS {
	case "darwin":
		return installMacOS(exe, logPath)
	case "linux":
		return installLinux(exe, logPath)
	case "windows":
		return installWindows(exe)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

// UninstallAutostart removes the autostart mechanism
func UninstallAutostart() error {
	switch runtime.GOOS {
	case "darwin":
		return uninstallMacOS()
	case "linux":
		return uninstallLinux()
	case "windows":
		return uninstallWindows()
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

// IsAutostartEnabled checks if autostart is currently configured
func IsAutostartEnabled() (bool, error) {
	var path string
	var err error

	switch runtime.GOOS {
	case "darwin":
		path, err = currentPathProvider.GetMacPlistPath()
	case "linux":
		path, err = currentPathProvider.GetLinuxUnitPath()
	case "windows":
		path, err = currentPathProvider.GetWindowsShortcutPath()
	default:
		return false, nil
	}

	if err != nil {
		return false, err
	}
	_, err = os.Stat(path)
	return err == nil, nil
}

// --- MacOS Implementation ---

func installMacOS(exe, logPath string) error {
	path, err := currentPathProvider.GetMacPlistPath()
	if err != nil {
		return err
	}

	plist := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.karanshah229.gistsync</string>
    <key>ProgramArguments</key>
    <array>
        <string>%s</string>
        <string>watch</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardOutPath</key>
    <string>%s</string>
    <key>StandardErrorPath</key>
    <string>%s</string>
</dict>
</plist>`, exe, logPath, logPath)

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	if err := os.WriteFile(path, []byte(plist), 0644); err != nil {
		return err
	}

	// Load the agent
	exec.Command("launchctl", "unload", path).Run() // Unload first if exists
	return exec.Command("launchctl", "load", path).Run()
}

func uninstallMacOS() error {
	path, err := currentPathProvider.GetMacPlistPath()
	if err != nil {
		return err
	}
	exec.Command("launchctl", "unload", path).Run()
	return os.Remove(path)
}

// --- Linux Implementation ---

func installLinux(exe, logPath string) error {
	path, err := currentPathProvider.GetLinuxUnitPath()
	if err != nil {
		return err
	}

	unit := fmt.Sprintf(`[Unit]
Description=GistSync Watcher
After=network.target

[Service]
ExecStart=%s watch
Restart=always
RestartSec=10
StandardOutput=append:%s
StandardError=append:%s

[Install]
WantedBy=default.target`, exe, logPath, logPath)

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	if err := os.WriteFile(path, []byte(unit), 0644); err != nil {
		return err
	}

	// Reload systemd and enable service
	exec.Command("systemctl", "--user", "daemon-reload").Run()
	exec.Command("systemctl", "--user", "stop", linuxUnitName).Run()
	return exec.Command("systemctl", "--user", "enable", "--now", linuxUnitName).Run()
}

func uninstallLinux() error {
	path, err := currentPathProvider.GetLinuxUnitPath()
	if err != nil {
		return err
	}
	exec.Command("systemctl", "--user", "disable", "--now", linuxUnitName).Run()
	return os.Remove(path)
}

// --- Windows Implementation ---

func installWindows(exe string) error {
	path, err := currentPathProvider.GetWindowsShortcutPath()
	if err != nil {
		return err
	}

	// PowerShell command to create a shortcut
	psCommand := fmt.Sprintf(`$WshShell = New-Object -ComObject WScript.Shell; $Shortcut = $WshShell.CreateShortcut('%s'); $Shortcut.TargetPath = '%s'; $Shortcut.Arguments = 'watch'; $Shortcut.Save()`, path, exe)
	
	cmd := exec.Command("powershell", "-Command", psCommand)
	return cmd.Run()
}

func uninstallWindows() error {
	path, err := currentPathProvider.GetWindowsShortcutPath()
	if err != nil {
		return err
	}
	return os.Remove(path)
}
