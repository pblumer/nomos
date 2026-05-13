package cli

import (
	"encoding/json"
	"fmt"

	"github.com/nomos/nomos/internal/app"
	"github.com/spf13/cobra"
)

func processCmd() *cobra.Command {
	c := &cobra.Command{Use: "process", Short: "Manage product-level BPMN process artifacts"}
	list := &cobra.Command{Use: "list", RunE: func(cmd *cobra.Command, args []string) error {
		p, _ := cmd.Flags().GetString("path")
		product, _ := cmd.Flags().GetString("product")
		outFmt, _ := cmd.Flags().GetString("format")
		var dto app.ProcessesDTO
		var err error
		if product != "" {
			dto, err = app.ListProductProcesses(p, product)
		} else {
			dto, err = app.ListProcesses(p)
		}
		if err != nil {
			return writeCLIError(cmd, outFmt, err)
		}
		if outFmt == "json" {
			return json.NewEncoder(cmd.OutOrStdout()).Encode(dto)
		}
		for _, pr := range dto.Items {
			fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\tproduct=%s\t%s\n", pr.ID, pr.Name, pr.RelatedProduct, pr.Status)
		}
		return nil
	}}
	list.Flags().String("path", ".", "")
	list.Flags().String("product", "", "Product ID")
	list.Flags().String("format", "text", "Output format: text or json")
	show := &cobra.Command{Use: "show <process-id>", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		p, _ := cmd.Flags().GetString("path")
		outFmt, _ := cmd.Flags().GetString("format")
		dto, err := app.GetProcess(p, args[0])
		if err != nil {
			return writeCLIError(cmd, outFmt, err)
		}
		if outFmt == "json" {
			return json.NewEncoder(cmd.OutOrStdout()).Encode(dto)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "%s %s product=%s tasks=%d status=%s\n", dto.ID, dto.Name, dto.RelatedProduct, len(dto.Tasks), dto.Validation.Status)
		return nil
	}}
	show.Flags().String("path", ".", "")
	show.Flags().String("format", "text", "Output format: text or json")
	create := &cobra.Command{Use: "create", RunE: func(cmd *cobra.Command, args []string) error {
		p, _ := cmd.Flags().GetString("path")
		product, _ := cmd.Flags().GetString("product")
		name, _ := cmd.Flags().GetString("name")
		id, _ := cmd.Flags().GetString("id")
		outFmt, _ := cmd.Flags().GetString("format")
		dto, err := app.CreateProductProcess(p, product, app.CreateProcessRequest{ID: id, Name: name})
		if err != nil {
			return writeCLIError(cmd, outFmt, err)
		}
		if outFmt == "json" {
			return json.NewEncoder(cmd.OutOrStdout()).Encode(dto)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "created %s\n", dto.ID)
		return nil
	}}
	create.Flags().String("path", ".", "")
	create.Flags().String("product", "", "Product ID")
	create.Flags().String("name", "", "Process name")
	create.Flags().String("id", "", "Process ID")
	create.Flags().String("format", "text", "Output format: text or json")
	_ = create.MarkFlagRequired("product")
	tasks := &cobra.Command{Use: "tasks <process-id>", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		p, _ := cmd.Flags().GetString("path")
		outFmt, _ := cmd.Flags().GetString("format")
		items, err := app.ProcessTasks(p, args[0])
		if err != nil {
			return writeCLIError(cmd, outFmt, err)
		}
		if outFmt == "json" {
			return json.NewEncoder(cmd.OutOrStdout()).Encode(map[string]any{"items": items, "count": len(items)})
		}
		for _, t := range items {
			fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\t%s\t%s\n", t.ID, t.ElementType, t.MappingStatus, t.ServiceRef)
		}
		return nil
	}}
	tasks.Flags().String("path", ".", "")
	tasks.Flags().String("format", "text", "Output format: text or json")
	validate := &cobra.Command{Use: "validate <process-id>", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		p, _ := cmd.Flags().GetString("path")
		outFmt, _ := cmd.Flags().GetString("format")
		dto, err := app.ValidateProcess(p, args[0])
		if err != nil {
			return writeCLIError(cmd, outFmt, err)
		}
		if outFmt == "json" {
			return json.NewEncoder(cmd.OutOrStdout()).Encode(dto)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "%s findings=%d\n", dto.Status, len(dto.Findings))
		return nil
	}}
	validate.Flags().String("path", ".", "")
	validate.Flags().String("format", "text", "Output format: text or json")
	c.AddCommand(list, show, create, tasks, validate)
	return c
}
