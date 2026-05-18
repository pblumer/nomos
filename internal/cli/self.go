package cli

import (
	"encoding/json"
	"fmt"

	"github.com/nomos/nomos/internal/selfmodel"
	"github.com/spf13/cobra"
)

func selfCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "self",
		Short: "Nomos Self-Model-Bundle verwalten (ADR-0014)",
	}
	c.AddCommand(selfImportCmd(), selfStatusCmd(), selfDiffCmd(), selfUpgradeCmd())
	return c
}

func selfImportCmd() *cobra.Command {
	var forceOverwrite bool
	cmd := &cobra.Command{
		Use:   "import",
		Short: "Self-Model-Bundle in den Cosmos importieren",
		RunE: func(cmd *cobra.Command, args []string) error {
			p, _ := cmd.Flags().GetString("path")
			if err := selfmodel.Import(p, forceOverwrite); err != nil {
				return err
			}
			meta, _ := selfmodel.LoadBundleMeta()
			fmt.Fprintf(cmd.OutOrStdout(), "Self-Model v%s importiert nach %s\n", meta.Version, p)
			return nil
		},
	}
	cmd.Flags().String("path", ".", "")
	cmd.Flags().BoolVar(&forceOverwrite, "force-overwrite-self", false, "Lokale Änderungen an Self-Model-Dateien überschreiben")
	return cmd
}

func selfStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Self-Model-Status des Cosmos anzeigen",
		RunE: func(cmd *cobra.Command, args []string) error {
			p, _ := cmd.Flags().GetString("path")
			outFmt, _ := cmd.Flags().GetString("format")
			st, err := selfmodel.Status(p)
			if err != nil {
				return err
			}
			if outFmt == "json" {
				return json.NewEncoder(cmd.OutOrStdout()).Encode(st)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Fall: %s\n%s\n", st.Case, st.Message)
			if st.WorkspaceVersion != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "Workspace:  v%s (%s)\n", st.WorkspaceVersion, st.WorkspaceChecksum[:8])
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Binary:     v%s (%s)\n", st.BinaryVersion, st.BinaryChecksum[:8])
			return nil
		},
	}
	cmd.Flags().String("path", ".", "")
	cmd.Flags().String("format", "text", "")
	return cmd
}

func selfDiffCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "diff",
		Short: "Unterschiede zwischen Binary-Bundle und Workspace anzeigen",
		RunE: func(cmd *cobra.Command, args []string) error {
			p, _ := cmd.Flags().GetString("path")
			st, err := selfmodel.Status(p)
			if err != nil {
				return err
			}
			switch st.Case {
			case "A":
				fmt.Fprintln(cmd.OutOrStdout(), "Self-Model noch nicht importiert.")
			case "B":
				fmt.Fprintln(cmd.OutOrStdout(), "Keine Unterschiede — Workspace ist aktuell.")
			default:
				fmt.Fprintf(cmd.OutOrStdout(), "Fall %s: %s\n", st.Case, st.Message)
				fmt.Fprintf(cmd.OutOrStdout(), "  Binary-Checksum:    %s\n", st.BinaryChecksum)
				fmt.Fprintf(cmd.OutOrStdout(), "  Workspace-Checksum: %s\n", st.WorkspaceChecksum)
				if st.ActualChecksum != "" && st.ActualChecksum != st.WorkspaceChecksum {
					fmt.Fprintf(cmd.OutOrStdout(), "  Actual-Checksum:    %s\n", st.ActualChecksum)
				}
			}
			return nil
		},
	}
	cmd.Flags().String("path", ".", "")
	return cmd
}

func selfUpgradeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "upgrade",
		Short: "Self-Model im Workspace auf die Binary-Version aktualisieren",
		RunE: func(cmd *cobra.Command, args []string) error {
			p, _ := cmd.Flags().GetString("path")
			if err := selfmodel.Import(p, true); err != nil {
				return err
			}
			meta, _ := selfmodel.LoadBundleMeta()
			fmt.Fprintf(cmd.OutOrStdout(), "Self-Model auf v%s aktualisiert.\n", meta.Version)
			return nil
		},
	}
	cmd.Flags().String("path", ".", "")
	return cmd
}
