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
	must(os.MkdirAll(filepath.Join(storage.ServicesDir(p), "user-account"), 0o755))
	must(os.MkdirAll(filepath.Join(storage.ServicesDir(p), "privileged-account"), 0o755))
	must(os.MkdirAll(filepath.Join(storage.ServicesDir(p), "rule-validation-api"), 0o755))
	must(os.WriteFile(storage.CosmosFile(p), []byte("id: cosmos-local\nname: Local Cosmos\nversion: 0.1.0\nstatus: draft\nowner: Platform Team\n"), 0o644))
	must(os.WriteFile(filepath.Join(storage.ServicesDir(p), "user-account/service.yaml"), []byte("name: user-account\nowner: Identity Team\nstatus: draft\n"), 0o644))
	must(os.WriteFile(filepath.Join(storage.ServicesDir(p), "privileged-account/service.yaml"), []byte("name: privileged-account\nowner: Identity Team\nstatus: draft\n"), 0o644))
	must(os.WriteFile(filepath.Join(storage.ServicesDir(p), "rule-validation-api/service.yaml"), []byte("name: rule-validation-api\nowner: Platform Team\nstatus: draft\n"), 0o644))
	return p
}

func TestAppUseCasesExposeNamespaceMetadata(t *testing.T) {
	p := createAppTestCosmos(t)
	cosmos, err := GetCosmos(p)
	if err != nil {
		t.Fatal(err)
	}
	if cosmos.ServiceCount != 3 {
		t.Fatalf("unexpected counts: %+v", cosmos)
	}
	services, err := ListServices(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(services.Services) != 3 {
		t.Fatalf("unexpected services: %+v", services.Services)
	}
	service, err := GetService(p, "user-account")
	if err != nil {
		t.Fatal(err)
	}
	if service.Name != "user-account" {
		t.Fatalf("unexpected service: %+v", service)
	}
	graph, err := BuildGraph(p)
	if err != nil {
		t.Fatal(err)
	}
	if graph.Format != "mermaid" {
		t.Fatalf("unexpected graph: %+v", graph)
	}
	validation, err := ValidateCosmos(p)
	if err != nil {
		t.Fatal(err)
	}
	_ = validation
}

func TestBuildNamespaceTree(t *testing.T) {
	tree, err := BuildNamespaceTree(createAppTestCosmos(t))
	if err != nil {
		t.Fatal(err)
	}
	body := strings.Join(flattenLabels(tree.Root), "|")
	for _, want := range []string{"Local Cosmos", "Services", "user-account", "rule-validation-api"} {
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
	must(os.MkdirAll(filepath.Join(storage.ServicesDir(p), "exchange"), 0o755))
	must(os.WriteFile(filepath.Join(storage.ServicesDir(p), "exchange", "service.yaml"), []byte("name: exchange\nowner: Mailing Team\nstatus: draft\n"), 0o644))
	must(os.MkdirAll(filepath.Join(storage.CatalogDir(p), "blueprints", "products"), 0o755))
	must(os.WriteFile(filepath.Join(storage.CatalogDir(p), "blueprints", "products", "account.yaml"), []byte(`id: PROD-ACC-MBX-001
type: product_blueprint
name: Benutzeraccount mit Mailbox
fulfillment:
  required_services:
    - service_ref: user-account
      role: primary
      required: true
      sla_ref: SLA-IDENTITY-ACCOUNT-STANDARD
      ola_ref: OLA-IDENTITY-OPS-STANDARD
    - service_ref: exchange
      role: primary
      required: true
    - service_ref: unknown-service
      role: supporting
      required: false
`), 0o644))

	// The tree exposes products flatly; fulfillment detail is on the product DTO.
	if _, err := BuildNamespaceTree(p); err != nil {
		t.Fatal(err)
	}
	bp, err := GetBlueprint(p, "PROD-ACC-MBX-001")
	if err != nil {
		t.Fatal(err)
	}
	if len(bp.Fulfillment.RequiredServices) != 3 {
		t.Fatalf("expected three fulfillment services: %+v", bp.Fulfillment.RequiredServices)
	}
	byRef := map[string]ProductRequiredServiceDTO{}
	for _, svc := range bp.Fulfillment.RequiredServices {
		byRef[svc.ServiceRef] = svc
	}
	first := byRef["user-account"]
	if first.TreeTarget != "service:user-account" || first.FulfillmentType != "local" || first.SLARef == "" || first.OLARef == "" {
		t.Fatalf("expected resolved local fulfillment target with SLA/OLA refs: %+v", first)
	}
	missing := byRef["unknown-service"]
	if missing.ResolutionStatus != "unresolved_service" || missing.FulfillmentType != "unresolved" || missing.TreeTarget != "" {
		t.Fatalf("expected unresolved fulfillment visible without target: %+v", missing)
	}
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
	if _, err := GetService(p, "does-not-exist"); err == nil || !strings.Contains(err.Error(), CodeServiceNotFound) {
		t.Fatalf("expected service error, got %v", err)
	}
}

func TestDeleteService(t *testing.T) {
	p := createAppTestCosmos(t)
	err := DeleteService(p, "user-account")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := os.Stat(filepath.Join(storage.ServicesDir(p), "user-account")); !os.IsNotExist(err) {
		t.Fatal("expected service directory to be removed")
	}
	if _, err := GetService(p, "user-account"); err == nil {
		t.Fatal("expected service not found after delete")
	}

	// Remaining services should still exist
	if _, err := GetService(p, "privileged-account"); err != nil {
		t.Fatalf("unexpected error for remaining service: %v", err)
	}
}

func TestDeleteService_NotFound(t *testing.T) {
	p := createAppTestCosmos(t)
	err := DeleteService(p, "does-not-exist")
	if err == nil {
		t.Fatal("expected error")
	}
	if ae, ok := AsAppError(err); !ok || ae.Code != CodeServiceNotFound {
		t.Fatalf("expected SERVICE_NOT_FOUND, got %v", err)
	}
}

func TestProductOfferingAndServiceDTOs(t *testing.T) {
	p := createAppTestCosmos(t)
	must := func(err error) {
		if err != nil {
			t.Fatal(err)
		}
	}
	must(os.MkdirAll(filepath.Join(storage.CatalogDir(p), "blueprints", "products"), 0o755))
	must(os.WriteFile(filepath.Join(storage.ServicesDir(p), "user-account", "service.yaml"), []byte(`name: user-account
owner: Identity Team
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
required_inputs:
  - person_reference
fulfillment:
  required_services:
    - service_ref: user-account
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
	if len(bp.Fulfillment.RequiredServices) != 1 || bp.Fulfillment.RequiredServices[0].ResolutionStatus != "resolved" {
		t.Fatalf("expected resolved fulfillment DTO: %+v", bp.Fulfillment)
	}
	row := bp.Fulfillment.RequiredServices[0]
	if row.ServiceRef != "user-account" || row.ResolvedService != "user-account" || row.FulfillmentType != "local" || row.SLARef != "SLA-IDENTITY-ACCOUNT-STANDARD" || row.OLARef != "OLA-IDENTITY-OPS-STANDARD" || row.TreeTarget != "service:user-account" {
		t.Fatalf("expected normalized fulfillment row with refs and tree target: %+v", row)
	}

	svc, err := GetService(p, "user-account")
	if err != nil {
		t.Fatal(err)
	}
	if len(svc.Capabilities) != 1 || len(svc.SupportedProducts) != 1 {
		t.Fatalf("expected service metadata in DTO: %+v", svc)
	}
}

func TestProductWorkflowDTOs(t *testing.T) {
	p := createAppTestCosmos(t)
	must := func(err error) {
		if err != nil {
			t.Fatal(err)
		}
	}
	must(os.MkdirAll(filepath.Join(storage.ServicesDir(p), "mailbox"), 0o755))
	must(os.WriteFile(filepath.Join(storage.ServicesDir(p), "mailbox", "service.yaml"), []byte("name: mailbox\nstatus: draft\n"), 0o644))
	must(os.MkdirAll(filepath.Join(storage.CatalogDir(p), "blueprints", "products"), 0o755))
	must(os.WriteFile(filepath.Join(storage.CatalogDir(p), "blueprints", "products", "account.yaml"), []byte(`id: PROD-ACC-MBX-001
type: product_blueprint
name: Benutzerkonto mit Mailbox
version: 0.1.0
status: draft
owner: Identity Team
summary: Product offering.
fulfillment:
  required_services:
    - service_ref: user-account
      role: primary
      required: true
    - service_ref: mailbox
      role: supporting
      required: true
    - service_ref: ghost
      role: dependency
      required: true
`), 0o644))

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
	if statuses["user-account"] != "resolved" || statuses["mailbox"] != "resolved" || statuses["ghost"] != "unresolved_service" {
		t.Fatalf("unexpected resolution statuses: %+v", statuses)
	}
	if fulfillmentTypes["user-account"] != "local" || fulfillmentTypes["mailbox"] != "local" {
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
	if !seen["user-account"] || !seen["mailbox"] {
		t.Fatalf("missing canonical service refs: %+v", refs.Services)
	}
}

func TestCreateProductOfferingAndAppendFulfillment(t *testing.T) {
	p := createAppTestCosmos(t)
	product, err := CreateProductOffering(p, CreateProductOfferingRequest{ID: "PROD-NEW-001", Name: "New Product", Summary: "Created."})
	if err != nil {
		t.Fatal(err)
	}
	if product.Version != "0.1.0" || product.Status != "draft" {
		t.Fatalf("unexpected created product defaults: %+v", product)
	}

	updated, err := AddProductFulfillmentService(p, "PROD-NEW-001", AddFulfillmentServiceRequest{ServiceRef: "user-account", Role: "primary", Required: true, Description: "Create account."})
	if err != nil {
		t.Fatal(err)
	}
	if len(updated.Fulfillment.RequiredServices) != 1 || updated.Fulfillment.RequiredServices[0].ResolutionStatus != "resolved" || updated.Fulfillment.RequiredServices[0].FulfillmentType != "local" {
		t.Fatalf("expected resolved local fulfillment service: %+v", updated.Fulfillment)
	}
	row := updated.Fulfillment.RequiredServices[0]
	if row.ServiceRef != "user-account" || row.ResolutionStatus != "resolved" || row.FulfillmentType != "local" {
		t.Fatalf("expected normalized fulfillment DTO: %+v", row)
	}
	if _, err := BuildNamespaceTree(p); err != nil {
		t.Fatal(err)
	}
	if _, err := CreateProductOffering(p, CreateProductOfferingRequest{ID: "PROD-NEW-001", Name: "Duplicate"}); err == nil {
		t.Fatal("expected duplicate product id to be rejected")
	}
	if _, err := AddProductFulfillmentService(p, "PROD-NEW-001", AddFulfillmentServiceRequest{ServiceRef: "user-account", Role: "supporting", Required: true}); err == nil {
		t.Fatal("expected duplicate fulfillment service ref to be rejected")
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
	must(os.WriteFile(filepath.Join(storage.ServicesDir(p), "user-account", "service.yaml"), []byte(`name: user-account
owner: Identity Team
status: draft
ola:
  name: Identity Account OLA
  target: 4h
  availability: business-hours
`), 0o644))
	must(os.WriteFile(filepath.Join(storage.ServicesDir(p), "rule-validation-api", "service.yaml"), []byte(`name: rule-validation-api
owner: Platform Team
status: draft
sla:
  name: Rule Validation SLA
  target: 8h
  availability: business-hours
`), 0o644))
	must(os.WriteFile(filepath.Join(storage.CatalogDir(p), "blueprints", "products", "levels.yaml"), []byte(`id: PROD-LEVELS-001
type: product_blueprint
name: Product With Levels
fulfillment:
  required_services:
    - service_ref: user-account
      role: primary
      required: true
      description: Inline OLA wins.
      ola:
        name: Inline Identity OLA
        target: 2h
        availability: business-hours
    - service_ref: rule-validation-api
      role: supporting
      required: false
      description: Falls back to service SLA.
    - service_ref: privileged-account
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
	if got := byRef["user-account"]; got.FulfillmentType != "local" || got.OLA == nil || got.OLA.Target != "2h" || got.ServiceLevelLabel != "OLA 2h" {
		t.Fatalf("expected inline local OLA DTO, got %+v", got)
	}
	if got := byRef["rule-validation-api"]; got.FulfillmentType != "local" || got.SLA == nil || got.SLA.Target != "8h" || got.ServiceLevelLabel != "SLA 8h" {
		t.Fatalf("expected fallback local SLA DTO, got %+v", got)
	}
	if got := byRef["privileged-account"]; got.SLA != nil || got.OLA != nil || got.ServiceLevelLabel != "" || got.FulfillmentType != "local" {
		t.Fatalf("expected valid DTO without service levels, got %+v", got)
	}
}
