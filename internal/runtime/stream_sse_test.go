package runtime_test

import (
	"strings"
	"testing"

	"github.com/levygit837-cyber/OrchestraOS/internal/sse"
)

func TestParseSSEStream_NormalChunks(t *testing.T) {
	t.Parallel()
	input := "data: {\"text\":\"hello\"}\n\ndata: {\"text\":\"world\"}\n\n"
	lines := make(chan sse.Line, 8)
	go sse.Parse(strings.NewReader(input), lines)

	var got []sse.Line
	for l := range lines {
		got = append(got, l)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(got))
	}
	if got[0].Data != `{"text":"hello"}` {
		t.Errorf("chunk 0: got %q", got[0].Data)
	}
	if got[1].Data != `{"text":"world"}` {
		t.Errorf("chunk 1: got %q", got[1].Data)
	}
}

func TestParseSSEStream_DoneSentinel(t *testing.T) {
	t.Parallel()
	input := "data: {\"x\":1}\n\ndata: [DONE]\n\n"
	lines := make(chan sse.Line, 8)
	go sse.Parse(strings.NewReader(input), lines)

	var got []sse.Line
	for l := range lines {
		got = append(got, l)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(got))
	}
	if got[0].IsDone {
		t.Error("first line should not be done")
	}
	if !got[1].IsDone {
		t.Error("second line should be done")
	}
}

func TestParseSSEStream_SkipsNonDataLines(t *testing.T) {
	t.Parallel()
	input := ": comment\nevent: ping\ndata: {\"ok\":true}\n\n"
	lines := make(chan sse.Line, 8)
	go sse.Parse(strings.NewReader(input), lines)

	var got []sse.Line
	for l := range lines {
		got = append(got, l)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 line, got %d", len(got))
	}
}

func TestParseSSEStream_EmptyInput(t *testing.T) {
	t.Parallel()
	lines := make(chan sse.Line, 8)
	go sse.Parse(strings.NewReader(""), lines)

	var got []sse.Line
	for l := range lines {
		got = append(got, l)
	}
	if len(got) != 0 {
		t.Fatalf("expected 0 lines, got %d", len(got))
	}
}
