package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/levygit837-cyber/OrchestraOS/internal/agent"
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
	"github.com/levygit837-cyber/OrchestraOS/internal/eventstore"
	"github.com/levygit837-cyber/OrchestraOS/internal/repository"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Manage runs",
	Long:  `Start, list, and manage task runs.`,
}

var runStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a new run",
	RunE: func(cmd *cobra.Command, args []string) error {
		workUnitID, _ := cmd.Flags().GetString("workunit-id")
		runtimeType, _ := cmd.Flags().GetString("runtime")

		// Get work unit to find task ID
		wuRepo := repository.NewWorkUnitRepository(getDB())
		wu, err := wuRepo.GetByID(workUnitID)
		if err != nil {
			return fmt.Errorf("failed to get work unit: %w", err)
		}
		if wu == nil {
			return fmt.Errorf("work unit not found: %s", workUnitID)
		}

		// Create run
		run := &domain.Run{
			ID:         uuid.New().String(),
			TaskID:     wu.TaskGraphID,
			WorkUnitID: workUnitID,
			Status:     domain.RunStatusCreated,
			Attempt:    1,
		}

		runRepo := repository.NewRunRepository(getDB())
		if err := runRepo.Create(run); err != nil {
			return fmt.Errorf("failed to create run: %w", err)
		}

		// Update run status to running
		if err := runRepo.UpdateStatus(run.ID, domain.RunStatusRunning, nil, nil); err != nil {
			return fmt.Errorf("failed to update run status: %w", err)
		}

		// Create event store
		eventStore, err := eventstore.NewStore(getDB())
		if err != nil {
			return fmt.Errorf("failed to create event store: %w", err)
		}

		// Create run.started event
		payload, _ := json.Marshal(map[string]interface{}{
			"run_id":       run.ID,
			"task_id":      wu.TaskGraphID,
			"work_unit_id": workUnitID,
			"runtime":      runtimeType,
		})

		event := &domain.EventEnvelope{
			Type:        "run.started",
			Version:     "v1",
			TaskID:      wu.TaskGraphID,
			RunID:       run.ID,
			WorkUnitID:  workUnitID,
			Priority:    domain.EventPriorityCheckpoint,
			RequiresAck: false,
			Payload:     payload,
		}

		if err := eventStore.Append(event); err != nil {
			return fmt.Errorf("failed to append event: %w", err)
		}

		// Create agent session
		agentID := fmt.Sprintf("agent-%s", uuid.New().String()[:8])
		session := &domain.AgentSession{
			ID:      uuid.New().String(),
			AgentID: agentID,
			RunID:   run.ID,
			Status:  domain.AgentSessionStatusStarting,
		}

		sessionRepo := repository.NewAgentSessionRepository(getDB())
		if err := sessionRepo.Create(session); err != nil {
			return fmt.Errorf("failed to create agent session: %w", err)
		}

		// Update session status
		if err := sessionRepo.UpdateStatus(session.ID, domain.AgentSessionStatusRunning); err != nil {
			return fmt.Errorf("failed to update session status: %w", err)
		}

		// Start runtime if fake
		if runtimeType == "fake" {
			fmt.Printf("Starting fake runtime for run %s...\n", run.ID)

			fakeRuntime := agent.NewFakeRuntime()
			config := agent.RuntimeConfig{
				RunID:      run.ID,
				WorkUnitID: workUnitID,
				TaskID:     wu.TaskGraphID,
				AgentID:    agentID,
				Prompt:     fmt.Sprintf("Execute work unit: %s", wu.Title),
				MaxSteps:   10,
				Timeout:    300,
			}

			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()

			if err := fakeRuntime.Start(ctx, config); err != nil {
				return fmt.Errorf("failed to start fake runtime: %w", err)
			}

			for {
				event, err := fakeRuntime.ReceiveEvent(ctx)
				if err != nil {
					return fmt.Errorf("failed to receive fake runtime event: %w", err)
				}
				if err := eventStore.Append(event); err != nil {
					return fmt.Errorf("failed to append fake runtime event %q: %w", event.Type, err)
				}
				if event.Type == "agent.heartbeat" {
					if err := sessionRepo.UpdateHeartbeat(session.ID); err != nil {
						return fmt.Errorf("failed to update fake runtime heartbeat: %w", err)
					}
				}
				if event.Type == "agent.checkpoint_reached" {
					if err := sessionRepo.UpdateCheckpoint(session.ID); err != nil {
						return fmt.Errorf("failed to update fake runtime checkpoint: %w", err)
					}
				}
				if event.Type == "agent.completed" {
					break
				}
			}

			// Update session and run status
			if err := sessionRepo.UpdateStatus(session.ID, domain.AgentSessionStatusStopped); err != nil {
				return fmt.Errorf("failed to stop fake runtime session: %w", err)
			}
			result := domain.RunResultSucceeded
			if err := runRepo.UpdateStatus(run.ID, domain.RunStatusCompleted, &result, nil); err != nil {
				return fmt.Errorf("failed to complete fake runtime run: %w", err)
			}

			// Create run.completed event
			payload, _ := json.Marshal(map[string]interface{}{
				"run_id":  run.ID,
				"result":  "succeeded",
				"runtime": runtimeType,
			})

			if err := eventStore.Append(&domain.EventEnvelope{
				Type:        "run.completed",
				Version:     "v1",
				TaskID:      wu.TaskGraphID,
				RunID:       run.ID,
				WorkUnitID:  workUnitID,
				AgentID:     agentID,
				Priority:    domain.EventPriorityCheckpoint,
				RequiresAck: false,
				Payload:     payload,
			}); err != nil {
				return fmt.Errorf("failed to append run.completed event: %w", err)
			}
		}

		fmt.Printf("Run started: %s (runtime: %s, agent: %s)\n", run.ID, runtimeType, agentID)
		return nil
	},
}

var runListCmd = &cobra.Command{
	Use:   "list",
	Short: "List runs",
	RunE: func(cmd *cobra.Command, args []string) error {
		taskID, _ := cmd.Flags().GetString("task-id")

		repo := repository.NewRunRepository(getDB())

		var runs []domain.Run
		var err error

		if taskID != "" {
			runs, err = repo.ListByTask(taskID)
		} else {
			runs, err = repo.List()
		}

		if err != nil {
			return fmt.Errorf("failed to list runs: %w", err)
		}

		if len(runs) == 0 {
			fmt.Println("No runs found")
			return nil
		}

		fmt.Printf("%-36s %-36s %-10s %-10s %-20s\n", "ID", "TASK ID", "STATUS", "ATTEMPT", "STARTED AT")
		for _, run := range runs {
			started := "-"
			if !run.StartedAt.IsZero() {
				started = run.StartedAt.Format("2006-01-02 15:04")
			}
			fmt.Printf("%-36s %-36s %-10s %-10d %-20s\n",
				run.ID,
				run.TaskID,
				run.Status,
				run.Attempt,
				started,
			)
		}
		return nil
	},
}

var runGetCmd = &cobra.Command{
	Use:   "get [id]",
	Short: "Get run details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		repo := repository.NewRunRepository(getDB())
		run, err := repo.GetByID(args[0])
		if err != nil {
			return fmt.Errorf("failed to get run: %w", err)
		}
		if run == nil {
			return fmt.Errorf("run not found: %s", args[0])
		}

		fmt.Printf("ID: %s\n", run.ID)
		fmt.Printf("Task ID: %s\n", run.TaskID)
		fmt.Printf("Work Unit ID: %s\n", run.WorkUnitID)
		fmt.Printf("Status: %s\n", run.Status)
		fmt.Printf("Attempt: %d\n", run.Attempt)
		if run.Result != nil {
			fmt.Printf("Result: %s\n", *run.Result)
		}
		if run.FailureReason != nil {
			fmt.Printf("Failure Reason: %s\n", *run.FailureReason)
		}
		if !run.StartedAt.IsZero() {
			fmt.Printf("Started At: %s\n", run.StartedAt.Format("2006-01-02 15:04:05"))
		}
		if run.FinishedAt != nil {
			fmt.Printf("Finished At: %s\n", run.FinishedAt.Format("2006-01-02 15:04:05"))
		}

		return nil
	},
}

func init() {
	runStartCmd.Flags().String("workunit-id", "", "Work unit ID to run (required)")
	runStartCmd.Flags().String("runtime", "fake", "Runtime type (fake, codex_cli)")
	runStartCmd.MarkFlagRequired("workunit-id")

	runListCmd.Flags().String("task-id", "", "Filter by task ID")

	runCmd.AddCommand(runStartCmd)
	runCmd.AddCommand(runListCmd)
	runCmd.AddCommand(runGetCmd)
}
