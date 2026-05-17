package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/levygit837-cyber/OrchestraOS/internal/bootstrap"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/apperrors"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/coordination"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/transition"
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
	"github.com/levygit837-cyber/OrchestraOS/internal/modules/agent"
	agentsessionmod "github.com/levygit837-cyber/OrchestraOS/internal/modules/agentsession"
	promptmod "github.com/levygit837-cyber/OrchestraOS/internal/modules/prompt"
	runmod "github.com/levygit837-cyber/OrchestraOS/internal/modules/run"

	workunitmod "github.com/levygit837-cyber/OrchestraOS/internal/modules/workunit"
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
		wuRepo := workunitmod.NewRepository(getDB())
		wu, err := wuRepo.GetByID(workUnitID)
		if err != nil {
			return fmt.Errorf("failed to get work unit: %w", err)
		}
		if wu == nil {
			return fmt.Errorf("work unit not found: %s", workUnitID)
		}

		runService := bootstrap.RunService(getDB())
		runResult, err := runService.Create(cmd.Context(), runmod.CreateRunInput{
			TaskID:     wu.TaskID,
			WorkUnitID: workUnitID,
			Attempt:    1,
		})
		if err != nil {
			return fmt.Errorf("failed to create run: %w", err)
		}
		run := runResult.Value
		if _, err := runService.Start(cmd.Context(), run.ID, transition.TransitionInput{
			Runtime: runtimeType,
		}); err != nil {
			return fmt.Errorf("failed to start run: %w", err)
		}

		// Resolve agent profile from work unit (default: code_worker)
		agentProfile := wu.AssignedAgentProfile
		if agentProfile == "" {
			agentProfile = "code_worker"
		}

		agentService := bootstrap.AgentService(getDB())
		agentEntity, err := agentService.FindOrCreate(cmd.Context(), agentProfile, agent.RuntimeType(runtimeType))
		if err != nil {
			_ = failStartedRun(context.Background(), runService, nil, run.ID, "", runtimeType, "", err)
			return fmt.Errorf("failed to find or create agent: %w", err)
		}
		agentID := agentEntity.ID

		sessionService := bootstrap.AgentSessionService(getDB())
		sessionResult, err := sessionService.Create(cmd.Context(), agentsessionmod.CreateAgentSessionInput{
			AgentID:    agentID,
			RunID:      run.ID,
			TaskID:     run.TaskID,
			WorkUnitID: run.WorkUnitID,
		})
		if err != nil {
			_ = failStartedRun(context.Background(), runService, sessionService, run.ID, "", runtimeType, agentID, err)
			return fmt.Errorf("failed to create agent session: %w", err)
		}
		session := sessionResult.Value
		connectionID := fmt.Sprintf("conn-%s", uuid.New().String())
		if _, err := sessionService.Connect(cmd.Context(), session.ID, connectionID, "", transition.TransitionInput{
			Runtime: runtimeType,
		}); err != nil {
			_ = failStartedRun(context.Background(), runService, sessionService, run.ID, session.ID, runtimeType, agentID, err)
			return fmt.Errorf("failed to connect agent session: %w", err)
		}

		preparedPrompt, err := coordination.NewPromptOrchestrator(getDB(), bootstrap.PromptService(getDB())).PrepareRunPrompt(cmd.Context(), promptmod.PrepareRunPromptInput{
			RunID:          run.ID,
			AgentSessionID: session.ID,
		})
		if err != nil {
			_ = failStartedRun(context.Background(), runService, sessionService, run.ID, session.ID, runtimeType, agentID, err)
			return fmt.Errorf("failed to prepare prompt: %w", err)
		}

		// Start runtime if fake or gemini
		if runtimeType == "fake" || runtimeType == "gemini" {
			fmt.Printf("Starting %s runtime for run %s...\n", runtimeType, run.ID)

			var runtime agent.Runtime
			if runtimeType == "fake" {
				runtime = agent.NewFakeRuntime()
			} else {
				runtime = agent.NewGeminiRuntime()
			}

			config := agent.RuntimeConfig{
				RunID:             run.ID,
				WorkUnitID:        workUnitID,
				TaskID:            wu.TaskID,
				AgentID:           agentID,
				Prompt:            preparedPrompt.CombinedPrompt,
				SystemPrompt:      preparedPrompt.SystemPrompt,
				TaskPrompt:        preparedPrompt.TaskPrompt,
				PromptSnapshotID:  preparedPrompt.PromptSnapshot.ID,
				ToolsetSnapshotID: preparedPrompt.ToolsetSnapshot.ID,
				PromptHash:        preparedPrompt.PromptHash,
				Toolset:           preparedPrompt.Toolset,
				MaxSteps:          10,
				Timeout:           300,
			}

			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()

			if err := runtime.Start(ctx, config); err != nil {
				return failStartedRun(ctx, runService, sessionService, run.ID, session.ID, runtimeType, agentID, fmt.Errorf("failed to start %s runtime: %w", runtimeType, err))
			}

			relay := bootstrap.RuntimeEventRelay(getDB())
			relayConfig := coordination.RelayConfig{
				SessionID:   session.ID,
				RunID:       run.ID,
				RuntimeType: runtimeType,
				AgentID:     agentID,
				OnEvent: func(event *domain.EventEnvelope) {
					fmt.Printf("[%s] %s\n", runtimeType, event.Type)
				},
			}

			finalStatus, err := relay.Run(ctx, runtime, relayConfig)
			if err != nil {
				return failStartedRun(ctx, runService, sessionService, run.ID, session.ID, runtimeType, agentID, fmt.Errorf("runtime relay failed: %w", err))
			}
			if finalStatus != domain.RunStatusCompleted {
				return failStartedRun(ctx, runService, sessionService, run.ID, session.ID, runtimeType, agentID, fmt.Errorf("runtime finished with unexpected status: %s", finalStatus))
			}
		}

		fmt.Printf("Run started: %s (runtime: %s, agent: %s, prompt_snapshot: %s, toolset_snapshot: %s)\n",
			run.ID,
			runtimeType,
			agentID,
			preparedPrompt.PromptSnapshot.ID,
			preparedPrompt.ToolsetSnapshot.ID,
		)
		return nil
	},
}

func failStartedRun(ctx context.Context, runService *runmod.RunService, sessionService *agentsessionmod.AgentSessionService, runID, sessionID, runtimeType, agentID string, cause error) error {
	if cause == nil {
		return nil
	}

	input := transition.TransitionInput{
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

		repo := runmod.NewRepository(getDB())

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
		repo := runmod.NewRepository(getDB())
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
	runStartCmd.Flags().String("runtime", "fake", "Runtime type (fake, codex_cli, gemini)")
	runStartCmd.MarkFlagRequired("workunit-id")

	runListCmd.Flags().String("task-id", "", "Filter by task ID")

	runCmd.AddCommand(runStartCmd)
	runCmd.AddCommand(runListCmd)
	runCmd.AddCommand(runGetCmd)
}


