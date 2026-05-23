package deepseek

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/levygit837-cyber/OrchestraOS/internal/apperrors"
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
	"github.com/levygit837-cyber/OrchestraOS/internal/runtime"
)

const (
	defaultModel    = "deepseek-chat"
	defaultEndpoint = "https://api.deepseek.com/v1"
	defaultTokens   = 4096
	defaultTemp     = 0.7
	defaultTimeout  = 60 * time.Second
)

// DeepSeek implements domain.Runtime and domain.StreamRuntime for DeepSeek.
type DeepSeek struct {
	apiKey, model, endpoint string
	maxTok                  int
	temp                    float64
	client                  *http.Client
}

// New creates a DeepSeek runtime from the given config.
func New(cfg runtime.Config) *DeepSeek {
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = defaultTimeout
	}
	return &DeepSeek{
		apiKey:   cfg.APIKey,
		model:    strOr(cfg.Model, defaultModel),
		endpoint: strOr(cfg.BaseURL, defaultEndpoint),
		maxTok:   intOr(cfg.MaxTokens, defaultTokens),
		temp:     floatOr(cfg.Temperature, defaultTemp),
		client:   &http.Client{Timeout: timeout},
	}
}

func (d *DeepSeek) Execute(ctx context.Context, wu *domain.WorkUnit, task *domain.Task) (*domain.RuntimeResult, error) {
	prompt := runtime.BuildPrompt(wu, task)
	start := time.Now()
	body, err := d.buildRequest(prompt)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.KindInternal, "deepseek.execute", err)
	}
	resp, err := d.doRequest(ctx, body)
	if err != nil {
		return nil, err
	}
	output, tokens, err := parseOpenAIResponse(resp)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.KindInternal, "deepseek.execute", err)
	}
	return &domain.RuntimeResult{
		Status: domain.RunResultSucceeded, Output: output, Provider: "deepseek",
		Model: d.model, TokensUsed: tokens, Latency: time.Since(start),
	}, nil
}

type openAIRequest struct {
	Model       string          `json:"model"`
	Messages    []openAIMessage `json:"messages"`
	MaxTokens   int             `json:"max_tokens"`
	Temperature float64         `json:"temperature"`
	Stream      bool            `json:"stream,omitempty"`
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIResponse struct {
	Choices []openAIChoice `json:"choices"`
	Usage   *openAIUsage   `json:"usage,omitempty"`
}

type openAIChoice struct {
	Message openAIMessage `json:"message"`
}

type openAIUsage struct {
	TotalTokens int `json:"total_tokens"`
}

func (d *DeepSeek) buildRequest(prompt domain.Prompt) ([]byte, error) {
	req := openAIRequest{
		Model: d.model,
		Messages: []openAIMessage{
			{Role: "system", Content: prompt.SystemMessage},
			{Role: "user", Content: prompt.UserMessage},
		},
		MaxTokens: d.maxTok, Temperature: d.temp,
	}
	return json.Marshal(req)
}

func (d *DeepSeek) doRequest(ctx context.Context, body []byte) ([]byte, error) {
	url := fmt.Sprintf("%s/chat/completions", d.endpoint)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, apperrors.Wrap(apperrors.KindInternal, "deepseek.execute", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+d.apiKey)
	resp, err := d.client.Do(req)
	if err != nil {
		return nil, classifyHTTPError("deepseek.execute", err)
	}
	defer func() { _ = resp.Body.Close() }()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.KindInternal, "deepseek.execute", err)
	}
	if err := classifyStatusCode("deepseek.execute", resp.StatusCode, respBody); err != nil {
		return nil, err
	}
	return respBody, nil
}

func parseOpenAIResponse(body []byte) (string, int, error) {
	var resp openAIResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return "", 0, err
	}
	if len(resp.Choices) == 0 {
		return "", 0, fmt.Errorf("empty response from provider")
	}
	tokens := 0
	if resp.Usage != nil {
		tokens = resp.Usage.TotalTokens
	}
	return resp.Choices[0].Message.Content, tokens, nil
}

func strOr(v, d string) string {
	if v != "" {
		return v
	}
	return d
}

func intOr(v, d int) int {
	if v != 0 {
		return v
	}
	return d
}

func floatOr(v, d float64) float64 {
	if v != 0 {
		return v
	}
	return d
}
