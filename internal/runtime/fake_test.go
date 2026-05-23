package runtime_test

import (
	"context"
	"strings"
	"testing"

	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
	"github.com/levygit837-cyber/OrchestraOS/internal/runtime"
)

func TestFakeExecuteStream_EmitsChunksAndFinal(t *testing.T) {
	t.Parallel()
	f := runtime.NewFake()
	wu := &domain.WorkUnit{ID: "wu-f1", Title: "hello"}
	task := &domain.Task{ID: "t-f1", Title: "task"}

	chunks, errs := f.ExecuteStream(context.Background(), wu, task)

	var parts []string
	var gotFinal bool
	for chunk := range chunks {
		if chunk.IsFinal {
			gotFinal = true
			continue
		}
		parts = append(parts, chunk.Delta)
		if chunk.Provider != "fake" {
			t.Errorf("expected provider fake, got %s", chunk.Provider)
		}
	}
	if err := <-errs; err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !gotFinal {
		t.Error("expected final chunk")
	}
	combined := strings.Join(parts, "")
	if combined != "fake execution completed for: hello" {
		t.Errorf("unexpected output: %q", combined)
	}
}
