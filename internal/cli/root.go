package cli

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/nomos/nomos/internal/app"
	"github.com/nomos/nomos/internal/fsx"
	"github.com/nomos/nomos/internal/model"
	"github.com/nomos/nomos/internal/server"
	"github.com/nomos/nomos/internal/storage"
	versionpkg "github.com/nomos/nomos/internal/version"
	"github.com/spf13/cobra"
)

func Execute() { _ = newRoot().Execute() }
func newRoot() *cobra.Command {
	root := &cobra.Command{Use: "nomos", Short: "Nomos Cosmos CLI", Long: "Nomos verwaltet lokale Cosmos Repositories."}
	root.AddCommand(versionCmd(), cosmosCmd(), domainCmd(), serviceCmd(), blueprintCmd(), instanceCmd(), namespaceCmd(), validateCmd(), graphCmd(), verifyCmd(), serveCmd(), servicegraphCmd())
	return root
}
func versionCmd() *cobra.Command {
	var format string
	var short bool

	c := &cobra.Command{Use: "version", RunE: func(cmd *cobra.Command, args []string) error {
		info := versionpkg.Get()

		// --short intentionally overrides --format for easy scripting.
		if short {
			_, err := fmt.Fprintln(cmd.OutOrStdout(), info.Version)
			return err
		}

		switch format {
		case "", "text":
			_, err := fmt.Fprintf(cmd.OutOrStdout(), "nomos version %s\ncommit:  %s\ndate:    %s\ndirty:   %s\nbuiltBy: %s\ngo:      %s\nos/arch: %s\n", info.Version, info.Commit, info.Date, info.Dirty, info.BuiltBy, info.Go, info.OSArch())
			return err
		case "json":
			enc := json.NewEncoder(cmd.OutOrStdout())
			return enc.Encode(info)
		default:
			return fmt.Errorf("unsupported format %q (supported: text, json)", format)
		}
	}}
	c.Flags().StringVar(&format, "format", "text", "Output format: text or json")
	c.Flags().BoolVar(&short, "short", false, "Print only the version")
	return c
}

func appendGitignore(p string) error {
	path := filepath.Join(p, ".gitignore")
	block := "# Nomos local generated data\n.nomos/cache/\n.nomos/index/\n"
	data, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	content := string(data)
	changed := false
	for _, line := range []string{"# Nomos local generated data", ".nomos/cache/", ".nomos/index/"} {
		if !strings.Contains(content, line) {
			changed = true
			break
		}
	}
	if !changed {
		return nil
	}
	if content != "" && !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	if content != "" {
		content += "\n"
	}
	content += block
	return os.WriteFile(path, []byte(content), 0o644)
}

func cosmosCmd() *cobra.Command {
	c := &cobra.Command{Use: "cosmos"}
	var force, git bool
	init := &cobra.Command{Use: "init <path>", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		p := args[0]
		if st, err := os.Stat(p); err == nil && st.IsDir() {
			entries, _ := os.ReadDir(p)
			if len(entries) > 0 && !force {
				return fmt.Errorf("Zielpfad ist nicht leer")
			}
		}
		for _, d := range []string{storage.DomainsDir(p), storage.CatalogDir(p), storage.EvidenceDir(p), storage.IndexDir(p), storage.CacheDir(p)} {
			if err := os.MkdirAll(d, 0o755); err != nil {
				return err
			}
		}
		co := model.Cosmos{ID: "cosmos-local", Type: "cosmos", Name: "Local Cosmos", Version: "0.1.0", Status: "draft", Owner: "unknown", Summary: "Lokaler Nomos Cosmos.", Domains: []string{}}
		if err := fsx.WriteYAML(storage.CosmosFile(p), co); err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(p, "README.md"), []byte("# Cosmos\n"), 0o644); err != nil {
			return err
		}
		if err := appendGitignore(p); err != nil {
			return err
		}
		if git {
			if err := os.WriteFile(filepath.Join(p, ".gitkeep"), []byte{}, 0o644); err != nil {
				return err
			}
			g := exec.Command("git", "init")
			g.Dir = p
			out, err := g.CombinedOutput()
			if err != nil {
				return fmt.Errorf("git init fehlgeschlagen: %w: %s", err, strings.TrimSpace(string(out)))
			}
		}
		fmt.Fprintln(cmd.OutOrStdout(), "Cosmos wurde erstellt:", p)
		return nil
	}}
	init.Flags().BoolVar(&force, "force", false, "")
	init.Flags().BoolVar(&git, "git", false, "")
	c.AddCommand(init)
	info := &cobra.Command{Use: "info", RunE: func(cmd *cobra.Command, args []string) error {
		p, _ := cmd.Flags().GetString("path")
		outFmt, _ := cmd.Flags().GetString("format")
		warnLegacyLayout(cmd, p)
		co, err := app.GetCosmos(p)
		if err != nil {
			return writeCLIError(cmd, outFmt, err)
		}
		if err := validateFormat(outFmt); err != nil {
			return writeCLIError(cmd, outFmt, err)
		}
		if outFmt == "json" {
			return json.NewEncoder(cmd.OutOrStdout()).Encode(co)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "id=%s name=%s version=%s status=%s owner=%s domains=%d services=%d\n", co.ID, co.Name, co.Version, co.Status, co.Owner, co.DomainCount, co.ServiceCount)
		return nil
	}}
	info.Flags().String("path", ".", "")
	info.Flags().String("format", "text", "Output format: text or json")
	c.AddCommand(info)
	doc := &cobra.Command{Use: "doctor", RunE: func(cmd *cobra.Command, args []string) error {
		p, _ := cmd.Flags().GetString("path")
		if p == "" {
			p = "."
		}
		doc, err := app.DoctorCosmos(p)
		if err != nil {
			return err
		}
		for _, check := range doc.Checks {
			switch check.Name {
			case ".nomos/cosmos.yaml":
				if check.Status == "ok" {
					fmt.Fprintln(cmd.OutOrStdout(), "OK .nomos/cosmos.yaml gefunden")
				} else {
					fmt.Fprintln(cmd.OutOrStdout(), "ERROR .nomos/cosmos.yaml fehlt")
				}
			case "git repository":
				if check.Status == "ok" {
					fmt.Fprintln(cmd.OutOrStdout(), "OK Git Repository gefunden")
				} else {
					fmt.Fprintln(cmd.OutOrStdout(), "WARNING Git Repository nicht initialisiert")
				}
			}
		}
		if doc.Status == "error" {
			return fmt.Errorf("cosmos doctor failed")
		}
		return nil
	}}
	doc.Flags().String("path", ".", "")
	c.AddCommand(doc)
	return c
}

func domainCmd() *cobra.Command {
	c := &cobra.Command{Use: "domain"}
	var force bool
	add := &cobra.Command{Use: "add <dns>", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		p, _ := cmd.Flags().GetString("path")
		owner, _ := cmd.Flags().GetString("owner")
		_, err := app.AddDomain(p, args[0], owner, force)
		return err
	}}
	add.Flags().String("path", ".", "")
	add.Flags().String("owner", "unknown", "")
	add.Flags().BoolVar(&force, "force", false, "")
	c.AddCommand(add)
	list := &cobra.Command{Use: "list", RunE: func(cmd *cobra.Command, args []string) error {
		p, _ := cmd.Flags().GetString("path")
		outFmt, _ := cmd.Flags().GetString("format")
		if err := validateFormat(outFmt); err != nil {
			return writeCLIError(cmd, outFmt, err)
		}
		warnLegacyLayout(cmd, p)
		domains, err := app.ListDomains(p)
		if err != nil {
			return writeCLIError(cmd, outFmt, err)
		}
		if outFmt == "json" {
			return json.NewEncoder(cmd.OutOrStdout()).Encode(domains)
		}
		for _, d := range domains.Domains {
			fmt.Fprintf(cmd.OutOrStdout(), "%-32s %-28s services=%d\n", d.Namespace.DisplayPath, d.Canonical, d.ServiceCount)
		}
		return nil
	}}
	list.Flags().String("path", ".", "Path to the Cosmos repository")
	list.Flags().String("format", "text", "Output format: text or json")
	get := &cobra.Command{Use: "get <domain>", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		p, _ := cmd.Flags().GetString("path")
		outFmt, _ := cmd.Flags().GetString("format")
		if err := validateFormat(outFmt); err != nil {
			return writeCLIError(cmd, outFmt, err)
		}
		domain, err := app.GetDomain(p, args[0])
		if err != nil {
			return writeCLIError(cmd, outFmt, err)
		}
		if outFmt == "json" {
			return json.NewEncoder(cmd.OutOrStdout()).Encode(domain)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "%s\nCanonical namespace: %s\nTree path: %s\nOwner: %s\nStatus: %s\nServices: %d\n", domain.DisplayName, domain.Canonical, domain.Namespace.DisplayPath, domain.Owner, domain.Status, domain.ServiceCount)
		for _, s := range domain.Services {
			fmt.Fprintf(cmd.OutOrStdout(), "  - %s\n", s.Name)
		}
		return nil
	}}
	get.Flags().String("path", ".", "Path to the Cosmos repository")
	get.Flags().String("format", "text", "Output format: text or json")
	del := &cobra.Command{Use: "delete <dns>", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		p, _ := cmd.Flags().GetString("path")
		if err := app.DeleteDomain(p, args[0]); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Domain deleted: %s\n", args[0])
		return nil
	}}
	del.Flags().String("path", ".", "Path to the Cosmos repository")
	c.AddCommand(list, get, del)
	return c
}
func serviceCmd() *cobra.Command {
	c := &cobra.Command{Use: "service"}
	var force bool
	add := &cobra.Command{Use: "add <name>", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		dom, _ := cmd.Flags().GetString("domain")
		p, _ := cmd.Flags().GetString("path")
		owner, _ := cmd.Flags().GetString("owner")
		_, err := app.AddService(p, dom, args[0], owner, force)
		return err
	}}
	add.Flags().String("domain", "", "")
	add.Flags().String("path", ".", "")
	add.Flags().String("owner", "unknown", "")
	add.Flags().BoolVar(&force, "force", false, "")
	_ = add.MarkFlagRequired("domain")
	list := &cobra.Command{Use: "list", RunE: func(cmd *cobra.Command, args []string) error {
		p, _ := cmd.Flags().GetString("path")
		dom, _ := cmd.Flags().GetString("domain")
		outFmt, _ := cmd.Flags().GetString("format")
		if err := validateFormat(outFmt); err != nil {
			return writeCLIError(cmd, outFmt, err)
		}
		items, err := app.ListServices(p, dom)
		if err != nil {
			return writeCLIError(cmd, outFmt, err)
		}
		if outFmt == "json" {
			return json.NewEncoder(cmd.OutOrStdout()).Encode(items)
		}
		for _, svc := range items.Services {
			fmt.Fprintf(cmd.OutOrStdout(), "%-28s %-30s %-12s %s\n", svc.Name, svc.Domain, svc.Status, svc.Owner)
		}
		return nil
	}}
	list.Flags().String("domain", "", "Canonical domain namespace")
	list.Flags().String("path", ".", "Path to the Cosmos repository")
	list.Flags().String("format", "text", "Output format: text or json")
	_ = list.MarkFlagRequired("domain")
	get := &cobra.Command{Use: "get <name>", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		p, _ := cmd.Flags().GetString("path")
		dom, _ := cmd.Flags().GetString("domain")
		outFmt, _ := cmd.Flags().GetString("format")
		if err := validateFormat(outFmt); err != nil {
			return writeCLIError(cmd, outFmt, err)
		}
		svc, err := app.GetService(p, dom, args[0])
		if err != nil {
			return writeCLIError(cmd, outFmt, err)
		}
		if outFmt == "json" {
			return json.NewEncoder(cmd.OutOrStdout()).Encode(svc)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "%s\nDomain: %s\nOwner: %s\nStatus: %s\nPath: %s\n", svc.Name, svc.Domain, svc.Owner, svc.Status, svc.Path)
		return nil
	}}
	get.Flags().String("domain", "", "Canonical domain namespace")
	get.Flags().String("path", ".", "Path to the Cosmos repository")
	get.Flags().String("format", "text", "Output format: text or json")
	_ = get.MarkFlagRequired("domain")
	del := &cobra.Command{Use: "delete <name>", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		p, _ := cmd.Flags().GetString("path")
		dom, _ := cmd.Flags().GetString("domain")
		if err := app.DeleteService(p, dom, args[0]); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Service deleted: %s/%s\n", dom, args[0])
		return nil
	}}
	del.Flags().String("domain", "", "Canonical domain namespace")
	del.Flags().String("path", ".", "Path to the Cosmos repository")
	_ = del.MarkFlagRequired("domain")
	c.AddCommand(add, list, get, del)
	return c
}

func blueprintCmd() *cobra.Command {
	c := &cobra.Command{Use: "blueprint"}
	list := &cobra.Command{Use: "list", RunE: func(cmd *cobra.Command, args []string) error {
		p, _ := cmd.Flags().GetString("path")
		outFmt, _ := cmd.Flags().GetString("format")
		if err := validateFormat(outFmt); err != nil {
			return writeCLIError(cmd, outFmt, err)
		}
		items, err := app.ListBlueprints(p)
		if err != nil {
			return writeCLIError(cmd, outFmt, err)
		}
		if outFmt == "json" {
			return json.NewEncoder(cmd.OutOrStdout()).Encode(items)
		}
		for _, b := range items.Blueprints {
			fmt.Fprintf(cmd.OutOrStdout(), "%-28s %-18s %-36s %s\n", b.ID, b.Type, b.Name, b.Version)
		}
		return nil
	}}
	list.Flags().String("path", ".", "Path to the Cosmos repository")
	list.Flags().String("format", "text", "Output format: text or json")
	show := &cobra.Command{Use: "show <id>", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		p, _ := cmd.Flags().GetString("path")
		outFmt, _ := cmd.Flags().GetString("format")
		if err := validateFormat(outFmt); err != nil {
			return writeCLIError(cmd, outFmt, err)
		}
		item, err := app.GetBlueprint(p, args[0])
		if err != nil {
			return writeCLIError(cmd, outFmt, err)
		}
		if outFmt == "json" {
			return json.NewEncoder(cmd.OutOrStdout()).Encode(item)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "%s\nType: %s\nName: %s\nVersion: %s\nStatus: %s\nOwner: %s\nPath: %s\n", item.ID, item.Type, item.Name, item.Version, item.Status, item.Owner, item.Path)
		return nil
	}}
	show.Flags().String("path", ".", "Path to the Cosmos repository")
	show.Flags().String("format", "text", "Output format: text or json")

	create := &cobra.Command{Use: "create", RunE: func(cmd *cobra.Command, args []string) error {
		p, _ := cmd.Flags().GetString("path")
		id, _ := cmd.Flags().GetString("id")
		bpType, _ := cmd.Flags().GetString("type")
		name, _ := cmd.Flags().GetString("name")
		version, _ := cmd.Flags().GetString("version")
		status, _ := cmd.Flags().GetString("status")
		owner, _ := cmd.Flags().GetString("owner")
		summary, _ := cmd.Flags().GetString("summary")
		serviceRef, _ := cmd.Flags().GetString("service-ref")
		serviceBlueprintRef, _ := cmd.Flags().GetString("service-blueprint-ref")

		if id == "" || name == "" {
			return fmt.Errorf("--id and --name are required")
		}

		bp := model.Blueprint{
			ID:      id,
			Type:    bpType,
			Name:    name,
			Version: version,
			Status:  status,
			Owner:   owner,
			Summary: summary,
		}
		if serviceRef != "" && serviceBlueprintRef != "" {
			bp.RequiredServices = []model.RequiredServiceRef{
				{ServiceRef: serviceRef, ServiceBlueprintRef: serviceBlueprintRef, Required: true},
			}
			bp.RequiredServiceBlueprints = []string{serviceBlueprintRef}
		}

		if err := app.CreateBlueprint(p, bp); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Blueprint created: %s\n", id)
		return nil
	}}
	create.Flags().String("path", ".", "Path to the Cosmos repository")
	create.Flags().String("id", "", "Blueprint ID (required)")
	create.Flags().String("type", "product_blueprint", "Blueprint type: product_blueprint or service_blueprint")
	create.Flags().String("name", "", "Blueprint name (required)")
	create.Flags().String("version", "0.1.0", "Version")
	create.Flags().String("status", "draft", "Status")
	create.Flags().String("owner", "", "Owner")
	create.Flags().String("summary", "", "Summary")
	create.Flags().String("service-ref", "", "Namespace service reference")
	create.Flags().String("service-blueprint-ref", "", "Service blueprint reference")

	delete := &cobra.Command{Use: "delete <id>", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		p, _ := cmd.Flags().GetString("path")
		if err := app.DeleteBlueprint(p, args[0]); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Blueprint deleted: %s\n", args[0])
		return nil
	}}
	delete.Flags().String("path", ".", "Path to the Cosmos repository")

	c.AddCommand(list, show, create, delete)
	return c
}

func instanceCmd() *cobra.Command {
	c := &cobra.Command{Use: "instance"}
	list := &cobra.Command{Use: "list", RunE: func(cmd *cobra.Command, args []string) error {
		p, _ := cmd.Flags().GetString("path")
		outFmt, _ := cmd.Flags().GetString("format")
		if err := validateFormat(outFmt); err != nil {
			return writeCLIError(cmd, outFmt, err)
		}
		items, err := app.ListInstances(p)
		if err != nil {
			return writeCLIError(cmd, outFmt, err)
		}
		if outFmt == "json" {
			return json.NewEncoder(cmd.OutOrStdout()).Encode(items)
		}
		for _, i := range items.Instances {
			fmt.Fprintf(cmd.OutOrStdout(), "%-32s %-18s %-28s %s\n", i.ID, i.Type, i.BlueprintRef, i.ComplianceStatus)
		}
		return nil
	}}
	list.Flags().String("path", ".", "Path to the Cosmos repository")
	list.Flags().String("format", "text", "Output format: text or json")
	show := &cobra.Command{Use: "show <id>", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		p, _ := cmd.Flags().GetString("path")
		outFmt, _ := cmd.Flags().GetString("format")
		if err := validateFormat(outFmt); err != nil {
			return writeCLIError(cmd, outFmt, err)
		}
		item, err := app.GetInstance(p, args[0])
		if err != nil {
			return writeCLIError(cmd, outFmt, err)
		}
		if outFmt == "json" {
			return json.NewEncoder(cmd.OutOrStdout()).Encode(item)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "%s\nType: %s\nBlueprint: %s@%s\nStatus: %s\nCompliance: %s\nPath: %s\n", item.ID, item.Type, item.BlueprintRef, item.BlueprintVersion, item.Status, item.ComplianceStatus, item.Path)
		return nil
	}}
	show.Flags().String("path", ".", "Path to the Cosmos repository")
	show.Flags().String("format", "text", "Output format: text or json")
	c.AddCommand(list, show)
	return c
}

func validateCmd() *cobra.Command {
	c := &cobra.Command{Use: "validate", RunE: func(cmd *cobra.Command, args []string) error {
		p, _ := cmd.Flags().GetString("path")
		if p == "" {
			p = "."
		}
		outFmt, _ := cmd.Flags().GetString("format")
		if err := validateFormat(outFmt); err != nil {
			return writeCLIError(cmd, outFmt, err)
		}
		warnLegacyLayout(cmd, p)
		res, err := app.ValidateCosmos(p)
		if err != nil {
			return writeCLIError(cmd, outFmt, err)
		}
		if outFmt == "json" {
			return json.NewEncoder(cmd.OutOrStdout()).Encode(res)
		} else {
			fmt.Fprintln(cmd.OutOrStdout(), "Nomos Validierung")
			for _, f := range res.Findings {
				fmt.Fprintf(cmd.OutOrStdout(), "%s %s\n", strings.ToUpper(f.Severity), f.Message)
			}
		}
		if len(res.Findings) > 0 {
			return fmt.Errorf("validation failed with %d finding(s)", len(res.Findings))
		}
		return nil
	}}
	c.Flags().String("path", ".", "Path to the Cosmos repository")
	c.Flags().String("format", "text", "Output format: text or json")
	return c
}
func graphCmd() *cobra.Command {
	c := &cobra.Command{Use: "graph", RunE: func(cmd *cobra.Command, args []string) error {
		p, _ := cmd.Flags().GetString("path")
		outFmt, _ := cmd.Flags().GetString("format")
		if err := validateFormat(outFmt); err != nil {
			return writeCLIError(cmd, outFmt, err)
		}
		g, err := app.BuildGraph(p)
		if err != nil {
			return writeCLIError(cmd, outFmt, err)
		}
		if outFmt == "json" {
			return json.NewEncoder(cmd.OutOrStdout()).Encode(g)
		}
		fmt.Fprint(cmd.OutOrStdout(), g.Content)
		return nil
	}}
	c.Flags().String("path", ".", "Path to the Cosmos repository")
	c.Flags().String("format", "text", "Output format: text or json")
	return c
}

func namespaceCmd() *cobra.Command {
	c := &cobra.Command{Use: "namespace"}
	tree := &cobra.Command{Use: "tree", RunE: func(cmd *cobra.Command, args []string) error {
		p, _ := cmd.Flags().GetString("path")
		outFmt, _ := cmd.Flags().GetString("format")
		if err := validateFormat(outFmt); err != nil {
			return writeCLIError(cmd, outFmt, err)
		}
		ns, err := app.BuildNamespaceTree(p)
		if err != nil {
			return writeCLIError(cmd, outFmt, err)
		}
		if outFmt == "json" {
			return json.NewEncoder(cmd.OutOrStdout()).Encode(ns)
		}
		printNamespaceNode(cmd, ns.Root, "")
		return nil
	}}
	tree.Flags().String("path", ".", "Path to the Cosmos repository")
	tree.Flags().String("format", "text", "Output format: text or json")
	c.AddCommand(tree)
	return c
}

func printNamespaceNode(cmd *cobra.Command, node app.NamespaceTreeNodeDTO, indent string) {
	fmt.Fprintf(cmd.OutOrStdout(), "%s%s\n", indent, node.Label)
	for _, child := range node.Children {
		printNamespaceNode(cmd, child, indent+"  ")
	}
}

func warnLegacyLayout(cmd *cobra.Command, p string) {
	if storage.LegacyLayoutDetected(p) {
		fmt.Fprintf(cmd.ErrOrStderr(), "WARNING legacy Nomos layout detected. Please run `nomos cosmos migrate --path %s` to move data into .nomos/.\n", p)
	}
}

func validateFormat(format string) error {
	if format == "" || format == "text" || format == "json" {
		return nil
	}
	return app.Error(app.CodeInvalidFormat, "Invalid format: "+format, 0, nil)
}

func writeCLIError(cmd *cobra.Command, format string, err error) error {
	if format == "json" {
		_ = json.NewEncoder(cmd.ErrOrStderr()).Encode(app.ErrorResponse(err))
	}
	return err
}

type cosmosTree struct {
	Cosmos  model.Cosmos
	Domains []domainNode
}
type domainNode struct {
	Name     string
	Services []serviceNode
}
type serviceNode struct{ Name string }

func loadCosmosTree(p string) (cosmosTree, error) {
	var co model.Cosmos
	cosmosFile, _ := storage.CosmosFileForRead(p)
	if err := fsx.ReadYAML(cosmosFile, &co); err != nil {
		return cosmosTree{}, err
	}
	domains, err := scanDomains(p)
	if err != nil {
		return cosmosTree{}, err
	}
	return cosmosTree{Cosmos: co, Domains: domains}, nil
}

func scanDomains(p string) ([]domainNode, error) {
	var domains []domainNode
	ents, err := os.ReadDir(storage.DomainsDirForRead(p))
	if err != nil {
		if os.IsNotExist(err) {
			return domains, nil
		}
		return nil, err
	}
	for _, e := range ents {
		if !e.IsDir() {
			continue
		}
		domainDir := filepath.Join(storage.DomainsDirForRead(p), e.Name())
		domainYAML := filepath.Join(domainDir, "domain.yaml")
		if _, err := os.Stat(domainYAML); err != nil {
			continue
		}
		var d model.Domain
		if err := fsx.ReadYAML(domainYAML, &d); err != nil {
			return nil, err
		}
		name := firstNonEmpty(d.Name, d.DNSName, e.Name())
		services, err := scanServices(domainDir)
		if err != nil {
			return nil, err
		}
		domains = append(domains, domainNode{Name: name, Services: services})
	}
	sort.Slice(domains, func(i, j int) bool { return domains[i].Name < domains[j].Name })
	return domains, nil
}
func scanServices(domainDir string) ([]serviceNode, error) {
	var services []serviceNode
	ents, err := os.ReadDir(filepath.Join(domainDir, "services"))
	if err != nil {
		if os.IsNotExist(err) {
			return services, nil
		}
		return nil, err
	}
	for _, e := range ents {
		if !e.IsDir() {
			continue
		}
		serviceYAML := filepath.Join(domainDir, "services", e.Name(), "service.yaml")
		if _, err := os.Stat(serviceYAML); err != nil {
			continue
		}
		var s model.Service
		if err := fsx.ReadYAML(serviceYAML, &s); err != nil {
			return nil, err
		}
		services = append(services, serviceNode{Name: firstNonEmpty(s.Name, e.Name())})
	}
	sort.Slice(services, func(i, j int) bool { return services[i].Name < services[j].Name })
	return services, nil
}
func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

var nonAlphaNum = regexp.MustCompile(`[^a-z0-9]+`)

func mermaidID(parts ...string) string {
	base := strings.ToLower(strings.Join(parts, "_"))
	base = nonAlphaNum.ReplaceAllString(base, "_")
	base = strings.Trim(base, "_")
	if base == "" {
		base = "node"
	}
	if r := rune(base[0]); unicode.IsDigit(r) {
		base = "n_" + base
	}
	return base
}
func mermaidLabel(label string) string {
	label = strings.ReplaceAll(label, "\"", "\\\"")
	label = strings.ReplaceAll(label, "\n", " ")
	label = strings.ReplaceAll(label, "\r", " ")
	return label
}
func verifyCmd() *cobra.Command {
	v := &cobra.Command{Use: "verify"}
	d := &cobra.Command{Use: "domain <dns>", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		dns := args[0]
		rec := "_nomos." + dns
		txt, err := net.DefaultResolver.LookupTXT(cmd.Context(), rec)
		p, _ := cmd.Flags().GetString("path")
		os.MkdirAll(storage.EvidenceDir(p), 0o755)
		status := "failed"
		exp := "nomos-domain=" + dns
		for _, t := range txt {
			if strings.Contains(t, exp) {
				status = "verified"
			}
		}
		ev := fmt.Sprintf("id: evidence-%s\ntype: evidence\nevidence_type: dns_verification\ndomain: %s\nrecord: %s\nstatus: %s\ntimestamp: \"%s\"\n", time.Now().UTC().Format("20060102-150405"), dns, rec, status, time.Now().UTC().Format(time.RFC3339))
		_ = os.WriteFile(filepath.Join(storage.EvidenceDir(p), strings.ReplaceAll(dns, ".", "-")+"-dns.yaml"), []byte(ev), 0o644)
		if err != nil || status != "verified" {
			return fmt.Errorf("DNS Verifikation fehlgeschlagen")
		}
		fmt.Fprintln(cmd.OutOrStdout(), "DNS Verifikation erfolgreich")
		return nil
	}}
	d.Flags().String("path", ".", "")
	v.AddCommand(d)
	return v
}
func newServeMux(cosmosPath string) http.Handler { return server.NewHandler(cosmosPath) }
func serveCmd() *cobra.Command {
	c := &cobra.Command{Use: "serve", RunE: func(cmd *cobra.Command, args []string) error {
		listen, _ := cmd.Flags().GetString("listen")
		p, _ := cmd.Flags().GetString("path")
		fmt.Fprintf(cmd.OutOrStdout(), "Nomos server listening on http://%s\nCosmos path: %s\n", listen, p)
		return http.ListenAndServe(listen, newServeMux(p))
	}}
	c.Flags().String("listen", "127.0.0.1:8080", "")
	c.Flags().String("path", ".", "Path to the Cosmos repository")
	return c
}
