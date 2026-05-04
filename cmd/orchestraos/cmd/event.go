package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
	"github.com/levygit837-cyber/OrchestraOS/internal/eventstore"
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

		eventStore, err := eventstore.NewStore(getDB())
		if err != nil {
			return fmt.Errorf("failed to create event store: %w", err)
		}

		var events []domain.EventEnvelope

		if taskID != "" {
			e, err := eventStore.ListByTask(taskID)
			if err != nil {
				return fmt.Errorf("failed to list events: %w", err)
			}
			events = e
		} else if runID != "" {
			e, err := eventStore.ListByRun(runID)
			if err != nil {
				return fmt.Errorf("failed to list events: %w", err)
			}
			events = e
		} else if workUnitID != "" {
			e, err := eventStore.ListByWorkUnit(workUnitID)
			if err != nil {
				return fmt.Errorf("failed to list events: %w", err)
			}
			events = e
		} else {
			e, err := eventStore.List()
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

		eventStore, err := eventstore.NewStore(getDB())
		if err != nil {
			return fmt.Errorf("failed to create event store: %w", err)
		}

		events, err := eventStore.Replay(taskID)
		if err != nil {
			return fmt.Errorf("failed to replay events: %w", err)
		}

		fmt.Printf("Replaying %d events for task %s:\n\n", len(events), taskID)

		for _, ev := range events {
			var payload map[string]interface{}
			json.Unmarshal(ev.Payload, &payload)

			fmt.Printf("[%d] %s (%s)\n", ev.Sequence, ev.Type, ev.CreatedAt.Format("15:04:05"))
			fmt.Printf("    Priority: %s | Ack: %v\n", ev.Priority, ev.RequiresAck)
			if len(payload) > 0 {
				payloadJSON, _ := json.MarshalIndent(payload, "    ", "  ")
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
