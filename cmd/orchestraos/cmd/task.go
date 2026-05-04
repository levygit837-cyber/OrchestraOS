package cmd

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
	"github.com/levygit837-cyber/OrchestraOS/internal/eventstore"
	"github.com/levygit837-cyber/OrchestraOS/internal/repository"
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

		task := &domain.Task{
			Title:       title,
			Description: description,
			Status:      domain.TaskStatusCreated,
			Priority:    domain.Priority(priority),
			RiskLevel:   domain.RiskLevel(riskLevel),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		repo := repository.NewTaskRepository(getDB())
		if err := repo.Create(task); err != nil {
			return fmt.Errorf("failed to create task: %w", err)
		}

		// Create event
		eventStore, err := eventstore.NewStore(getDB())
		if err != nil {
			return fmt.Errorf("failed to create event store: %w", err)
		}

		payload, _ := json.Marshal(map[string]interface{}{
			"task_id": task.ID,
			"title":   task.Title,
			"status":  task.Status,
		})

		event := &domain.EventEnvelope{
			Type:        "task.created",
			Version:     "v1",
			TaskID:      task.ID,
			Priority:    domain.EventPriorityNotification,
			RequiresAck: false,
			Payload:     payload,
		}

		if err := eventStore.Append(event); err != nil {
			return fmt.Errorf("failed to append event: %w", err)
		}

		fmt.Printf("Task created: %s\n", task.ID)
		return nil
	},
}

var taskListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all tasks",
	RunE: func(cmd *cobra.Command, args []string) error {
		repo := repository.NewTaskRepository(getDB())
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
		repo := repository.NewTaskRepository(getDB())
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

func init() {
	taskCreateCmd.Flags().String("title", "", "Task title (required)")
	taskCreateCmd.Flags().String("description", "", "Task description")
	taskCreateCmd.Flags().String("priority", "P2", "Task priority (P0-P3)")
	taskCreateCmd.Flags().String("risk-level", "low", "Risk level (low, medium, high, critical)")
	taskCreateCmd.MarkFlagRequired("title")

	taskCmd.AddCommand(taskCreateCmd)
	taskCmd.AddCommand(taskListCmd)
	taskCmd.AddCommand(taskGetCmd)
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
