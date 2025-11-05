package backend

import (
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

// Model bilgisi
type LLMModel struct {
	Name string
	URL  string
	HashURL string
}

var SystemModelDir = "/usr/share/hypr-release/ai/LLM/"

// getRemoteChecksum : modelin sha256 değerini curl ile çeker
func getRemoteChecksum(hashURL string) (string, error) {
	cmd := exec.Command("curl", "-fsSL", hashURL)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("curl checksum fetch failed: %v", err)
	}
	sum := strings.TrimSpace(string(out))
	sum = strings.Split(sum, " ")[0] // bazen "hash  filename" biçiminde gelir
	return sum, nil
}

// calcLocalChecksum : indirilen dosyanın SHA256 değerini hesaplar
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

func DownloadModel(m LLMModel) error {
	fmt.Printf("[hyprrelease-llm] downloading model: %s\n", m.Name)

	resp, err := http.Get(m.URL)
	if err != nil {
		return fmt.Errorf("download failed: %v", err)
	}
	defer resp.Body.Close()

	os.MkdirAll(SystemModelDir, 0755)
	dest := filepath.Join(SystemModelDir, m.Name)
	out, _ := os.Create(dest)
	defer out.Close()
	io.Copy(out, resp.Body)

	remoteHash, err := getRemoteChecksum(m.HashURL)
	if err != nil {
		fmt.Println("⚠️ unable to fetch remote hash:", err)
	} else {
		localHash, _ := calcLocalChecksum(dest)
		if localHash != remoteHash {
			os.Remove(dest)
			return fmt.Errorf("checksum mismatch for %s\nexpected: %s\ngot: %s", m.Name, remoteHash, localHash)
		}
		fmt.Println("✅ checksum verified:", localHash)
	}
	fmt.Println("[hyprrelease-llm] model installed at", dest)
	return nil
}

func EnsureDefaultModels() {
	models := []LLMModel{
		{
			Name:    "mistral-7b.Q4_K_M.gguf",
			URL:     "https://huggingface.co/TheBloke/Mistral-7B-GGUF/resolve/main/mistral-7b.Q4_K_M.gguf",
			HashURL: "https://huggingface.co/TheBloke/Mistral-7B-GGUF/resolve/main/mistral-7b.Q4_K_M.gguf.sha256",
		},
		{
			Name:    "phi3-mini.Q4_K_M.gguf",
			URL:     "https://huggingface.co/TheBloke/phi-3-mini-GGUF/resolve/main/phi3-mini.Q4_K_M.gguf",
			HashURL: "https://huggingface.co/TheBloke/phi-3-mini-GGUF/resolve/main/phi3-mini.Q4_K_M.gguf.sha256",
		},
	}

	for _, m := range models {
		path := filepath.Join(SystemModelDir, m.Name)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			if err := DownloadModel(m); err != nil {
				fmt.Println("⚠️ model download failed:", err)
			}
		} else {
			fmt.Println("[hyprrelease-llm] model already exists:", m.Name)
		}
	}
}
