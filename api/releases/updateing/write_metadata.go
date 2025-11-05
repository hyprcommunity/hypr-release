package updateing

import (
	"fmt"
	"os"
	"time"

	"hypr-release/internal/dotfiles"
)

// WriteMetaFile : Registry bilgileriyle /etc/hyprland-release dosyasını oluşturur veya günceller.
func WriteMetaFile(dotfileName string, versionMain, versionBuild, branch, releaseChannel, commitsBehind string) error {
	d := dotfiles.GetDotfileByName(dotfileName)
	if d == nil {
		return fmt.Errorf("dotfile not found: %s", dotfileName)
	}

	file := "/etc/hyprland-release"
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

	err := os.WriteFile(file, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("failed to write /etc/hyprland-release: %v", err)
	}
	return nil
}
