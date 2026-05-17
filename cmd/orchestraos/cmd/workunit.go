package cmd

import (
	"fmt"

	"github.com/levygit837-cyber/OrchestraOS/internal/bootstrap"
	workunitmod "github.com/levygit837-cyber/OrchestraOS/internal/modules/workunit"
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
		ownedPaths, _ := cmd.Flags().GetStringArray("owned-path")
		readPaths, _ := cmd.Flags().GetStringArray("read-path")
		acceptanceCriteria, _ := cmd.Flags().GetStringArray("acceptance")
		validationPlan, _ := cmd.Flags().GetStringArray("validation")
		dependsOn, _ := cmd.Flags().GetStringArray("depends-on")

		service := bootstrap.WorkUnitService(getDB())
		result, err := service.Create(cmd.Context(), workunitmod.CreateWorkUnitInput{
			TaskID:               taskID,
			Title:                title,
			Objective:            objective,
			AssignedAgentProfile: agentProfile,
			OwnedPaths:           ownedPaths,
			ReadPaths:            readPaths,
			AcceptanceCriteria:   acceptanceCriteria,
			ValidationPlan:       validationPlan,
			DependsOn:            dependsOn,
		})
		if err != nil {
			return fmt.Errorf("failed to create work unit: %w", err)
		}

		fmt.Printf("Work unit created: %s (task: %s)\n", result.Value.ID, taskID)
		return nil
	},
}

var workUnitListCmd = &cobra.Command{
	Use:   "list",
	Short: "List work units for a task",
	RunE: func(cmd *cobra.Command, args []string) error {
		taskID, _ := cmd.Flags().GetString("task-id")

		repo := workunitmod.NewRepository(getDB())
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
	workUnitCreateCmd.Flags().String("objective", "", "Work unit objective (required)")
	workUnitCreateCmd.Flags().String("agent-profile", "default", "Agent profile to use")
	workUnitCreateCmd.Flags().StringArray("owned-path", nil, "Path owned by this work unit (repeatable)")
	workUnitCreateCmd.Flags().StringArray("read-path", nil, "Path read by this work unit (repeatable)")
	workUnitCreateCmd.Flags().StringArray("acceptance", nil, "Acceptance criterion (repeatable)")
	workUnitCreateCmd.Flags().StringArray("validation", nil, "Validation step (repeatable)")
	workUnitCreateCmd.Flags().StringArray("depends-on", nil, "Dependency work unit UUID (repeatable)")
	_ = workUnitCreateCmd.MarkFlagRequired("task-id")
	_ = workUnitCreateCmd.MarkFlagRequired("title")
	_ = workUnitCreateCmd.MarkFlagRequired("objective")

	workUnitListCmd.Flags().String("task-id", "", "Task ID to list work units for (required)")
	_ = workUnitListCmd.MarkFlagRequired("task-id")

	workUnitCmd.AddCommand(workUnitCreateCmd)
	workUnitCmd.AddCommand(workUnitListCmd)
}
