package gemini

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
	defaultModel    = "gemini-2.0-flash"
	defaultEndpoint = "https://generativelanguage.googleapis.com/v1beta"
	defaultTokens   = 4096
	defaultTemp     = 0.7
	defaultTimeout  = 60 * time.Second
)

// Gemini implements domain.Runtime and domain.StreamRuntime for Google Gemini.
type Gemini struct {
	apiKey, model, endpoint string
	maxTok                  int
	temp                    float64
	client                  *http.Client
}

// New creates a Gemini runtime from the given config.
func New(cfg runtime.Config) *Gemini {
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = defaultTimeout
	}
	return &Gemini{
		apiKey:   cfg.APIKey,
		model:    strOr(cfg.Model, defaultModel),
		endpoint: strOr(cfg.BaseURL, defaultEndpoint),
		maxTok:   intOr(cfg.MaxTokens, defaultTokens),
		temp:     floatOr(cfg.Temperature, defaultTemp),
		client:   &http.Client{Timeout: timeout},
	}
}

func (g *Gemini) Execute(ctx context.Context, wu *domain.WorkUnit, task *domain.Task) (*domain.RuntimeResult, error) {
	prompt := runtime.BuildPrompt(wu, task)
	start := time.Now()
	body, err := g.buildRequest(prompt)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.KindInternal, "gemini.execute", err)
	}
	resp, err := g.doRequest(ctx, body)
	if err != nil {
		return nil, err
	}
	output, tokens, err := g.parseResponse(resp)
	if err != nil {
		return nil, err
	}
	return &domain.RuntimeResult{
		Status: domain.RunResultSucceeded, Output: output, Provider: "gemini",
		Model: g.model, TokensUsed: tokens, Latency: time.Since(start),
	}, nil
}

type geminiRequest struct {
	Contents         []geminiContent        `json:"contents"`
	SystemInstruct   *geminiContent         `json:"systemInstruction,omitempty"`
	GenerationConfig geminiGenerationConfig `json:"generationConfig"`
}

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
	Role  string       `json:"role,omitempty"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiGenerationConfig struct {
	MaxOutputTokens int     `json:"maxOutputTokens"`
	Temperature     float64 `json:"temperature"`
}

type geminiResponse struct {
	Candidates []geminiCandidate  `json:"candidates"`
	UsageMeta  *geminiUsageMetada `json:"usageMetadata,omitempty"`
}

type geminiCandidate struct {
	Content geminiContent `json:"content"`
}

type geminiUsageMetada struct {
	TotalTokenCount int `json:"totalTokenCount"`
}

func (g *Gemini) buildRequest(prompt domain.Prompt) ([]byte, error) {
	req := geminiRequest{
		SystemInstruct: &geminiContent{Parts: []geminiPart{{Text: prompt.SystemMessage}}},
		Contents: []geminiContent{{
			Role: "user", Parts: []geminiPart{{Text: prompt.UserMessage}},
		}},
		GenerationConfig: geminiGenerationConfig{MaxOutputTokens: g.maxTok, Temperature: g.temp},
	}
	return json.Marshal(req)
}

func (g *Gemini) doRequest(ctx context.Context, body []byte) ([]byte, error) {
	url := fmt.Sprintf("%s/models/%s:generateContent?key=%s", g.endpoint, g.model, g.apiKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, apperrors.Wrap(apperrors.KindInternal, "gemini.execute", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := g.client.Do(req)
	if err != nil {
		return nil, classifyHTTPError("gemini.execute", err)
	}
	defer func() { _ = resp.Body.Close() }()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.KindInternal, "gemini.execute", err)
	}
	if err := classifyStatusCode("gemini.execute", resp.StatusCode, respBody); err != nil {
		return nil, err
	}
	return respBody, nil
}

func (g *Gemini) parseResponse(body []byte) (string, int, error) {
	var resp geminiResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return "", 0, apperrors.Wrap(apperrors.KindInternal, "gemini.execute", err)
	}
	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", 0, apperrors.New(apperrors.KindInternal, "gemini.execute", "empty response from Gemini")
	}
	tokens := 0
	if resp.UsageMeta != nil {
		tokens = resp.UsageMeta.TotalTokenCount
	}
	return resp.Candidates[0].Content.Parts[0].Text, tokens, nil
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
