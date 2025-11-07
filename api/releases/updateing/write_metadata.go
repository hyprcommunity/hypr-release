package updateing

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/hyprcommunity/hypr-release/api/releases/summaryofversion"
)

// WriteMetaFile : Registry bilgileriyle hyprland-release metadata dosyasını oluşturur veya günceller.
func WriteMetaFile(dotfileName string, versionMain, versionBuild, branch, releaseChannel, commitsBehind string) error {
	d := summaryofversion.GetDotfileByName(dotfileName)
	if d == nil {
		return fmt.Errorf("dotfile not found: %s", dotfileName)
	}

	content := fmt.Sprintf(`# Hyprland Release Metadata
HYPRLAND_DOTFILES_NAME="%s"
HYPRLAND_DOTFILES_AUTHOR="%s"
HYPRLAND_DOTFILES_BRANCH="%s"
HYPRLAND_VERSION_MAIN="%s"
HYPRLAND_VERSION_BUILD="%s"
HYPRLAND_BRANCH="%s"
HYPRLAND_RELEASE_CHANNEL="%s"
HYPRLAND_COMMITS_BEHIND="%s"
HYPRLAND_REMOTE_URL="%s"
HYPRLAND_INSTALL_DATE="%s"
`,
		d.Name,
		d.Author,
		d.Branch,
		versionMain,
		versionBuild,
		branch,
		releaseChannel,
		commitsBehind,
		d.Repo,
		time.Now().Format("2006-01-02 15:04:05"),
	)

	// Kullanıcı dizinine yazılacak varsayılan yol
	home, err := os.UserHomeDir()
	if err != nil {
		home = "/tmp"
	}
	userPath := filepath.Join(home, ".config", "hypr-release", "hyprland-release")
	os.MkdirAll(filepath.Dir(userPath), 0755)

	// Önce kullanıcı dizinine yazmayı dene
	if err := os.WriteFile(userPath, []byte(content), 0644); err == nil {
		fmt.Printf("[hyprrelease] metadata written to %s\n", userPath)
		return nil
	}

	// Eğer başarısız olursa /etc altına yazmayı dene
	systemPath := "/etc/hyprland-release"
	if err := os.WriteFile(systemPath, []byte(content), 0644); err == nil {
		fmt.Printf("[hyprrelease] metadata written to %s\n", systemPath)
		return nil
	}

	return fmt.Errorf("failed to write metadata to either %s or %s", userPath, systemPath)
}
