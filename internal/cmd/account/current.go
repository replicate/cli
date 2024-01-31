package account

import (
	"encoding/json"
	"fmt"

	"github.com/cli/browser"
	"github.com/spf13/cobra"

	"github.com/replicate/cli/internal/client"
	"github.com/replicate/cli/internal/util"
)

// CurrentCmd represents the get current account command
var CurrentCmd = &cobra.Command{
	Use:   "current",
	Short: "Show the current account",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		r8, err := client.NewClient()
		if err != nil {
			return err
		}

		account, err := r8.GetCurrentAccount(ctx)
		if err != nil {
			return fmt.Errorf("failed to get account: %w", err)
		}

		if cmd.Flags().Changed("web") {
			if util.IsTTY() {
				fmt.Println("Opening in browser...")
			}

			url := "https://replicate.com/" + account.Username
			err := browser.OpenURL(url)
			if err != nil {
				return fmt.Errorf("failed to open browser: %w", err)
			}

			return nil
		}

		if cmd.Flags().Changed("json") || !util.IsTTY() {
			bytes, err := json.MarshalIndent(account, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal account: %w", err)
			}
			fmt.Println(string(bytes))
			return nil
		}

		fmt.Printf("Type: %s\n", account.Type)
		fmt.Printf("Username: %s\n", account.Username)
		fmt.Printf("Name: %s\n", account.Name)
		fmt.Printf("GitHub URL: %s\n", account.GithubURL)

		return nil
	},
}

func init() {
	AddCurrentAccountFlags(CurrentCmd)
}

func AddCurrentAccountFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Emit JSON")
	cmd.Flags().Bool("web", false, "View on web")
	cmd.MarkFlagsMutuallyExclusive("json", "web")
}
