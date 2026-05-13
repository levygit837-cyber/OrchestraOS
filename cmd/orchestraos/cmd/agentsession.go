package cmd

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/levygit837-cyber/OrchestraOS/internal/bootstrap"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/transition"
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
	agentsessionmod "github.com/levygit837-cyber/OrchestraOS/internal/modules/agentsession"
	runmod "github.com/levygit837-cyber/OrchestraOS/internal/modules/run"
	"github.com/spf13/cobra"
)

var agentSessionCmd = &cobra.Command{
	Use:   "agentsession",
	Short: "Manage agent sessions",
	Long:  `Create, list, and manage agent sessions.`,
}

var agentSessionCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new agent session",
	RunE: func(cmd *cobra.Command, args []string) error {
		runID, _ := cmd.Flags().GetString("run-id")
		agentID, _ := cmd.Flags().GetString("agent-id")
		connectionID, _ := cmd.Flags().GetString("connection-id")
		sandboxID, _ := cmd.Flags().GetString("sandbox-id")

		db := getDB()
		run, err := runmod.NewRepository(db).GetByID(runID)
		if err != nil {
			return fmt.Errorf("failed to get run: %w", err)
		}
		if run == nil {
			return fmt.Errorf("run not found: %s", runID)
		}

		service := bootstrap.AgentSessionService(db)
		result, err := service.Create(cmd.Context(), agentsessionmod.CreateAgentSessionInput{
			RunID:        runID,
			TaskID:       run.TaskID,
			WorkUnitID:   run.WorkUnitID,
			AgentID:      agentID,
			ConnectionID: connectionID,
			SandboxID:    sandboxID,
		})
		if err != nil {
			return fmt.Errorf("failed to create agent session: %w", err)
		}

		fmt.Printf("Agent session created: %s (run: %s, agent: %s)\n", result.Value.ID, runID, agentID)
		return nil
	},
}

var agentSessionGetCmd = &cobra.Command{
	Use:   "get [id]",
	Short: "Get agent session details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		repo := agentsessionmod.NewRepository(getDB())
		session, err := repo.GetByID(args[0])
		if err != nil {
			return fmt.Errorf("failed to get agent session: %w", err)
		}
		if session == nil {
			return fmt.Errorf("agent session not found: %s", args[0])
		}

		fmt.Printf("ID: %s\n", session.ID)
		fmt.Printf("Agent ID: %s\n", session.AgentID)
		fmt.Printf("Run ID: %s\n", session.RunID)
		fmt.Printf("Sandbox ID: %s\n", session.SandboxID)
		fmt.Printf("Connection ID: %s\n", session.ConnectionID)
		fmt.Printf("Status: %s\n", session.Status)
		if session.LastSeenEventID != "" {
			fmt.Printf("Last Seen Event: %s\n", session.LastSeenEventID)
		}
		if session.LastHeartbeatAt != nil {
			fmt.Printf("Last Heartbeat: %s\n", session.LastHeartbeatAt.Format("2006-01-02 15:04:05"))
		}
		if session.LastCheckpointAt != nil {
			fmt.Printf("Last Checkpoint: %s\n", session.LastCheckpointAt.Format("2006-01-02 15:04:05"))
		}

		return nil
	},
}

var agentSessionStatusCmd = &cobra.Command{
	Use:   "status [id]",
	Short: "Update agent session status",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		status, _ := cmd.Flags().GetString("status")

		service := bootstrap.AgentSessionService(getDB())
		if err := updateAgentSessionStatus(context.Background(), service, args[0], domain.AgentSessionStatus(status)); err != nil {
			return fmt.Errorf("failed to update status: %w", err)
		}

		fmt.Printf("Agent session %s status updated to: %s\n", args[0], status)
		return nil
	},
}

var agentSessionHeartbeatCmd = &cobra.Command{
	Use:   "heartbeat [id]",
	Short: "Update agent session heartbeat",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		service := bootstrap.AgentSessionService(getDB())
		if _, err := service.Heartbeat(cmd.Context(), args[0], domain.HeartbeatInput{
			Payload: map[string]interface{}{"source": "cli"},
		}); err != nil {
			return fmt.Errorf("failed to update heartbeat: %w", err)
		}

		fmt.Printf("Heartbeat updated for agent session: %s\n", args[0])
		return nil
	},
}

var agentSessionCheckpointCmd = &cobra.Command{
	Use:   "checkpoint [id]",
	Short: "Record a manual debug/test checkpoint",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		service := bootstrap.AgentSessionService(getDB())
		if _, err := service.Checkpoint(cmd.Context(), args[0], domain.CheckpointInput{
			CheckpointID:   "cli-debug-" + uuid.New().String(),
			CurrentGoal:    "debug/manual checkpoint",
			MinimalSummary: "manual debug checkpoint recorded from CLI",
			Source:         "cli_debug",
			Ledger: map[string]interface{}{
				"source":        "cli",
				"debug":         true,
				"pending_todos": []interface{}{},
			},
		}); err != nil {
			return fmt.Errorf("failed to update checkpoint: %w", err)
		}

		fmt.Printf("Checkpoint updated for agent session: %s\n", args[0])
		return nil
	},
}

func init() {
	agentSessionCreateCmd.Flags().String("run-id", "", "Run ID (required)")
	agentSessionCreateCmd.Flags().String("agent-id", "", "Agent ID (required)")
	agentSessionCreateCmd.Flags().String("connection-id", "", "Connection ID")
	agentSessionCreateCmd.Flags().String("sandbox-id", "", "Sandbox ID")
	agentSessionCreateCmd.MarkFlagRequired("run-id")
	agentSessionCreateCmd.MarkFlagRequired("agent-id")

	agentSessionStatusCmd.Flags().String("status", "", "New status (required)")
	agentSessionStatusCmd.MarkFlagRequired("status")

	agentSessionCmd.AddCommand(agentSessionCreateCmd)
	agentSessionCmd.AddCommand(agentSessionGetCmd)
	agentSessionCmd.AddCommand(agentSessionStatusCmd)
	agentSessionCmd.AddCommand(agentSessionHeartbeatCmd)
	agentSessionCmd.AddCommand(agentSessionCheckpointCmd)
}

func updateAgentSessionStatus(ctx context.Context, service *agentsessionmod.AgentSessionService, sessionID string, status domain.AgentSessionStatus) error {
	switch status {
	case domain.AgentSessionStatusRunning:
		_, err := service.Resume(ctx, sessionID, transition.TransitionInput{})
		return err
	case domain.AgentSessionStatusDisconnected:
		_, err := service.Disconnect(ctx, sessionID, transition.TransitionInput{Justification: "manual status update"})
		return err
	case domain.AgentSessionStatusStopped:
		_, err := service.Stop(ctx, sessionID, transition.TransitionInput{Justification: "manual status update"})
		return err
	case domain.AgentSessionStatusFailed:
		_, err := service.Fail(ctx, sessionID, transition.TransitionInput{FailureReason: "manual status update"})
		return err
	case domain.AgentSessionStatusPaused, domain.AgentSessionStatusWaitingApproval, domain.AgentSessionStatusStopping:
		return fmt.Errorf("manual status %q is not exposed as a service command yet", status)
	default:
		return fmt.Errorf("unknown agent session status %q", status)
	}
}
