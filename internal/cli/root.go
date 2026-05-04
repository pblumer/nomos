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

	"github.com/nomos/nomos/internal/cosmosfs"
	"github.com/nomos/nomos/internal/fsx"
	"github.com/nomos/nomos/internal/graph"
	"github.com/nomos/nomos/internal/model"
	"github.com/nomos/nomos/internal/server"
	"github.com/nomos/nomos/internal/validate"
	versionpkg "github.com/nomos/nomos/internal/version"
	"github.com/spf13/cobra"
)

func Execute() { _ = newRoot().Execute() }
func newRoot() *cobra.Command {
	root := &cobra.Command{Use: "nomos", Short: "Nomos Cosmos CLI", Long: "Nomos verwaltet lokale Cosmos Repositories."}
	root.AddCommand(versionCmd(), cosmosCmd(), domainCmd(), serviceCmd(), validateCmd(), graphCmd(), verifyCmd(), serveCmd())
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
		for _, d := range []string{"domains", ".nomos/cache", ".nomos/index", ".nomos/evidence"} {
			if err := os.MkdirAll(filepath.Join(p, d), 0o755); err != nil {
				return err
			}
		}
		co := model.Cosmos{ID: "cosmos-local", Type: "cosmos", Name: "Local Cosmos", Version: "0.1.0", Status: "draft", Owner: "unknown", Summary: "Lokaler Nomos Cosmos.", Domains: []string{}}
		if err := fsx.WriteYAML(filepath.Join(p, "cosmos.yaml"), co); err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(p, "README.md"), []byte("# Cosmos\n"), 0o644); err != nil {
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
		if p == "" {
			p = "."
		}
		var co model.Cosmos
		if err := fsx.ReadYAML(filepath.Join(p, "cosmos.yaml"), &co); err != nil {
			return err
		}
		domains, err := scanDomains(p)
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "id=%s name=%s version=%s status=%s owner=%s domains=%d\n", co.ID, co.Name, co.Version, co.Status, co.Owner, len(domains))
		return nil
	}}
	info.Flags().String("path", ".", "")
	c.AddCommand(info)
	doc := &cobra.Command{Use: "doctor", RunE: func(cmd *cobra.Command, args []string) error {
		p, _ := cmd.Flags().GetString("path")
		if p == "" {
			p = "."
		}
		if _, e := os.Stat(filepath.Join(p, "cosmos.yaml")); e != nil {
			fmt.Fprintln(cmd.OutOrStdout(), "ERROR cosmos.yaml fehlt")
			return fmt.Errorf("cosmos.yaml fehlt")
		}
		fmt.Fprintln(cmd.OutOrStdout(), "OK cosmos.yaml gefunden")
		if _, e := os.Stat(filepath.Join(p, ".git")); e != nil {
			fmt.Fprintln(cmd.OutOrStdout(), "WARNING Git Repository nicht initialisiert")
		} else {
			fmt.Fprintln(cmd.OutOrStdout(), "OK Git Repository gefunden")
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
		dns := args[0]
		if !strings.Contains(dns, ".") {
			return fmt.Errorf("ungueltiger DNS Name")
		}
		p, _ := cmd.Flags().GetString("path")
		owner, _ := cmd.Flags().GetString("owner")
		ddir := filepath.Join(p, "domains", dns)
		if _, e := os.Stat(ddir); e == nil && !force {
			return fmt.Errorf("Domain existiert bereits")
		}
		if err := os.MkdirAll(filepath.Join(ddir, "services"), 0o755); err != nil {
			return err
		}
		d := model.Domain{ID: "domain-" + strings.ReplaceAll(dns, ".", "-"), Type: "domain", Name: dns, Version: "0.1.0", Status: "draft", Owner: owner, DNSName: dns, Summary: "Nomos Domaene " + dns + "."}
		if err := fsx.WriteYAML(filepath.Join(ddir, "domain.yaml"), d); err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(ddir, "README.md"), []byte("# Domain\n"), 0o644); err != nil {
			return err
		}
		return nil
	}}
	add.Flags().String("path", ".", "")
	add.Flags().String("owner", "unknown", "")
	add.Flags().BoolVar(&force, "force", false, "")
	c.AddCommand(add)
	list := &cobra.Command{Use: "list", RunE: func(cmd *cobra.Command, args []string) error {
		p, _ := cmd.Flags().GetString("path")
		ents, _ := os.ReadDir(filepath.Join(p, "domains"))
		for _, e := range ents {
			if e.IsDir() {
				fmt.Fprintln(cmd.OutOrStdout(), e.Name())
			}
		}
		return nil
	}}
	list.Flags().String("path", ".", "Path to the Cosmos repository")
	c.AddCommand(list)
	return c
}
func serviceCmd() *cobra.Command {
	c := &cobra.Command{Use: "service"}
	var force bool
	add := &cobra.Command{Use: "add <name>", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		dom, _ := cmd.Flags().GetString("domain")
		p, _ := cmd.Flags().GetString("path")
		owner, _ := cmd.Flags().GetString("owner")
		sdir := filepath.Join(p, "domains", dom, "services", name)
		if _, e := os.Stat(sdir); e == nil && !force {
			return fmt.Errorf("Service existiert bereits")
		}
		for _, d := range []string{"capabilities", "requirements", "rules", "processes", "skills", "findings", "evidence"} {
			if err := os.MkdirAll(filepath.Join(sdir, d), 0o755); err != nil {
				return err
			}
		}
		s := model.Service{ID: "service-" + name, Type: "service", Name: name, Version: "0.1.0", Status: "draft", Owner: owner, Summary: "Nomos Service " + name + "."}
		if err := fsx.WriteYAML(filepath.Join(sdir, "service.yaml"), s); err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(sdir, "README.md"), []byte("# Service\n"), 0o644); err != nil {
			return err
		}
		return nil
	}}
	add.Flags().String("domain", "", "")
	add.Flags().String("path", ".", "")
	add.Flags().String("owner", "unknown", "")
	add.Flags().BoolVar(&force, "force", false, "")
	_ = add.MarkFlagRequired("domain")
	c.AddCommand(add)
	return c
}

func validateCmd() *cobra.Command {
	c := &cobra.Command{Use: "validate", RunE: func(cmd *cobra.Command, args []string) error {
		p, _ := cmd.Flags().GetString("path")
		if p == "" {
			p = "."
		}
		res, _ := validate.Validate(p)
		outFmt, _ := cmd.Flags().GetString("format")
		if outFmt != "text" && outFmt != "json" {
			return fmt.Errorf("ungueltiges format: %s", outFmt)
		}
		if outFmt == "json" {
			b, _ := json.MarshalIndent(res, "", "  ")
			fmt.Fprintln(cmd.OutOrStdout(), string(b))
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
		tree, err := cosmosfs.LoadTree(p)
		if err != nil {
			return err
		}
		fmt.Fprint(cmd.OutOrStdout(), graph.Mermaid(tree))
		return nil
	}}
	c.Flags().String("path", ".", "Path to the Cosmos repository")
	return c
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
	if err := fsx.ReadYAML(filepath.Join(p, "cosmos.yaml"), &co); err != nil {
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
	ents, err := os.ReadDir(filepath.Join(p, "domains"))
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
		domainDir := filepath.Join(p, "domains", e.Name())
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
		os.MkdirAll(filepath.Join(p, ".nomos/evidence"), 0o755)
		status := "failed"
		exp := "nomos-domain=" + dns
		for _, t := range txt {
			if strings.Contains(t, exp) {
				status = "verified"
			}
		}
		ev := fmt.Sprintf("id: evidence-%s\ntype: evidence\nevidence_type: dns_verification\ndomain: %s\nrecord: %s\nstatus: %s\ntimestamp: \"%s\"\n", time.Now().UTC().Format("20060102-150405"), dns, rec, status, time.Now().UTC().Format(time.RFC3339))
		_ = os.WriteFile(filepath.Join(p, ".nomos/evidence", strings.ReplaceAll(dns, ".", "-")+"-dns.yaml"), []byte(ev), 0o644)
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
