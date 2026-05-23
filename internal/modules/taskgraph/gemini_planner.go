package taskgraph

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/apperrors"
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"

	"google.golang.org/genai"
)

const (
	llmGeminiPlanner   = "llm_gemini_v1"
	plannerMaxSteps    = 1
	plannerTemperature = float32(0.1)
	plannerTimeout     = 30 * time.Second
)

// GeminiPlanner uses the Gemini API to decompose a Task into WorkUnits.
type GeminiPlanner struct {
	apiKey string
	model  string
}

// plannerOutput is the JSON structure returned by the LLM.
type plannerOutput struct {
	Rationale string            `json:"rationale"`
	WorkUnits []plannerWorkUnit `json:"work_units"`
}

type plannerWorkUnit struct {
	Title                string   `json:"title"`
	Objective            string   `json:"objective"`
	AssignedAgentProfile string   `json:"assigned_agent_profile"`
	AcceptanceCriteria   []string `json:"acceptance_criteria"`
	ValidationPlan       []string `json:"validation_plan"`
	DependsOn            []int    `json:"depends_on"`
	OwnedPaths           []string `json:"owned_paths"`
	ReadPaths            []string `json:"read_paths"`
}

// NewGeminiPlanner creates a new Gemini-based planner.
func NewGeminiPlanner() (*GeminiPlanner, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("GOOGLE_API_KEY")
	}
	if apiKey == "" {
		return nil, apperrors.New(apperrors.CodeRuntime, "gemini_planner.new", "GEMINI_API_KEY or GOOGLE_API_KEY environment variable is required")
	}

	model := os.Getenv("GEMINI_MODEL")
	if model == "" {
		model = "gemini-3-flash-preview"
	}

	return &GeminiPlanner{
		apiKey: apiKey,
		model:  model,
	}, nil
}

// Plan decomposes the given task into a GraphPlan using Gemini.
func (p *GeminiPlanner) Plan(ctx context.Context, task *domain.Task) (*GraphPlan, error) {
	op := "gemini_planner.plan"

	prompt, err := PlannerPrompt(task)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodeValidation, op+":prompt", err)
	}

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  p.apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodeRuntime, op+":client", err)
	}

	config := &genai.GenerateContentConfig{
		Temperature:      genai.Ptr(plannerTemperature),
		ResponseMIMEType: "application/json",
		ResponseSchema:   buildPlannerResponseSchema(),
	}

	ctx, cancel := context.WithTimeout(ctx, plannerTimeout)
	defer cancel()

	resp, err := client.Models.GenerateContent(ctx, p.model, []*genai.Content{genai.NewContentFromText(prompt, genai.RoleUser)}, config)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodeRuntime, op+":generate", err)
	}

	if resp == nil || len(resp.Candidates) == 0 {
		return nil, apperrors.New(apperrors.CodeRuntime, op, "empty response from gemini api")
	}

	candidate := resp.Candidates[0]
	if candidate.Content == nil || len(candidate.Content.Parts) == 0 {
		return nil, apperrors.New(apperrors.CodeRuntime, op, "candidate content is empty")
	}

	text := resp.Text()
	if text == "" {
		return nil, apperrors.New(apperrors.CodeRuntime, op, "response text is empty")
	}

	var output plannerOutput
	if err := json.Unmarshal([]byte(text), &output); err != nil {
		return nil, apperrors.Wrap(apperrors.CodeValidation, op+":unmarshal", fmt.Errorf("invalid JSON response: %w", err))
	}

	return p.convertToGraphPlan(task, &output)
}

// convertToGraphPlan transforms planner output into a GraphPlan with UUIDs and resolved dependencies.
func (p *GeminiPlanner) convertToGraphPlan(task *domain.Task, output *plannerOutput) (*GraphPlan, error) {
	op := "gemini_planner.convert"

	if output == nil {
		return nil, apperrors.New(apperrors.CodeValidation, op, "planner output is nil")
	}

	graphID := uuid.New().String()
	wuCount := len(output.WorkUnits)

	workUnits := make([]PlanWorkUnit, wuCount)
	idByIndex := make(map[int]string, wuCount)

	// First pass: generate IDs
	for i := range output.WorkUnits {
		idByIndex[i] = uuid.New().String()
	}

	// Second pass: build work units with resolved dependencies
	for i, wu := range output.WorkUnits {
		deps := make([]string, 0, len(wu.DependsOn))
		for _, depIdx := range wu.DependsOn {
			if depIdx < 0 || depIdx >= wuCount {
				return nil, apperrors.New(apperrors.CodeValidation, op,
					fmt.Sprintf("work unit %d depends_on index %d out of bounds (0-%d)", i, depIdx, wuCount-1))
			}
			deps = append(deps, idByIndex[depIdx])
		}

		profile := wu.AssignedAgentProfile
		if profile == "" {
			profile = "default"
		}

		workUnits[i] = PlanWorkUnit{
			ID:                   idByIndex[i],
			TaskID:               task.ID,
			TaskGraphID:          graphID,
			Title:                wu.Title,
			Objective:            wu.Objective,
			AssignedAgentProfile: profile,
			OwnedPaths:           wu.OwnedPaths,
			ReadPaths:            wu.ReadPaths,
			AcceptanceCriteria:   wu.AcceptanceCriteria,
			ValidationPlan:       wu.ValidationPlan,
			DependsOn:            deps,
		}
	}

	// Build edges
	edges := make([]TaskGraphEdgeInfo, 0)
	for i, wu := range output.WorkUnits {
		for _, depIdx := range wu.DependsOn {
			edges = append(edges, TaskGraphEdgeInfo{
				From:   idByIndex[depIdx],
				To:     workUnits[i].ID,
				Type:   "blocks",
				Reason: "semantic dependency from LLM planner",
			})
		}
	}

	// Build nodes
	nodes := make([]TaskGraphNodeInfo, 0, wuCount)
	for _, wu := range workUnits {
		nodes = append(nodes, TaskGraphNodeInfo{
			ID:                 wu.ID,
			Title:              wu.Title,
			Objective:          wu.Objective,
			AgentProfile:       wu.AssignedAgentProfile,
			OwnedPaths:         wu.OwnedPaths,
			ReadPaths:          wu.ReadPaths,
			AcceptanceCriteria: wu.AcceptanceCriteria,
			ValidationPlan:     wu.ValidationPlan,
		})
	}

	rationale := output.Rationale
	if rationale == "" {
		rationale = fmt.Sprintf("LLM decomposition via %s", p.model)
	}

	return &GraphPlan{
		GraphID:   graphID,
		WorkUnits: workUnits,
		Nodes:     nodes,
		Edges:     edges,
		Rationale: rationale,
	}, nil
}

// buildPlannerResponseSchema constructs the genai.Schema for structured generation.
func buildPlannerResponseSchema() *genai.Schema {
	stringSchema := &genai.Schema{Type: genai.TypeString}
	stringListSchema := &genai.Schema{
		Type:  genai.TypeArray,
		Items: stringSchema,
	}
	integerSchema := &genai.Schema{Type: genai.TypeInteger}
	integerListSchema := &genai.Schema{
		Type:  genai.TypeArray,
		Items: integerSchema,
	}
	profileEnumSchema := &genai.Schema{
		Type: genai.TypeString,
		Enum: []string{"code_worker", "docs_writer", "reviewer", "debugger", "default"},
	}

	workUnitSchema := &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"title":                  stringSchema,
			"objective":              stringSchema,
			"assigned_agent_profile": profileEnumSchema,
			"acceptance_criteria":    stringListSchema,
			"validation_plan":        stringListSchema,
			"depends_on":             integerListSchema,
			"owned_paths":            stringListSchema,
			"read_paths":             stringListSchema,
		},
		Required: []string{
			"title",
			"objective",
			"assigned_agent_profile",
			"acceptance_criteria",
			"validation_plan",
		},
	}

	workUnitListSchema := &genai.Schema{
		Type:     genai.TypeArray,
		Items:    workUnitSchema,
		MinItems: genai.Ptr(int64(1)),
		MaxItems: genai.Ptr(int64(maxPlannerWorkUnits)),
	}

	return &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"rationale":  stringSchema,
			"work_units": workUnitListSchema,
		},
		Required: []string{"rationale", "work_units"},
	}
}
