// internal/llm/provider.go
// Abstracción multi-provider LLM — opera local (Ollama) o conectado (OpenRouter)
// Principio anfibio: misma interfaz, diferente backend
package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Provider es la interfaz que todo backend LLM debe implementar
type Provider interface {
	Chat(ctx context.Context, messages []Message) (*Response, error)
	Name() string
}

// Message representa un mensaje en la conversación
type Message struct {
	Role    string `json:"role"`    // "system", "user", "assistant"
	Content string `json:"content"`
}

// Response representa la respuesta del modelo
type Response struct {
	Content      string        `json:"content"`
	Model        string        `json:"model"`
	Usage        Usage         `json:"usage"`
	FinishReason string        `json:"finish_reason"`
}

// Usage muestra tokens consumidos
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// Config para crear providers
type Config struct {
	Provider string  // "ollama" o "openrouter"
	Model    string  // "llama3.2", "gpt-4", etc.
	APIKey   string  // requerido para OpenRouter
	BaseURL  string  // URL base (default según provider)
}

// New crea un provider según la configuración
func New(cfg Config) (Provider, error) {
	switch cfg.Provider {
	case "ollama":
		return NewOllamaProvider(cfg), nil
	case "openrouter":
		return NewOpenRouterProvider(cfg), nil
	default:
		return nil, fmt.Errorf("provider desconocido: %s", cfg.Provider)
	}
}

// --- Ollama Provider (local, offline-first) ---

type OllamaProvider struct {
	baseURL string
	model   string
	client  *http.Client
}

func NewOllamaProvider(cfg Config) *OllamaProvider {
	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	return &OllamaProvider{
		baseURL: baseURL,
		model:   cfg.Model,
		client:  &http.Client{Timeout: 60 * time.Second},
	}
}

func (p *OllamaProvider) Name() string { return "ollama" }

func (p *OllamaProvider) Chat(ctx context.Context, messages []Message) (*Response, error) {
	reqBody := map[string]any{
		"model":    p.model,
		"messages": messages,
		"stream":   false,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/api/chat", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ollama request failed: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
		Model  string `json:"model"`
		Done   bool   `json:"done"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &Response{
		Content:      result.Message.Content,
		Model:        result.Model,
		FinishReason: "stop",
	}, nil
}

// --- OpenRouter Provider (conectado, multi-modelo) ---

type OpenRouterProvider struct {
	apiKey  string
	model   string
	baseURL string
	client  *http.Client
}

func NewOpenRouterProvider(cfg Config) *OpenRouterProvider {
	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = "https://openrouter.ai/api/v1"
	}
	return &OpenRouterProvider{
		apiKey:  cfg.APIKey,
		model:   cfg.Model,
		baseURL: baseURL,
		client:  &http.Client{Timeout: 60 * time.Second},
	}
}

func (p *OpenRouterProvider) Name() string { return "openrouter" }

func (p *OpenRouterProvider) Chat(ctx context.Context, messages []Message) (*Response, error) {
	if p.apiKey == "" {
		return nil, fmt.Errorf("OpenRouter requiere API key")
	}

	reqBody := map[string]any{
		"model":    p.model,
		"messages": messages,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/chat/completions", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	req.Header.Set("HTTP-Referer", "https://github.com/Isaacko0/red-de-intercambio-federada-isaacko")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("openrouter request failed: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
			FinishReason string `json:"finish_reason"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		} `json:"usage"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("sin respuesta de OpenRouter")
	}

	return &Response{
		Content:      result.Choices[0].Message.Content,
		Model:        p.model,
		FinishReason: result.Choices[0].FinishReason,
		Usage: Usage{
			PromptTokens:     result.Usage.PromptTokens,
			CompletionTokens: result.Usage.CompletionTokens,
			TotalTokens:      result.Usage.TotalTokens,
		},
	}, nil
}
