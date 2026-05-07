package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/nomos/nomos/internal/app"
	"github.com/nomos/nomos/internal/fsx"
	"github.com/nomos/nomos/internal/model"
)

func servicegraphCmd() *cobra.Command {
	c := &cobra.Command{Use: "servicegraph", Short: "Manage servicegraphs"}

	list := &cobra.Command{Use: "list", Short: "List all servicegraphs", RunE: func(cmd *cobra.Command, args []string) error {
		p, _ := cmd.Flags().GetString("path")
		outFmt, _ := cmd.Flags().GetString("format")
		dto, err := app.ListServicegraphs(p)
		if err != nil {
			return err
		}
		if outFmt == "json" {
			return printJSON(dto)
		}
		fmt.Printf("Servicegraphs: %d\n", dto.Count)
		for _, sg := range dto.Servicegraphs {
			fmt.Printf("  %s  %s  [%s]  nodes=%d edges=%d\n", sg.ID, sg.Name, sg.Status, sg.NodeCount, sg.EdgeCount)
		}
		return nil
	}}
	list.Flags().StringP("path", "p", ".", "Cosmos path")
	list.Flags().StringP("format", "f", "text", "Output format (text|json)")

	get := &cobra.Command{Use: "get <id>", Args: cobra.ExactArgs(1), Short: "Get a servicegraph", RunE: func(cmd *cobra.Command, args []string) error {
		p, _ := cmd.Flags().GetString("path")
		outFmt, _ := cmd.Flags().GetString("format")
		dto, err := app.GetServicegraph(p, args[0])
		if err != nil {
			return err
		}
		if outFmt == "json" {
			return printJSON(dto)
		}
		fmt.Printf("ID:      %s\nName:    %s\nVersion: %s\nStatus:  %s\nOwner:   %s\nProduct: %s\nNodes:   %d\nEdges:   %d\nRules:   %d\n",
			dto.ID, dto.Name, dto.Version, dto.Status, dto.Owner, dto.RelatedProduct, dto.NodeCount, dto.EdgeCount, dto.RuleCount)
		return nil
	}}
	get.Flags().StringP("path", "p", ".", "Cosmos path")
	get.Flags().StringP("format", "f", "text", "Output format (text|json)")

	create := &cobra.Command{Use: "create <file.yaml>", Args: cobra.ExactArgs(1), Short: "Create a servicegraph from YAML file", RunE: func(cmd *cobra.Command, args []string) error {
		p, _ := cmd.Flags().GetString("path")
		var sg model.Servicegraph
		if err := fsx.ReadYAML(args[0], &sg); err != nil {
			return fmt.Errorf("reading %s: %w", args[0], err)
		}
		if err := app.CreateServicegraph(p, sg); err != nil {
			return err
		}
		fmt.Printf("Created servicegraph: %s\n", sg.ID)
		return nil
	}}
	create.Flags().StringP("path", "p", ".", "Cosmos path")

	del := &cobra.Command{Use: "delete <id>", Args: cobra.ExactArgs(1), Short: "Delete a servicegraph", RunE: func(cmd *cobra.Command, args []string) error {
		p, _ := cmd.Flags().GetString("path")
		if err := app.DeleteServicegraph(p, args[0]); err != nil {
			return err
		}
		fmt.Printf("Deleted: %s\n", args[0])
		return nil
	}}
	del.Flags().StringP("path", "p", ".", "Cosmos path")

	validateCmd := &cobra.Command{Use: "validate <id>", Args: cobra.ExactArgs(1), Short: "Validate a servicegraph", RunE: func(cmd *cobra.Command, args []string) error {
		p, _ := cmd.Flags().GetString("path")
		dto, err := app.GetServicegraph(p, args[0])
		if err != nil {
			return err
		}
		_ = dto
		res, err := app.ValidateCosmos(p)
		if err != nil {
			return err
		}
		if res.Status == "ok" {
			fmt.Println("Validation passed")
		} else {
			fmt.Printf("Validation: %s\n", res.Status)
			for _, f := range res.Findings {
				fmt.Printf("  [%s] %s: %s\n", f.Severity, f.Code, f.Message)
			}
		}
		return nil
	}}
	validateCmd.Flags().StringP("path", "p", ".", "Cosmos path")

	mermaid := &cobra.Command{Use: "mermaid <id>", Args: cobra.ExactArgs(1), Short: "Output Mermaid diagram", RunE: func(cmd *cobra.Command, args []string) error {
		p, _ := cmd.Flags().GetString("path")
		dto, err := app.GetServicegraphMermaid(p, args[0])
		if err != nil {
			return err
		}
		fmt.Print(dto.Content)
		return nil
	}}
	mermaid.Flags().StringP("path", "p", ".", "Cosmos path")

	execution := &cobra.Command{Use: "execution <id>", Args: cobra.ExactArgs(1), Short: "Output execution order", RunE: func(cmd *cobra.Command, args []string) error {
		p, _ := cmd.Flags().GetString("path")
		outFmt, _ := cmd.Flags().GetString("format")
		dto, err := app.GetServicegraphExecutionOrder(p, args[0])
		if err != nil {
			return err
		}
		if outFmt == "json" {
			return printJSON(dto)
		}
		for i, group := range dto.Steps {
			fmt.Printf("Step %d: %v\n", i+1, group)
		}
		return nil
	}}
	execution.Flags().StringP("path", "p", ".", "Cosmos path")
	execution.Flags().StringP("format", "f", "text", "Output format (text|json)")

	c.AddCommand(list, get, create, del, validateCmd, mermaid, execution)
	return c
}

func printJSON(v interface{}) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
