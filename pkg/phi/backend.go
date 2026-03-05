package phi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/sipeed/picoclaw/pkg/config"
)

const (
	OllamaDefaultURL = "http://localhost:11434"
	MLXDefaultURL    = "http://localhost:8080"
)

// ollamaClient is shared across Ollama API calls to reuse connections.
var ollamaClient = &http.Client{Timeout: 5 * time.Second}

// BackendStatus reports the health of a local inference backend.
type BackendStatus struct {
	Installed bool   `json:"installed"`
	Running   bool   `json:"running"`
	Version   string `json:"version,omitempty"`
	Error     string `json:"error,omitempty"`
}

// BackendAPIBase returns the OpenAI-compatible API base URL for the given backend.
func BackendAPIBase(backend string) string {
	switch backend {
	case config.BackendOllama:
		return OllamaDefaultURL + "/v1"
	case config.BackendMLX:
		return MLXDefaultURL + "/v1"
	default:
		return ""
	}
}

// CheckBackend probes the given backend and returns its status.
func CheckBackend(backend string) BackendStatus {
	switch backend {
	case config.BackendOllama:
		return checkOllama()
	case config.BackendMLX:
		return BackendStatus{Error: "MLX support coming soon"}
	default:
		return BackendStatus{Error: fmt.Sprintf("unknown backend: %s", backend)}
	}
}

func checkOllama() BackendStatus {
	status := BackendStatus{}

	// Check if ollama is installed
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, "ollama", "--version").CombinedOutput()
	if err != nil {
		status.Error = "ollama is not installed. Install from https://ollama.com"
		return status
	}
	status.Installed = true
	status.Version = strings.TrimSpace(string(out))

	// Check if ollama is running by listing models
	if _, err := OllamaListModels(); err != nil {
		status.Error = "ollama is installed but not running"
		return status
	}
	status.Running = true

	return status
}

// CheckModelReady returns true if the specified model tag is already pulled in Ollama.
func CheckModelReady(ollamaTag string) bool {
	models, err := OllamaListModels()
	if err != nil {
		return false
	}

	for _, name := range models {
		if name == ollamaTag || name == ollamaTag+":latest" ||
			strings.TrimSuffix(name, ":latest") == strings.TrimSuffix(ollamaTag, ":latest") {
			return true
		}
	}
	return false
}

// PullModel pulls a model using ollama pull. It calls the progress callback
// with status lines from ollama's output.
func PullModel(ctx context.Context, ollamaTag string, progress func(string)) error {
	pullCtx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(pullCtx, "ollama", "pull", ollamaTag)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("creating stdout pipe: %w", err)
	}
	cmd.Stderr = cmd.Stdout // merge stderr into stdout

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("starting ollama pull: %w", err)
	}

	buf := make([]byte, 4096)
	for {
		n, readErr := stdout.Read(buf)
		if n > 0 && progress != nil {
			progress(strings.TrimSpace(string(buf[:n])))
		}
		if readErr != nil {
			break
		}
	}

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("ollama pull failed: %w", err)
	}
	return nil
}

// WarmupModel sends a trivial prompt to verify the model responds correctly.
func WarmupModel(ctx context.Context, ollamaTag string) error {
	warmupCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	body := fmt.Sprintf(`{"model":%q,"messages":[{"role":"user","content":"hello"}],"max_tokens":16}`, ollamaTag)
	req, err := http.NewRequestWithContext(warmupCtx, "POST",
		OllamaDefaultURL+"/v1/chat/completions",
		strings.NewReader(body))
	if err != nil {
		return fmt.Errorf("creating warmup request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := (&http.Client{Timeout: 30 * time.Second}).Do(req)
	if err != nil {
		return fmt.Errorf("warmup request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("warmup returned status %d", resp.StatusCode)
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("parsing warmup response: %w", err)
	}
	if len(result.Choices) == 0 {
		return fmt.Errorf("warmup returned no choices")
	}

	return nil
}

// OllamaListModels returns the list of locally available model tags.
func OllamaListModels() ([]string, error) {
	resp, err := ollamaClient.Get(OllamaDefaultURL + "/api/tags")
	if err != nil {
		return nil, fmt.Errorf("connecting to ollama: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("parsing ollama response: %w", err)
	}

	models := make([]string, len(result.Models))
	for i, m := range result.Models {
		models[i] = m.Name
	}
	return models, nil
}
