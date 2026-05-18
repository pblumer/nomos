package cli

import (
	"context"
	"fmt"
	"net/http"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/nomos/nomos/internal/mcpserver"
	"github.com/spf13/cobra"
)

func mcpCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mcp",
		Short: "Nomos MCP-Server starten (stdio oder HTTP)",
	}
	cmd.AddCommand(mcpServeStdioCmd(), mcpServeHTTPCmd())
	return cmd
}

func mcpServeStdioCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "MCP-Server über stdio starten (für Claude Code / KI-Agenten)",
		RunE: func(cmd *cobra.Command, args []string) error {
			p, _ := cmd.Flags().GetString("path")
			srv := mcpserver.New(p)
			fmt.Fprintln(cmd.ErrOrStderr(), "Nomos MCP server starting (stdio)")
			return srv.Run(context.Background(), &mcp.StdioTransport{})
		},
	}
	cmd.Flags().String("path", ".", "Pfad zum Cosmos-Workspace")
	return cmd
}

func mcpServeHTTPCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "http",
		Short: "MCP-Server über HTTP/SSE starten",
		RunE: func(cmd *cobra.Command, args []string) error {
			p, _ := cmd.Flags().GetString("path")
			listen, _ := cmd.Flags().GetString("listen")
			srv := mcpserver.New(p)
			handler := mcp.NewStreamableHTTPHandler(func(_ *http.Request) *mcp.Server { return srv }, nil)
			fmt.Fprintf(cmd.OutOrStdout(), "Nomos MCP HTTP server on http://%s/mcp\n", listen)
			return http.ListenAndServe(listen, handler)
		},
	}
	cmd.Flags().String("path", ".", "Pfad zum Cosmos-Workspace")
	cmd.Flags().String("listen", "127.0.0.1:7374", "Listen-Adresse")
	return cmd
}
