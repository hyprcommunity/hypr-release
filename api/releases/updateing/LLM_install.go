package updateing

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// LLMModel : information about available language models
type LLMModel struct {
	Name string
	URL  string
	Size string
	Desc string
}

// Default model directory
var SystemModelDir = "/usr/share/hypr-release/ai/LLM/"

// ✅ CheckOrInstallWingman : ensures Wingman is installed or attempts auto-install
func CheckOrInstallWingman() error {
	_, err := exec.LookPath("wingman")
	if err == nil {
		fmt.Println("✅ Wingman detected on system.")
		return nil
	}

	fmt.Println("⚠️  Wingman CLI not found on system.")
	fmt.Println("↪  Attempting to install automatically...")

	// Kullanıcıya bilgi notu
	fmt.Println(`
You can manually install Wingman using one of the following:
  • Arch Linux (AUR):    yay -S wingman-bin
  • Go source install:   go install github.com/adrianliechti/wingman/cmd/wingman@latest
  • GitHub releases:     https://github.com/adrianliechti/wingman/releases
`)

	// Otomatik Go ile kurulum denemesi
	cmd := exec.Command("go", "install", "github.com/adrianliechti/wingman/cmd/wingman@latest")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("⚠️ automatic Wingman install failed: %v", err)
	}

	// Yeniden kontrol
	_, err = exec.LookPath("wingman")
	if err != nil {
		return fmt.Errorf("⚠️ Wingman installation unsuccessful; please install manually")
	}

	fmt.Println("✅ Wingman successfully installed and ready for use.")
	return nil
}

// calcLocalChecksum : calculates the SHA256 checksum of a file
func calcLocalChecksum(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// DownloadModel : downloads the selected model and verifies checksum
func DownloadModel(m LLMModel) error {
	fmt.Printf("[hyprrelease-llm] Downloading model: %s\n", m.Name)

	resp, err := http.Get(m.URL)
	if err != nil {
		return fmt.Errorf("download failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("bad response from server: %s", resp.Status)
	}

	os.MkdirAll(SystemModelDir, 0755)
	dest := filepath.Join(SystemModelDir, m.Name)

	out, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("cannot create file: %v", err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("write failed: %v", err)
	}

	localHash, err := calcLocalChecksum(dest)
	if err != nil {
		fmt.Println("⚠️  Checksum calculation failed:", err)
	} else {
		fmt.Println("✅ Local SHA256 checksum:", localHash)
	}

	fmt.Println("[hyprrelease-llm] Model installed at:", dest)
	return nil
}

// ListAvailableModels : returns a list of available LLM models to choose from
func ListAvailableModels() []LLMModel {
	return []LLMModel{
		{
			Name: "TinyLlama-1.1B-Chat-v1.0.Q4_K_M.gguf",
			URL:  "https://huggingface.co/TheBloke/TinyLlama-1.1B-Chat-v1.0-GGUF/resolve/main/TinyLlama-1.1B-Chat-v1.0.Q4_K_M.gguf",
			Size: "1.1 GB",
			Desc: "Fast and lightweight. Ideal for local reasoning and README parsing.",
		},
		{
			Name: "Phi-3-mini-4k-instruct-IQ4_NL.gguf",
			URL:  "https://huggingface.co/bartowski/Phi-3-mini-4k-instruct-GGUF/resolve/main/Phi-3-mini-4k-instruct-IQ4_NL.gguf",
			Size: "2.17 GB",
			Desc: "Balanced quality and performance. Recommended default choice.",
		},
		{
			Name: "Phi-3-mini-4k-instruct-Q4_K_M.gguf",
			URL:  "https://huggingface.co/bartowski/Phi-3-mini-4k-instruct-GGUF/resolve/main/Phi-3-mini-4k-instruct-Q4_K_M.gguf",
			Size: "2.39 GB",
			Desc: "High-quality quantization. Slightly larger memory footprint.",
		},
	}
}

// SelectAndInstallModel : lets the user choose an LLM to install interactively
func SelectAndInstallModel() {
	models := ListAvailableModels()

	fmt.Println("Available LLM models for HyprRelease AI:")
	fmt.Println("----------------------------------------")
	for i, m := range models {
		fmt.Printf("[%d] %s  (%s)\n    %s\n", i+1, m.Name, m.Size, m.Desc)
	}
	fmt.Println("----------------------------------------")

	fmt.Print("Select model number to install: ")
	reader := bufio.NewReader(os.Stdin)
	choiceRaw, _ := reader.ReadString('\n')
	choiceRaw = strings.TrimSpace(choiceRaw)

	var index int
	fmt.Sscanf(choiceRaw, "%d", &index)
	if index <= 0 || index > len(models) {
		fmt.Println("Invalid selection.")
		return
	}

	selected := models[index-1]
	fmt.Printf("\nYou selected: %s (%s)\n", selected.Name, selected.Size)

	path := filepath.Join(SystemModelDir, selected.Name)
	if _, err := os.Stat(path); err == nil {
		fmt.Println("[hyprrelease-llm] Model already exists, skipping download.")
		localHash, _ := calcLocalChecksum(path)
		fmt.Println("ℹ️  Existing model SHA256:", localHash)
		return
	}

	if err := DownloadModel(selected); err != nil {
		fmt.Println("⚠️  Model download failed:", err)
	} else {
		fmt.Println("✅ Model successfully installed.")
	}

	// ✅ Wingman kontrolü (model indirildikten sonra)
	if err := CheckOrInstallWingman(); err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("🧠 Wingman environment ready for AI-based automation.")
	}
}

// ListInstalledModels : lists models already installed on the system
func ListInstalledModels() {
	fmt.Println("Installed LLM models in:", SystemModelDir)
	files, err := os.ReadDir(SystemModelDir)
	if err != nil {
		fmt.Println("⚠️  Cannot read model directory:", err)
		return
	}

	found := false
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".gguf") {
			found = true
			fullPath := filepath.Join(SystemModelDir, f.Name())
			size, _ := os.Stat(fullPath)
			fmt.Printf("• %s (%.2f MB)\n", f.Name(), float64(size.Size())/1024/1024)
		}
	}
	if !found {
		fmt.Println("No models found.")
	}
}
