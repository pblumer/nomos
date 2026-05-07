package server

import (
	"net/http"

	"github.com/nomos/nomos/internal/app"
)

type domainFormState struct {
	Mode, Error, Parent, Segment, Owner, Canonical, ServiceName, Namespace, Label string
	Force                                                                         bool
}

func (h *handler) createTopLevelDomainPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	_ = r.ParseForm()
	namespaceName := r.FormValue("namespace")
	label := first(r.FormValue("label"), r.FormValue("segment"))
	canonical := first(r.FormValue("canonical"), r.FormValue("dns"), r.FormValue("domain"))
	var err error
	if namespaceName != "" || label != "" {
		var dto app.DomainDTO
		dto, err = app.AddDomainInNamespace(h.cosmosPath, namespaceName, label, r.FormValue("owner"), r.FormValue("force") != "")
		canonical = dto.Canonical
	} else {
		_, err = app.AddDomain(h.cosmosPath, canonical, r.FormValue("owner"), r.FormValue("force") != "")
	}
	if err != nil {
		h.renderDomainsPage(w, r, statusOf(err), domainFormState{Mode: first(r.FormValue("mode"), "top-level"), Error: err.Error(), Canonical: canonical, Namespace: namespaceName, Label: label, Owner: r.FormValue("owner"), Force: r.FormValue("force") != ""})
		return
	}
	http.Redirect(w, r, "/domains?selected=domain:"+canonical, http.StatusSeeOther)
}

func (h *handler) createChildDomainPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	_ = r.ParseForm()
	parent := r.FormValue("parent")
	segment := r.FormValue("segment")
	dto, err := app.AddChildDomain(h.cosmosPath, parent, segment, r.FormValue("owner"), r.FormValue("force") != "")
	if err != nil {
		h.renderDomainsPage(w, r, statusOf(err), domainFormState{Mode: "child", Error: err.Error(), Parent: parent, Segment: segment, Owner: r.FormValue("owner"), Force: r.FormValue("force") != ""})
		return
	}
	http.Redirect(w, r, "/domains?selected=domain:"+dto.Canonical, http.StatusSeeOther)
}

func (h *handler) createServicePage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	_ = r.ParseForm()
	domain := r.FormValue("domain")
	name := r.FormValue("name")
	_, err := app.AddService(h.cosmosPath, domain, name, r.FormValue("owner"), r.FormValue("force") != "")
	if err != nil {
		h.renderDomainsPage(w, r, statusOf(err), domainFormState{Mode: "service", Error: err.Error(), Parent: domain, ServiceName: name, Owner: r.FormValue("owner"), Force: r.FormValue("force") != ""})
		return
	}
	http.Redirect(w, r, "/domains?selected=service:"+domain+"/"+name, http.StatusSeeOther)
}

func (h *handler) renderDomainsPage(w http.ResponseWriter, r *http.Request, status int, form domainFormState) {
	ex, err := buildDomainsExplorer(h.cosmosPath, r.URL.Query().Get("selected"))
	if err != nil {
		h.errorPage(w, r, statusOf(err), "Domains unavailable", err.Error())
		return
	}
	if form.Owner == "" {
		form.Owner = "unknown"
	}
	if status != http.StatusOK {
		w.WriteHeader(status)
	}
	h.page(w, "domains", map[string]any{"ActiveNav": "domains", "PageTitle": "Domains", "Explorer": ex, "DomainForm": form})
}
