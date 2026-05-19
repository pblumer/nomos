package cli

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/nomos/nomos/internal/idgen"
	"github.com/nomos/nomos/internal/idmigrate"
)

// idCmd liefert das `nomos id`-Subkommando mit new/check/migrate.
func idCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "id",
		Short: "ID-Tools (ADR-0020): generieren, prüfen, migrieren",
	}
	c.AddCommand(idNewCmd(), idCheckCmd(), idMigrateCmd())
	return c
}

func idNewCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "new <artefact-type>",
		Short: "Eine frische ID für einen Artefakttyp erzeugen",
		Long: "Bekannte Artefakttypen: " + strings.Join(sortedTypes(), ", ") +
			"\n\nBeispiel: nomos id new product_blueprint",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := idgen.NewForType(args[0])
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), id)
			return nil
		},
	}
	return c
}

func idCheckCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "check",
		Short: "Listet Artefakte mit Legacy-IDs (ohne Änderungen)",
		RunE: func(cmd *cobra.Command, args []string) error {
			p, _ := cmd.Flags().GetString("path")
			outFmt, _ := cmd.Flags().GetString("format")
			if err := validateFormat(outFmt); err != nil {
				return err
			}
			cands, err := idmigrate.Scan(p)
			if err != nil {
				return err
			}
			if outFmt == "json" {
				return json.NewEncoder(cmd.OutOrStdout()).Encode(map[string]any{
					"count":      len(cands),
					"candidates": cands,
				})
			}
			if len(cands) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "OK keine Legacy-IDs gefunden")
				return nil
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%d Artefakte mit Legacy-IDs:\n", len(cands))
			for _, k := range cands {
				fmt.Fprintf(cmd.OutOrStdout(), "  %-40s %-22s %s\n", k.OldID, k.ArtefactType, k.Path)
			}
			fmt.Fprintln(cmd.OutOrStdout(), "\nMit `nomos id migrate --path "+p+"` migrieren.")
			return nil
		},
	}
	c.Flags().String("path", ".", "Pfad zum Cosmos-Repository")
	c.Flags().String("format", "text", "Output format: text or json")
	return c
}

func idMigrateCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "migrate",
		Short: "Migriert Legacy-IDs auf das ADR-0020-Format",
		Long: "Erzeugt für jedes Artefakt mit Legacy-ID eine neue ADR-0020-konforme ID,\n" +
			"schreibt sie in die YAML-Dateien, benennt Dateien um, aktualisiert\n" +
			"Cross-References und persistiert die Alt→Neu-Zuordnung in\n" +
			".nomos/id-history.yaml. Mit --dry-run kann der Plan vorab geprüft werden.",
		RunE: func(cmd *cobra.Command, args []string) error {
			p, _ := cmd.Flags().GetString("path")
			outFmt, _ := cmd.Flags().GetString("format")
			dryRun, _ := cmd.Flags().GetBool("dry-run")
			if err := validateFormat(outFmt); err != nil {
				return err
			}
			cands, err := idmigrate.Scan(p)
			if err != nil {
				return err
			}
			plan, err := idmigrate.BuildPlan(cands)
			if err != nil {
				return err
			}
			if outFmt == "json" {
				out := map[string]any{
					"dry_run": dryRun,
					"steps":   plan.Steps,
				}
				if !dryRun {
					if _, err := idmigrate.Apply(p, plan, false); err != nil {
						return err
					}
					out["applied"] = true
				}
				return json.NewEncoder(cmd.OutOrStdout()).Encode(out)
			}
			if len(plan.Steps) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "OK keine Migration nötig")
				return nil
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Plan (%d Artefakte):\n", len(plan.Steps))
			for _, s := range plan.Steps {
				rename := ""
				if s.NewPath != "" && s.NewPath != s.Path {
					rename = " → Datei: " + s.NewPath
				}
				fmt.Fprintf(cmd.OutOrStdout(), "  %-30s → %s  [%s]%s\n", s.OldID, s.NewID, s.ArtefactType, rename)
			}
			if dryRun {
				fmt.Fprintln(cmd.OutOrStdout(), "\n(dry-run, keine Änderungen geschrieben)")
				return nil
			}
			if _, err := idmigrate.Apply(p, plan, false); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "\nMigration abgeschlossen. History: .nomos/id-history.yaml\n")
			return nil
		},
	}
	c.Flags().String("path", ".", "Pfad zum Cosmos-Repository")
	c.Flags().String("format", "text", "Output format: text or json")
	c.Flags().Bool("dry-run", false, "Plan nur anzeigen, keine Dateien ändern")
	return c
}

func sortedTypes() []string {
	t := idgen.KnownTypes()
	sort.Strings(t)
	return t
}
