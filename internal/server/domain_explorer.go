package server

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/nomos/nomos/internal/app"
)

type domainTreePageData struct {
	Title              string
	Subtitle           string
	Cosmos             cosmosSummaryView
	NamespaceTree      app.NamespaceTreeDTO
	SelectedKind       string
	SelectedDomain     *domainDetailView
	SelectedService    *serviceDetailView
	SelectedCosmos     *cosmosSummaryView
	SelectedDisplayKey string
}

type cosmosSummaryView struct {
	Name, ID, Version, Status, Owner, Path, Link string
	DomainCount, ServiceCount                    int
	Selected                                     bool
}
type domainTreeNodeView struct {
	Name, Owner, Status, Path string
	ServiceCount              int
	Services                  []serviceTreeNodeView
	Link                      string
	Selected, Expanded        bool
}
type serviceTreeNodeView struct {
	Name, Domain, Owner, Status, Path, Link string
	Selected                                bool
}
type domainDetailView struct {
	Name, Canonical, DisplayName, DisplayPath, TreePath, Owner, Status, Path string
	ServiceCount                                                             int
	Services                                                                 []serviceTreeNodeView
	PageLink                                                                 string
}
type serviceDetailView struct{ Name, Domain, Owner, Status, Path, PageLink string }

func buildDomainsExplorer(path, selected string) (domainTreePageData, error) {
	cosmos, err := app.GetCosmos(path)
	if err != nil {
		return domainTreePageData{}, err
	}
	nsTree, err := app.BuildNamespaceTree(path)
	if err != nil {
		return domainTreePageData{}, err
	}
	view := domainTreePageData{Title: "Domains Explorer", Subtitle: "Browse domains and services like a repository tree", NamespaceTree: nsTree}
	view.Cosmos = cosmosSummaryView{Name: fallback(cosmos.Name, "Local Cosmos"), ID: fallback(cosmos.ID, "n/a"), Version: fallback(cosmos.Version, "n/a"), Status: fallback(cosmos.Status, "unknown"), Owner: fallback(cosmos.Owner, "unknown"), Path: cosmos.Path, DomainCount: cosmos.DomainCount, ServiceCount: cosmos.ServiceCount, Link: "/domains?selected=cosmos"}
	kind, domainName, serviceName := parseSelection(selected)
	knownSelection := false
	if kind == "domain" {
		d, err := app.GetDomain(path, domainName)
		if err == nil {
			view.SelectedDomain = domainDetail(d)
			knownSelection = true
		}
	}
	if kind == "service" {
		s, err := app.GetService(path, domainName, serviceName)
		if err == nil {
			sv := serviceDetail(s)
			view.SelectedService = &sv
			knownSelection = true
		}
	}
	if kind == "cosmos" || !knownSelection {
		view.Cosmos.Selected = true
		kind = "cosmos"
		view.SelectedCosmos = &view.Cosmos
	}
	view.SelectedKind = kind
	view.SelectedDisplayKey = selected
	markSelection(&view.NamespaceTree.Root, kind, domainName, serviceName)
	return view, nil
}

func domainDetail(d app.DomainDTO) *domainDetailView {
	dd := &domainDetailView{Name: d.Name, Canonical: d.Canonical, DisplayName: d.DisplayName, DisplayPath: d.Namespace.DisplayPath, TreePath: d.Namespace.TreePath, Owner: d.Owner, Status: d.Status, Path: d.Path, ServiceCount: d.ServiceCount, PageLink: "/domains/" + d.Canonical}
	for _, s := range d.Services {
		dd.Services = append(dd.Services, serviceTreeNodeView{Name: s.Name, Domain: s.Domain, Owner: s.Owner, Status: s.Status, Path: s.Path, Link: "/domains?selected=service:" + s.Domain + "/" + s.Name})
	}
	return dd
}
func serviceDetail(s app.ServiceDTO) serviceDetailView {
	return serviceDetailView{Name: s.Name, Domain: s.Domain, Owner: s.Owner, Status: s.Status, Path: s.Path, PageLink: fmt.Sprintf("/services/%s/%s", s.Domain, s.Name)}
}

func markSelection(n *app.NamespaceTreeNodeDTO, kind, domain, service string) {
	for i := range n.Children {
		markSelection(&n.Children[i], kind, domain, service)
	}
}

func parseSelection(value string) (kind, domain, service string) {
	if value == "" || value == "cosmos" {
		return "cosmos", "", ""
	}
	if strings.HasPrefix(value, "domain:") {
		d := strings.TrimPrefix(value, "domain:")
		if d != "" && !strings.Contains(d, "/") {
			return "domain", d, ""
		}
	}
	if strings.HasPrefix(value, "service:") {
		s := strings.TrimPrefix(value, "service:")
		parts := strings.Split(s, "/")
		if len(parts) == 2 && parts[0] != "" && parts[1] != "" && filepath.Clean(s) == s {
			return "service", parts[0], parts[1]
		}
	}
	return "cosmos", "", ""
}
func fallback(v, d string) string {
	if strings.TrimSpace(v) == "" {
		return d
	}
	return v
}
