package runtime

import (
	"context"
	"net/http"
	"time"

	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
)

type Config struct {
	Provider    string
	Model       string
	APIKey      string
	BaseURL     string
	MaxTokens   int
	Temperature float64
	Timeout     time.Duration
}

type Prompt struct {
	SystemMessage string
	UserMessage   string
	WorkUnitID    string
	TaskID        string
}

type Result struct {
	Status        domain.RunResult
	Output        string
	FailureReason string
	Provider      string
	Model         string
	TokensUsed    int
	Latency       time.Duration
}

type StreamChunk struct {
	Delta         string
	ThinkingDelta string
	TokensUsed    int
	IsThinking    bool
	IsFinal       bool
	Provider      string
	Model         string
}

type Runtime interface {
	Execute(ctx context.Context, wu *domain.WorkUnit, task *domain.Task) (*Result, error)
}

type StreamRuntime interface {
	Runtime
	ExecuteStream(ctx context.Context, wu *domain.WorkUnit, task *domain.Task) (<-chan StreamChunk, <-chan error)
}

type resolvedConfig struct {
	apiKey, model, endpoint string
	maxTok                  int
	temp                    float64
	client                  *http.Client
}

func resolveConfig(cfg Config, defModel, defEndpoint string, defTok int, defTemp float64, defTimeout time.Duration) resolvedConfig {
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = defTimeout
	}
	return resolvedConfig{
		apiKey: cfg.APIKey, model: strOr(cfg.Model, defModel),
		endpoint: strOr(cfg.BaseURL, defEndpoint), maxTok: intOr(cfg.MaxTokens, defTok),
		temp: floatOr(cfg.Temperature, defTemp), client: &http.Client{Timeout: timeout},
	}
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
