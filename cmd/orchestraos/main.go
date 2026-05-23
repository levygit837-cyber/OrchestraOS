package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	orchestraos "github.com/levygit837-cyber/OrchestraOS/internal"
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
	"github.com/levygit837-cyber/OrchestraOS/internal/executor"
	"github.com/levygit837-cyber/OrchestraOS/internal/planner"
	"github.com/levygit837-cyber/OrchestraOS/internal/runtime"
	"github.com/levygit837-cyber/OrchestraOS/internal/store"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	switch cmd {
	case "run":
		if err := runTask(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, "Usage: orchestraos run [--provider fake|gemini|deepseek] \"<title>\" \"<criteria1>\" ...\n")
}

func runTask(args []string) error {
	provider := "fake"
	remaining := args

	if len(remaining) > 1 && remaining[0] == "--provider" {
		provider = remaining[1]
		remaining = remaining[2:]
	}

	if len(remaining) < 3 {
		return fmt.Errorf("need at least: title + 2 acceptance criteria")
	}

	title := remaining[0]
	criteria := remaining[1:]

	ctx := context.Background()
	s := store.NewMemory()
	p := planner.NewHeuristic()
	rt := buildRuntime(provider)
	ex := executor.New(s, rt)
	orch := orchestraos.NewOrchestrator(s, p, ex)

	task := &domain.Task{
		ID:                 uuid.New().String(),
		Title:              title,
		Description:        "Created via CLI",
		Status:             domain.TaskStatusCreated,
		Priority:           domain.TaskPriorityP1,
		RiskLevel:          domain.TaskRiskLevelLow,
		AcceptanceCriteria: criteria,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	if err := s.CreateTask(ctx, task); err != nil {
		return err
	}

	fmt.Printf("Task created: %s\n", task.ID)
	fmt.Printf("  Title: %s\n", task.Title)
	fmt.Printf("  Provider: %s\n", provider)
	fmt.Printf("  Criteria: %s\n", strings.Join(criteria, " | "))
	fmt.Println()

	result, err := orch.RunTask(ctx, task.ID)
	if err != nil {
		return err
	}

	fmt.Printf("Result: %s\n", result.Status)
	fmt.Printf("  Runs: %d\n", len(result.RunIDs))
	if len(result.Errors) > 0 {
		fmt.Printf("  Errors: %s\n", strings.Join(result.Errors, "; "))
	}

	return nil
}

func buildRuntime(provider string) runtime.Runtime {
	switch provider {
	case "gemini":
		return runtime.NewGemini(runtime.Config{
			APIKey: os.Getenv("GEMINI_API_KEY"),
		})
	case "deepseek":
		return runtime.NewDeepSeek(runtime.Config{
			APIKey: os.Getenv("DEEPSEEK_API_KEY"),
		})
	default:
		return runtime.NewFake()
	}
}
