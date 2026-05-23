package runtime_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
	"github.com/levygit837-cyber/OrchestraOS/internal/runtime"
)

func deepseekSSEChunk(content, reasoning string) string {
	rc := ""
	if reasoning != "" {
		rc = fmt.Sprintf(`,"reasoning_content":%q`, reasoning)
	}
	return fmt.Sprintf(`{"choices":[{"delta":{"content":%q%s}}]}`, content, rc)
}

func deepseekTestWU() *domain.WorkUnit {
	return &domain.WorkUnit{
		ID:                 "wu-d1",
		Title:              "test deepseek",
		Objective:          "test",
		AcceptanceCriteria: []string{"pass"},
	}
}

func deepseekTestTask() *domain.Task {
	return &domain.Task{ID: "t-d1", Title: "test task", Description: "desc"}
}

func TestDeepSeekExecuteStream_NormalChunks(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprintf(w, "data: %s\n\n", deepseekSSEChunk("Hello", ""))
		fmt.Fprintf(w, "data: %s\n\n", deepseekSSEChunk(" world", ""))
		fmt.Fprint(w, "data: [DONE]\n\n")
	}))
	defer srv.Close()

	d := runtime.NewDeepSeek(runtime.Config{APIKey: "test", BaseURL: srv.URL})
	chunks, errs := d.ExecuteStream(context.Background(), deepseekTestWU(), deepseekTestTask())

	var deltas []string
	var gotFinal bool
	for chunk := range chunks {
		if chunk.IsFinal {
			gotFinal = true
			continue
		}
		deltas = append(deltas, chunk.Delta)
	}
	if err := <-errs; err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !gotFinal {
		t.Error("expected final chunk")
	}
	if len(deltas) != 2 || deltas[0] != "Hello" || deltas[1] != " world" {
		t.Errorf("unexpected deltas: %v", deltas)
	}
}

func TestDeepSeekExecuteStream_ReasoningContent(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprintf(w, "data: %s\n\n", deepseekSSEChunk("", "thinking step"))
		fmt.Fprintf(w, "data: %s\n\n", deepseekSSEChunk("answer", ""))
		fmt.Fprint(w, "data: [DONE]\n\n")
	}))
	defer srv.Close()

	d := runtime.NewDeepSeek(runtime.Config{APIKey: "test", BaseURL: srv.URL})
	chunks, errs := d.ExecuteStream(context.Background(), deepseekTestWU(), deepseekTestTask())

	var thinking, content []string
	for chunk := range chunks {
		if chunk.IsFinal {
			continue
		}
		if chunk.IsThinking {
			thinking = append(thinking, chunk.ThinkingDelta)
		}
		if chunk.Delta != "" {
			content = append(content, chunk.Delta)
		}
	}
	if err := <-errs; err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(thinking) != 1 || thinking[0] != "thinking step" {
		t.Errorf("unexpected thinking: %v", thinking)
	}
	if len(content) != 1 || content[0] != "answer" {
		t.Errorf("unexpected content: %v", content)
	}
}

func TestDeepSeekExecuteStream_HTTPError(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, "unauthorized")
	}))
	defer srv.Close()

	d := runtime.NewDeepSeek(runtime.Config{APIKey: "bad", BaseURL: srv.URL})
	chunks, errs := d.ExecuteStream(context.Background(), deepseekTestWU(), deepseekTestTask())

	for range chunks {
	}
	if err := <-errs; err == nil {
		t.Fatal("expected error for HTTP 401")
	}
}

func TestDeepSeekExecuteStream_InvalidJSON(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, "data: {broken\n\n")
	}))
	defer srv.Close()

	d := runtime.NewDeepSeek(runtime.Config{APIKey: "test", BaseURL: srv.URL})
	chunks, errs := d.ExecuteStream(context.Background(), deepseekTestWU(), deepseekTestTask())

	for range chunks {
	}
	if err := <-errs; err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestDeepSeekExecuteStream_RateLimitError(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		fmt.Fprint(w, "rate limited")
	}))
	defer srv.Close()

	d := runtime.NewDeepSeek(runtime.Config{APIKey: "test", BaseURL: srv.URL})
	chunks, errs := d.ExecuteStream(context.Background(), deepseekTestWU(), deepseekTestTask())

	for range chunks {
	}
	err := <-errs
	if err == nil {
		t.Fatal("expected error for HTTP 429")
	}
}
