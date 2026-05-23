// LLM AGENT: BEFORE MODIFYING THIS FILE, READ:
//   1. README.md  in this directory -> purpose, file map, dependencies
//   2. CONTRACTS.md in this directory -> invariants, state machine, boundary rules
// Ignoring these files will cause architecture test failures.

package agent

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/apperrors"
	dbcore "github.com/levygit837-cyber/OrchestraOS/internal/core/db"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/serialization"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/transition"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/validation"
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
	"github.com/lib/pq"
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
	RuntimeType            RuntimeType
}

func NewAgentService(database *sql.DB) *AgentService {
	return &AgentService{db: database}
}

func (s *AgentService) Create(ctx context.Context, input CreateAgentInput) (*transition.OperationResult[*Agent], error) {
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

	if input.ID == "" {
		input.ID = uuid.New().String()
	}
	now := time.Now().UTC()
	agent := &Agent{
		ID:                     input.ID,
		Name:                   input.Name,
		Profile:                input.Profile,
		Capabilities:           input.Capabilities,
		AllowedTools:           input.AllowedTools,
		DefaultPromptFragments: input.DefaultPromptFragments,
		RuntimeType:            input.RuntimeType,
		Status:                 AgentStatusActive,
		CreatedAt:              now,
		UpdatedAt:              now,
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

	return &transition.OperationResult[*Agent]{Value: agent, Event: &appendResult.Event, Duplicate: appendResult.Duplicate}, nil
}

func (s *AgentService) GetByID(ctx context.Context, id string) (*Agent, error) {
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

func (s *AgentService) FindOrCreate(ctx context.Context, profile string, runtimeType RuntimeType) (*Agent, error) {
	op := "agent_service.find_or_create"

	if err := ValidateProfile(profile, op); err != nil {
		return nil, err
	}
	if err := ValidateRuntimeType(runtimeType, op); err != nil {
		return nil, err
	}

	tx, err := dbcore.BeginTx(ctx, s.db, "agent_service.find_or_create")
	if err != nil {
		return nil, err
	}
	defer dbcore.RollbackTx(tx)

	repo := NewRepository(tx)

	// Try to find existing active agent first (fast path)
	agent, err := repo.FindByProfileAndRuntime(profile, runtimeType)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodePersistence, op, err)
	}
	if agent != nil {
		if err := dbcore.CommitTx(tx, "agent_service.find_or_create"); err != nil {
			return nil, err
		}
		return agent, nil
	}

	// Atomically create; if a concurrent call already created it, the unique
	// constraint will raise a violation that we handle by selecting again.
	now := time.Now().UTC()
	agent = &Agent{
		ID:          uuid.New().String(),
		Name:        profile + " agent",
		Profile:     profile,
		RuntimeType: runtimeType,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := repo.Create(agent); err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			agent, findErr := repo.FindByProfileAndRuntime(profile, runtimeType)
			if findErr != nil {
				return nil, apperrors.Wrap(apperrors.CodePersistence, op, findErr)
			}
			if agent != nil {
				return agent, nil
			}
		}
		return nil, apperrors.Wrap(apperrors.CodePersistence, op, err)
	}

	if err := dbcore.CommitTx(tx, "agent_service.find_or_create"); err != nil {
		return nil, err
	}
	return agent, nil
}
