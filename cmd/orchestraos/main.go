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
		fmt.Fprintf(os.Stderr, "Usage: orchestraos run \"<task title>\" \"<criteria1>\" \"<criteria2>\" ...\n")
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
		fmt.Fprintf(os.Stderr, "Unknown command: %s\nUsage: orchestraos run \"<title>\" \"<criteria1>\" ...\n", cmd)
		os.Exit(1)
	}
}

func runTask(args []string) error {
	if len(args) < 3 {
		return fmt.Errorf("need at least: title + 2 acceptance criteria")
	}

	title := args[0]
	criteria := args[1:]

	ctx := context.Background()

	s := store.NewMemory()
	p := planner.NewHeuristic()
	rt := runtime.NewFake()
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
