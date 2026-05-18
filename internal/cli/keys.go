package cli

import (
	"fmt"

	"github.com/nomos/nomos/internal/app"
	"github.com/spf13/cobra"
)

func keyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "key",
		Short: "Manage API keys for this cosmos",
	}
	cmd.AddCommand(keyCreateCmd(), keyListCmd(), keyRevokeCmd())
	return cmd
}

func keyCreateCmd() *cobra.Command {
	var path string
	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Generate a new API key",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := app.CreateKey(path, args[0])
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Key created for %q — store it now, it will not be shown again:\n\n  %s\n\n", result.Name, result.Key)
			return nil
		},
	}
	cmd.Flags().StringVar(&path, "path", ".", "cosmos path")
	return cmd
}

func keyListCmd() *cobra.Command {
	var path string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List API keys",
		RunE: func(cmd *cobra.Command, args []string) error {
			keys, err := app.ListKeys(path)
			if err != nil {
				return err
			}
			if len(keys) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No API keys configured.")
				return nil
			}
			for _, k := range keys {
				fmt.Fprintf(cmd.OutOrStdout(), "  %-20s  created %s\n", k.Name, k.Created.Format("2006-01-02"))
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&path, "path", ".", "cosmos path")
	return cmd
}

func keyRevokeCmd() *cobra.Command {
	var path string
	cmd := &cobra.Command{
		Use:   "revoke <name>",
		Short: "Revoke an API key by name",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := app.RevokeKey(path, args[0]); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Key %q revoked.\n", args[0])
			return nil
		},
	}
	cmd.Flags().StringVar(&path, "path", ".", "cosmos path")
	return cmd
}
