package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/levygit837-cyber/OrchestraOS/internal/bootstrap"
	orchestratormod "github.com/levygit837-cyber/OrchestraOS/internal/modules/orchestrator"
	"github.com/spf13/cobra"
)

var taskRunCmd = &cobra.Command{
	Use:   "run",
	Short: "Run a task end-to-end via OrchestratorService",
	Long:  `Execute a complete task by orchestrating decomposition, agent spawning, runtime execution, and completion.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		taskID, _ := cmd.Flags().GetString("id")
		runtimeType, _ := cmd.Flags().GetString("runtime")
		planner, _ := cmd.Flags().GetString("planner")
		maxSteps, _ := cmd.Flags().GetInt("max-steps")
		timeout, _ := cmd.Flags().GetInt("timeout")

		if taskID == "" {
			return fmt.Errorf("--id is required")
		}

		// Resolve planner strategy fallback
		if planner == "" {
			planner = os.Getenv("ORCHESTRAOS_PLANNER_STRATEGY")
		}
		if planner == "" {
			planner = "local_heuristic_v1"
		}

		// Resolve defaults
		if maxSteps <= 0 {
			maxSteps = 10
		}
		if timeout <= 0 {
			timeout = 300
		}

		fmt.Printf("Orchestrating task: %s\n", taskID)
		fmt.Printf("  → Runtime: %s | Planner: %s | MaxSteps: %d | Timeout: %ds\n",
			runtimeType, planner, maxSteps, timeout)

		orchService := bootstrap.OrchestratorService(getDB())

		start := time.Now()
		result, err := orchService.RunTask(cmd.Context(), taskID, orchestratormod.RunTaskOptions{
			RuntimeType:     runtimeType,
			PlannerStrategy: planner,
			MaxSteps:        maxSteps,
			TimeoutSeconds:  timeout,
		})
		elapsed := time.Since(start)

		if err != nil {
			fmt.Printf("\n  → Task failed: %v\n", err)
			return fmt.Errorf("task run failed: %w", err)
		}

		fmt.Printf("\n  → Task %s\n", result.Status)
		fmt.Printf("  → Runs: %d | Reviews: %d | Time: %.1fs\n",
			len(result.RunIDs), len(result.ReviewIDs), elapsed.Seconds())

		if len(result.RunIDs) > 0 {
			fmt.Println("  → Runs created:")
			for _, rid := range result.RunIDs {
				fmt.Printf("      - %s\n", rid)
			}
		}

		if len(result.ReviewIDs) > 0 {
			fmt.Println("  → Reviews pending:")
			for _, rid := range result.ReviewIDs {
				fmt.Printf("      - %s\n", rid)
			}
		}

		return nil
	},
}

func init() {
	taskRunCmd.Flags().String("id", "", "Task ID to execute (required)")
	taskRunCmd.Flags().String("runtime", "fake", "Runtime type (fake, gemini, codex_cli)")
	taskRunCmd.Flags().String("planner", "", "Planner strategy (local_heuristic_v1, llm_gemini_v1). Falls back to ORCHESTRAOS_PLANNER_STRATEGY env var")
	taskRunCmd.Flags().Int("max-steps", 10, "Maximum steps per work unit")
	taskRunCmd.Flags().Int("timeout", 300, "Timeout in seconds per work unit")
	taskRunCmd.MarkFlagRequired("id")

	taskCmd.AddCommand(taskRunCmd)
}
