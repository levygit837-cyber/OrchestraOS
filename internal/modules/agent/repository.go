package agent

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/apperrors"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/db"
)

// Repository handles agent persistence
type Repository struct {
	db db.DBTX
}

// NewRepository creates a new agent repository
func NewRepository(db db.DBTX) *Repository {
	return &Repository{db: db}
}

// Create inserts a new agent
func (r *Repository) Create(agent *Agent) error {
	if agent.ID == "" {
		agent.ID = uuid.New().String()
	}

	now := time.Now().UTC()

	_, err := r.db.Exec(
		QueryInsert,
		agent.ID,
		agent.Name,
		agent.Profile,
		textArray(agent.Capabilities),
		textArray(agent.AllowedTools),
		textArray(agent.DefaultPromptFragments),
		agent.RuntimeType,
		"active", // status
		now,
		now,
	)
	if err != nil {
		return apperrors.Wrap(apperrors.CodePersistence, "repository.create", err)
	}

	return nil
}

// GetByID retrieves an agent by ID
func (r *Repository) GetByID(ctx context.Context, id string) (*Agent, error) {
	row := r.db.QueryRowContext(ctx, QueryGetByID, id)
	return r.scanAgent(row)
}

// FindByProfileAndRuntime finds an active agent by profile and runtime type
func (r *Repository) FindByProfileAndRuntime(profile string, runtimeType RuntimeType) (*Agent, error) {
	row := r.db.QueryRow(QueryFindByProfileAndRuntime, profile, runtimeType)
	return r.scanAgent(row)
}

// List retrieves all agents
func (r *Repository) List() ([]*Agent, error) {
	rows, err := r.db.Query(QueryList)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodePersistence, "repository.list", err)
	}
	defer rows.Close()

	var agents []*Agent
	for rows.Next() {
		agent, err := r.scanAgent(rows)
		if err != nil {
			return nil, err
		}
		agents = append(agents, agent)
	}

	return agents, nil
}

func (r *Repository) scanAgent(scanner interface {
	Scan(dest ...interface{}) error
}) (*Agent, error) {
	var agent Agent
	var capabilities, allowedTools, defaultPromptFragments []string
	var status string
	var createdAt, updatedAt time.Time

	err := scanner.Scan(
		&agent.ID,
		&agent.Name,
		&agent.Profile,
		&capabilities,
		&allowedTools,
		&defaultPromptFragments,
		&agent.RuntimeType,
		&status,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, apperrors.Wrap(apperrors.CodePersistence, "repository.scan_agent", err)
	}

	agent.Capabilities = capabilities
	agent.AllowedTools = allowedTools
	agent.DefaultPromptFragments = defaultPromptFragments

	return &agent, nil
}

func textArray(arr []string) interface{} {
	return arr
}
