package cmd

import (
	"github.com/spf13/cobra"

	"github.com/replicate/cli/internal/cmd/account"
)

var AccountCmd = &cobra.Command{
	Use:     "account",
	Short:   `Alias for "accounts current"`,
	Aliases: []string{"profile", "whoami"},
	RunE:    account.CurrentCmd.RunE,
}

func init() {
	account.AddCurrentAccountFlags(AccountCmd)
}
