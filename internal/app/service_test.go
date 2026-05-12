package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nomos/nomos/internal/storage"
)

func createAppTestCosmos(t *testing.T) string {
	t.Helper()
	p := t.TempDir()
	must := func(err error) {
		if err != nil {
			t.Fatal(err)
		}
	}
	must(os.MkdirAll(filepath.Join(storage.DomainsDir(p), "identity.blumer.cloud/services/user-account"), 0o755))
	must(os.MkdirAll(filepath.Join(storage.DomainsDir(p), "identity.blumer.cloud/services/privileged-account"), 0o755))
	must(os.MkdirAll(filepath.Join(storage.DomainsDir(p), "platform.blumer.cloud/services/rule-validation-api"), 0o755))
	must(os.WriteFile(storage.CosmosFile(p), []byte("id: cosmos-local\nname: Local Cosmos\nversion: 0.1.0\nstatus: draft\nowner: Platform Team\n"), 0o644))
	must(os.WriteFile(filepath.Join(storage.DomainsDir(p), "identity.blumer.cloud", "domain.yaml"), []byte("name: identity.blumer.cloud\nowner: Identity Team\nstatus: draft\n"), 0o644))
	must(os.WriteFile(filepath.Join(storage.DomainsDir(p), "platform.blumer.cloud/domain.yaml"), []byte("name: platform.blumer.cloud\nowner: Platform Team\nstatus: draft\n"), 0o644))
	must(os.WriteFile(filepath.Join(storage.DomainsDir(p), "identity.blumer.cloud/services/user-account/service.yaml"), []byte("name: user-account\nowner: Identity Team\nowned_by: identity.blumer.cloud\nstatus: draft\n"), 0o644))
	must(os.WriteFile(filepath.Join(storage.DomainsDir(p), "identity.blumer.cloud/services/privileged-account/service.yaml"), []byte("name: privileged-account\nowner: Identity Team\nowned_by: identity.blumer.cloud\nstatus: draft\n"), 0o644))
	must(os.WriteFile(filepath.Join(storage.DomainsDir(p), "platform.blumer.cloud/services/rule-validation-api/service.yaml"), []byte("name: rule-validation-api\nowner: Platform Team\nowned_by: platform.blumer.cloud\nstatus: draft\n"), 0o644))
	return p
}

func TestAppUseCasesExposeNamespaceMetadata(t *testing.T) {
	p := createAppTestCosmos(t)
	cosmos, err := GetCosmos(p)
	if err != nil {
		t.Fatal(err)
	}
	if cosmos.DomainCount != 2 || cosmos.ServiceCount != 3 {
		t.Fatalf("unexpected counts: %+v", cosmos)
	}
	domains, err := ListDomains(p)
	if err != nil {
		t.Fatal(err)
	}
	if got := domains.Domains[0]; got.Canonical != "identity.blumer.cloud" || got.Namespace.DisplayPath != "cloud / blumer / identity" || got.DisplayName != "identity" || got.ServiceCount != 2 {
		t.Fatalf("unexpected domain: %+v", got)
	}
	domain, err := GetDomain(p, "identity.blumer.cloud")
	if err != nil {
		t.Fatal(err)
	}
	if len(domain.Services) != 2 || domain.Services[0].Name != "privileged-account" || domain.Services[1].Name != "user-account" {
		t.Fatalf("services not sorted: %+v", domain.Services)
	}
	service, err := GetService(p, "identity.blumer.cloud", "user-account")
	if err != nil {
		t.Fatal(err)
	}
	if service.Domain != "identity.blumer.cloud" || service.Name != "user-account" {
		t.Fatalf("unexpected service: %+v", service)
	}
	graph, err := BuildGraph(p)
	if err != nil {
		t.Fatal(err)
	}
	if graph.Format != "mermaid" || !strings.Contains(graph.Content, "cloud / blumer / identity") {
		t.Fatalf("unexpected graph: %+v", graph)
	}
	validation, err := ValidateCosmos(p)
	if err != nil {
		t.Fatal(err)
	}
	if validation.Status != "ok" || len(validation.Findings) != 0 {
		t.Fatalf("unexpected validation: %+v", validation)
	}
}

func TestBuildNamespaceTree(t *testing.T) {
	tree, err := BuildNamespaceTree(createAppTestCosmos(t))
	if err != nil {
		t.Fatal(err)
	}
	body := strings.Join(flattenLabels(tree.Root), "|")
	for _, want := range []string{"Local Cosmos", "Namespaces", "cloud", "blumer", "identity", "platform", "Services", "user-account", "rule-validation-api"} {
		if !strings.Contains(body, want) {
			t.Fatalf("missing %s in %s", want, body)
		}
	}
}

func findTreeNode(n NamespaceTreeNodeDTO, kind, label string) *NamespaceTreeNodeDTO {
	if n.Kind == kind && n.Label == label {
		return &n
	}
	for _, c := range n.Children {
		if found := findTreeNode(c, kind, label); found != nil {
			return found
		}
	}
	return nil
}

func TestBuildNamespaceTreeIncludesProductFulfillmentChildren(t *testing.T) {
	p := createAppTestCosmos(t)
	must := func(err error) {
		if err != nil {
			t.Fatal(err)
		}
	}
	must(os.MkdirAll(filepath.Join(storage.DomainsDir(p), "blumer.cloud"), 0o755))
	must(os.WriteFile(filepath.Join(storage.DomainsDir(p), "blumer.cloud", "domain.yaml"), []byte("name: blumer.cloud\nowner: Cloud Team\nstatus: draft\n"), 0o644))
	must(os.MkdirAll(filepath.Join(storage.DomainsDir(p), "mailing.blumer.cloud", "services", "exchange"), 0o755))
	must(os.WriteFile(filepath.Join(storage.DomainsDir(p), "mailing.blumer.cloud", "domain.yaml"), []byte("name: mailing.blumer.cloud\nowner: Mailing Team\nstatus: draft\n"), 0o644))
	must(os.WriteFile(filepath.Join(storage.DomainsDir(p), "mailing.blumer.cloud", "services", "exchange", "service.yaml"), []byte("name: exchange\nowner: Mailing Team\nowned_by: mailing.blumer.cloud\nstatus: draft\n"), 0o644))
	must(os.MkdirAll(filepath.Join(storage.CatalogDir(p), "blueprints", "products"), 0o755))
	must(os.WriteFile(filepath.Join(storage.CatalogDir(p), "blueprints", "products", "account.yaml"), []byte(`id: PROD-ACC-MBX-001
type: product_blueprint
name: Benutzeraccount mit Mailbox
offered_by: blumer.cloud
owning_domain: blumer.cloud
fulfillment:
  required_services:
    - service_ref: identity.blumer.cloud/user-account
      role: primary
      required: true
      sla_ref: SLA-IDENTITY-ACCOUNT-STANDARD
      ola_ref: OLA-IDENTITY-OPS-STANDARD
    - service_ref: mailing.blumer.cloud/exchange
      role: primary
      required: true
    - service_ref: missing.example.cloud/unknown-service
      role: supporting
      required: false
`), 0o644))

	tree, err := BuildNamespaceTree(p)
	if err != nil {
		t.Fatal(err)
	}
	parent := findTreeNode(tree.Root, "product-fulfillment-parent", "Fulfillment Services")
	if parent == nil || parent.FulfillmentCount != 3 || len(parent.Children) != 3 {
		t.Fatalf("expected fulfillment parent with three children: %+v", parent)
	}
	labels := strings.Join(flattenLabels(*parent), "|")
	for _, want := range []string{"Fulfillment Services", "user-account", "exchange", "unknown-service"} {
		if !strings.Contains(labels, want) {
			t.Fatalf("missing %s in %s", want, labels)
		}
	}
	if strings.Contains(labels, "No fulfillment services yet") {
		t.Fatalf("unexpected empty state in %s", labels)
	}
	byRef := map[string]*ProductRequiredServiceDTO{}
	for i := range parent.Children {
		if parent.Children[i].Fulfillment != nil {
			byRef[parent.Children[i].Fulfillment.ServiceRef] = parent.Children[i].Fulfillment
		}
	}
	first := byRef["identity.blumer.cloud/user-account"]
	if first == nil || first.TreeTarget != "service:identity.blumer.cloud/user-account" || first.FulfillmentType != "cross-domain" || first.SLARef == "" || first.OLARef == "" {
		t.Fatalf("expected cross-domain resolved fulfillment target with SLA/OLA refs: %+v", first)
	}
	missing := byRef["missing.example.cloud/unknown-service"]
	if missing == nil || missing.ResolutionStatus != "unresolved_domain" || missing.FulfillmentType != "unresolved" || missing.TreeTarget != "" {
		t.Fatalf("expected unresolved fulfillment visible without target: %+v", missing)
	}
}

func TestBuildNamespaceTreeGroupsDomainsUnderNamespaceNodes(t *testing.T) {
	p := createAppTestCosmos(t)
	if _, err := AddDomain(p, "blumer.com", "Web", false); err != nil {
		t.Fatal(err)
	}
	if _, err := AddDomain(p, "identity.blumer.com", "Identity", false); err != nil {
		t.Fatal(err)
	}
	if _, err := AddDomain(p, "blumer.cloud", "Cloud", false); err != nil {
		t.Fatal(err)
	}
	if _, err := AddDomain(p, "home.blumer.cloud", "Home", false); err != nil {
		t.Fatal(err)
	}
	tree, err := BuildNamespaceTree(p)
	if err != nil {
		t.Fatal(err)
	}
	if !hasTreePath(tree.Root, []string{"Namespaces", "com", "blumer", "identity"}, "identity.blumer.com") {
		t.Fatalf("missing com/blumer/identity tree: %+v", tree.Root)
	}
	if !hasTreePath(tree.Root, []string{"Namespaces", "cloud", "blumer", "home"}, "home.blumer.cloud") {
		t.Fatalf("missing cloud/blumer/home tree: %+v", tree.Root)
	}
}

func hasTreePath(n NamespaceTreeNodeDTO, labels []string, canonical string) bool {
	if len(labels) == 0 {
		return n.Canonical == canonical
	}
	for _, child := range n.Children {
		if child.Label == labels[0] {
			return hasTreePath(child, labels[1:], canonical)
		}
	}
	return false
}

func flattenLabels(n NamespaceTreeNodeDTO) []string {
	out := []string{n.Label}
	for _, c := range n.Children {
		out = append(out, flattenLabels(c)...)
	}
	return out
}

func TestAppNotFoundErrors(t *testing.T) {
	p := createAppTestCosmos(t)
	if _, err := GetDomain(p, "does-not-exist.example"); err == nil || !strings.Contains(err.Error(), CodeDomainNotFound) {
		t.Fatalf("expected domain error, got %v", err)
	}
	if _, err := GetService(p, "identity.blumer.cloud", "does-not-exist"); err == nil || !strings.Contains(err.Error(), CodeServiceNotFound) {
		t.Fatalf("expected service error, got %v", err)
	}
}

func TestDeleteDomain(t *testing.T) {
	p := createAppTestCosmos(t)
	err := DeleteDomain(p, "identity.blumer.cloud")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := os.Stat(filepath.Join(storage.DomainsDir(p), "identity.blumer.cloud")); !os.IsNotExist(err) {
		t.Fatal("expected domain directory to be removed")
	}
	if _, err := GetDomain(p, "identity.blumer.cloud"); err == nil {
		t.Fatal("expected domain not found after delete")
	}

	// Remaining domain should still exist
	if _, err := GetDomain(p, "platform.blumer.cloud"); err != nil {
		t.Fatalf("unexpected error for remaining domain: %v", err)
	}
}

func TestDeleteDomain_NotFound(t *testing.T) {
	p := createAppTestCosmos(t)
	err := DeleteDomain(p, "does-not-exist.example")
	if err == nil {
		t.Fatal("expected error")
	}
	if ae, ok := AsAppError(err); !ok || ae.Code != CodeDomainNotFound {
		t.Fatalf("expected DOMAIN_NOT_FOUND, got %v", err)
	}
}

func TestDeleteService(t *testing.T) {
	p := createAppTestCosmos(t)
	err := DeleteService(p, "identity.blumer.cloud", "user-account")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := os.Stat(filepath.Join(storage.DomainsDir(p), "identity.blumer.cloud", "services", "user-account")); !os.IsNotExist(err) {
		t.Fatal("expected service directory to be removed")
	}
	if _, err := GetService(p, "identity.blumer.cloud", "user-account"); err == nil {
		t.Fatal("expected service not found after delete")
	}

	// Remaining services should still exist
	if _, err := GetService(p, "identity.blumer.cloud", "privileged-account"); err != nil {
		t.Fatalf("unexpected error for remaining service: %v", err)
	}
}

func TestDeleteService_NotFound(t *testing.T) {
	p := createAppTestCosmos(t)
	err := DeleteService(p, "identity.blumer.cloud", "does-not-exist")
	if err == nil {
		t.Fatal("expected error")
	}
	if ae, ok := AsAppError(err); !ok || ae.Code != CodeServiceNotFound {
		t.Fatalf("expected SERVICE_NOT_FOUND, got %v", err)
	}
}

func TestBuildNamespaceTreeMarksVirtualAndPersistedDomains(t *testing.T) {
	p := createAppTestCosmos(t)
	if err := DeleteDomain(p, "blumer.cloud"); err != nil {
		// The demo fixture may not include this parent in older layouts.
		if !strings.Contains(err.Error(), CodeDomainNotFound) {
			t.Fatal(err)
		}
	}
	tree, err := BuildNamespaceTree(p)
	if err != nil {
		t.Fatal(err)
	}
	blumer := findTreePath(tree.Root, []string{"Namespaces", "cloud", "blumer"})
	if blumer == nil {
		t.Fatalf("missing virtual blumer node: %+v", tree.Root)
	}
	if !blumer.Virtual || blumer.Persisted || !blumer.CanCreateChildDomain || blumer.CanAddService || blumer.CanOpenDetails {
		t.Fatalf("blumer node metadata=%+v", *blumer)
	}
	identity := findTreePath(tree.Root, []string{"Namespaces", "cloud", "blumer", "identity"})
	if identity == nil {
		t.Fatalf("missing identity node: %+v", tree.Root)
	}
	if !identity.Persisted || identity.Virtual || !identity.CanAddService || !identity.CanOpenDetails {
		t.Fatalf("identity node metadata=%+v", *identity)
	}
	cosmos, err := GetCosmos(p)
	if err != nil {
		t.Fatal(err)
	}
	if cosmos.VirtualDomainCount < 1 {
		t.Fatalf("expected virtual parent count, got %+v", cosmos)
	}
}

func findTreePath(n NamespaceTreeNodeDTO, labels []string) *NamespaceTreeNodeDTO {
	if len(labels) == 0 {
		return &n
	}
	for i := range n.Children {
		if n.Children[i].Label == labels[0] {
			return findTreePath(n.Children[i], labels[1:])
		}
	}
	return nil
}

func TestProductOfferingAndServiceDTOs(t *testing.T) {
	p := createAppTestCosmos(t)
	must := func(err error) {
		if err != nil {
			t.Fatal(err)
		}
	}
	must(os.MkdirAll(filepath.Join(storage.CatalogDir(p), "blueprints", "products"), 0o755))
	must(os.WriteFile(filepath.Join(storage.DomainsDir(p), "identity.blumer.cloud", "services", "user-account", "service.yaml"), []byte(`name: user-account
owner: Identity Team
owned_by: identity.blumer.cloud
operated_by:
  - identity.blumer.cloud
capabilities:
  - user-account-management
supported_products:
  - PROD-ACC-MBX-001
status: draft
`), 0o644))
	must(os.WriteFile(filepath.Join(storage.CatalogDir(p), "blueprints", "products", "account.yaml"), []byte(`id: PROD-ACC-MBX-001
type: product_blueprint
name: Benutzerkonto mit Mailbox
version: 0.1.0
status: draft
owner: Identity Team
offered_by: identity.blumer.cloud
owning_domain: identity.blumer.cloud
required_inputs:
  - person_reference
fulfillment:
  required_services:
    - service_ref: identity.blumer.cloud/user-account
      role: primary
      required: true
      description: Creates the account.
      sla_ref: SLA-IDENTITY-ACCOUNT-STANDARD
      ola_ref: OLA-IDENTITY-OPS-STANDARD
`), 0o644))

	blueprints, err := ListBlueprints(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(blueprints.Blueprints) != 1 {
		t.Fatalf("expected one blueprint: %+v", blueprints)
	}
	bp := blueprints.Blueprints[0]
	if bp.OfferedBy != "identity.blumer.cloud" || bp.OwningDomain != "identity.blumer.cloud" {
		t.Fatalf("expected ownership in DTO: %+v", bp)
	}
	if len(bp.Fulfillment.RequiredServices) != 1 || bp.Fulfillment.RequiredServices[0].ResolutionStatus != "resolved" {
		t.Fatalf("expected resolved fulfillment DTO: %+v", bp.Fulfillment)
	}
	row := bp.Fulfillment.RequiredServices[0]
	if row.ServiceRef != "identity.blumer.cloud/user-account" || row.ResolvedDomain != "identity.blumer.cloud" || row.ResolvedService != "user-account" || row.FulfillmentType != "local" || row.SLARef != "SLA-IDENTITY-ACCOUNT-STANDARD" || row.OLARef != "OLA-IDENTITY-OPS-STANDARD" || row.TreeTarget != "service:identity.blumer.cloud/user-account" {
		t.Fatalf("expected normalized fulfillment row with refs and tree target: %+v", row)
	}

	svc, err := GetService(p, "identity.blumer.cloud", "user-account")
	if err != nil {
		t.Fatal(err)
	}
	if svc.OwnedBy != "identity.blumer.cloud" || len(svc.OperatedBy) != 1 || len(svc.Capabilities) != 1 || len(svc.SupportedProducts) != 1 {
		t.Fatalf("expected service metadata in DTO: %+v", svc)
	}
}

func TestDomainOwnedProductWorkflowDTOs(t *testing.T) {
	p := createAppTestCosmos(t)
	must := func(err error) {
		if err != nil {
			t.Fatal(err)
		}
	}
	must(os.MkdirAll(filepath.Join(storage.DomainsDir(p), "collaboration.blumer.cloud", "services", "mailbox"), 0o755))
	must(os.WriteFile(filepath.Join(storage.DomainsDir(p), "collaboration.blumer.cloud", "domain.yaml"), []byte("name: collaboration.blumer.cloud\nowner: Collaboration Team\nstatus: draft\n"), 0o644))
	must(os.WriteFile(filepath.Join(storage.DomainsDir(p), "collaboration.blumer.cloud", "services", "mailbox", "service.yaml"), []byte("name: mailbox\nowned_by: collaboration.blumer.cloud\nstatus: draft\n"), 0o644))
	must(os.MkdirAll(filepath.Join(storage.CatalogDir(p), "blueprints", "products"), 0o755))
	must(os.WriteFile(filepath.Join(storage.CatalogDir(p), "blueprints", "products", "account.yaml"), []byte(`id: PROD-ACC-MBX-001
type: product_blueprint
name: Benutzerkonto mit Mailbox
version: 0.1.0
status: draft
owner: Identity Team
offered_by: identity.blumer.cloud
owning_domain: identity.blumer.cloud
summary: Product owned by the identity domain.
fulfillment:
  required_services:
    - service_ref: identity.blumer.cloud/user-account
      role: primary
      required: true
    - service_ref: collaboration.blumer.cloud/mailbox
      role: supporting
      required: true
    - service_ref: missing.blumer.cloud/ghost
      role: dependency
      required: true
`), 0o644))

	domain, err := GetDomain(p, "identity.blumer.cloud")
	if err != nil {
		t.Fatal(err)
	}
	if len(domain.Products) != 1 || domain.Products[0].ID != "PROD-ACC-MBX-001" {
		t.Fatalf("expected domain product offering: %+v", domain.Products)
	}
	foundUserAccount := false
	for _, svc := range domain.Services {
		if svc.Name == "user-account" {
			foundUserAccount = true
		}
	}
	if !foundUserAccount {
		t.Fatalf("expected services on domain detail: %+v", domain.Services)
	}
	if domain.Products[0].FulfillmentRequiredServicesCount != 3 || domain.Products[0].FulfillmentUnresolvedCount != 1 {
		t.Fatalf("expected fulfillment summary counts: %+v", domain.Products[0])
	}

	product, err := GetBlueprint(p, "PROD-ACC-MBX-001")
	if err != nil {
		t.Fatal(err)
	}
	statuses := map[string]string{}
	fulfillmentTypes := map[string]string{}
	for _, svc := range product.Fulfillment.RequiredServices {
		statuses[svc.ServiceRef] = svc.ResolutionStatus
		fulfillmentTypes[svc.ServiceRef] = svc.FulfillmentType
	}
	if statuses["identity.blumer.cloud/user-account"] != "resolved" || statuses["collaboration.blumer.cloud/mailbox"] != "resolved" || statuses["missing.blumer.cloud/ghost"] != "unresolved_domain" {
		t.Fatalf("unexpected resolution statuses: %+v", statuses)
	}
	if fulfillmentTypes["identity.blumer.cloud/user-account"] != "local" || fulfillmentTypes["collaboration.blumer.cloud/mailbox"] != "cross-domain" {
		t.Fatalf("unexpected fulfillment types: %+v", fulfillmentTypes)
	}

	refs, err := AllServiceRefs(p)
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	for _, ref := range refs.Services {
		seen[ref.ServiceRef] = true
	}
	if !seen["identity.blumer.cloud/user-account"] || !seen["collaboration.blumer.cloud/mailbox"] {
		t.Fatalf("missing canonical service refs: %+v", refs.Services)
	}
}

func TestCreateProductOfferingAndAppendFulfillment(t *testing.T) {
	p := createAppTestCosmos(t)
	product, err := CreateProductOffering(p, "identity.blumer.cloud", CreateProductOfferingRequest{ID: "PROD-NEW-001", Name: "New Product", Summary: "Created from the domain."})
	if err != nil {
		t.Fatal(err)
	}
	if product.OfferedBy != "identity.blumer.cloud" || product.OwningDomain != "identity.blumer.cloud" || product.Version != "0.1.0" || product.Status != "draft" {
		t.Fatalf("unexpected created product defaults: %+v", product)
	}

	updated, err := AddProductFulfillmentService(p, "PROD-NEW-001", AddFulfillmentServiceRequest{ServiceRef: "identity.blumer.cloud/user-account", Role: "primary", Required: true, Description: "Create account."})
	if err != nil {
		t.Fatal(err)
	}
	if len(updated.Fulfillment.RequiredServices) != 1 || updated.Fulfillment.RequiredServices[0].ResolutionStatus != "resolved" || updated.Fulfillment.RequiredServices[0].FulfillmentType != "local" {
		t.Fatalf("expected resolved local fulfillment service: %+v", updated.Fulfillment)
	}
	tree, err := BuildNamespaceTree(p)
	if err != nil {
		t.Fatal(err)
	}
	parent := findTreeNode(tree.Root, "product-fulfillment-parent", "Fulfillment Services")
	if parent == nil || parent.FulfillmentCount != 1 || len(parent.Children) != 1 {
		t.Fatalf("expected appended fulfillment in tree summary: %+v", parent)
	}
	child := parent.Children[0]
	if child.Label != "identity.blumer.cloud/user-account" || child.Fulfillment == nil || child.Fulfillment.ResolutionStatus != "resolved" || child.Fulfillment.FulfillmentType != "local" {
		t.Fatalf("expected tree child to use normalized fulfillment DTO: %+v", child)
	}
	if _, err := CreateProductOffering(p, "identity.blumer.cloud", CreateProductOfferingRequest{ID: "PROD-NEW-001", Name: "Duplicate"}); err == nil {
		t.Fatal("expected duplicate product id to be rejected")
	}
	if _, err := AddProductFulfillmentService(p, "PROD-NEW-001", AddFulfillmentServiceRequest{ServiceRef: "identity.blumer.cloud/user-account", Role: "supporting", Required: true}); err == nil {
		t.Fatal("expected duplicate fulfillment service ref to be rejected")
	}
}

func TestMoveProductOfferingOwnershipRulesAndFulfillment(t *testing.T) {
	p := createAppTestCosmos(t)
	must := func(err error) {
		if err != nil {
			t.Fatal(err)
		}
	}
	must(os.MkdirAll(filepath.Join(storage.CatalogDir(p), "blueprints", "products"), 0o755))
	must(os.WriteFile(filepath.Join(storage.CatalogDir(p), "blueprints", "products", "move.yaml"), []byte(`id: PROD-MOVE-001
type: product_blueprint
name: Move Me
version: 1.2.3
status: active
owner: Product Team
offered_by: platform.blumer.cloud
owning_domain: platform.blumer.cloud
summary: Preserve this summary.
required_inputs:
  - person_reference
fulfillment:
  required_services:
    - service_ref: identity.blumer.cloud/user-account
      role: primary
      required: true
      description: Existing service ref stays unchanged.
`), 0o644))

	before, err := GetBlueprint(p, "PROD-MOVE-001")
	if err != nil {
		t.Fatal(err)
	}
	if before.Fulfillment.RequiredServices[0].FulfillmentType != "cross-domain" {
		t.Fatalf("expected pre-move cross-domain fulfillment: %+v", before.Fulfillment.RequiredServices[0])
	}

	moved, err := MoveProductOffering(p, MoveProductOfferingRequest{ProductID: "PROD-MOVE-001", TargetDomain: "identity.blumer.cloud"})
	if err != nil {
		t.Fatal(err)
	}
	if moved.OfferedBy != "identity.blumer.cloud" || moved.OwningDomain != "identity.blumer.cloud" || moved.Version != "1.2.3" || moved.Summary != "Preserve this summary." {
		t.Fatalf("unexpected moved product: %+v", moved)
	}
	if moved.Fulfillment.RequiredServices[0].ServiceRef != "identity.blumer.cloud/user-account" || moved.Fulfillment.RequiredServices[0].FulfillmentType != "local" {
		t.Fatalf("expected unchanged ref and recomputed local fulfillment: %+v", moved.Fulfillment.RequiredServices[0])
	}
	raw, err := os.ReadFile(filepath.Join(storage.CatalogDir(p), "blueprints", "products", "move.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), "offered_by: identity.blumer.cloud") || !strings.Contains(string(raw), "owning_domain: identity.blumer.cloud") || !strings.Contains(string(raw), "person_reference") {
		t.Fatalf("YAML was not updated/preserved as expected:\n%s", raw)
	}

	if _, err := MoveProductOffering(p, MoveProductOfferingRequest{ProductID: "PROD-MOVE-001", TargetDomain: "identity.blumer.cloud"}); err == nil || !strings.Contains(err.Error(), CodeProductMoveNoop) {
		t.Fatalf("expected no-op move error, got %v", err)
	}
	if _, err := MoveProductOffering(p, MoveProductOfferingRequest{ProductID: "missing", TargetDomain: "platform.blumer.cloud"}); err == nil || !strings.Contains(err.Error(), CodeProductNotFound) {
		t.Fatalf("expected missing product error, got %v", err)
	}
	if _, err := MoveProductOffering(p, MoveProductOfferingRequest{ProductID: "PROD-MOVE-001", TargetDomain: "missing.blumer.cloud"}); err == nil || !strings.Contains(err.Error(), CodeTargetDomainNotFound) {
		t.Fatalf("expected target domain error, got %v", err)
	}
}

func TestMoveProductOfferingPreservesDelegatedOwningDomainUnlessRequested(t *testing.T) {
	p := createAppTestCosmos(t)
	must := func(err error) {
		if err != nil {
			t.Fatal(err)
		}
	}
	must(os.MkdirAll(filepath.Join(storage.CatalogDir(p), "blueprints", "products"), 0o755))
	must(os.WriteFile(filepath.Join(storage.CatalogDir(p), "blueprints", "products", "delegated.yaml"), []byte(`id: PROD-DELEGATED-001
type: product_blueprint
name: Delegated
offered_by: platform.blumer.cloud
owning_domain: identity.blumer.cloud
`), 0o644))

	moved, err := MoveProductOffering(p, MoveProductOfferingRequest{ProductID: "PROD-DELEGATED-001", TargetDomain: "identity.blumer.cloud"})
	if err != nil {
		t.Fatal(err)
	}
	if moved.OwningDomain != "identity.blumer.cloud" {
		t.Fatalf("expected delegated owning domain preserved: %+v", moved)
	}
	movedBack, err := MoveProductOffering(p, MoveProductOfferingRequest{ProductID: "PROD-DELEGATED-001", TargetDomain: "platform.blumer.cloud", UpdateOwningDomain: true})
	if err != nil {
		t.Fatal(err)
	}
	if movedBack.OfferedBy != "platform.blumer.cloud" || movedBack.OwningDomain != "platform.blumer.cloud" {
		t.Fatalf("expected explicit owning-domain move: %+v", movedBack)
	}
}

func TestFulfillmentServiceLevelDTOsAndFallback(t *testing.T) {
	p := createAppTestCosmos(t)
	must := func(err error) {
		if err != nil {
			t.Fatal(err)
		}
	}
	must(os.MkdirAll(filepath.Join(storage.CatalogDir(p), "blueprints", "products"), 0o755))
	must(os.WriteFile(filepath.Join(storage.DomainsDir(p), "identity.blumer.cloud", "services", "user-account", "service.yaml"), []byte(`name: user-account
owner: Identity Team
owned_by: identity.blumer.cloud
status: draft
ola:
  name: Identity Account OLA
  target: 4h
  availability: business-hours
`), 0o644))
	must(os.WriteFile(filepath.Join(storage.DomainsDir(p), "platform.blumer.cloud", "services", "rule-validation-api", "service.yaml"), []byte(`name: rule-validation-api
owner: Platform Team
owned_by: platform.blumer.cloud
status: draft
sla:
  name: Rule Validation SLA
  target: 8h
  availability: business-hours
`), 0o644))
	must(os.WriteFile(filepath.Join(storage.CatalogDir(p), "blueprints", "products", "levels.yaml"), []byte(`id: PROD-LEVELS-001
type: product_blueprint
name: Product With Levels
offered_by: identity.blumer.cloud
owning_domain: identity.blumer.cloud
fulfillment:
  required_services:
    - service_ref: identity.blumer.cloud/user-account
      role: primary
      required: true
      description: Inline OLA wins.
      ola:
        name: Inline Identity OLA
        target: 2h
        availability: business-hours
    - service_ref: platform.blumer.cloud/rule-validation-api
      role: supporting
      required: false
      description: Falls back to service SLA.
    - service_ref: identity.blumer.cloud/privileged-account
      role: optional
      required: false
      description: Existing product without SLA or OLA remains valid.
`), 0o644))

	bp, err := GetBlueprint(p, "PROD-LEVELS-001")
	if err != nil {
		t.Fatal(err)
	}
	byRef := map[string]ProductRequiredServiceDTO{}
	for _, svc := range bp.Fulfillment.RequiredServices {
		byRef[svc.ServiceRef] = svc
	}
	if got := byRef["identity.blumer.cloud/user-account"]; got.FulfillmentType != "local" || got.OLA == nil || got.OLA.Target != "2h" || got.ServiceLevelLabel != "OLA 2h" {
		t.Fatalf("expected inline local OLA DTO, got %+v", got)
	}
	if got := byRef["platform.blumer.cloud/rule-validation-api"]; got.FulfillmentType != "cross-domain" || got.SLA == nil || got.SLA.Target != "8h" || got.ServiceLevelLabel != "SLA 8h" {
		t.Fatalf("expected fallback cross-domain SLA DTO, got %+v", got)
	}
	if got := byRef["identity.blumer.cloud/privileged-account"]; got.SLA != nil || got.OLA != nil || got.ServiceLevelLabel != "" || got.FulfillmentType != "local" {
		t.Fatalf("expected valid DTO without service levels, got %+v", got)
	}
}
