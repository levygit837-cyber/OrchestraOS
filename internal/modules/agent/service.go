// LLM AGENT: BEFORE MODIFYING THIS FILE, READ:
//   1. README.md  in this directory -> purpose, file map, dependencies
//   2. CONTRACTS.md in this directory -> invariants, state machine, boundary rules
// Ignoring these files will cause architecture test failures.

package agent

import (
	"context"
	"database/sql"

	"github.com/levygit837-cyber/OrchestraOS/internal/core/apperrors"
	dbcore "github.com/levygit837-cyber/OrchestraOS/internal/core/db"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/serialization"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/transition"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/validation"
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
)

type AgentService struct {
	db *sql.DB
}

type CreateAgentInput struct {
	ID                     string
	Name                   string
	Profile                string
	Capabilities           []string
	AllowedTools           []string
	DefaultPromptFragments []string
	RuntimeType            domain.AgentRuntimeType
}

func NewAgentService(database *sql.DB) *AgentService {
	return &AgentService{db: database}
}

func (s *AgentService) Create(ctx context.Context, input CreateAgentInput) (*transition.OperationResult[*domain.Agent], error) {
	op := "agent_service.create"

	if err := ValidateName(input.Name, op); err != nil {
		return nil, err
	}
	if err := ValidateProfile(input.Profile, op); err != nil {
		return nil, err
	}
	if err := ValidateRuntimeType(input.RuntimeType, op); err != nil {
		return nil, err
	}

	tx, err := dbcore.BeginTx(ctx, s.db, "agent_service.begin_create")
	if err != nil {
		return nil, err
	}
	defer dbcore.RollbackTx(tx)

	agent := &domain.Agent{
		ID:                     input.ID,
		Name:                   input.Name,
		Profile:                input.Profile,
		Capabilities:           input.Capabilities,
		AllowedTools:           input.AllowedTools,
		DefaultPromptFragments: input.DefaultPromptFragments,
		RuntimeType:            input.RuntimeType,
	}

	if err := NewRepository(tx).Create(agent); err != nil {
		return nil, apperrors.Wrap(apperrors.CodePersistence, "agent_service.create_projection", err)
	}

	payload, err := serialization.MarshalPayload("agent_service.create_payload", map[string]interface{}{
		"agent_id":      agent.ID,
		"name":          agent.Name,
		"profile":       agent.Profile,
		"runtime_type":  agent.RuntimeType,
		"capabilities":  agent.Capabilities,
		"allowed_tools": agent.AllowedTools,
	})
	if err != nil {
		return nil, err
	}

	appendResult, err := transition.AppendServiceEvent(ctx, tx, &domain.EventEnvelope{
		ID:          "", // Let transition generate
		Type:        EventCreated,
		Version:     transition.EventVersionV1,
		AgentID:     agent.ID,
		Priority:    domain.EventPriorityNotification,
		RequiresAck: false,
		Payload:     payload,
	})
	if err != nil {
		return nil, err
	}

	if err := dbcore.CommitTx(tx, "agent_service.commit_create"); err != nil {
		return nil, err
	}

	return &transition.OperationResult[*domain.Agent]{Value: agent, Event: &appendResult.Event, Duplicate: appendResult.Duplicate}, nil
}

func (s *AgentService) GetByID(ctx context.Context, id string) (*domain.Agent, error) {
	op := "agent_service.get_by_id"
	if err := validation.RequiredUUID(id, "agent_id", op); err != nil {
		return nil, err
	}
	agent, err := NewRepository(s.db).GetByID(ctx, id)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodePersistence, op, err)
	}
	if agent == nil {
		return nil, apperrors.New(apperrors.CodeNotFound, op, "agent not found")
	}
	return agent, nil
}

func (s *AgentService) FindOrCreate(ctx context.Context, profile string, runtimeType domain.AgentRuntimeType) (*domain.Agent, error) {
	op := "agent_service.find_or_create"

	if err := ValidateProfile(profile, op); err != nil {
		return nil, err
	}
	if err := ValidateRuntimeType(runtimeType, op); err != nil {
		return nil, err
	}

	// Try to find existing active agent
	agent, err := NewRepository(s.db).FindByProfileAndRuntime(profile, runtimeType)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodePersistence, op, err)
	}
	if agent != nil {
		return agent, nil
	}

	// Create new agent
	result, err := s.Create(ctx, CreateAgentInput{
		Name:        profile + " agent",
		Profile:     profile,
		RuntimeType: runtimeType,
	})
	if err != nil {
		return nil, err
	}
	return result.Value, nil
}
