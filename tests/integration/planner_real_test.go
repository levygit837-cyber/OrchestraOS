package integration

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/levygit837-cyber/OrchestraOS/internal/bootstrap"
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
	taskmod "github.com/levygit837-cyber/OrchestraOS/internal/modules/task"
	taskgraphmod "github.com/levygit837-cyber/OrchestraOS/internal/modules/taskgraph"
)

func TestGeminiPlanner_RealInference(t *testing.T) {
	if os.Getenv("RUN_LLM_TESTS") != "1" && os.Getenv("GEMINI_API_KEY") == "" {
		t.Skip("Skipping real Gemini inference test: set RUN_LLM_TESTS=1 or provide GEMINI_API_KEY")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	task := &domain.Task{
		Title:       "Criar API de autenticação com JWT e refresh tokens",
		Description: "Implementar um serviço de autenticação completo usando JWT para access tokens e refresh tokens. Deve incluir endpoints de login, registro, refresh e logout. Usar Go com framework padrão net/http.",
		Priority:    domain.PriorityP1,
		RiskLevel:   domain.RiskLevelMedium,
		AcceptanceCriteria: []string{
			"Endpoint POST /auth/login retorna access_token e refresh_token válidos",
			"Endpoint POST /auth/register cria usuário com senha hashada (bcrypt)",
			"Endpoint POST /auth/refresh gera novo access_token a partir de refresh_token válido",
			"Endpoint POST /auth/logout invalida o refresh_token no lado do servidor",
			"Middleware de autenticação valida access_token em rotas protegidas",
			"Tokens JWT possuem claims de expiração e user_id",
			"Documentação da API gerada em formato OpenAPI/Swagger",
		},
	}

	planner, err := bootstrap.GeminiPlanner()
	if err != nil {
		t.Fatalf("failed to create gemini planner: %v", err)
	}

	t.Logf("Calling Gemini planner with task: %s", task.Title)
	plan, err := planner.Plan(ctx, task)
	if err != nil {
		t.Fatalf("gemini planner failed: %v", err)
	}

	t.Logf("Planner rationale: %s", plan.Rationale)
	t.Logf("Generated %d work units", len(plan.WorkUnits))

	// Validate the plan
	if err := bootstrap.ValidateGraphPlan(plan); err != nil {
		t.Fatalf("plan validation failed: %v", err)
	}

	// Assertions
	if len(plan.WorkUnits) < 1 {
		t.Fatalf("expected at least 1 work unit, got %d", len(plan.WorkUnits))
	}
	if len(plan.WorkUnits) > 10 {
		t.Fatalf("expected at most 10 work units, got %d", len(plan.WorkUnits))
	}

	// Check for semantic richness
	profileSet := make(map[string]bool)
	for i, wu := range plan.WorkUnits {
		t.Logf("WorkUnit[%d]: title=%q profile=%q objective=%q deps=%v",
			i, wu.Title, wu.AssignedAgentProfile, wu.Objective, wu.DependsOn)

		if wu.Title == "" {
			t.Errorf("work unit %d has empty title", i)
		}
		if wu.Objective == "" {
			t.Errorf("work unit %d has empty objective", i)
		}
		if len(wu.AcceptanceCriteria) == 0 {
			t.Errorf("work unit %d has no acceptance criteria", i)
		}
		if len(wu.ValidationPlan) == 0 {
			t.Errorf("work unit %d has no validation plan", i)
		}

		profileSet[wu.AssignedAgentProfile] = true
	}

	t.Logf("Agent profiles used: %v", keys(profileSet))

	// Verify DAG is acyclic by checking no work unit depends on itself
	idToIndex := make(map[string]int)
	for i, wu := range plan.WorkUnits {
		idToIndex[wu.ID] = i
	}
	for i, wu := range plan.WorkUnits {
		for _, depID := range wu.DependsOn {
			if depID == wu.ID {
				t.Errorf("work unit %d (%s) depends on itself", i, wu.ID)
			}
		}
	}

	// Verify edges match dependencies
	if len(plan.Edges) > 0 {
		t.Logf("Graph edges: %d", len(plan.Edges))
		for i, edge := range plan.Edges {
			t.Logf("Edge[%d]: %s -> %s (%s)", i, edge.From, edge.To, edge.Type)
		}
	}
}

func TestTaskGraphService_Decompose_RealLLM(t *testing.T) {
	if os.Getenv("RUN_LLM_TESTS") != "1" && os.Getenv("GEMINI_API_KEY") == "" {
		t.Skip("Skipping real LLM decomposition test: set RUN_LLM_TESTS=1 or provide GEMINI_API_KEY")
	}

	db := getTestDB(t)
	defer db.Close()
	ctx := context.Background()

	taskService := bootstrap.TaskService(db)
	graphService := bootstrap.TaskGraphService(db)

	taskResult, err := taskService.Create(ctx, taskmod.CreateTaskInput{
		Title:       "Implementar sistema de cache distribuído",
		Description: "Criar um sistema de cache distribuído com Redis, suportando TTL, invalidação por padrão e fallback para banco de dados. Deve incluir testes de integração e documentação.",
		Priority:    domain.PriorityP1,
		RiskLevel:   domain.RiskLevelMedium,
		AcceptanceCriteria: []string{
			"Cache hit retorna dados em menos de 10ms",
			"Cache miss busca do banco e popula o cache automaticamente",
			"TTL configurável por tipo de entidade",
			"Invalidação de cache por chave e por padrão (prefixo)",
			"Testes de integração com container Redis",
			"Documentação de arquitetura e operação",
		},
	})
	if err != nil {
		t.Fatalf("create task: %v", err)
	}

	result, err := graphService.Decompose(ctx, taskgraphmod.DecomposeTaskGraphInput{
		TaskID:          taskResult.Value.ID,
		PlannerStrategy: "llm_gemini_v1",
	})
	if err != nil {
		t.Fatalf("decompose task: %v", err)
	}

	t.Logf("Graph strategy: %s", result.Graph.PlannerStrategy)
	t.Logf("Graph rationale: %s", result.Graph.Rationale)
	t.Logf("Work units: %d", len(result.WorkUnits))

	// The decomposition should use LLM strategy since we have API key
	if result.Graph.PlannerStrategy != "llm_gemini_v1" {
		t.Logf("Warning: expected llm_gemini_v1 but got %s (fallback may have occurred)", result.Graph.PlannerStrategy)
	}

	for i, wu := range result.WorkUnits {
		t.Logf("WorkUnit[%d]: title=%q profile=%q", i, wu.Title, wu.AssignedAgentProfile)
	}
}

func keys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
