package check

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/hyprcommunity/hypr-release/api/releases/summaryofversion"
)

type TestingStatus struct {
	Branch          string
	ReleaseChannel  string
	Source          string
	KeywordDetected bool
}

// CheckTestingStatus : GH releases varsa oradan alır, yoksa git branch'e göre belirler.
func CheckTestingStatus(dotfileName, repoPath string) (TestingStatus, string, error) {
	var log strings.Builder
	status := TestingStatus{}

	d := dotfiles.GetDotfileByName(dotfileName)
	if d == nil {
		return status, "", fmt.Errorf("dotfile not found: %s", dotfileName)
	}

	branchCmd := exec.Command("git", "-C", repoPath, "rev-parse", "--abbrev-ref", "HEAD")
	branchOut, err := branchCmd.Output()
	if err != nil {
		return status, "", fmt.Errorf("cannot detect branch: %v", err)
	}
	branch := strings.TrimSpace(string(branchOut))
	status.Branch = branch

	log.WriteString(fmt.Sprintf("🔎 Detected branch: %s\n", branch))

	keywords := map[string]string{
		"main":     "stable",
		"master":   "stable",
		"stable":   "stable",
		"release":  "stable",
		"beta":     "beta",
		"testing":  "testing",
		"test":     "testing",
		"rc":       "release-candidate",
		"nightly":  "nightly",
		"dev":      "development",
		"preview":  "preview",
		"edge":     "unstable",
	}

	if d.HasReleases {
		status.Source = "GitHub releases"
		for key, ch := range keywords {
			if strings.Contains(strings.ToLower(branch), key) {
				status.ReleaseChannel = ch
				status.KeywordDetected = true
				break
			}
		}
		if !status.KeywordDetected {
			status.ReleaseChannel = "stable"
		}
	} else {
		status.Source = "git branch"
		for key, ch := range keywords {
			if strings.Contains(strings.ToLower(branch), key) {
				status.ReleaseChannel = ch
				status.KeywordDetected = true
				break
			}
		}
		if !status.KeywordDetected {
			status.ReleaseChannel = "unknown-version"
		}
	}

	log.WriteString(fmt.Sprintf("💡 Release channel detected: %s (%s)\n", status.ReleaseChannel, status.Source))
	return status, log.String(), nil
}
