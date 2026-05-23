package decomposer

import (
	"context"
	"fmt"
	"time"

	"github.com/levygit837-cyber/OrchestraOS/internal/apperrors"
	"github.com/levygit837-cyber/OrchestraOS/internal/daggen"
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
	"github.com/levygit837-cyber/OrchestraOS/internal/retry"
)

// EventHandler receives pipeline lifecycle events.
type EventHandler func(eventType, taskID string, detail string)

// PipelineConfig controls decomposition behavior.
type PipelineConfig struct {
	Retry   retry.Config
	OnEvent EventHandler
}

// DefaultPipelineConfig returns sensible defaults.
func DefaultPipelineConfig() PipelineConfig {
	return PipelineConfig{
		Retry: retry.Config{
			MaxRetries:    3,
			InitialDelay:  200 * time.Millisecond,
			MaxDelay:      10 * time.Second,
			BackoffFactor: 2.0,
		},
	}
}

// Pipeline orchestrates the full decomposition flow:
// strategy → validate DAG → build work units.
type Pipeline struct {
	strategy Strategy
	cfg      PipelineConfig
}

// NewPipeline creates a Pipeline with the given strategy and config.
func NewPipeline(strategy Strategy, cfg PipelineConfig) *Pipeline {
	return &Pipeline{strategy: strategy, cfg: cfg}
}

// Run decomposes a task into a validated DAG and work units.
func (p *Pipeline) Run(ctx context.Context, task *domain.Task) (*PipelineResult, error) {
	req := buildRequest(task)
	p.emit(EventDecompositionStarted, task.ID, task.Title)

	result, err := p.decompose(ctx, req)
	if err != nil {
		p.emit(EventDecompositionFailed, task.ID, err.Error())
		return nil, apperrors.Wrap(apperrors.KindDecomposition, "decomposer.pipeline", err)
	}

	p.emit(EventGraphBuildStarted, task.ID, result.Rationale)
	graph, err := p.buildAndValidate(task.ID, result)
	if err != nil {
		return nil, err
	}

	p.emit(EventWorkUnitsCreated, task.ID, formatWUCount(result.WorkUnits))
	wus, err := daggen.BuildWorkUnits(task, graph, result.WorkUnits)
	if err != nil {
		return nil, err
	}

	p.emit(EventDecompositionCompleted, task.ID, p.strategy.Name())
	return &PipelineResult{
		Graph: graph, WorkUnits: wus,
		Rationale: result.Rationale, Strategy: p.strategy.Name(),
	}, nil
}

func (p *Pipeline) decompose(ctx context.Context, req *domain.DecompositionRequest) (*domain.DecompositionResult, error) {
	var result *domain.DecompositionResult
	err := retry.Do(ctx, func(ctx context.Context) error {
		r, err := p.strategy.Decompose(ctx, req)
		if err != nil {
			p.emit(EventRetryAttempted, req.TaskID, err.Error())
			return err
		}
		result = r
		return nil
	}, p.cfg.Retry)
	return result, err
}

func (p *Pipeline) buildAndValidate(taskID string, result *domain.DecompositionResult) (*domain.DAGGraph, error) {
	graph, err := daggen.BuildGraph(result)
	if err != nil {
		p.emit(EventGraphRejected, taskID, err.Error())
		return nil, apperrors.Wrap(apperrors.KindGraphValidation, "decomposer.pipeline", err)
	}
	p.emit(EventGraphValidated, taskID, graph.ID)
	return graph, nil
}

func (p *Pipeline) emit(eventType, taskID, detail string) {
	if p.cfg.OnEvent != nil {
		p.cfg.OnEvent(eventType, taskID, detail)
	}
}

func buildRequest(task *domain.Task) *domain.DecompositionRequest {
	return &domain.DecompositionRequest{
		TaskID:   task.ID,
		RawInput: task.Title + ": " + task.Description,
		Context: domain.TaskContext{
			TaskID:      task.ID,
			Intent:      task.Title,
			RawInput:    task.Description,
			Domains:     []string{},
			Constraints: task.AcceptanceCriteria,
		},
		CreatedAt: time.Now(),
	}
}

func formatWUCount(specs []domain.WUSpec) string {
	if len(specs) == 1 {
		return "1 work unit"
	}
	return fmt.Sprintf("%d work units", len(specs))
}

// PipelineResult holds the output of a successful pipeline run.
type PipelineResult struct {
	Graph     *domain.DAGGraph
	WorkUnits []domain.WorkUnit
	Rationale string
	Strategy  string
}

// ToPlan converts a PipelineResult into a planner-compatible plan.
func (pr *PipelineResult) ToPlan() *Plan {
	return &Plan{
		GraphID:   pr.Graph.ID,
		WorkUnits: pr.WorkUnits,
		Rationale: pr.Rationale,
		Strategy:  pr.Strategy,
	}
}

// Plan is the decomposer output compatible with the planner interface.
type Plan struct {
	GraphID   string
	WorkUnits []domain.WorkUnit
	Rationale string
	Strategy  string
}

// PlannerAdapter wraps a Pipeline to implement the planner.Planner interface.
type PlannerAdapter struct {
	pipeline *Pipeline
}

// NewPlannerAdapter creates an adapter from a Pipeline.
func NewPlannerAdapter(p *Pipeline) *PlannerAdapter {
	return &PlannerAdapter{pipeline: p}
}
