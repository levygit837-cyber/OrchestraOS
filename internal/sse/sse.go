package sse

import (
	"bufio"
	"io"
	"strings"
)

const (
	dataPrefix    = "data: "
	doneSentinel  = "[DONE]"
)

// Line represents a single parsed SSE data line.
type Line struct {
	Data   string
	IsDone bool
}

// Parse reads an SSE stream and emits parsed data lines.
func Parse(r io.Reader, lines chan<- Line) {
	defer close(lines)
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		text := scanner.Text()
		if !strings.HasPrefix(text, dataPrefix) {
			continue
		}
		data := strings.TrimSpace(strings.TrimPrefix(text, dataPrefix))
		if data == doneSentinel {
			lines <- Line{IsDone: true}
			return
		}
		if data == "" {
			continue
		}
		lines <- Line{Data: data}
	}
}
