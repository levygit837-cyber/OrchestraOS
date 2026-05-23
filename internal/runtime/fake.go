package runtime

import (
	"context"

	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
)

// Fake is a test runtime that always succeeds.
type Fake struct{}

func NewFake() *Fake { return &Fake{} }

func (f *Fake) Execute(_ context.Context, wu *domain.WorkUnit, _ *domain.Task) (*Result, error) {
	return &Result{
		Status:   domain.RunResultSucceeded,
		Output:   "fake execution completed for: " + wu.Title,
		Provider: "fake",
		Model:    "fake",
	}, nil
}

// ExecuteStream emits simulated streaming chunks for testing.
func (f *Fake) ExecuteStream(_ context.Context, wu *domain.WorkUnit, _ *domain.Task) (<-chan StreamChunk, <-chan error) {
	chunks := make(chan StreamChunk, 4)
	errs := make(chan error, 1)
	go fakeStreamLoop(wu.Title, chunks, errs)
	return chunks, errs
}

func fakeStreamLoop(title string, chunks chan<- StreamChunk, errs chan<- error) {
	defer close(chunks)
	defer close(errs)

	parts := []string{"fake ", "execution ", "completed for: ", title}
	for _, p := range parts {
		chunks <- StreamChunk{Delta: p, Provider: "fake", Model: "fake"}
	}
	chunks <- StreamChunk{IsFinal: true, Provider: "fake", Model: "fake"}
}
