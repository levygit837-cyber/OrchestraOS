package decomposer_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/levygit837-cyber/OrchestraOS/internal/decomposer"
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
)

// deepseekRuntime implements decomposer.AgentRuntime using the DeepSeek API.
type deepseekRuntime struct {
	apiKey string
	client *http.Client
}

func (d *deepseekRuntime) Execute(ctx context.Context, prompt *domain.Prompt) (string, error) {
	body, err := json.Marshal(map[string]interface{}{
		"model": "deepseek-chat",
		"messages": []map[string]string{
			{"role": "system", "content": prompt.SystemMessage},
			{"role": "user", "content": prompt.UserMessage},
		},
		"max_tokens":  4096,
		"temperature": 0.3,
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://api.deepseek.com/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+d.apiKey)

	resp, err := d.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("deepseek request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("deepseek status %d: %s", resp.StatusCode, respBody)
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("parse response: %w", err)
	}
	if len(result.Choices) == 0 {
		return "", fmt.Errorf("empty choices from deepseek")
	}
	return result.Choices[0].Message.Content, nil
}

func TestIntegration_DeepSeekDAGGeneration(t *testing.T) {
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		t.Skip("DEEPSEEK_API_KEY not set, skipping integration test")
	}

	rt := &deepseekRuntime{
		apiKey: apiKey,
		client: &http.Client{Timeout: 90 * time.Second},
	}

	var events []string
	cfg := decomposer.DefaultPipelineConfig()
	cfg.OnEvent = func(eventType, taskID, detail string) {
		events = append(events, eventType)
		t.Logf("EVENT: %s | task=%s | %s", eventType, taskID, detail)
	}

	strategy := decomposer.NewAgentStrategy(rt)
	pipeline := decomposer.NewPipeline(strategy, cfg)

	task := &domain.Task{
		ID:    "task-integration-001",
		Title: "User Authentication and Profile Dashboard",
		Description: "Build a user authentication system with login, signup, " +
			"and password reset. Then create a profile dashboard that shows " +
			"user info and recent activity. The auth system needs JWT tokens " +
			"and the dashboard needs to fetch data from a REST API.",
		AcceptanceCriteria: []string{
			"User can sign up with email and password",
			"User can login and receive a JWT token",
			"User can reset their password via email",
			"Dashboard displays user profile information",
			"Dashboard shows recent activity feed from API",
		},
		Status: domain.TaskStatusCreated,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	result, err := pipeline.Run(ctx, task)
	if err != nil {
		t.Fatalf("Pipeline.Run failed: %v", err)
	}

	// Validate the DAG graph
	if result.Graph == nil {
		t.Fatal("expected non-nil graph")
	}
	if result.Graph.ID == "" {
		t.Error("expected non-empty graph ID")
	}
	if result.Graph.Status != "validated" {
		t.Errorf("expected graph status=validated, got %s", result.Graph.Status)
	}
	if len(result.Graph.Nodes) < 2 {
		t.Errorf("expected at least 2 nodes, got %d", len(result.Graph.Nodes))
	}

	// Validate work units
	if len(result.WorkUnits) < 2 {
		t.Errorf("expected at least 2 work units, got %d", len(result.WorkUnits))
	}

	// Log the decomposition result
	t.Logf("Strategy: %s", result.Strategy)
	t.Logf("Rationale: %s", result.Rationale)
	t.Logf("Graph ID: %s", result.Graph.ID)
	t.Logf("Nodes: %d", len(result.Graph.Nodes))
	t.Logf("Edges: %d", len(result.Graph.Edges))
	t.Logf("Work Units: %d", len(result.WorkUnits))

	for i, wu := range result.WorkUnits {
		t.Logf("  WU[%d]: %s (status=%s, agent=%s)",
			i, wu.Title, wu.Status, wu.AssignedAgentProfile)
	}

	// Verify every WU references the parent task
	for _, wu := range result.WorkUnits {
		if wu.TaskID != task.ID {
			t.Errorf("WU %s has TaskID=%s, expected %s", wu.ID, wu.TaskID, task.ID)
		}
	}

	// Verify events were emitted
	if len(events) == 0 {
		t.Error("expected events to be emitted")
	}
	expectedEvents := []string{
		decomposer.EventDecompositionStarted,
		decomposer.EventGraphBuildStarted,
		decomposer.EventGraphValidated,
		decomposer.EventWorkUnitsCreated,
		decomposer.EventDecompositionCompleted,
	}
	for _, expected := range expectedEvents {
		found := false
		for _, e := range events {
			if e == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected event %s not found in emitted events: %v", expected, events)
		}
	}

	// Validate domain context separation
	domains := make(map[string]int)
	for _, node := range result.Graph.Nodes {
		domains[node.Context.Domain]++
	}
	t.Logf("Domains found: %v", domains)
	if len(domains) < 2 {
		t.Errorf("expected at least 2 distinct domains, got %d: %v", len(domains), domains)
	}
}

func TestIntegration_DeepSeekMultiDomainTask(t *testing.T) {
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		t.Skip("DEEPSEEK_API_KEY not set, skipping integration test")
	}

	rt := &deepseekRuntime{
		apiKey: apiKey,
		client: &http.Client{Timeout: 90 * time.Second},
	}

	cfg := decomposer.DefaultPipelineConfig()
	strategy := decomposer.NewAgentStrategy(rt)
	pipeline := decomposer.NewPipeline(strategy, cfg)

	task := &domain.Task{
		ID:    "task-integration-002",
		Title: "E-commerce Order Processing",
		Description: "Implement an order processing pipeline: validate payment " +
			"through a payment gateway, update inventory in the database, " +
			"send confirmation email to the user, and update the order " +
			"tracking dashboard in the UI.",
		AcceptanceCriteria: []string{
			"Payment is validated via payment gateway API",
			"Inventory is decremented in the database",
			"Confirmation email is sent to the customer",
			"Order tracking dashboard shows the new order",
		},
		Status: domain.TaskStatusCreated,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	result, err := pipeline.Run(ctx, task)
	if err != nil {
		t.Fatalf("Pipeline.Run failed: %v", err)
	}

	if len(result.WorkUnits) < 3 {
		t.Errorf("expected at least 3 work units for multi-domain task, got %d",
			len(result.WorkUnits))
	}

	// Log the decomposition
	t.Logf("Rationale: %s", result.Rationale)
	for i, wu := range result.WorkUnits {
		t.Logf("  WU[%d]: %s | objective=%s", i, wu.Title, wu.Objective)
	}

	// Verify DAG is valid with dependencies
	if len(result.Graph.Edges) == 0 {
		t.Log("Warning: no edges in graph (tasks may be independent)")
	}
}
