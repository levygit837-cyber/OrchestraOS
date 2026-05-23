package cmd

import (
	"fmt"
	"os"

	"github.com/levygit837-cyber/OrchestraOS/internal/bootstrap"
	"github.com/spf13/cobra"
)

var taskCmd = &cobra.Command{
	Use:   "task",
	Short: "Manage tasks",
	Long:  `Create, list, and manage tasks.`,
}

var taskCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new task",
	RunE: func(cmd *cobra.Command, args []string) error {
		title, _ := cmd.Flags().GetString("title")
		description, _ := cmd.Flags().GetString("description")
		priority, _ := cmd.Flags().GetString("priority")
		riskLevel, _ := cmd.Flags().GetString("risk-level")
		acceptanceCriteria, _ := cmd.Flags().GetStringArray("acceptance")

		service := bootstrap.TaskService(getDB())
		result, err := service.Create(cmd.Context(), bootstrap.CreateTaskInput{
			Title:              title,
			Description:        description,
			Priority:           bootstrap.Priority(priority),
			RiskLevel:          bootstrap.RiskLevel(riskLevel),
			AcceptanceCriteria: acceptanceCriteria,
		})
		if err != nil {
			return fmt.Errorf("failed to create task: %w", err)
		}

		fmt.Printf("Task created: %s\n", result.Value.ID)
		return nil
	},
}

var taskListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all tasks",
	RunE: func(cmd *cobra.Command, args []string) error {
		repo := bootstrap.TaskRepository(getDB())
		tasks, err := repo.List()
		if err != nil {
			return fmt.Errorf("failed to list tasks: %w", err)
		}

		if len(tasks) == 0 {
			fmt.Println("No tasks found")
			return nil
		}

		fmt.Printf("%-36s %-20s %-10s %-10s %-20s\n", "ID", "TITLE", "STATUS", "PRIORITY", "CREATED AT")
		for _, task := range tasks {
			fmt.Printf("%-36s %-20s %-10s %-10s %-20s\n",
				task.ID,
				truncate(task.Title, 20),
				task.Status,
				task.Priority,
				task.CreatedAt.Format("2006-01-02 15:04"),
			)
		}
		return nil
	},
}

var taskGetCmd = &cobra.Command{
	Use:   "get [id]",
	Short: "Get task details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		repo := bootstrap.TaskRepository(getDB())
		task, err := repo.GetByID(args[0])
		if err != nil {
			return fmt.Errorf("failed to get task: %w", err)
		}
		if task == nil {
			return fmt.Errorf("task not found: %s", args[0])
		}

		fmt.Printf("ID: %s\n", task.ID)
		fmt.Printf("Title: %s\n", task.Title)
		fmt.Printf("Description: %s\n", task.Description)
		fmt.Printf("Status: %s\n", task.Status)
		fmt.Printf("Priority: %s\n", task.Priority)
		fmt.Printf("Risk Level: %s\n", task.RiskLevel)
		fmt.Printf("Created At: %s\n", task.CreatedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("Updated At: %s\n", task.UpdatedAt.Format("2006-01-02 15:04:05"))
		return nil
	},
}

var taskGraphCmd = &cobra.Command{
	Use:   "graph",
	Short: "Manage task graphs",
}

var taskGraphCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a task graph from task acceptance criteria",
	RunE: func(cmd *cobra.Command, args []string) error {
		taskID, _ := cmd.Flags().GetString("task-id")
		replaceActive, _ := cmd.Flags().GetBool("replace-active")
		createdBy, _ := cmd.Flags().GetString("created-by")

		service := bootstrap.TaskGraphService(getDB())
		result, err := service.Decompose(cmd.Context(), bootstrap.DecomposeTaskGraphInput{
			TaskID:        taskID,
			ReplaceActive: replaceActive,
			CreatedBy:     createdBy,
		})
		if err != nil {
			return fmt.Errorf("failed to create task graph: %w", err)
		}

		fmt.Printf("Task graph created: %s (task: %s, version: %d, work units: %d)\n",
			result.Graph.ID,
			result.Graph.TaskID,
			result.Graph.Version,
			len(result.WorkUnits),
		)
		return nil
	},
}

var taskGraphListCmd = &cobra.Command{
	Use:   "list",
	Short: "List task graphs for a task",
	RunE: func(cmd *cobra.Command, args []string) error {
		taskID, _ := cmd.Flags().GetString("task-id")

		service := bootstrap.TaskGraphService(getDB())
		graphs, err := service.ListByTask(cmd.Context(), taskID)
		if err != nil {
			return fmt.Errorf("failed to list task graphs: %w", err)
		}
		if len(graphs) == 0 {
			fmt.Println("No task graphs found")
			return nil
		}
		fmt.Printf("%-36s %-8s %-12s %-20s %-10s %-10s\n", "ID", "VERSION", "STATUS", "PLANNER", "NODES", "EDGES")
		for _, graph := range graphs {
			fmt.Printf("%-36s %-8d %-12s %-20s %-10d %-10d\n",
				graph.ID,
				graph.Version,
				graph.Status,
				truncate(graph.PlannerStrategy, 20),
				graph.NodeCount,
				graph.EdgeCount,
			)
		}
		return nil
	},
}

func init() {
	taskCreateCmd.Flags().String("title", "", "Task title (required)")
	taskCreateCmd.Flags().String("description", "", "Task description")
	taskCreateCmd.Flags().String("priority", "P2", "Task priority (P0-P3)")
	taskCreateCmd.Flags().String("risk-level", "low", "Risk level (low, medium, high, critical)")
	taskCreateCmd.Flags().StringArray("acceptance", nil, "Acceptance criterion (repeatable)")
	_ = taskCreateCmd.MarkFlagRequired("title")

	taskGraphCreateCmd.Flags().String("task-id", "", "Task ID to decompose (required)")
	taskGraphCreateCmd.Flags().Bool("replace-active", false, "Supersede the active task graph before creating a new one")
	taskGraphCreateCmd.Flags().String("created-by", "cli", "Actor creating the task graph")
	taskGraphCreateCmd.Flags().String("planner", os.Getenv("ORCHESTRAOS_PLANNER_STRATEGY"), "Planner strategy (local_heuristic_v1, llm_gemini_v1)")
	_ = taskGraphCreateCmd.MarkFlagRequired("task-id")

	taskGraphListCmd.Flags().String("task-id", "", "Task ID to list graphs for (required)")
	_ = taskGraphListCmd.MarkFlagRequired("task-id")

	taskGraphCmd.AddCommand(taskGraphCreateCmd)
	taskGraphCmd.AddCommand(taskGraphListCmd)

	taskCmd.AddCommand(taskCreateCmd)
	taskCmd.AddCommand(taskListCmd)
	taskCmd.AddCommand(taskGetCmd)
	taskCmd.AddCommand(taskGraphCmd)
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
