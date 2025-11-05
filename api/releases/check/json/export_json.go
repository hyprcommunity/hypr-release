package api

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type HyprJSON struct {
	ReleaseMeta map[string]string `json:"release_meta"`
	SystemMeta  map[string]string `json:"system_meta"`
}

// ExportJSON : /etc/hyprland-release ve /etc/hyprland-system-release içeriklerini JSON olarak birleştirir
func ExportJSON() (string, error) {
	releaseFile := "/etc/hyprland-release"
	systemFile := "/etc/hyprland-system-release"

	readFile := func(path string) (map[string]string, error) {
		data := make(map[string]string)
		f, err := os.Open(path)
		if err != nil {
			return data, err
		}
		defer f.Close()
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if strings.HasPrefix(line, "#") || line == "" {
				continue
			}
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := strings.Trim(parts[0], "\" ")
				val := strings.Trim(parts[1], "\" ")
				val = strings.Trim(val, "\"")
				data[key] = strings.Trim(val, "\"")
			}
		}
		return data, scanner.Err()
	}

	releaseMeta, _ := readFile(releaseFile)
	systemMeta, _ := readFile(systemFile)

	full := HyprJSON{
		ReleaseMeta: releaseMeta,
		SystemMeta:  systemMeta,
	}

	out, err := json.MarshalIndent(full, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to encode JSON: %v", err)
	}

	return string(out), nil
}
