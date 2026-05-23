package runtime

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
)

const (
	deepseekDefaultModel    = "deepseek-chat"
	deepseekDefaultEndpoint = "https://api.deepseek.com/v1"
	deepseekDefaultTokens   = 4096
	deepseekDefaultTemp     = 0.7
	deepseekDefaultTimeout  = 60 * time.Second
)

// DeepSeek implements Runtime using the DeepSeek API (OpenAI-compatible).
type DeepSeek struct {
	apiKey   string
	model    string
	endpoint string
	maxTok   int
	temp     float64
	client   *http.Client
}

func NewDeepSeek(cfg Config) *DeepSeek {
	model := cfg.Model
	if model == "" {
		model = deepseekDefaultModel
	}
	endpoint := cfg.BaseURL
	if endpoint == "" {
		endpoint = deepseekDefaultEndpoint
	}
	maxTok := cfg.MaxTokens
	if maxTok == 0 {
		maxTok = deepseekDefaultTokens
	}
	temp := cfg.Temperature
	if temp == 0 {
		temp = deepseekDefaultTemp
	}
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = deepseekDefaultTimeout
	}
	return &DeepSeek{
		apiKey:   cfg.APIKey,
		model:    model,
		endpoint: endpoint,
		maxTok:   maxTok,
		temp:     temp,
		client:   &http.Client{Timeout: timeout},
	}
}

func (d *DeepSeek) Execute(ctx context.Context, wu *domain.WorkUnit, task *domain.Task) (*Result, error) {
	prompt := BuildPrompt(wu, task)
	start := time.Now()

	body, err := d.buildRequest(prompt)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.KindInternal, "runtime.deepseek", err)
	}

	resp, err := d.doRequest(ctx, body)
	if err != nil {
		return nil, err
	}

	output, tokens, err := parseOpenAIResponse(resp)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.KindInternal, "runtime.deepseek", err)
	}

	return &Result{
		Status:     domain.RunResultSucceeded,
		Output:     output,
		Provider:   "deepseek",
		Model:      d.model,
		TokensUsed: tokens,
		Latency:    time.Since(start),
	}, nil
}

type openAIRequest struct {
	Model       string          `json:"model"`
	Messages    []openAIMessage `json:"messages"`
	MaxTokens   int             `json:"max_tokens"`
	Temperature float64         `json:"temperature"`
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

func (d *DeepSeek) buildRequest(prompt Prompt) ([]byte, error) {
	req := openAIRequest{
		Model: d.model,
		Messages: []openAIMessage{
			{Role: "system", Content: prompt.SystemMessage},
			{Role: "user", Content: prompt.UserMessage},
		},
		MaxTokens:   d.maxTok,
		Temperature: d.temp,
	}
	return json.Marshal(req)
}

func (d *DeepSeek) doRequest(ctx context.Context, body []byte) ([]byte, error) {
	url := fmt.Sprintf("%s/chat/completions", d.endpoint)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, apperrors.Wrap(apperrors.KindInternal, "runtime.deepseek", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+d.apiKey)

	resp, err := d.client.Do(req)
	if err != nil {
		return nil, classifyHTTPError("runtime.deepseek", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.KindInternal, "runtime.deepseek", err)
	}

	if err := classifyStatusCode("runtime.deepseek", resp.StatusCode, respBody); err != nil {
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
