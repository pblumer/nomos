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
	NamespaceTree      namespaceTreeView
	SelectedKind       string
	SelectedDomain     *domainDetailView
	SelectedService    *serviceDetailView
	SelectedNamespace  *namespaceDetailView
	SelectedCosmos     *cosmosSummaryView
	SelectedDisplayKey string
	Form               domainFormState
}

type domainFormState struct {
	Mode, Error, Parent, Segment, Owner, Canonical, ServiceName string
	Force                                                       bool
}

type namespaceTreeView struct{ Root namespaceNodeView }
type namespaceNodeView struct {
	Label, Kind, Canonical, DisplayPath, TreePath string
	Domain                                        *app.DomainDTO
	Service                                       *app.ServiceDTO
	Children                                      []namespaceNodeView
	Selected                                      bool
}

type cosmosSummaryView struct {
	Name, ID, Version, Status, Owner, Path, Link string
	DomainCount, ServiceCount                    int
	Selected                                     bool
}
type namespaceDetailView struct{ Label, Canonical, DisplayPath, TreePath string }
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
	appTree, err := app.BuildNamespaceTree(path)
	if err != nil {
		return domainTreePageData{}, err
	}
	kind, domainName, serviceName := parseSelection(selected)
	view := domainTreePageData{Title: "Domains Explorer", Subtitle: "Browse domains and services like a repository tree", NamespaceTree: namespaceTreeView{Root: namespaceNodeFromDTO(appTree.Root, "", kind, domainName, serviceName)}}
	view.Cosmos = cosmosSummaryView{Name: fallback(cosmos.Name, "Local Cosmos"), ID: fallback(cosmos.ID, "n/a"), Version: fallback(cosmos.Version, "n/a"), Status: fallback(cosmos.Status, "unknown"), Owner: fallback(cosmos.Owner, "unknown"), Path: cosmos.Path, DomainCount: cosmos.DomainCount, ServiceCount: cosmos.ServiceCount, Link: "/domains?selected=cosmos"}
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
	if kind == "namespace" {
		if ns := findNamespace(&view.NamespaceTree.Root, domainName); ns != nil {
			view.SelectedNamespace = ns
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
	return view, nil
}

func namespaceNodeFromDTO(n app.NamespaceTreeNodeDTO, parentTreePath, selectedKind, selectedDomain, selectedService string) namespaceNodeView {
	v := namespaceNodeView{Label: n.Label, Kind: n.Kind, Canonical: n.Canonical, DisplayPath: n.DisplayPath, TreePath: n.TreePath, Domain: n.Domain, Service: n.Service}
	if n.Kind == "namespace" {
		v.TreePath = joinTreePath(parentTreePath, n.Label)
		v.DisplayPath = strings.ReplaceAll(v.TreePath, "/", " / ")
		v.Canonical = canonicalFromTreePath(v.TreePath)
	}
	switch selectedKind {
	case "domain":
		v.Selected = v.Kind == "domain" && v.Canonical == selectedDomain
	case "service":
		v.Selected = v.Kind == "service" && v.Canonical == selectedDomain+"/"+selectedService
	case "namespace":
		v.Selected = v.Kind == "namespace" && v.TreePath == selectedDomain
	}
	childParent := v.TreePath
	for _, child := range n.Children {
		v.Children = append(v.Children, namespaceNodeFromDTO(child, childParent, selectedKind, selectedDomain, selectedService))
	}
	return v
}

func joinTreePath(parent, label string) string {
	if parent == "" {
		return label
	}
	return parent + "/" + label
}
func canonicalFromTreePath(treePath string) string {
	parts := strings.Split(treePath, "/")
	for i, j := 0, len(parts)-1; i < j; i, j = i+1, j-1 {
		parts[i], parts[j] = parts[j], parts[i]
	}
	return strings.Join(parts, ".")
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

func findNamespace(n *namespaceNodeView, treePath string) *namespaceDetailView {
	for i := range n.Children {
		c := &n.Children[i]
		if c.Kind == "namespace" && c.TreePath == treePath {
			return &namespaceDetailView{Label: c.Label, Canonical: c.Canonical, DisplayPath: c.DisplayPath, TreePath: c.TreePath}
		}
		if found := findNamespace(c, treePath); found != nil {
			return found
		}
	}
	return nil
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
	if strings.HasPrefix(value, "namespace:") {
		n := strings.Trim(strings.TrimPrefix(value, "namespace:"), "/")
		if n != "" && filepath.Clean(n) == n {
			return "namespace", n, ""
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
