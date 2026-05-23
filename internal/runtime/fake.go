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
