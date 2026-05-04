package server

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/nomos/nomos/internal/cosmosfs"
)

type domainTreePageData struct {
	Title              string
	Subtitle           string
	Cosmos             cosmosSummaryView
	Domains            []domainTreeNodeView
	SelectedKind       string
	SelectedDomain     *domainDetailView
	SelectedService    *serviceDetailView
	SelectedCosmos     *cosmosSummaryView
	SelectedDisplayKey string
}

type cosmosSummaryView struct {
	Name         string
	ID           string
	Version      string
	Status       string
	Owner        string
	Path         string
	DomainCount  int
	ServiceCount int
	Link         string
	Selected     bool
}

type domainTreeNodeView struct {
	Name         string
	Owner        string
	Status       string
	Path         string
	ServiceCount int
	Services     []serviceTreeNodeView
	Link         string
	Selected     bool
	Expanded     bool
}

type serviceTreeNodeView struct {
	Name     string
	Domain   string
	Owner    string
	Status   string
	Path     string
	Link     string
	Selected bool
}

type domainDetailView struct {
	Name         string
	Owner        string
	Status       string
	Path         string
	ServiceCount int
	Services     []serviceTreeNodeView
	PageLink     string
}

type serviceDetailView struct {
	Name     string
	Domain   string
	Owner    string
	Status   string
	Path     string
	PageLink string
}

func buildDomainsExplorer(tree cosmosfs.Tree, selected string) domainTreePageData {
	view := domainTreePageData{Title: "Domains Explorer", Subtitle: "Browse domains and services like a repository tree"}
	serviceTotal := 0
	for _, d := range tree.Domains {
		serviceTotal += len(d.Services)
	}
	view.Cosmos = cosmosSummaryView{Name: fallback(tree.Cosmos.Name, "Local Cosmos"), ID: fallback(tree.Cosmos.ID, "n/a"), Version: fallback(tree.Cosmos.Version, "n/a"), Status: fallback(tree.Cosmos.Status, "unknown"), Owner: fallback(tree.Cosmos.Owner, "unknown"), Path: tree.Path, DomainCount: len(tree.Domains), ServiceCount: serviceTotal, Link: "/domains?selected=cosmos"}

	kind, domainName, serviceName := parseSelection(selected)
	knownSelection := false
	for _, d := range tree.Domains {
		dv := domainTreeNodeView{Name: d.Name, Owner: fallback(d.Metadata.Owner, "unknown"), Status: fallback(d.Metadata.Status, "unknown"), Path: d.Path, ServiceCount: len(d.Services), Link: "/domains?selected=domain:" + d.Name, Expanded: true}
		if kind == "domain" && domainName == d.Name {
			dv.Selected = true
			knownSelection = true
		}
		for _, s := range d.Services {
			sv := serviceTreeNodeView{Name: s.Name, Domain: d.Name, Owner: fallback(s.Metadata.Owner, "unknown"), Status: fallback(s.Metadata.Status, "unknown"), Path: s.Path, Link: "/domains?selected=service:" + d.Name + "/" + s.Name}
			if kind == "service" && domainName == d.Name && serviceName == s.Name {
				sv.Selected = true
				dv.Selected = true
				knownSelection = true
			}
			dv.Services = append(dv.Services, sv)
		}
		view.Domains = append(view.Domains, dv)
	}
	if kind == "cosmos" || !knownSelection {
		view.Cosmos.Selected = true
		kind = "cosmos"
	}
	view.SelectedKind = kind
	view.SelectedDisplayKey = selected
	if view.Cosmos.Selected {
		view.SelectedCosmos = &view.Cosmos
		return view
	}
	for _, d := range view.Domains {
		if kind == "domain" && d.Name == domainName {
			dd := domainDetailView{Name: d.Name, Owner: d.Owner, Status: d.Status, Path: d.Path, ServiceCount: d.ServiceCount, Services: d.Services, PageLink: "/domains/" + d.Name}
			view.SelectedDomain = &dd
			return view
		}
		if kind == "service" && d.Name == domainName {
			for _, s := range d.Services {
				if s.Name == serviceName {
					sd := serviceDetailView{Name: s.Name, Domain: d.Name, Owner: s.Owner, Status: s.Status, Path: s.Path, PageLink: fmt.Sprintf("/services/%s/%s", d.Name, s.Name)}
					view.SelectedService = &sd
					return view
				}
			}
		}
	}
	view.Cosmos.Selected = true
	view.SelectedCosmos = &view.Cosmos
	return view
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
