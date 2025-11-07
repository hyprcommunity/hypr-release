package bridge

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/hyprcommunity/hypr-release/api/releases/check"
	"github.com/hyprcommunity/hypr-release/api/releases/summaryofversion"
	"github.com/hyprcommunity/hypr-release/api/releases/updateing"
)

// Bridge GUI ile CLI backend arasındaki soyut katmandır.
// GUI yalnızca bu interface ile konuşur.
type Bridge struct {
	ReleasePath string
}

// NewBridge: Varsayılan bir Bridge oluşturur.
func NewBridge() *Bridge {
	return &Bridge{
		ReleasePath: "/etc/hyprland-release",
	}
}

//
// ──────────────────────────── 1. SYSTEM CHECK ────────────────────────────
//

// SystemInfo: sistem bileşenleri ve log çıktısını döndürür.
func (b *Bridge) SystemInfo() (string, error) {
	components, logText, err := check.CheckHyprSystem()
	if err != nil {
		return "", fmt.Errorf("system check failed: %v", err)
	}

	result := map[string]any{
		"components": components,
		"log":        logText,
	}
	data, _ := json.MarshalIndent(result, "", "  ")
	return string(data), nil
}

//
// ──────────────────────────── 2. DOTFILES REGISTRY ────────────────────────────
//

// DotfileEntry: GUI'de gösterilecek sade model
type DotfileEntry struct {
	Name    string `json:"name"`
	Author  string `json:"author"`
	RepoURL string `json:"repo"`
	Branch  string `json:"branch"`
}

// GetDotfiles: Registry listesini döndürür.
func (b *Bridge) GetDotfiles() ([]DotfileEntry, error) {
	var entries []DotfileEntry
	for _, d := range summaryofversion.Registry {
		entries = append(entries, DotfileEntry{
			Name:    d.Name,
			Author:  d.Author,
			RepoURL: d.Repo,
			Branch:  d.Branch,
		})
	}
	return entries, nil
}

//
// ──────────────────────────── 3. INSTALL / UPDATE ────────────────────────────
//

// InstallDotfile: seçili dotfile’ı kurar.
func (b *Bridge) InstallDotfile(name string) error {
	fmt.Printf("[install] installing %s...\n", name)
	if err := updateing.InstallFromRegistry(name); err != nil {
		return fmt.Errorf("install failed: %v", err)
	}
	return nil
}

// UpdateDotfile: mevcut dotfile’ı günceller.
func (b *Bridge) UpdateDotfile(name string) error {
	fmt.Printf("[update] updating %s...\n", name)
	if err := updateing.UpdateDotfileAndSystem(name); err != nil {
		return fmt.Errorf("update failed: %v", err)
	}
	return nil
}

//
// ──────────────────────────── 4. RELEASE / VERSION ────────────────────────────
//

// CheckRelease: GitHub ve git tag’leri üzerinden versiyon bilgisi alır.
func (b *Bridge) CheckRelease(dotfileName, repoPath string) (string, error) {
	result, logText, err := check.CheckAll(dotfileName, repoPath)
	if err != nil {
		return "", fmt.Errorf("version check failed: %v", err)
	}

	data := map[string]any{
		"version_info": result,
		"log":          logText,
	}
	out, _ := json.MarshalIndent(data, "", "  ")
	return string(out), nil
}

// ExportReleaseJSON: release metadata dosyalarını birleştirip JSON döndürür.
func (b *Bridge) ExportReleaseJSON() (string, error) {
	release := filepath.Join("/etc", "hyprland-release")
	system := filepath.Join("/etc", "hyprland-system-release")

	data := make(map[string]any)
	files := []string{release, system}
	for _, f := range files {
		if _, err := os.Stat(f); err != nil {
			continue
		}
		content, err := os.ReadFile(f)
		if err != nil {
			continue
		}

		// /etc/hyprland-release text dosyası olduğundan JSON parse denemesi korumalı olmalı
		lines := strings.Split(string(content), "\n")
		for _, line := range lines {
			if strings.Contains(line, "=") {
				parts := strings.SplitN(line, "=", 2)
				if len(parts) == 2 {
					key := strings.TrimSpace(parts[0])
					val := strings.Trim(strings.TrimSpace(parts[1]), `"`)
					data[key] = val
				}
			}
		}
	}

	out, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", err
	}
	return string(out), nil
}

//
// ──────────────────────────── 5. UTILITIES ────────────────────────────
//

// RunCommand: CLI komutlarını güvenli şekilde çalıştırır (log amaçlı)
func (b *Bridge) RunCommand(cmd string, args ...string) (string, error) {
	command := exec.Command(cmd, args...)
	out, err := command.CombinedOutput()
	if err != nil {
		return string(out), fmt.Errorf("command failed: %v", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// SaveLog: GUI loglarını /tmp içine kaydeder
func (b *Bridge) SaveLog(content string) error {
	path := filepath.Join(os.TempDir(), fmt.Sprintf("hyprrelease_%d.log", time.Now().Unix()))
	return os.WriteFile(path, []byte(content), 0644)
}
