package backend

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/nomic-ai/go-tiny-llm"
	"hyprrelease/summaryofversion"
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
	content, err := os.ReadFile(readmePath)
	if err != nil {
		return err
	}

	modelPath := filepath.Join(SystemModelDir, "mistral-7b.Q4_K_M.gguf")
	if _, err := os.Stat(modelPath); os.IsNotExist(err) {
		fmt.Println("⚠️ LLM model not found, fallback to regex parser.")
		return parseReadmeRegex(string(content))
	}

	fmt.Println("[hyprrelease] loading AI model:", modelPath)
	model := llm.LoadModel(modelPath)

	prompt := `
You are an installation step extractor.
Read the following README and output ONLY the shell commands to install the project.
Use one command per line, no explanations.
---
` + string(content)

	raw := model.Predict(prompt)
	cmds := strings.Split(strings.TrimSpace(raw), "\n")

	if len(cmds) == 0 {
		fmt.Println("[hyprrelease] AI found no commands, fallback to regex parser.")
		return parseReadmeRegex(string(content))
	}

	fmt.Println("[AI extracted install steps]:")
	for i, c := range cmds {
		fmt.Printf("%d. %s\n", i+1, c)
	}

	fmt.Print("Proceed with installation? [Y/n]: ")
	var resp string
	fmt.Scanln(&resp)
	if strings.ToLower(resp) != "y" && resp != "" {
		return fmt.Errorf("installation aborted by user")
	}

	for _, c := range cmds {
		c = strings.TrimSpace(c)
		if c == "" {
			continue
		}
		if strings.Contains(c, "sudo") || strings.Contains(c, "rm ") {
			fmt.Printf("⚠️ skipped dangerous command: %s\n", c)
			continue
		}
		runCmd(c)
	}

	return nil
}

// Regex fallback
func parseReadmeRegex(content string) error {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "git clone") ||
			strings.Contains(line, "./install") ||
			strings.Contains(line, "make install") {
			fmt.Println("→ executing:", line)
			runCmd(line)
		}
	}
	return nil
}

// ------------------------------------------------------------
// AI tabanlı güvenli dosya seçimi
func aiSafeFileInstall(repoPath string) error {
	modelPath := filepath.Join(SystemModelDir, "mistral-7b.Q4_K_M.gguf")
	if _, err := os.Stat(modelPath); os.IsNotExist(err) {
		return fmt.Errorf("no AI model found")
	}

	fmt.Println("[hyprrelease] AI-based safe file analysis")
	model := llm.LoadModel(modelPath)

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

	prompt := `
You are a configuration installer AI.
Analyze the following file tree and select ONLY safe configuration files and scripts to copy into ~/.config/hypr/.
Ignore LICENSE, README, binaries, caches, wallpapers.
Return one relative path per line, no explanations.
---
` + strings.Join(structure, "\n")

	raw := model.Predict(prompt)
	files := strings.Split(strings.TrimSpace(raw), "\n")

	if len(files) == 0 {
		return fmt.Errorf("AI returned no file list")
	}

	fmt.Println("[AI selected safe files]:")
	for _, f := range files {
		fmt.Println(" →", f)
	}

	fmt.Print("Proceed with AI-selected file copy? [Y/n]: ")
	var resp string
	fmt.Scanln(&resp)
	if strings.ToLower(resp) != "y" && resp != "" {
		return fmt.Errorf("user aborted")
	}

	home, _ := os.UserHomeDir()
	target := filepath.Join(home, ".config", "hypr")

	for _, rel := range files {
		src := filepath.Join(repoPath, rel)
		dest := filepath.Join(target, rel)
		if _, err := os.Stat(src); err != nil {
			continue
		}
		os.MkdirAll(filepath.Dir(dest), 0755)
		s, _ := os.Open(src)
		defer s.Close()
		d, _ := os.Create(dest)
		io.Copy(d, s)
		d.Close()
		fmt.Println("→ copied:", rel)
	}
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
