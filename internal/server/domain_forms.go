package server

import (
	"net/http"

	"github.com/nomos/nomos/internal/app"
)

func (h *handler) contextualFormPost(w http.ResponseWriter, r *http.Request) bool {
	switch r.URL.Path {
	case "/domains/create", "/domains/create-top-level", "/domains/create-advanced":
		canonical := first(r.FormValue("canonical"), r.FormValue("dns"), r.FormValue("domain"))
		_, err := app.AddDomain(h.cosmosPath, canonical, r.FormValue("owner"), r.FormValue("force") != "")
		if err != nil {
			h.renderDomainsPage(w, r, statusOf(err), domainFormState{Mode: first(r.FormValue("mode"), "top-level"), Error: err.Error(), Canonical: canonical, Owner: r.FormValue("owner"), Force: r.FormValue("force") != ""})
			return true
		}
		http.Redirect(w, r, "/domains?selected=domain:"+canonical, http.StatusSeeOther)
		return true
	case "/domains/create-child":
		parent := r.FormValue("parent")
		segment := r.FormValue("segment")
		dto, err := app.AddChildDomain(h.cosmosPath, parent, segment, r.FormValue("owner"), r.FormValue("force") != "")
		if err != nil {
			h.renderDomainsPage(w, r, statusOf(err), domainFormState{Mode: "child", Error: err.Error(), Parent: parent, Segment: segment, Owner: r.FormValue("owner"), Force: r.FormValue("force") != ""})
			return true
		}
		http.Redirect(w, r, "/domains?selected=domain:"+dto.Canonical, http.StatusSeeOther)
		return true
	case "/services/create":
		domain := r.FormValue("domain")
		name := r.FormValue("name")
		_, err := app.AddService(h.cosmosPath, domain, name, r.FormValue("owner"), r.FormValue("force") != "")
		if err != nil {
			h.renderDomainsPage(w, r, statusOf(err), domainFormState{Mode: "service", Error: err.Error(), Parent: domain, ServiceName: name, Owner: r.FormValue("owner"), Force: r.FormValue("force") != ""})
			return true
		}
		http.Redirect(w, r, "/domains?selected=service:"+domain+"/"+name, http.StatusSeeOther)
		return true
	}
	return false
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
	ex.Form = form
	if status != http.StatusOK {
		w.WriteHeader(status)
	}
	h.page(w, "domains", map[string]any{"ActiveNav": "domains", "PageTitle": "Domains", "Explorer": ex})
}
