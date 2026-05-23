package runtime

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/levygit837-cyber/OrchestraOS/internal/apperrors"
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
	"github.com/levygit837-cyber/OrchestraOS/internal/sse"
)

type geminiStreamPart struct {
	Text    string `json:"text"`
	Thought bool   `json:"thought,omitempty"`
}

type geminiStreamCandidate struct {
	Content struct {
		Parts []geminiStreamPart `json:"parts"`
	} `json:"content"`
}

type geminiStreamResponse struct {
	Candidates []geminiStreamCandidate `json:"candidates"`
	UsageMeta  *geminiUsageMetada      `json:"usageMetadata,omitempty"`
}

// ExecuteStream opens a streaming connection to Gemini and emits chunks.
func (g *Gemini) ExecuteStream(ctx context.Context, wu *domain.WorkUnit, task *domain.Task) (<-chan StreamChunk, <-chan error) {
	chunks := make(chan StreamChunk, 16)
	errs := make(chan error, 1)
	go g.runStream(ctx, wu, task, chunks, errs)
	return chunks, errs
}

func (g *Gemini) runStream(ctx context.Context, wu *domain.WorkUnit, task *domain.Task, chunks chan<- StreamChunk, errs chan<- error) {
	defer close(chunks)
	defer close(errs)

	body, err := g.buildRequest(BuildPrompt(wu, task))
	if err != nil {
		errs <- apperrors.Wrap(apperrors.KindStreamInitFailed, "runtime.gemini.stream", err)
		return
	}

	resp, err := g.doStreamRequest(ctx, body)
	if err != nil {
		errs <- err
		return
	}
	defer func() { _ = resp.Body.Close() }()

	lines := make(chan sse.Line, 16)
	go sse.Parse(resp.Body, lines)
	g.consumeGeminiSSE(lines, chunks, errs)
}

func (g *Gemini) doStreamRequest(ctx context.Context, body []byte) (*http.Response, error) {
	url := fmt.Sprintf("%s/models/%s:streamGenerateContent?alt=sse&key=%s", g.endpoint, g.model, g.apiKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, apperrors.Wrap(apperrors.KindStreamInitFailed, "runtime.gemini.stream", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, classifyStreamInitError("runtime.gemini.stream", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		defer func() { _ = resp.Body.Close() }()
		return nil, classifyStreamStatusCode("runtime.gemini.stream", resp)
	}
	return resp, nil
}

func (g *Gemini) consumeGeminiSSE(lines <-chan sse.Line, chunks chan<- StreamChunk, errs chan<- error) {
	for line := range lines {
		if line.IsDone {
			break
		}
		if err := g.emitGeminiChunk(line.Data, chunks); err != nil {
			errs <- err
			return
		}
	}
	chunks <- StreamChunk{IsFinal: true, Provider: "gemini", Model: g.model}
}

func (g *Gemini) emitGeminiChunk(data string, chunks chan<- StreamChunk) error {
	var sr geminiStreamResponse
	if err := json.Unmarshal([]byte(data), &sr); err != nil {
		return newStreamInterruptedError("runtime.gemini.stream", "invalid JSON chunk: "+err.Error())
	}
	for _, cand := range sr.Candidates {
		for _, part := range cand.Content.Parts {
			chunk := StreamChunk{Provider: "gemini", Model: g.model}
			if part.Thought {
				chunk.ThinkingDelta = part.Text
				chunk.IsThinking = true
			} else {
				chunk.Delta = part.Text
			}
			if sr.UsageMeta != nil {
				chunk.TokensUsed = sr.UsageMeta.TotalTokenCount
			}
			chunks <- chunk
		}
	}
	return nil
}
