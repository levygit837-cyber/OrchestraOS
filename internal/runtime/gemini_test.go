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

func geminiSSEChunk(text string, thought bool) string {
	thoughtField := ""
	if thought {
		thoughtField = `,"thought":true`
	}
	return fmt.Sprintf(`{"candidates":[{"content":{"parts":[{"text":%q%s}]}}]}`, text, thoughtField)
}

func geminiTestWU() *domain.WorkUnit {
	return &domain.WorkUnit{
		ID:                 "wu-g1",
		Title:              "test gemini",
		Objective:          "test",
		AcceptanceCriteria: []string{"pass"},
	}
}

func geminiTestTask() *domain.Task {
	return &domain.Task{ID: "t-g1", Title: "test task", Description: "desc"}
}

func TestGeminiExecuteStream_NormalChunks(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprintf(w, "data: %s\n\n", geminiSSEChunk("Hello", false))
		fmt.Fprintf(w, "data: %s\n\n", geminiSSEChunk(" world", false))
		fmt.Fprint(w, "data: [DONE]\n\n")
	}))
	defer srv.Close()

	g := runtime.NewGemini(runtime.Config{APIKey: "test", BaseURL: srv.URL})
	chunks, errs := g.ExecuteStream(context.Background(), geminiTestWU(), geminiTestTask())

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

func TestGeminiExecuteStream_ThinkingChunks(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprintf(w, "data: %s\n\n", geminiSSEChunk("reasoning...", true))
		fmt.Fprintf(w, "data: %s\n\n", geminiSSEChunk("answer", false))
		fmt.Fprint(w, "data: [DONE]\n\n")
	}))
	defer srv.Close()

	g := runtime.NewGemini(runtime.Config{APIKey: "test", BaseURL: srv.URL})
	chunks, errs := g.ExecuteStream(context.Background(), geminiTestWU(), geminiTestTask())

	var thinking, content []string
	for chunk := range chunks {
		if chunk.IsFinal {
			continue
		}
		if chunk.IsThinking {
			thinking = append(thinking, chunk.ThinkingDelta)
		} else {
			content = append(content, chunk.Delta)
		}
	}
	if err := <-errs; err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(thinking) != 1 || thinking[0] != "reasoning..." {
		t.Errorf("unexpected thinking: %v", thinking)
	}
	if len(content) != 1 || content[0] != "answer" {
		t.Errorf("unexpected content: %v", content)
	}
}

func TestGeminiExecuteStream_HTTPError(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "server error")
	}))
	defer srv.Close()

	g := runtime.NewGemini(runtime.Config{APIKey: "test", BaseURL: srv.URL})
	chunks, errs := g.ExecuteStream(context.Background(), geminiTestWU(), geminiTestTask())

	for range chunks {
	}
	if err := <-errs; err == nil {
		t.Fatal("expected error for HTTP 500")
	}
}

func TestGeminiExecuteStream_InvalidJSON(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, "data: not-json\n\n")
	}))
	defer srv.Close()

	g := runtime.NewGemini(runtime.Config{APIKey: "test", BaseURL: srv.URL})
	chunks, errs := g.ExecuteStream(context.Background(), geminiTestWU(), geminiTestTask())

	for range chunks {
	}
	if err := <-errs; err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}
