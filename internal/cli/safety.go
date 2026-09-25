package cli

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/sekai-labs/michibiki/pkg/model"
	"github.com/sekai-labs/michibiki/pkg/safety"
)

var (
	flagCommitConfirmSec int
)

var configCmd = &cobra.Command{
	Use:     "config",
	Aliases: []string{"cfg"},
	Short:   "Device configuration management, safety pre-flight inspection, and commit-confirm",
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show the current running configuration of the connected device",
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, _, err := resolveNetworkService(cmd)
		if err != nil {
			return err
		}
		defer svc.Provider().Disconnect(cmd.Context())

		cfgState, err := svc.GetConfigState(cmd.Context())
		if err != nil {
			return fmt.Errorf("failed to retrieve running config: %w", err)
		}
		cfg := cfgState.RunningConfig
		return formatOutput(cmd, map[string]string{"configuration": cfg}, func() {
			fmt.Println(cfg)
		})
	},
}

var configDiffCmd = &cobra.Command{
	Use:   "diff <candidate-file>",
	Short: "Diff candidate configuration against device running configuration",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		candidateFile := args[0]
		candidateBytes, err := os.ReadFile(candidateFile)
		if err != nil {
			return fmt.Errorf("failed to read candidate configuration '%s': %w", candidateFile, err)
		}
		candidate := string(candidateBytes)

		prov, _, err := resolveProvider(cmd)
		if err != nil {
			return err
		}
		defer prov.Disconnect(cmd.Context())

		running, err := prov.GetRunningConfig(cmd.Context())
		if err != nil {
			return fmt.Errorf("failed to retrieve running config: %w", err)
		}

		diffResult := computeSimpleDiff(running, candidate)

		return formatOutput(cmd, map[string]any{
			"candidate_file": candidateFile,
			"diff":           diffResult,
		}, func() {
			fmt.Printf("--- running config\n+++ %s\n", candidateFile)
			fmt.Println(diffResult)
		})
	},
}

var configValidateCmd = &cobra.Command{
	Use:   "validate <candidate-file>",
	Short: "Inspect candidate configuration for syntax errors and safety hazards",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		candidateFile := args[0]
		candidateBytes, err := os.ReadFile(candidateFile)
		if err != nil {
			return fmt.Errorf("failed to read candidate configuration '%s': %w", candidateFile, err)
		}
		candidate := string(candidateBytes)

		prov, profile, err := resolveProvider(cmd)
		if err != nil {
			return err
		}
		defer prov.Disconnect(cmd.Context())

		valResult, err := prov.ValidateConfig(cmd.Context(), candidate)
		if err != nil {
			return fmt.Errorf("provider validation failed: %w", err)
		}

		engine := safety.NewSafetyEngine()
		warnings := engine.InspectCandidate(candidate, profile.Address)
		for _, w := range warnings {
			valResult.Warnings = append(valResult.Warnings, fmt.Sprintf("[%s] %s", w.Severity, w.Message))
		}

		return formatOutput(cmd, valResult, func() {
			if valResult.Valid {
				fmt.Printf("✓ VALIDATION PASSED for candidate configuration '%s'\n", candidateFile)
			} else {
				fmt.Printf("✗ VALIDATION FAILED for candidate configuration '%s'\n", candidateFile)
				for _, errStr := range valResult.Errors {
					fmt.Printf("  ERROR: %s\n", errStr)
				}
			}

			if len(valResult.Warnings) > 0 {
				fmt.Printf("\nSafety Warnings (%d):\n", len(valResult.Warnings))
				for _, warnStr := range valResult.Warnings {
					fmt.Printf("  ⚠  %s\n", warnStr)
				}
			}
		})
	},
}

var configCommitCmd = &cobra.Command{
	Use:   "commit <candidate-file>",
	Short: "Apply candidate configuration with optional commit-confirm rollback timer",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		candidateFile := args[0]
		candidateBytes, err := os.ReadFile(candidateFile)
		if err != nil {
			return fmt.Errorf("failed to read candidate configuration '%s': %w", candidateFile, err)
		}
		candidate := string(candidateBytes)

		prov, profile, err := resolveProvider(cmd)
		if err != nil {
			return err
		}
		defer prov.Disconnect(cmd.Context())

		engine := safety.NewSafetyEngine()
		warnings := engine.InspectCandidate(candidate, profile.Address)
		if len(warnings) > 0 {
			fmt.Printf("Detected %d safety warnings in candidate configuration:\n", len(warnings))
			for _, w := range warnings {
				fmt.Printf("  ⚠ [%s] %s\n", w.Severity, w.Message)
			}
		}

		req := model.ConfigApplyRequest{
			CandidateConfig: candidate,
			ConfirmSeconds:  flagCommitConfirmSec,
			Description:     fmt.Sprintf("Applied from %s via michibiki CLI", candidateFile),
		}

		res, err := prov.ApplyConfig(cmd.Context(), req)
		if err != nil {
			return fmt.Errorf("failed to apply configuration: %w", err)
		}

		return formatOutput(cmd, res, func() {
			if res.Success {
				fmt.Printf("✓ Configuration successfully applied to '%s'\n", profile.Name)
				if res.RollbackID != "" {
					fmt.Printf("\nCOMMIT-CONFIRM ACTIVE: Rollback timer started (%d seconds).\n", flagCommitConfirmSec)
					fmt.Printf("To confirm changes permanently, run:\n")
					fmt.Printf("  michibiki config rollback %s\n", res.RollbackID)
				}
			} else {
				fmt.Printf("✗ FAILED to apply configuration: %s\n", res.Message)
			}
		})
	},
}

var configRollbackCmd = &cobra.Command{
	Use:   "rollback [id]",
	Short: "Revert configuration to previous revision or trigger immediate rollback",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		prov, _, err := resolveProvider(cmd)
		if err != nil {
			return err
		}
		defer prov.Disconnect(cmd.Context())

		rollbackID := ""
		if len(args) > 0 {
			rollbackID = args[0]
		}

		if err := prov.RollbackConfig(cmd.Context(), rollbackID); err != nil {
			return fmt.Errorf("rollback failed: %w", err)
		}

		res := map[string]any{
			"status":      "reverted",
			"rollback_id": rollbackID,
			"timestamp":   time.Now().Format(time.RFC3339),
		}

		return formatOutput(cmd, res, func() {
			fmt.Printf("✓ Configuration rollback triggered successfully (ID: %s)\n", rollbackID)
		})
	},
}

func computeSimpleDiff(oldStr, newStr string) string {
	oldLines := strings.Split(oldStr, "\n")
	newLines := strings.Split(newStr, "\n")

	var diffLines []string
	i, j := 0, 0
	for i < len(oldLines) && j < len(newLines) {
		if oldLines[i] == newLines[j] {
			diffLines = append(diffLines, "  "+oldLines[i])
			i++
			j++
		} else {
			diffLines = append(diffLines, "- "+oldLines[i])
			diffLines = append(diffLines, "+ "+newLines[j])
			i++
			j++
		}
	}
	for ; i < len(oldLines); i++ {
		diffLines = append(diffLines, "- "+oldLines[i])
	}
	for ; j < len(newLines); j++ {
		diffLines = append(diffLines, "+ "+newLines[j])
	}

	return strings.Join(diffLines, "\n")
}

func init() {
	configCommitCmd.Flags().IntVar(&flagCommitConfirmSec, "confirm", 0, "Commit-confirm rollback timeout in seconds (0 = immediate permanent commit)")

	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configDiffCmd)
	configCmd.AddCommand(configValidateCmd)
	configCmd.AddCommand(configCommitCmd)
	configCmd.AddCommand(configRollbackCmd)
}
