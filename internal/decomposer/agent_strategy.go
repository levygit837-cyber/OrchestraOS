package decomposer

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/levygit837-cyber/OrchestraOS/internal/apperrors"
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
)

// AgentRuntime defines the LLM execution contract for decomposition.
type AgentRuntime interface {
	Execute(ctx context.Context, prompt *domain.Prompt) (string, error)
}

// AgentStrategy uses an LLM agent to decompose tasks.
type AgentStrategy struct {
	runtime AgentRuntime
}

// NewAgentStrategy creates an agent-based decomposition strategy.
func NewAgentStrategy(rt AgentRuntime) *AgentStrategy {
	return &AgentStrategy{runtime: rt}
}

func (s *AgentStrategy) Name() string { return "agent_llm_v1" }

func (s *AgentStrategy) Decompose(ctx context.Context, req *domain.DecompositionRequest) (*domain.DecompositionResult, error) {
	prompt := buildDecomposePrompt(req)

	raw, err := s.runtime.Execute(ctx, prompt)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.KindGraphGeneration, "decomposer.agent", err)
	}

	return parseDecompositionResponse(req, raw)
}

func buildDecomposePrompt(req *domain.DecompositionRequest) *domain.Prompt {
	return &domain.Prompt{
		SystemMessage: systemPrompt,
		UserMessage:   formatUserPrompt(req),
		TaskID:        req.TaskID,
	}
}

const systemPrompt = `You are a task decomposition agent for OrchestraOS.
Given a task in natural language, break it into a DAG of work units.

Rules:
1. Each work unit must have a single context domain (e.g., "auth", "runtime", "database", "api", "ui").
2. Do not mix domains within a single work unit.
3. Define dependencies between work units where one depends on another.
4. Each work unit must have clear acceptance criteria.
5. Return valid JSON matching the expected schema.

Output JSON schema:
{
  "rationale": "string",
  "work_units": [
    {
      "title": "string",
      "objective": "string",
      "domain": "string",
      "description": "string",
      "acceptance_criteria": ["string"],
      "depends_on_indices": [int],
      "suggested_agent": "string"
    }
  ]
}`

func formatUserPrompt(req *domain.DecompositionRequest) string {
	return "Decompose this task:\n\n" + req.RawInput +
		"\n\nConstraints: " + formatList(req.Context.Constraints)
}

func formatList(items []string) string {
	if len(items) == 0 {
		return "none"
	}
	result := ""
	for i, item := range items {
		if i > 0 {
			result += "; "
		}
		result += item
	}
	return result
}

type llmResponse struct {
	Rationale string      `json:"rationale"`
	WorkUnits []llmWUSpec `json:"work_units"`
}

type llmWUSpec struct {
	Title              string   `json:"title"`
	Objective          string   `json:"objective"`
	Domain             string   `json:"domain"`
	Description        string   `json:"description"`
	AcceptanceCriteria []string `json:"acceptance_criteria"`
	DependsOnIndices   []int    `json:"depends_on_indices"`
	SuggestedAgent     string   `json:"suggested_agent"`
}

func parseDecompositionResponse(req *domain.DecompositionRequest, raw string) (*domain.DecompositionResult, error) {
	cleaned := stripCodeFences(raw)
	var resp llmResponse
	if err := json.Unmarshal([]byte(cleaned), &resp); err != nil {
		return nil, apperrors.Wrap(apperrors.KindDecomposition, "decomposer.agent.parse", err)
	}
	if len(resp.WorkUnits) == 0 {
		return nil, apperrors.New(apperrors.KindDecomposition, "decomposer.agent.parse", "LLM returned zero work units")
	}

	graphID := uuid.New().String()
	specs := convertLLMSpecs(resp.WorkUnits, graphID)

	return &domain.DecompositionResult{
		RequestID: uuid.New().String(),
		TaskID:    req.TaskID,
		Graph: domain.DAGGraph{
			ID:     graphID,
			TaskID: req.TaskID,
		},
		WorkUnits: specs,
		Rationale: resp.Rationale,
		Strategy:  "agent_llm_v1",
		CreatedAt: time.Now(),
	}, nil
}

func convertLLMSpecs(llmSpecs []llmWUSpec, graphID string) []domain.WUSpec {
	nodeIDs := make([]string, len(llmSpecs))
	for i := range llmSpecs {
		nodeIDs[i] = uuid.New().String()
	}

	specs := make([]domain.WUSpec, len(llmSpecs))
	for i, ls := range llmSpecs {
		deps := resolveDeps(ls.DependsOnIndices, nodeIDs)
		specs[i] = domain.WUSpec{
			NodeID:    nodeIDs[i],
			Title:     ls.Title,
			Objective: ls.Objective,
			Context: domain.WUContext{
				Domain:      ls.Domain,
				Description: ls.Description,
			},
			AcceptanceCriteria: ls.AcceptanceCriteria,
			DependsOn:          deps,
			SuggestedAgent:     ls.SuggestedAgent,
		}
	}
	return specs
}

func resolveDeps(indices []int, nodeIDs []string) []string {
	var deps []string
	for _, idx := range indices {
		if idx >= 0 && idx < len(nodeIDs) {
			deps = append(deps, nodeIDs[idx])
		}
	}
	if deps == nil {
		return []string{}
	}
	return deps
}

func stripCodeFences(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```") {
		if i := strings.Index(s, "\n"); i != -1 {
			s = s[i+1:]
		}
	}
	if strings.HasSuffix(s, "```") {
		s = s[:len(s)-3]
	}
	return strings.TrimSpace(s)
}
