package updateing

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
        "github.com/hyprcommunity/hypr-release/api/releases/summaryofversion" 
)

const SystemModelDir = "/usr/share/hypr-release/ai/LLM/"

// InstallFromRegistry : summaryofversion/registry.go'dan dotfile indirip kurar
func InstallFromRegistry(name string) error {
	var selected *summaryofversion.Dotfile
	for _, d := range summaryofversion.Registry {
		if strings.EqualFold(d.Name, name) {
			selected = &d
			break
		}
	}

	if selected == nil {
		return fmt.Errorf("dotfile '%s' not found in registry", name)
	}

	fmt.Printf("[hyprrelease] selected: %s (%s)\n", selected.Name, selected.Repo)
	fmt.Printf("[hyprrelease] default branch: %s\n", selected.Branch)

	// Kullanıcıya farklı branch seçme fırsatı ver
	fmt.Print("Enter a branch or tag to install (leave empty to use default): ")
	reader := bufio.NewReader(os.Stdin)
	userBranch, _ := reader.ReadString('\n')
	userBranch = strings.TrimSpace(userBranch)

	if userBranch != "" {
		fmt.Printf("[hyprrelease] overriding branch: %s → %s\n", selected.Branch, userBranch)
		selected.Branch = userBranch
	} else {
		fmt.Printf("[hyprrelease] using default branch: %s\n", selected.Branch)
	}

	targetDir := filepath.Join(os.TempDir(), "hyprrelease-dotfiles", selected.Name)
	os.RemoveAll(targetDir)
	os.MkdirAll(targetDir, 0755)

	fmt.Printf("[hyprrelease] cloning %s (branch: %s)...\n", selected.Repo, selected.Branch)
	cmd := exec.Command("git", "clone", "--depth=1", "-b", selected.Branch, selected.Repo, targetDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to clone: %v", err)
	}

	fmt.Println("[hyprrelease] repository cloned successfully")
	return InstallRepo(targetDir)
}
// ------------------------------------------------------------
// InstallRepo : akıllı kurulum (betik, README, AI-safe kopya)
func InstallRepo(repoPath string) error {
	fmt.Println("[hyprrelease] starting intelligent installation")

	// 1️⃣ install.sh veya hyprrelease.sh varsa çalıştır
	if err := runInstallerScript(repoPath); err == nil {
		return nil
	}

	// 2️⃣ README varsa AI analizli kurulum
	readme := findReadme(repoPath)
	if readme != "" {
		if err := installFromReadme(readme, repoPath); err == nil {
			return nil
		}
	}

	// 3️⃣ fallback: AI dosya seçimiyle güvenli kopyalama
	if err := aiSafeFileInstall(repoPath); err != nil {
		fmt.Println("⚠️ AI safe-copy failed, using default safe filter.")
		if err2 := defaultCopy(repoPath); err2 != nil {
			return fmt.Errorf("fallback copy failed: %v", err2)
		}
	}
	fmt.Println("[hyprrelease] installation complete")
	return nil
}

// ------------------------------------------------------------
// install.sh veya hyprrelease.sh
func runInstallerScript(repoPath string) error {
	scripts := []string{"hyprrelease.sh", "install.sh"}
	for _, s := range scripts {
		script := filepath.Join(repoPath, s)
		if _, err := os.Stat(script); err == nil {
			fmt.Printf("[hyprrelease] running %s\n", s)
			cmd := exec.Command("bash", script, "install")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				fmt.Println("⚠️ script failed:", err)
				continue
			}
			fmt.Println("[hyprrelease] script install complete")
			return nil
		}
	}
	return fmt.Errorf("no installer script found")
}

// ------------------------------------------------------------
// README analizli kurulum
func installFromReadme(readmePath, repoPath string) error {
    // 🔧 README içeriğini oku
    content, err := os.ReadFile(readmePath)
    if err != nil {
        return fmt.Errorf("failed to read README: %w", err)
    }

    // 🔍 Model dizininden .gguf dosyasını bul
    files, err := os.ReadDir(SystemModelDir)
    if err != nil {
        fmt.Printf("⚠️ cannot read model directory: %v\n", err)
        fmt.Println("↪ fallback to regex parser.")
        return parseReadmeRegex(string(content))
    }

    var modelPath string
    for _, f := range files {
        if !f.IsDir() && strings.HasSuffix(f.Name(), ".gguf") {
            modelPath = filepath.Join(SystemModelDir, f.Name())
            break
        }
    }

    // 🔄 Model bulunamadıysa regex parser’a geç
    if modelPath == "" {
        fmt.Println("⚠️ no LLM model found, fallback to regex parser.")
        return parseReadmeRegex(string(content))
    }

    fmt.Printf("[hyprrelease] using Wingman with model: %s\n", filepath.Base(modelPath))

    // 🧠 Wingman prompt
    prompt := `
You are an installation step extractor.
Analyze the following README and output ONLY the shell commands to install the project.
List each command on its own line. No explanations, no comments.
---
` + string(content)

    // 🚀 Wingman CLI çağrısı
    cmd := exec.Command("wingman", "ask", prompt)
    output, err := cmd.CombinedOutput()
    if err != nil {
        fmt.Printf("⚠️ Wingman failed: %v\nOutput: %s\n", err, string(output))
        fmt.Println("↪ fallback to regex parser.")
        return parseReadmeRegex(string(content))
    }

    raw := strings.TrimSpace(string(output))
    cmds := strings.Split(raw, "\n")

    if len(cmds) == 0 || raw == "" {
        fmt.Println("[hyprrelease] AI found no install commands, fallback to regex parser.")
        return parseReadmeRegex(string(content))
    }

    // 📋 Komutları yazdır
    fmt.Println("[AI extracted install steps]:")
    for i, c := range cmds {
        fmt.Printf("%d. %s\n", i+1, strings.TrimSpace(c))
    }

    // ☑️ Kullanıcı onayı
    fmt.Print("Proceed with installation? [Y/n]: ")
    reader := bufio.NewReader(os.Stdin)
    resp, _ := reader.ReadString('\n')
    resp = strings.TrimSpace(resp)
    if strings.ToLower(resp) != "y" && resp != "" {
        return fmt.Errorf("installation aborted by user")
    }

    // 🧱 Komutları sırayla çalıştır
    for _, c := range cmds {
        c = strings.TrimSpace(c)
        if c == "" {
            continue
        }

        lower := strings.ToLower(c)
        if strings.HasPrefix(lower, "sudo ") ||
            strings.Contains(lower, "rm ") ||
            strings.Contains(lower, ":(){ :|:& };:") ||
            strings.Contains(lower, "mkfs") ||
            strings.Contains(lower, "dd if=") {
            fmt.Printf("⚠️ skipped dangerous command: %s\n", c)
            continue
        }

        parts := strings.Fields(c)
        if len(parts) == 0 {
            continue
        }

        execCmd := exec.Command(parts[0], parts[1:]...)
        execCmd.Stdout = os.Stdout
        execCmd.Stderr = os.Stderr
        fmt.Printf("→ executing: %s\n", c)

        if err := execCmd.Run(); err != nil {
            return fmt.Errorf("command failed (%s): %w", c, err)
        }
    }

    return nil
}

// Basit fallback regex parser
func parseReadmeRegex(content string) error {
    lines := strings.Split(content, "\n")
    for _, line := range lines {
        trimmed := strings.TrimSpace(line)
        if strings.HasPrefix(trimmed, "git clone") ||
            strings.Contains(trimmed, "github.com/hyprcommunity/hypr-release/install"") ||
            strings.Contains(trimmed, "make install") {
            fmt.Println("→ executing (regex):", trimmed)
            parts := strings.Fields(trimmed)
            if len(parts) == 0 {
                continue
            }
            cmd := exec.Command(parts[0], parts[1:]...)
            cmd.Stdout = os.Stdout
            cmd.Stderr = os.Stderr
            if err := cmd.Run(); err != nil {
                return fmt.Errorf("fallback command failed (%s): %w", trimmed, err)
            }
        }
    }
    return nil
}

// ------------------------------------------------------------
// AI tabanlı güvenli dosya seçimi
func aiSafeFileInstall(repoPath string) error {
    // 🔍 Model dizini taraması (sadece bilgilendirme amaçlı)
    files, err := os.ReadDir(SystemModelDir)
    if err != nil {
        return fmt.Errorf("failed to read model directory: %v", err)
    }

    var modelPath string
    for _, f := range files {
        if strings.HasSuffix(f.Name(), ".gguf") {
            modelPath = filepath.Join(SystemModelDir, f.Name())
            break
        }
    }
    if modelPath == "" {
        fmt.Println("⚠️ no .gguf model found, continuing with Wingman LLM backend.")
    } else {
        fmt.Printf("[hyprrelease] using AI model (gguf detected): %s\n", filepath.Base(modelPath))
    }

    // 🔧 Dosya ağacını çıkar
    var structure []string
    filepath.WalkDir(repoPath, func(path string, d fs.DirEntry, err error) error {
        if err == nil {
            rel, _ := filepath.Rel(repoPath, path)
            if rel != "." {
                structure = append(structure, rel)
            }
        }
        return nil
    })

    // 🧠 LLM prompt
    prompt := `
You are a configuration installer AI.
From this file tree, select ONLY configuration and script files safe to copy into ~/.config/hypr/.
Prefer .conf, .ini, .json, .lua, .sh, .desktop files.
Ignore LICENSE, README, cache, images, fonts, binaries, build artifacts.
Return one relative path per line, no comments, no explanations.
---
` + strings.Join(structure, "\n")

    // 🚀 Wingman CLI çağrısı
    cmd := exec.Command("wingman", "ask", prompt)
    output, err := cmd.CombinedOutput()
    if err != nil {
        return fmt.Errorf("wingman failed: %v\nOutput: %s", err, string(output))
    }

    raw := strings.TrimSpace(string(output))
    filesList := strings.Split(raw, "\n")
    if len(filesList) == 0 || raw == "" {
        return fmt.Errorf("AI returned no file list")
    }

    fmt.Println("[AI selected safe files]:")
    for _, f := range filesList {
        fmt.Println(" →", f)
    }

    // ☑️ Kullanıcı onayı
    fmt.Print("Proceed with AI-selected file copy? [Y/n]: ")
    var resp string
    fmt.Scanln(&resp)
    if strings.ToLower(strings.TrimSpace(resp)) != "y" && resp != "" {
        return fmt.Errorf("user aborted installation")
    }

    // 🎯 Hedef dizin
    home, err := os.UserHomeDir()
    if err != nil {
        return fmt.Errorf("cannot resolve home directory: %v", err)
    }
    target := filepath.Join(home, ".config", "hypr")

    // 📁 Dosyaları güvenli şekilde kopyala
    for _, rel := range filesList {
        rel = strings.TrimSpace(rel)
        if rel == "" {
            continue
        }

        src := filepath.Join(repoPath, rel)
        dest := filepath.Join(target, rel)

        info, err := os.Stat(src)
        if err != nil || info.IsDir() {
            fmt.Printf("⚠️ skipping invalid: %s\n", rel)
            continue
        }

        if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
            fmt.Printf("⚠️ failed to create directory for %s: %v\n", rel, err)
            continue
        }

        in, err := os.Open(src)
        if err != nil {
            fmt.Printf("⚠️ failed to open source %s: %v\n", rel, err)
            continue
        }

        out, err := os.Create(dest)
        if err != nil {
            in.Close()
            fmt.Printf("⚠️ failed to create destination %s: %v\n", rel, err)
            continue
        }

        if _, err := io.Copy(out, in); err != nil {
            fmt.Printf("⚠️ copy error for %s: %v\n", rel, err)
        } else {
            fmt.Println("→ copied:", rel)
        }

        in.Close()
        out.Close()
    }

    fmt.Println("✅ AI-selected configuration files successfully copied.")
    return nil
}

// ------------------------------------------------------------
// Klasik kopyalama fallback
func defaultCopy(repoPath string) error {
	home, _ := os.UserHomeDir()
	target := filepath.Join(home, ".config", "hypr")
	fmt.Println("[hyprrelease] default safe filter copy")

	return filepath.WalkDir(repoPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		exts := []string{".conf", ".ini", ".json", ".sh", ".png"}
		for _, e := range exts {
			if strings.HasSuffix(path, e) {
				rel, _ := filepath.Rel(repoPath, path)
				dest := filepath.Join(target, rel)
				os.MkdirAll(filepath.Dir(dest), 0755)
				src, _ := os.Open(path)
				defer src.Close()
				dst, _ := os.Create(dest)
				defer dst.Close()
				io.Copy(dst, src)
				fmt.Println("→ copied:", rel)
			}
		}
		return nil
	})
}

// ------------------------------------------------------------
// Yardımcı fonksiyonlar
func findReadme(repoPath string) string {
	candidates := []string{"README.md", "README", "readme.md", "readme"}
	for _, f := range candidates {
		path := filepath.Join(repoPath, f)
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	return ""
}

func runCmd(cmdline string) {
	parts := strings.Fields(cmdline)
	if len(parts) == 0 {
		return
	}
	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Println("⚠️ command failed:", err)
	}
}
