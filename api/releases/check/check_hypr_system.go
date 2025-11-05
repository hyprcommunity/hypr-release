package hyprland

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

type HyprComponent struct {
	Name            string
	Version         string
	Path            string
	RemoteVersion   string
	PackageVersion  string
	UpdateAvailable bool
	Source          string
	PackageSource   string
}

func CheckHyprSystem() ([]HyprComponent, string, error) {
	components := map[string]string{
		"hyprland":  "hyprwm/Hyprland",
		"hyprctl":   "hyprwm/Hyprland",
		"hyprpaper": "hyprwm/hyprpaper",
		"hypridle":  "hyprwm/hypridle",
		"hyprlock":  "hyprwm/hyprlock",
	}
	var results []HyprComponent
	var log bytes.Buffer

	for name, repo := range components {
		pathCmd := exec.Command("which", name)
		pathOut, err := pathCmd.Output()
		if err != nil {
			log.WriteString(fmt.Sprintf("⚠️ %s not found in PATH\n", name))
			continue
		}
		path := strings.TrimSpace(string(pathOut))

		verCmd := exec.Command(name, "--version")
		verOut, _ := verCmd.Output()
		localVer := strings.TrimSpace(string(verOut))
		if localVer == "" {
			localVer = "unknown"
		}

		remoteVer := getRemoteVersion(repo)
		pkgVer, pkgSrc := getPackageManagerVersion(name)

		updateAvailable := false
		if remoteVer != "unknown" && remoteVer != localVer {
			updateAvailable = true
		} else if pkgVer != "unknown" && pkgVer != localVer {
			updateAvailable = true
		}

		results = append(results, HyprComponent{
			Name:            name,
			Version:         localVer,
			Path:            path,
			RemoteVersion:   remoteVer,
			PackageVersion:  pkgVer,
			UpdateAvailable: updateAvailable,
			Source:          repo,
			PackageSource:   pkgSrc,
		})

		if updateAvailable {
			log.WriteString(fmt.Sprintf("⬆️  %s update available: %s → %s (%s)\n", name, localVer, remoteVer, pkgSrc))
		} else {
			log.WriteString(fmt.Sprintf("✅ %s up to date (%s)\n", name, localVer))
		}
	}

	if err := WriteHyprSystemMeta(results); err != nil {
		log.WriteString(fmt.Sprintf("⚠️ system meta write failed: %v\n", err))
	}
	appendReleaseInfo()
	return results, log.String(), nil
}

// github / git fallback
func getRemoteVersion(repo string) string {
	ghCmd := exec.Command("gh", "release", "list", "--repo", repo, "--limit", "1", "--json", "tagName")
	data, err := ghCmd.Output()
	if err == nil {
		var rel []map[string]string
		_ = json.Unmarshal(data, &rel)
		if len(rel) > 0 {
			return rel[0]["tagName"]
		}
	}
	gitCmd := exec.Command("git", "ls-remote", "--tags", fmt.Sprintf("https://github.com/%s.git", repo))
	gitOut, err := gitCmd.Output()
	if err == nil {
		lines := strings.Split(strings.TrimSpace(string(gitOut)), "\n")
		if len(lines) > 0 {
			last := lines[len(lines)-1]
			if idx := strings.LastIndex(last, "/"); idx != -1 {
				return strings.TrimSpace(last[idx+1:])
			}
		}
	}
	return "unknown"
}

// distro paket yöneticilerini kontrol eder
func getPackageManagerVersion(pkg string) (string, string) {
	candidates := map[string][]string{
		"arch":   {"pacman", "-Si", pkg},
		"debian": {"apt-cache", "policy", pkg},
		"fedora": {"dnf", "info", pkg},
		"void":   {"xbps-query", "-R", pkg},
		"gentoo": {"emerge", "-pv", pkg},
	}
	for distro, args := range candidates {
		cmd := exec.Command(args[0], args[1:]...)
		out, err := cmd.Output()
		if err == nil {
			text := string(out)
			if strings.Contains(text, "Version") || strings.Contains(text, "Version :") {
				for _, line := range strings.Split(text, "\n") {
					if strings.Contains(strings.ToLower(line), "version") {
						return strings.TrimSpace(strings.Split(line, ":")[1]), distro
					}
				}
			}
		}
	}
	return "unknown", "none"
}
