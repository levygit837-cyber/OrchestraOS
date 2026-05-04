package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/levygit837-cyber/OrchestraOS/internal/agent"
	"github.com/levygit837-cyber/OrchestraOS/internal/apperrors"
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
	"github.com/levygit837-cyber/OrchestraOS/internal/repository"
	"github.com/levygit837-cyber/OrchestraOS/internal/services"
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

		runService := services.NewRunService(getDB())
		runResult, err := runService.Create(cmd.Context(), services.CreateRunInput{
			TaskID:     wu.TaskGraphID,
			WorkUnitID: workUnitID,
			Attempt:    1,
		})
		if err != nil {
			return fmt.Errorf("failed to create run: %w", err)
		}
		run := runResult.Value
		if _, err := runService.Start(cmd.Context(), run.ID, services.TransitionInput{
			Runtime: runtimeType,
		}); err != nil {
			return fmt.Errorf("failed to start run: %w", err)
		}

		agentID := fmt.Sprintf("agent-%s", uuid.New().String()[:8])
		sessionService := services.NewAgentSessionService(getDB())
		sessionResult, err := sessionService.Create(cmd.Context(), services.CreateAgentSessionInput{
			AgentID: agentID,
			RunID:   run.ID,
		})
		if err != nil {
			_ = failStartedRun(context.Background(), runService, sessionService, run.ID, "", runtimeType, agentID, err)
			return fmt.Errorf("failed to create agent session: %w", err)
		}
		session := sessionResult.Value
		connectionID := fmt.Sprintf("conn-%s", uuid.New().String())
		if _, err := sessionService.Connect(cmd.Context(), session.ID, connectionID, "", services.TransitionInput{
			Runtime: runtimeType,
		}); err != nil {
			_ = failStartedRun(context.Background(), runService, sessionService, run.ID, session.ID, runtimeType, agentID, err)
			return fmt.Errorf("failed to connect agent session: %w", err)
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
				return failStartedRun(ctx, runService, sessionService, run.ID, session.ID, runtimeType, agentID, fmt.Errorf("failed to start fake runtime: %w", err))
			}

			for {
				event, err := fakeRuntime.ReceiveEvent(ctx)
				if err != nil {
					return failStartedRun(ctx, runService, sessionService, run.ID, session.ID, runtimeType, agentID, fmt.Errorf("failed to receive fake runtime event: %w", err))
				}
				switch event.Type {
				case "agent.heartbeat":
					payload, err := decodePayloadMap(event)
					if err != nil {
						return failStartedRun(ctx, runService, sessionService, run.ID, session.ID, runtimeType, agentID, err)
					}
					if _, err := sessionService.Heartbeat(ctx, session.ID, services.HeartbeatInput{
						EventID: event.ID,
						Payload: payload,
					}); err != nil {
						return failStartedRun(ctx, runService, sessionService, run.ID, session.ID, runtimeType, agentID, fmt.Errorf("failed to persist fake runtime heartbeat: %w", err))
					}
				case "agent.checkpoint_reached":
					if _, err := sessionService.CheckpointFromEvent(ctx, session.ID, event); err != nil {
						return failStartedRun(ctx, runService, sessionService, run.ID, session.ID, runtimeType, agentID, fmt.Errorf("failed to persist fake runtime checkpoint: %w", err))
					}
				default:
					eventService := services.NewEventService(getDB())
					if _, err := eventService.Append(ctx, event); err != nil {
						return failStartedRun(ctx, runService, sessionService, run.ID, session.ID, runtimeType, agentID, fmt.Errorf("failed to append fake runtime event %q: %w", event.Type, err))
					}
				}
				if err := maybeAutoCheckpointRuntimeEvent(ctx, sessionService, session.ID, runtimeType, event); err != nil {
					return failStartedRun(ctx, runService, sessionService, run.ID, session.ID, runtimeType, agentID, fmt.Errorf("failed to persist automatic checkpoint for %q: %w", event.Type, err))
				}
				if event.Type == "agent.completed" {
					break
				}
			}

			if _, err := sessionService.Stop(context.Background(), session.ID, services.TransitionInput{
				Runtime: runtimeType,
			}); err != nil {
				return failStartedRun(ctx, runService, sessionService, run.ID, session.ID, runtimeType, agentID, fmt.Errorf("failed to stop fake runtime session: %w", err))
			}

			if _, err := runService.Validate(context.Background(), run.ID, services.TransitionInput{
				Runtime: runtimeType,
			}); err != nil {
				return failStartedRun(ctx, runService, sessionService, run.ID, session.ID, runtimeType, agentID, fmt.Errorf("failed to validate fake runtime run: %w", err))
			}
			if _, err := runService.Complete(context.Background(), run.ID, services.TransitionInput{
				Runtime:       runtimeType,
				AgentID:       agentID,
				EvidenceRefs:  []string{"fake-runtime:agent.completed"},
				Justification: "fake runtime completed with agent.completed event",
			}); err != nil {
				return failStartedRun(ctx, runService, sessionService, run.ID, session.ID, runtimeType, agentID, fmt.Errorf("failed to complete fake runtime run: %w", err))
			}
		}

		fmt.Printf("Run started: %s (runtime: %s, agent: %s)\n", run.ID, runtimeType, agentID)
		return nil
	},
}

func failStartedRun(ctx context.Context, runService *services.RunService, sessionService *services.AgentSessionService, runID, sessionID, runtimeType, agentID string, cause error) error {
	if cause == nil {
		return nil
	}

	input := services.TransitionInput{
		Runtime:       runtimeType,
		AgentID:       agentID,
		FailureReason: cause.Error(),
		Justification: "runtime failed after operational state was started",
	}

	var compensationErrs []error
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		input.FailureReason = "runtime timed out"
		recoverableState, _ := json.Marshal(map[string]interface{}{
			"runtime":      runtimeType,
			"reason":       cause.Error(),
			"timed_out_at": time.Now().UTC().Format(time.RFC3339Nano),
		})
		if sessionID != "" {
			if _, err := sessionService.Timeout(context.Background(), sessionID, recoverableState, input); err != nil && !ignorableCompensationTransition(err) {
				compensationErrs = append(compensationErrs, fmt.Errorf("session timeout: %w", err))
			}
		}
		if runID != "" {
			if _, err := runService.Timeout(context.Background(), runID, input); err != nil && !ignorableCompensationTransition(err) {
				compensationErrs = append(compensationErrs, fmt.Errorf("run timeout: %w", err))
			}
		}
	} else {
		if sessionID != "" {
			if _, err := sessionService.Fail(context.Background(), sessionID, input); err != nil && !ignorableCompensationTransition(err) {
				compensationErrs = append(compensationErrs, fmt.Errorf("session fail: %w", err))
			}
		}
		if runID != "" {
			if _, err := runService.Fail(context.Background(), runID, input); err != nil && !ignorableCompensationTransition(err) {
				compensationErrs = append(compensationErrs, fmt.Errorf("run fail: %w", err))
			}
		}
	}

	if len(compensationErrs) > 0 {
		return fmt.Errorf("%w; compensation failed: %w", cause, errors.Join(compensationErrs...))
	}
	return cause
}

func ignorableCompensationTransition(err error) bool {
	var appErr *apperrors.Error
	return errors.As(err, &appErr) && appErr.Code == apperrors.CodeInvalidTransition
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

func decodePayloadMap(event *domain.EventEnvelope) (map[string]interface{}, error) {
	var payload map[string]interface{}
	if len(event.Payload) == 0 {
		return map[string]interface{}{}, nil
	}
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return nil, fmt.Errorf("failed to decode event payload %s: %w", event.ID, err)
	}
	return payload, nil
}

func maybeAutoCheckpointRuntimeEvent(ctx context.Context, sessionService *services.AgentSessionService, sessionID, runtimeType string, event *domain.EventEnvelope) error {
	trigger, ok := checkpointTriggerForRuntimeEvent(event.Type)
	if !ok {
		return nil
	}
	payload, err := decodePayloadMap(event)
	if err != nil {
		return err
	}
	currentGoal := event.Type
	if value, ok := payload["current_goal"].(string); ok && value != "" {
		currentGoal = value
	}
	summary := fmt.Sprintf("automatic checkpoint for runtime event %s", event.Type)
	if value, ok := payload["summary"].(string); ok && value != "" {
		summary = value
	} else if value, ok := payload["reason"].(string); ok && value != "" {
		summary = value
	}
	ledger := map[string]interface{}{
		"runtime_event_type": event.Type,
		"runtime_event_id":   event.ID,
		"runtime_payload":    payload,
		"pending_todos":      []interface{}{},
	}
	if value, ok := payload["ledger"].(map[string]interface{}); ok {
		ledger = value
	}
	_, _, err = sessionService.AutomaticCheckpoint(ctx, sessionID, services.AutoCheckpointInput{
		EventID:        uuid.NewSHA1(uuid.NameSpaceURL, []byte("orchestraos:auto_checkpoint:"+sessionID+":"+event.ID+":"+string(trigger))).String(),
		Trigger:        trigger,
		CurrentGoal:    currentGoal,
		MinimalSummary: summary,
		Ledger:         ledger,
		EvidenceRefs:   []string{"event:" + event.ID},
		OccurredAt:     event.CreatedAt,
		SourceEventID:  event.ID,
		Runtime:        runtimeType,
		Extra: map[string]interface{}{
			"runtime_event_type": event.Type,
		},
	})
	return err
}

func checkpointTriggerForRuntimeEvent(eventType string) (services.CheckpointTrigger, bool) {
	switch eventType {
	case "agent.tool_requested":
		return services.CheckpointTriggerToolRequest, true
	case "agent.tool_executed", "tool.completed":
		return services.CheckpointTriggerToolExecuted, true
	case "agent.completed":
		return services.CheckpointTriggerBeforeCompletion, true
	default:
		return "", false
	}
}
