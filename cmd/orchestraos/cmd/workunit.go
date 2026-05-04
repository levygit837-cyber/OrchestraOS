package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
	"github.com/levygit837-cyber/OrchestraOS/internal/eventstore"
	"github.com/levygit837-cyber/OrchestraOS/internal/repository"
	"github.com/spf13/cobra"
)

var workUnitCmd = &cobra.Command{
	Use:   "workunit",
	Short: "Manage work units",
	Long:  `Create, list, and manage work units.`,
}

var workUnitCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new work unit",
	RunE: func(cmd *cobra.Command, args []string) error {
		taskID, _ := cmd.Flags().GetString("task-id")
		title, _ := cmd.Flags().GetString("title")
		objective, _ := cmd.Flags().GetString("objective")
		agentProfile, _ := cmd.Flags().GetString("agent-profile")

		wu := &domain.WorkUnit{
			TaskGraphID:          taskID,
			Title:                title,
			Objective:            objective,
			AssignedAgentProfile: agentProfile,
			Status:               domain.WorkUnitStatusCreated,
		}

		repo := repository.NewWorkUnitRepository(getDB())
		if err := repo.Create(wu); err != nil {
			return fmt.Errorf("failed to create work unit: %w", err)
		}

		// Create event
		eventStore, err := eventstore.NewStore(getDB())
		if err != nil {
			return fmt.Errorf("failed to create event store: %w", err)
		}

		payload, _ := json.Marshal(map[string]interface{}{
			"work_unit_id": wu.ID,
			"task_id":      taskID,
			"title":        title,
			"status":       wu.Status,
		})

		event := &domain.EventEnvelope{
			Type:        "work_unit.created",
			Version:     "v1",
			TaskID:      taskID,
			WorkUnitID:  wu.ID,
			Priority:    domain.EventPriorityNotification,
			RequiresAck: false,
			Payload:     payload,
		}

		if err := eventStore.Append(event); err != nil {
			return fmt.Errorf("failed to append event: %w", err)
		}

		fmt.Printf("Work unit created: %s (task: %s)\n", wu.ID, taskID)
		return nil
	},
}

var workUnitListCmd = &cobra.Command{
	Use:   "list",
	Short: "List work units for a task",
	RunE: func(cmd *cobra.Command, args []string) error {
		taskID, _ := cmd.Flags().GetString("task-id")

		repo := repository.NewWorkUnitRepository(getDB())
		workUnits, err := repo.ListByTask(taskID)
		if err != nil {
			return fmt.Errorf("failed to list work units: %w", err)
		}

		if len(workUnits) == 0 {
			fmt.Println("No work units found")
			return nil
		}

		fmt.Printf("%-36s %-20s %-15s %-15s\n", "ID", "TITLE", "STATUS", "AGENT PROFILE")
		for _, wu := range workUnits {
			fmt.Printf("%-36s %-20s %-15s %-15s\n",
				wu.ID,
				truncate(wu.Title, 20),
				wu.Status,
				wu.AssignedAgentProfile,
			)
		}
		return nil
	},
}

func init() {
	workUnitCreateCmd.Flags().String("task-id", "", "Parent task ID (required)")
	workUnitCreateCmd.Flags().String("title", "", "Work unit title (required)")
	workUnitCreateCmd.Flags().String("objective", "", "Work unit objective")
	workUnitCreateCmd.Flags().String("agent-profile", "default", "Agent profile to use")
	workUnitCreateCmd.MarkFlagRequired("task-id")
	workUnitCreateCmd.MarkFlagRequired("title")

	workUnitListCmd.Flags().String("task-id", "", "Task ID to list work units for (required)")
	workUnitListCmd.MarkFlagRequired("task-id")

	workUnitCmd.AddCommand(workUnitCreateCmd)
	workUnitCmd.AddCommand(workUnitListCmd)
}
