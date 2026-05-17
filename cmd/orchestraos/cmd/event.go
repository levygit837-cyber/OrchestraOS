package cmd

import (
	"encoding/json"
	"fmt"

	eventmod "github.com/levygit837-cyber/OrchestraOS/internal/core/event"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/statemachine"
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
	"github.com/spf13/cobra"
)

var eventCmd = &cobra.Command{
	Use:   "event",
	Short: "Manage events",
	Long:  `List and view events from the event store.`,
}

var eventListCmd = &cobra.Command{
	Use:   "list",
	Short: "List events",
	RunE: func(cmd *cobra.Command, args []string) error {
		taskID, _ := cmd.Flags().GetString("task-id")
		runID, _ := cmd.Flags().GetString("run-id")
		workUnitID, _ := cmd.Flags().GetString("workunit-id")

		eventService := eventmod.NewService(getDB())

		var events []domain.EventEnvelope

		if taskID != "" {
			e, err := eventService.ListByTask(cmd.Context(), taskID)
			if err != nil {
				return fmt.Errorf("failed to list events: %w", err)
			}
			events = e
		} else if runID != "" {
			e, err := eventService.ListByRun(cmd.Context(), runID)
			if err != nil {
				return fmt.Errorf("failed to list events: %w", err)
			}
			events = e
		} else if workUnitID != "" {
			e, err := eventService.ListByWorkUnit(cmd.Context(), workUnitID)
			if err != nil {
				return fmt.Errorf("failed to list events: %w", err)
			}
			events = e
		} else {
			e, err := eventService.List(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to list events: %w", err)
			}
			events = e
		}

		if len(events) == 0 {
			fmt.Println("No events found")
			return nil
		}

		fmt.Printf("%-36s %-30s %-10s %-10s %-20s\n", "ID", "TYPE", "SEQUENCE", "PRIORITY", "CREATED AT")
		for _, ev := range events {
			fmt.Printf("%-36s %-30s %-10d %-10s %-20s\n",
				ev.ID,
				truncate(ev.Type, 30),
				ev.Sequence,
				ev.Priority,
				ev.CreatedAt.Format("2006-01-02 15:04"),
			)
		}
		return nil
	},
}

var eventReplayCmd = &cobra.Command{
	Use:   "replay [task-id]",
	Short: "Replay events for a task",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		taskID := args[0]

		eventService := eventmod.NewService(getDB())
		events, err := eventService.ListByTask(cmd.Context(), taskID)
		if err != nil {
			return fmt.Errorf("failed to replay events: %w", err)
		}

		fmt.Printf("Replaying %d events for task %s:\n\n", len(events), taskID)
		state, err := statemachine.ProjectStrict(events)
		if err != nil {
			return fmt.Errorf("failed to reconstruct replay state: %w", err)
		}
		if state.TaskStatus != "" {
			fmt.Printf("Reconstructed task status: %s\n", state.TaskStatus)
		}
		if state.LastCheckpoint != nil {
			fmt.Printf("Last checkpoint: %s (sequence %d)\n", state.LastCheckpoint.ID, state.LastCheckpoint.Sequence)
		}
		fmt.Println()

		for _, ev := range events {
			var payload map[string]interface{}
			if err := json.Unmarshal(ev.Payload, &payload); err != nil {
				return fmt.Errorf("failed to decode payload for event %s: %w", ev.ID, err)
			}

			fmt.Printf("[%d] %s (%s)\n", ev.Sequence, ev.Type, ev.CreatedAt.Format("15:04:05"))
			fmt.Printf("    Priority: %s | Ack: %v\n", ev.Priority, ev.RequiresAck)
			if len(payload) > 0 {
				payloadJSON, err := json.MarshalIndent(payload, "    ", "  ")
				if err != nil {
					return fmt.Errorf("failed to format payload for event %s: %w", ev.ID, err)
				}
				fmt.Printf("    Payload: %s\n", payloadJSON)
			}
			fmt.Println()
		}

		return nil
	},
}

func init() {
	eventListCmd.Flags().String("task-id", "", "Filter by task ID")
	eventListCmd.Flags().String("run-id", "", "Filter by run ID")
	eventListCmd.Flags().String("workunit-id", "", "Filter by work unit ID")

	eventCmd.AddCommand(eventListCmd)
	eventCmd.AddCommand(eventReplayCmd)
}
