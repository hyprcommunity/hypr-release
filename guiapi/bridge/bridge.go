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
    // Config alanları istenirse buraya eklenebilir
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

// SystemInfo: sistem sürümü, kernel, distro, hyprland verilerini döner.
func (b *Bridge) SystemInfo() (map[string]string, error) {
    data, err := check.GetSystemMeta() // Örnek fonksiyon; check paketinde bu tür bir çağrı tanımlanmalı
    if err != nil {
        return nil, fmt.Errorf("system check failed: %v", err)
    }
    return data, nil
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
    for _, d := range dotfiles.Registry {
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
    fmt.Printf("[install] starting installation for %s...\n", name)
    if err := backend.Install(name); err != nil {
        return fmt.Errorf("install failed: %v", err)
    }
    return nil
}

// UpdateDotfile: mevcut dotfile’ı günceller.
func (b *Bridge) UpdateDotfile(name string) error {
    fmt.Printf("[update] checking updates for %s...\n", name)
    if err := backend.Update(name); err != nil {
        return fmt.Errorf("update failed: %v", err)
    }
    return nil
}

//
// ──────────────────────────── 4. RELEASE / VERSION ────────────────────────────
//

// CheckRelease: Git üzerinden Hyprland versiyonunu denetler.
func (b *Bridge) CheckRelease() (string, error) {
    version, commitsBehind, err := check.CheckVersion()
    if err != nil {
        return "", err
    }
    return fmt.Sprintf("%s (behind %d commits)", version, commitsBehind), nil
}

// ExportReleaseJSON: /etc/hyprland-release ve system meta’yı birleştirir, JSON döner.
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
        var part map[string]any
        if err := json.Unmarshal(content, &part); err == nil {
            for k, v := range part {
                data[k] = v
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
