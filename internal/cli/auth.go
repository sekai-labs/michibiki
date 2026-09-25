package cli

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/sekai-labs/michibiki/internal/handler/auth"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage authentication and encrypted token sessions",
	Long:  "Commands to import, inspect, and clear temporary encrypted session tokens for passwordless and env-free access.",
}

var authTokenCmd = &cobra.Command{
	Use:   "token",
	Short: "Manage encrypted token sessions",
}

var authTokenImportCmd = &cobra.Command{
	Use:   "import",
	Short: "Import credentials from a token file into a temporary encrypted session",
	Long: `Reads credentials from a token file (-f/--token-file) and encrypts them using
AES-256-GCM into a temporary machine-isolated session store. Subsequent michibiki commands
will automatically use this session token.

Supported file formats:
  - JSON credentials: {"token":"..."}, {"api_key":"...", "api_secret":"..."}
  - Colon-delimited key: "api_key:api_secret"
  - Raw token string (one line)`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if flagTokenFile == "" {
			return fmt.Errorf("token file must be specified with -f or --token-file")
		}
		authHdlr := auth.NewAuthHandler()
		creds, err := authHdlr.ParseTokenFile(flagTokenFile)
		if err != nil {
			return fmt.Errorf("failed to parse token file: %w", err)
		}

		targetDevice := flagDevice
		if err := authHdlr.StoreTempEncryptedToken(targetDevice, creds); err != nil {
			return fmt.Errorf("failed to store encrypted session token: %w", err)
		}

		hint := auth.ObfuscateCredential(creds)
		sessionPath := authHdlr.GetTempSessionPath()
		deviceLabel := "global (all devices)"
		if targetDevice != "" {
			deviceLabel = fmt.Sprintf("device '%s'", targetDevice)
		}

		fmt.Printf("Successfully imported token session:\n")
		fmt.Printf("  Target:    %s\n", deviceLabel)
		fmt.Printf("  Token:     %s\n", hint)
		fmt.Printf("  Session:   %s (mode: 0600, AES-256-GCM)\n", sessionPath)
		return nil
	},
}

var authTokenStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check the status of the current temporary encrypted token session",
	RunE: func(cmd *cobra.Command, args []string) error {
		authHdlr := auth.NewAuthHandler()
		status, err := authHdlr.SessionStatus()
		if err != nil {
			return fmt.Errorf("failed to check session status: %w", err)
		}

		if !status.Active {
			fmt.Println("No active temporary token session found.")
			fmt.Println("Use 'michibiki auth token import -f <file>' or pass '-f <file>' directly.")
			return nil
		}

		fmt.Println("Active Encrypted Token Session:")
		if status.Device != "" {
			fmt.Printf("  Device:    %s\n", status.Device)
		} else {
			fmt.Printf("  Device:    global\n")
		}
		fmt.Printf("  Token:     %s\n", status.TokenHint)
		fmt.Printf("  Path:      %s\n", status.Path)
		fmt.Printf("  Created:   %s\n", status.CreatedAt.Format(time.RFC1123))
		return nil
	},
}

var authTokenClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear and wipe the temporary encrypted token session",
	RunE: func(cmd *cobra.Command, args []string) error {
		authHdlr := auth.NewAuthHandler()
		if err := authHdlr.ClearTempEncryptedTokens(); err != nil {
			return fmt.Errorf("failed to clear session token: %w", err)
		}
		fmt.Println("Temporary encrypted token session cleared.")
		return nil
	},
}

func init() {
	authTokenCmd.AddCommand(authTokenImportCmd)
	authTokenCmd.AddCommand(authTokenStatusCmd)
	authTokenCmd.AddCommand(authTokenClearCmd)
	authCmd.AddCommand(authTokenCmd)
}
