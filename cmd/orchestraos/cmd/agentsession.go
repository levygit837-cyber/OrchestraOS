package cmd

import (
	"context"
	"fmt"

	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
	"github.com/levygit837-cyber/OrchestraOS/internal/orchestration"
	"github.com/levygit837-cyber/OrchestraOS/internal/repository"
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

		// Validate run exists
		runRepo := repository.NewRunRepository(getDB())
		run, err := runRepo.GetByID(runID)
		if err != nil {
			return fmt.Errorf("failed to get run: %w", err)
		}
		if run == nil {
			return fmt.Errorf("run not found: %s", runID)
		}

		// Create agent session
		session := &domain.AgentSession{
			RunID:   runID,
			AgentID: agentID,
			Status:  domain.AgentSessionStatusStarting,
		}

		repo := repository.NewAgentSessionRepository(getDB())
		if err := repo.Create(session); err != nil {
			return fmt.Errorf("failed to create agent session: %w", err)
		}

		fmt.Printf("Agent session created: %s (run: %s, agent: %s)\n", session.ID, runID, agentID)
		return nil
	},
}

var agentSessionGetCmd = &cobra.Command{
	Use:   "get [id]",
	Short: "Get agent session details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		repo := repository.NewAgentSessionRepository(getDB())
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

		repo := repository.NewAgentSessionRepository(getDB())
		session, err := repo.GetByID(args[0])
		if err != nil {
			return fmt.Errorf("failed to get agent session: %w", err)
		}
		if session == nil {
			return fmt.Errorf("agent session not found: %s", args[0])
		}

		commander := orchestration.NewCommander(getDB())
		if err := commander.TransitionAgentSession(context.Background(), session.ID, domain.AgentSessionStatus(status), orchestration.TransitionOptions{}); err != nil {
			return fmt.Errorf("failed to update status: %w", err)
		}

		fmt.Printf("Agent session %s status updated to: %s\n", session.ID, status)
		return nil
	},
}

var agentSessionHeartbeatCmd = &cobra.Command{
	Use:   "heartbeat [id]",
	Short: "Update agent session heartbeat",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		repo := repository.NewAgentSessionRepository(getDB())

		if err := repo.UpdateHeartbeat(args[0]); err != nil {
			return fmt.Errorf("failed to update heartbeat: %w", err)
		}

		fmt.Printf("Heartbeat updated for agent session: %s\n", args[0])
		return nil
	},
}

var agentSessionCheckpointCmd = &cobra.Command{
	Use:   "checkpoint [id]",
	Short: "Update agent session checkpoint",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		repo := repository.NewAgentSessionRepository(getDB())

		if err := repo.UpdateCheckpoint(args[0]); err != nil {
			return fmt.Errorf("failed to update checkpoint: %w", err)
		}

		fmt.Printf("Checkpoint updated for agent session: %s\n", args[0])
		return nil
	},
}

func init() {
	agentSessionCreateCmd.Flags().String("run-id", "", "Run ID (required)")
	agentSessionCreateCmd.Flags().String("agent-id", "", "Agent ID (required)")
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
