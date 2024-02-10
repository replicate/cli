package auth

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/replicate/cli/internal/client"
	"github.com/replicate/cli/internal/config"
	"github.com/spf13/cobra"
)

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login --token-stdin",
	Short: "Log in to Replicate",
	Long: `Log in to Replicate

You can find your Replicate API token at https://replicate.com/account`,
	Example: `
	# Log in with environment variable
	$ echo $REPLICATE_API_TOKEN | replicate auth login --token-stdin

	# Log in with token file
	$ replicate auth login --token-stdin < path/to/token`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		tokenStdin, err := cmd.Flags().GetBool("token-stdin")
		if err != nil {
			return err
		}

		var token string
		if tokenStdin {
			token, err = readTokenFromStdin()
			if err != nil {
				return fmt.Errorf("failed to read token from stdin: %w", err)
			}
			if token == "" {
				return fmt.Errorf("no token provided (empty string)")
			}
		} else {
			return fmt.Errorf("token must be passed to stdin with --token-stdin flag")
		}
		token = strings.TrimSpace(token)

		ok, err := client.VerifyToken(ctx, token)
		if err != nil {
			return fmt.Errorf("error verifying token: %w", err)
		}
		if !ok {
			return fmt.Errorf("invalid token")
		}

		if err := config.SetAPIToken(token); err != nil {
			return fmt.Errorf("failed to set API token: %w", err)
		}

		fmt.Printf("Token saved to configuration file: %s\n", config.ConfigFilePath)

		return nil
	},
}

func readTokenFromStdin() (string, error) {
	tokenBytes, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", fmt.Errorf("Failed to read token from stdin: %w", err)
	}
	return string(tokenBytes), nil
}

func init() {
	loginCmd.Flags().Bool("token-stdin", false, "Take the token from stdin.")
	_ = loginCmd.MarkFlagRequired("token-stdin")
}
