package hyprland

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// WriteHyprSystemMeta : hyprland çekirdek araçlarının sürüm bilgilerini kaydeder
func WriteHyprSystemMeta(components []HyprComponent) error {
	file := "/etc/hyprland-system-release"
	var b strings.Builder

	b.WriteString("# Hyprland System Release Metadata\n")
	b.WriteString(fmt.Sprintf("HYPRLAND_SYSTEM_CHECK_DATE=\"%s\"\n", time.Now().Format("2006-01-02 15:04:05")))

	for _, c := range components {
		upper := strings.ToUpper(c.Name)
		b.WriteString(fmt.Sprintf("HYPRLAND_%s_VERSION=\"%s\"\n", upper, c.Version))
		b.WriteString(fmt.Sprintf("HYPRLAND_%s_REMOTE_VERSION=\"%s\"\n", upper, c.RemoteVersion))
		b.WriteString(fmt.Sprintf("HYPRLAND_%s_PATH=\"%s\"\n", upper, c.Path))
		if c.UpdateAvailable {
			b.WriteString(fmt.Sprintf("HYPRLAND_%s_UPDATE=\"true\"\n", upper))
		} else {
			b.WriteString(fmt.Sprintf("HYPRLAND_%s_UPDATE=\"false\"\n", upper))
		}
		b.WriteString(fmt.Sprintf("HYPRLAND_%s_SOURCE=\"%s\"\n", upper, c.Source))
		b.WriteString("\n")
	}

	err := os.WriteFile(file, []byte(b.String()), 0644)
	if err != nil {
		return fmt.Errorf("failed to write %s: %v", file, err)
	}
	return nil
}
