package check

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/hyprcommunity/hypr-release/api/releases/summaryofversion"
)

// Release bilgisi
type Release struct {
	Tag         string `json:"tagName"`
	Name        string `json:"name"`
	Body        string `json:"body"`
	PublishedAt string `json:"publishedAt"`
}

// CheckAll : GH release varsa onu ana sürüm, git tag'i alt sürüm olarak kullanır.
// GH release yoksa git tag ana sürüm olur.
func CheckAll(dotfileName, repoPath string) (map[string]string, string, error) {
	result := make(map[string]string)
	var output strings.Builder

	d := dotfiles.GetDotfileByName(dotfileName)
	if d == nil {
		return nil, "", fmt.Errorf("dotfile not found: %s", dotfileName)
	}

	output.WriteString(fmt.Sprintf("🔍 Checking repository: %s (%s)\n", d.Name, d.Repo))

	// GH komutu
	ghCmd := exec.Command("gh", "release", "list", "--repo", d.Repo, "--limit", "3", "--json", "tagName,publishedAt")
	data, err := ghCmd.Output()
	if err != nil {
		output.WriteString(fmt.Sprintf("⚠️ gh release list failed: %v\n", err))
	}

	var releases []Release
	_ = json.Unmarshal(data, &releases)

	// git tag
	tagCmd := exec.Command("git", "-C", repoPath, "describe", "--tags", "--abbrev=7", "--always")
	tagOut, _ := tagCmd.Output()
	gitTag := strings.TrimSpace(string(tagOut))

	if len(releases) > 0 {
		mainVer := releases[0].Tag
		buildVer := gitTag
		result["version_main"] = mainVer
		result["version_build"] = buildVer
		output.WriteString(fmt.Sprintf("Main version (GitHub): %s\n", mainVer))
		output.WriteString(fmt.Sprintf("Build version (git):  %s\n", buildVer))
	} else {
		mainVer := gitTag
		result["version_main"] = mainVer
		result["version_build"] = mainVer
		output.WriteString(fmt.Sprintf("Main version (git): %s\n", mainVer))
	}

	// commit farkı
	diffCmd := exec.Command("git", "-C", repoPath, "rev-list", "HEAD..origin/master", "--count")
	diffOut, _ := diffCmd.Output()
	result["commits_behind"] = strings.TrimSpace(string(diffOut))
	output.WriteString(fmt.Sprintf("Commits behind origin/master: %s\n", result["commits_behind"]))

	return result, output.String(), nil
}
