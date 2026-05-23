package deepseek

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/levygit837-cyber/OrchestraOS/internal/apperrors"
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
	"github.com/levygit837-cyber/OrchestraOS/internal/runtime"
	"github.com/levygit837-cyber/OrchestraOS/internal/sse"
)

type openAIStreamDelta struct {
	Content          string `json:"content"`
	ReasoningContent string `json:"reasoning_content,omitempty"`
}

type openAIStreamChoice struct {
	Delta openAIStreamDelta `json:"delta"`
}

type openAIStreamResponse struct {
	Choices []openAIStreamChoice `json:"choices"`
	Usage   *openAIUsage         `json:"usage,omitempty"`
}

func (d *DeepSeek) ExecuteStream(ctx context.Context, wu *domain.WorkUnit, task *domain.Task) (<-chan domain.StreamChunk, <-chan error) {
	chunks := make(chan domain.StreamChunk, 16)
	errs := make(chan error, 1)
	go d.runStream(ctx, wu, task, chunks, errs)
	return chunks, errs
}

func (d *DeepSeek) runStream(ctx context.Context, wu *domain.WorkUnit, task *domain.Task, chunks chan<- domain.StreamChunk, errs chan<- error) {
	defer close(chunks)
	defer close(errs)

	prompt := runtime.BuildPrompt(wu, task)
	body, err := d.buildStreamRequest(prompt)
	if err != nil {
		errs <- apperrors.Wrap(apperrors.KindStreamInitFailed, "deepseek.stream", err)
		return
	}

	resp, err := d.doStreamRequest(ctx, body)
	if err != nil {
		errs <- err
		return
	}
	defer func() { _ = resp.Body.Close() }()

	lines := make(chan sse.Line, 16)
	go sse.Parse(resp.Body, lines)
	d.consumeSSE(lines, chunks, errs)
}

func (d *DeepSeek) buildStreamRequest(prompt domain.Prompt) ([]byte, error) {
	req := openAIRequest{
		Model:  d.model,
		Stream: true,
		Messages: []openAIMessage{
			{Role: "system", Content: prompt.SystemMessage},
			{Role: "user", Content: prompt.UserMessage},
		},
		MaxTokens:   d.maxTok,
		Temperature: d.temp,
	}
	return json.Marshal(req)
}

func (d *DeepSeek) doStreamRequest(ctx context.Context, body []byte) (*http.Response, error) {
	url := fmt.Sprintf("%s/chat/completions", d.endpoint)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, apperrors.Wrap(apperrors.KindStreamInitFailed, "deepseek.stream", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+d.apiKey)

	resp, err := d.client.Do(req)
	if err != nil {
		return nil, classifyStreamInitError("deepseek.stream", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		defer func() { _ = resp.Body.Close() }()
		return nil, classifyStreamStatusCode("deepseek.stream", resp)
	}
	return resp, nil
}

func (d *DeepSeek) consumeSSE(lines <-chan sse.Line, chunks chan<- domain.StreamChunk, errs chan<- error) {
	for line := range lines {
		if line.IsDone {
			break
		}
		if err := d.emitChunk(line.Data, chunks); err != nil {
			errs <- err
			return
		}
	}
	chunks <- domain.StreamChunk{IsFinal: true, Provider: "deepseek", Model: d.model}
}

func (d *DeepSeek) emitChunk(data string, chunks chan<- domain.StreamChunk) error {
	var sr openAIStreamResponse
	if err := json.Unmarshal([]byte(data), &sr); err != nil {
		return apperrors.New(apperrors.KindStreamInterrupted, "deepseek.stream", "invalid JSON chunk: "+err.Error())
	}
	for _, choice := range sr.Choices {
		chunk := domain.StreamChunk{Provider: "deepseek", Model: d.model}
		if choice.Delta.ReasoningContent != "" {
			chunk.ThinkingDelta = choice.Delta.ReasoningContent
			chunk.IsThinking = true
		}
		if choice.Delta.Content != "" {
			chunk.Delta = choice.Delta.Content
		}
		if sr.Usage != nil {
			chunk.TokensUsed = sr.Usage.TotalTokens
		}
		chunks <- chunk
	}
	return nil
}
