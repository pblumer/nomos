import { useEffect, useMemo, useState } from "react";

type Product = {
  id: string;
  name: string;
  version: string;
};

type ProductResponse = {
  items: Product[];
  count: number;
};

type ProductValidation = {
  is_valid: boolean;
  errors: string[];
};

type ProductDetail = {
  id: string;
  name: string;
  version?: string;
  description?: string;
  requirements?: string[];
  rules?: string[];
  validation?: ProductValidation;
};

type ProductRequirementsResponse = {
  product_id: string;
  items: string[];
  count: number;
};

type ProductRulesResponse = {
  product_id: string;
  items: string[];
  count: number;
};

type RuleDetail = {
  id: string;
  name: string;
  description?: string;
  category?: string;
  severity?: string;
  scope?: string;
  enforcement?: string;
  type?: string;
  condition?: string;
  effect?: string;
  rationale?: string;
  evidence?: string;
  version?: string;
  validation?: {
    type?: string;
    field?: string;
    operator?: string;
    value?: string | number | boolean;
  };
};

type RuleEditorState = {
  name: string;
  description: string;
  category: string;
  severity: string;
  scope: string;
  enforcement: string;
  type: string;
  condition: string;
  effect: string;
  rationale: string;
  evidence: string;
  version: string;
  validationType: string;
  validationField: string;
  validationOperator: string;
  validationValue: string;
};

type RequirementDetail = {
  id: string;
  name: string;
  description?: string;
  category?: string;
  scope?: string;
  source?: string;
  priority?: string;
  version?: string;
};

type RequirementEditorState = {
  name: string;
  description: string;
  category: string;
  scope: string;
  source: string;
  priority: string;
  version: string;
};

type ProductSummary = {
  product_id: string;
  name: string;
  version: string;
  requirements_count: number;
  rules_count: number;
  validation_is_valid: boolean;
  validation_error_count: number;
};

type View = "dashboard" | "catalog" | "detail";

const navItems: Array<{ key: View; label: string; icon: string }> = [
  { key: "dashboard", label: "Dashboard", icon: "dashboard" },
  { key: "catalog", label: "Product Catalog", icon: "inventory_2" },
];

const secondaryNav = [
  { label: "Audit Log", icon: "history" },
  { label: "Settings", icon: "settings" },
];

function App() {
  const [products, setProducts] = useState<Product[]>([]);
  const [selectedProduct, setSelectedProduct] = useState<ProductDetail | null>(null);
  const [productSummary, setProductSummary] = useState<ProductSummary | null>(null);
  const [view, setView] = useState<View>("dashboard");
  const [activeTab, setActiveTab] = useState<"requirements" | "rules" | "variants">("requirements");
  const [requirements, setRequirements] = useState<string[]>([]);
  const [requirementsFilter, setRequirementsFilter] = useState("");
  const [newRequirement, setNewRequirement] = useState("");
  const [rules, setRules] = useState<string[]>([]);
  const [rulesFilter, setRulesFilter] = useState("");
  const [rulesSort, setRulesSort] = useState<"asc" | "desc">("asc");
  const [newRule, setNewRule] = useState("");
  const [variants, setVariants] = useState<Array<{ id: string; name: string; description?: string; context?: string; validity_conditions?: string[]; requirement_ids?: string[]; rule_ids?: string[] }>>([]);
  const [newVariant, setNewVariant] = useState({ id: "", name: "", description: "", context: "", validityConditions: "", requirementIds: "", ruleIds: "" });
  const [selectedRuleId, setSelectedRuleId] = useState<string | null>(null);
  const [ruleDetails, setRuleDetails] = useState<Record<string, RuleDetail>>({});
  const [ruleEditor, setRuleEditor] = useState<RuleEditorState>({ name: "", description: "", category: "", severity: "", scope: "", enforcement: "", type: "", condition: "", effect: "", rationale: "", evidence: "", version: "", validationType: "", validationField: "", validationOperator: "", validationValue: "" });
  const [selectedRequirementId, setSelectedRequirementId] = useState<string | null>(null);
  const [requirementDetails, setRequirementDetails] = useState<Record<string, RequirementDetail>>({});
  const [requirementEditor, setRequirementEditor] = useState<RequirementEditorState>({ name: "", description: "", category: "", scope: "", source: "", priority: "", version: "" });
  const [toast, setToast] = useState<{ type: "success" | "error"; message: string } | null>(null);
  const [newProduct, setNewProduct] = useState({ id: "", name: "", version: "", description: "" });
  const [editProduct, setEditProduct] = useState({ name: "", version: "", description: "" });
  const [searchQuery, setSearchQuery] = useState("");
  const apiBaseUrl = import.meta.env.VITE_API_BASE_URL ?? "";

  const readErrorMessage = async (response: Response) => {
    try { const payload = (await response.json()) as { detail?: string }; if (payload?.detail) return payload.detail; } catch { /* ignore */ }
    return "Request failed";
  };

  const toRuleEditorState = (detail: RuleDetail): RuleEditorState => ({
    name: detail.name ?? "", description: detail.description ?? "", category: detail.category ?? "", severity: detail.severity ?? "", scope: detail.scope ?? "", enforcement: detail.enforcement ?? "", type: detail.type ?? "", condition: detail.condition ?? "", effect: detail.effect ?? "", rationale: detail.rationale ?? "", evidence: detail.evidence ?? "", version: detail.version ?? "",
    validationType: detail.validation?.type ? String(detail.validation.type) : "",
    validationField: detail.validation?.field ? String(detail.validation.field) : "",
    validationOperator: detail.validation?.operator ? String(detail.validation.operator) : "",
    validationValue: detail.validation?.value !== undefined ? String(detail.validation.value) : "",
  });

  const toRequirementEditorState = (detail: RequirementDetail): RequirementEditorState => ({
    name: detail.name ?? "", description: detail.description ?? "", category: detail.category ?? "", scope: detail.scope ?? "", source: detail.source ?? "", priority: detail.priority ?? "", version: detail.version ?? "",
  });

  const loadProducts = async () => {
    try { const response = await fetch(`${apiBaseUrl}/api/v1/products`); if (!response.ok) return; const data = (await response.json()) as ProductResponse; setProducts(data.items); } catch { /* ignore */ }
  };

  useEffect(() => { void loadProducts(); }, [apiBaseUrl]);

  useEffect(() => { if (toast) { const timer = setTimeout(() => setToast(null), 3000); return () => clearTimeout(timer); } }, [toast]);

  const loadProductSummary = async (productId: string) => {
    try { const response = await fetch(`${apiBaseUrl}/api/v1/products/${productId}/summary`); if (!response.ok) return; const data = (await response.json()) as ProductSummary; setProductSummary(data); } catch { /* ignore */ }
  };

  const loadProductDetail = async (productId: string) => {
    try {
      const response = await fetch(`${apiBaseUrl}/api/v1/products/${productId}`);
      if (!response.ok) return;
      const data = (await response.json()) as ProductDetail;
      setSelectedProduct(data);
      setEditProduct({ name: data.name ?? "", version: data.version ?? "", description: data.description ?? "" });
      setProductSummary(null); setActiveTab("requirements"); setRequirements([]); setRequirementsFilter(""); setNewRequirement(""); setRules([]); setRulesFilter(""); setRulesSort("asc"); setNewRule(""); setVariants([]); setNewVariant({ id: "", name: "", description: "", context: "", validityConditions: "", requirementIds: "", ruleIds: "" }); setSelectedRuleId(null); setRuleDetails({}); setRuleEditor({ name: "", description: "", category: "", severity: "", scope: "", enforcement: "", type: "", condition: "", effect: "", rationale: "", evidence: "", version: "", validationType: "", validationField: "", validationOperator: "", validationValue: "" }); setSelectedRequirementId(null); setRequirementDetails({}); setRequirementEditor({ name: "", description: "", category: "", scope: "", source: "", priority: "", version: "" });
      setView("detail");
      void loadProductSummary(productId);
    } catch { /* ignore */ }
  };

  const createProduct = async () => {
    const payload = { id: newProduct.id.trim(), name: newProduct.name.trim(), version: newProduct.version.trim(), description: newProduct.description.trim() };
    if (!payload.id || !payload.name || !payload.version) { setToast({ type: "error", message: "id, name and version are required" }); return; }
    try { const response = await fetch(`${apiBaseUrl}/api/v1/products`, { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify(payload) }); if (!response.ok) { setToast({ type: "error", message: await readErrorMessage(response) }); return; } setNewProduct({ id: "", name: "", version: "", description: "" }); await loadProducts(); await loadProductDetail(payload.id); setToast({ type: "success", message: "Product created" }); } catch { setToast({ type: "error", message: "Request failed" }); }
  };

  const saveProduct = async () => {
    if (!selectedProduct) return;
    const payload = { name: editProduct.name.trim(), version: editProduct.version.trim(), description: editProduct.description.trim() };
    if (!payload.name || !payload.version) { setToast({ type: "error", message: "name and version are required" }); return; }
    try { const response = await fetch(`${apiBaseUrl}/api/v1/products/${selectedProduct.id}`, { method: "PUT", headers: { "Content-Type": "application/json" }, body: JSON.stringify(payload) }); if (!response.ok) { setToast({ type: "error", message: await readErrorMessage(response) }); return; } await loadProducts(); await loadProductDetail(selectedProduct.id); setToast({ type: "success", message: "Product saved" }); } catch { setToast({ type: "error", message: "Request failed" }); }
  };

  const removeProduct = async () => {
    if (!selectedProduct) return;
    let confirmed = true;
    try { confirmed = typeof window.confirm === "function" ? window.confirm(`Delete product ${selectedProduct.name}?`) : true; } catch { confirmed = true; }
    if (!confirmed) return;
    try { const response = await fetch(`${apiBaseUrl}/api/v1/products/${selectedProduct.id}`, { method: "DELETE" }); if (!response.ok) { setToast({ type: "error", message: await readErrorMessage(response) }); return; } setSelectedProduct(null); setProductSummary(null); setRequirements([]); setRules([]); setSelectedRuleId(null); setRuleDetails({}); setRuleEditor({ name: "", description: "", category: "", severity: "", scope: "", enforcement: "", type: "", condition: "", effect: "", rationale: "", evidence: "", version: "", validationType: "", validationField: "", validationOperator: "", validationValue: "" }); await loadProducts(); setView("catalog"); setToast({ type: "success", message: "Product deleted" }); } catch { setToast({ type: "error", message: "Request failed" }); }
  };

  const loadRequirements = async () => { if (!selectedProduct) return; try { const response = await fetch(`${apiBaseUrl}/api/v1/products/${selectedProduct.id}/requirements`); if (!response.ok) return; const data = (await response.json()) as ProductRequirementsResponse; setRequirements(data.items); } catch { /* ignore */ } };
  const addRequirement = async () => { if (!selectedProduct) return; const value = newRequirement.trim(); if (value === "") return; try { const response = await fetch(`${apiBaseUrl}/api/v1/products/${selectedProduct.id}/requirements`, { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify({ value }) }); if (!response.ok) { setToast({ type: "error", message: await readErrorMessage(response) }); return; } setNewRequirement(""); await loadRequirements(); await loadProductSummary(selectedProduct.id); setToast({ type: "success", message: "Requirement added" }); } catch { setToast({ type: "error", message: "Request failed" }); } };
  const removeRequirement = async (item: string) => { if (!selectedProduct) return; try { const response = await fetch(`${apiBaseUrl}/api/v1/products/${selectedProduct.id}/requirements/${encodeURIComponent(item)}`, { method: "DELETE" }); if (!response.ok) { setToast({ type: "error", message: await readErrorMessage(response) }); return; } await loadRequirements(); await loadProductSummary(selectedProduct.id); setToast({ type: "success", message: "Requirement removed" }); } catch { setToast({ type: "error", message: "Request failed" }); } };
  const loadRules = async () => { if (!selectedProduct) return; try { const response = await fetch(`${apiBaseUrl}/api/v1/products/${selectedProduct.id}/rules`); if (!response.ok) return; const data = (await response.json()) as ProductRulesResponse; setRules(data.items); setSelectedRuleId(null); setRuleEditor({ name: "", description: "", category: "", severity: "", scope: "", enforcement: "", type: "", condition: "", effect: "", rationale: "", evidence: "", version: "", validationType: "", validationField: "", validationOperator: "", validationValue: "" }); } catch { /* ignore */ } };
  const loadRuleDetail = async (ruleId: string) => { if (ruleDetails[ruleId]) { setSelectedRuleId(ruleId); setRuleEditor(toRuleEditorState(ruleDetails[ruleId])); return; } try { const response = await fetch(`${apiBaseUrl}/api/v1/rules/${ruleId}`); if (!response.ok) { setToast({ type: "error", message: await readErrorMessage(response) }); return; } const data = (await response.json()) as RuleDetail; setRuleDetails((current) => ({ ...current, [ruleId]: data })); setSelectedRuleId(ruleId); setRuleEditor(toRuleEditorState(data)); } catch { setToast({ type: "error", message: "Request failed" }); } };
  const saveRuleDetail = async () => { if (!selectedRuleId) return; const payload: RuleDetail = { id: selectedRuleId, name: ruleEditor.name.trim(), description: ruleEditor.description.trim(), category: ruleEditor.category.trim(), severity: ruleEditor.severity.trim(), scope: ruleEditor.scope.trim(), enforcement: ruleEditor.enforcement.trim(), type: ruleEditor.type.trim(), condition: ruleEditor.condition.trim(), effect: ruleEditor.effect.trim(), rationale: ruleEditor.rationale.trim(), evidence: ruleEditor.evidence.trim(), version: ruleEditor.version.trim() }; if (ruleEditor.validationType.trim() || ruleEditor.validationField.trim() || ruleEditor.validationOperator.trim() || ruleEditor.validationValue.trim()) { payload.validation = { type: ruleEditor.validationType.trim(), field: ruleEditor.validationField.trim(), operator: ruleEditor.validationOperator.trim(), value: ruleEditor.validationValue.trim() }; } try { const response = await fetch(`${apiBaseUrl}/api/v1/rules/${selectedRuleId}`, { method: "PUT", headers: { "Content-Type": "application/json" }, body: JSON.stringify(payload) }); if (!response.ok) { setToast({ type: "error", message: await readErrorMessage(response) }); return; } const data = (await response.json()) as RuleDetail; setRuleDetails((current) => ({ ...current, [selectedRuleId]: data })); setRuleEditor(toRuleEditorState(data)); setToast({ type: "success", message: "Rule saved" }); } catch { setToast({ type: "error", message: "Request failed" }); } };
  const loadRequirementDetail = async (reqId: string) => { if (requirementDetails[reqId]) { setSelectedRequirementId(reqId); setRequirementEditor(toRequirementEditorState(requirementDetails[reqId])); return; } try { const response = await fetch(`${apiBaseUrl}/api/v1/requirements/${reqId}`); if (!response.ok) { setToast({ type: "error", message: await readErrorMessage(response) }); return; } const data = (await response.json()) as RequirementDetail; setRequirementDetails((current) => ({ ...current, [reqId]: data })); setSelectedRequirementId(reqId); setRequirementEditor(toRequirementEditorState(data)); } catch { setToast({ type: "error", message: "Request failed" }); } };
  const saveRequirementDetail = async () => { if (!selectedRequirementId) return; const payload: RequirementDetail = { id: selectedRequirementId, name: requirementEditor.name.trim(), description: requirementEditor.description.trim(), category: requirementEditor.category.trim(), scope: requirementEditor.scope.trim(), source: requirementEditor.source.trim(), priority: requirementEditor.priority.trim(), version: requirementEditor.version.trim() }; try { const response = await fetch(`${apiBaseUrl}/api/v1/requirements/${selectedRequirementId}`, { method: "PUT", headers: { "Content-Type": "application/json" }, body: JSON.stringify(payload) }); if (!response.ok) { setToast({ type: "error", message: await readErrorMessage(response) }); return; } const data = (await response.json()) as RequirementDetail; setRequirementDetails((current) => ({ ...current, [selectedRequirementId]: data })); setRequirementEditor(toRequirementEditorState(data)); setToast({ type: "success", message: "Requirement saved" }); } catch { setToast({ type: "error", message: "Request failed" }); } };
  const addRule = async () => { if (!selectedProduct) return; const value = newRule.trim(); if (value === "") return; try { const response = await fetch(`${apiBaseUrl}/api/v1/products/${selectedProduct.id}/rules`, { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify({ value }) }); if (!response.ok) { setToast({ type: "error", message: await readErrorMessage(response) }); return; } setNewRule(""); await loadRules(); await loadProductSummary(selectedProduct.id); setToast({ type: "success", message: "Rule added" }); } catch { setToast({ type: "error", message: "Request failed" }); } };
  const removeRule = async (item: string) => { if (!selectedProduct) return; try { const response = await fetch(`${apiBaseUrl}/api/v1/products/${selectedProduct.id}/rules/${encodeURIComponent(item)}`, { method: "DELETE" }); if (!response.ok) { setToast({ type: "error", message: await readErrorMessage(response) }); return; } await loadRules(); await loadProductSummary(selectedProduct.id); setToast({ type: "success", message: "Rule removed" }); } catch { setToast({ type: "error", message: "Request failed" }); } };
  const loadVariants = async () => { if (!selectedProduct) return; try { const response = await fetch(`${apiBaseUrl}/api/v1/products/${selectedProduct.id}/variants`); if (!response.ok) return; const data = (await response.json()) as { items: Array<{ id: string; name: string; description?: string; context?: string; validity_conditions?: string[]; requirement_ids?: string[]; rule_ids?: string[] }> }; setVariants(data.items); } catch { /* ignore */ } };
  const addVariant = async () => { if (!selectedProduct) return; const payload: Record<string, unknown> = { id: newVariant.id.trim(), name: newVariant.name.trim(), description: newVariant.description.trim(), context: newVariant.context.trim() }; const vc = newVariant.validityConditions.split(",").map((s) => s.trim()).filter(Boolean); if (vc.length > 0) payload.validity_conditions = vc; const reqIds = newVariant.requirementIds.split(",").map((s) => s.trim()).filter(Boolean); if (reqIds.length > 0) payload.requirement_ids = reqIds; const rIds = newVariant.ruleIds.split(",").map((s) => s.trim()).filter(Boolean); if (rIds.length > 0) payload.rule_ids = rIds; if (!payload.id) { setToast({ type: "error", message: "Variant id is required" }); return; } try { const response = await fetch(`${apiBaseUrl}/api/v1/products/${selectedProduct.id}/variants`, { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify(payload) }); if (!response.ok) { setToast({ type: "error", message: await readErrorMessage(response) }); return; } setNewVariant({ id: "", name: "", description: "", context: "", validityConditions: "", requirementIds: "", ruleIds: "" }); await loadVariants(); await loadProductSummary(selectedProduct.id); setToast({ type: "success", message: "Variant added" }); } catch { setToast({ type: "error", message: "Request failed" }); } };
  const removeVariant = async (variantId: string) => { if (!selectedProduct) return; try { const response = await fetch(`${apiBaseUrl}/api/v1/products/${selectedProduct.id}/variants/${encodeURIComponent(variantId)}`, { method: "DELETE" }); if (!response.ok) { setToast({ type: "error", message: await readErrorMessage(response) }); return; } await loadVariants(); await loadProductSummary(selectedProduct.id); setToast({ type: "success", message: "Variant removed" }); } catch { setToast({ type: "error", message: "Request failed" }); } };

  const visibleRequirements = useMemo(() => {
    const source = requirements.length > 0 ? requirements : selectedProduct?.requirements ?? [];
    if (requirementsFilter.trim() === "") return source;
    const query = requirementsFilter.toLowerCase();
    return source.filter((item) => item.toLowerCase().includes(query));
  }, [requirements, requirementsFilter, selectedProduct]);

  const visibleRules = useMemo(() => {
    const source = rules.length > 0 ? rules : selectedProduct?.rules ?? [];
    const filtered = rulesFilter.trim() ? source.filter((item) => item.toLowerCase().includes(rulesFilter.toLowerCase())) : source;
    const sorted = [...filtered].sort((left, right) => left.localeCompare(right));
    return rulesSort === "asc" ? sorted : sorted.reverse();
  }, [rules, rulesFilter, rulesSort, selectedProduct]);

  const selectedRuleDetail = selectedRuleId ? ruleDetails[selectedRuleId] ?? null : null;

  const filteredProducts = useMemo(() => {
    if (!searchQuery.trim()) return products;
    const q = searchQuery.toLowerCase();
    return products.filter((p) => p.id.toLowerCase().includes(q) || p.name.toLowerCase().includes(q));
  }, [products, searchQuery]);

  const statusBadge = (valid: boolean | undefined, errors: number) => {
    if (valid === false || errors > 0) return <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-red-50 text-red-700 border border-red-100">Review Required</span>;
    if (valid === true) return <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-emerald-50 text-emerald-700 border border-emerald-100">Active</span>;
    return <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-slate-100 text-slate-600 border border-slate-200">Draft</span>;
  };

  return (
    <div className="min-h-screen bg-background font-sans">
      <aside className="fixed left-0 top-0 w-64 h-full bg-slate-900 border-r border-slate-800 shadow-xl flex flex-col z-50">
        <div className="px-6 py-6">
          <div className="flex items-center gap-3">
            <div className="w-8 h-8 bg-indigo-500 rounded-lg flex items-center justify-center">
              <span aria-hidden="true" className="material-symbols-outlined text-white text-sm">account_balance</span>
            </div>
            <div>
              <h1 className="text-xl font-black tracking-tighter text-white">Nomos</h1>
              <p className="text-[10px] text-slate-400 font-medium tracking-widest uppercase">Governance Platform</p>
            </div>
          </div>
        </div>
        <nav className="flex-1 px-3 space-y-1 overflow-y-auto">
          {navItems.map((item) => (
            <button key={item.key} type="button" onClick={() => setView(item.key)} className={`w-full flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm font-medium transition-all ${view === item.key ? "bg-indigo-500/10 text-indigo-400" : "text-slate-400 hover:text-slate-100 hover:bg-slate-800/50"}`}>
              <span aria-hidden="true" className="material-symbols-outlined text-lg">{item.icon}</span>
              {item.label}
            </button>
          ))}
          <div className="pt-4 mt-4 border-t border-slate-800 space-y-1">
            {secondaryNav.map((item) => (
              <button key={item.label} type="button" className="w-full flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm font-medium text-slate-400 hover:text-slate-100 hover:bg-slate-800/50 transition-all">
                <span aria-hidden="true" className="material-symbols-outlined text-lg">{item.icon}</span>
                {item.label}
              </button>
            ))}
          </div>
        </nav>
        <div className="px-6 py-4 border-t border-slate-800">
          <div className="flex items-center gap-3">
            <div className="w-8 h-8 rounded-full bg-indigo-600 flex items-center justify-center text-white text-xs font-bold">JD</div>
            <div className="flex-1 min-w-0">
              <p className="text-sm font-medium text-white truncate">John Doe</p>
              <p className="text-xs text-slate-500 truncate">Compliance Officer</p>
            </div>
          </div>
        </div>
      </aside>
      <div className="pl-64 min-h-screen">
        <header className="fixed top-0 right-0 left-64 h-16 z-40 bg-white/80 backdrop-blur-md border-b border-slate-200 flex items-center justify-between px-8">
          <div className="flex items-center flex-1 max-w-xl">
            <div className="relative w-full">
              <span aria-hidden="true" className="material-symbols-outlined absolute left-3 top-1/2 -translate-y-1/2 text-slate-400">search</span>
              <input type="text" className="w-full pl-10 pr-4 py-2 bg-slate-100 border-transparent focus:bg-white focus:ring-2 focus:ring-indigo-500 rounded-lg text-sm transition-all" placeholder="Search catalog, requirements or rules..." value={searchQuery} onChange={(e) => setSearchQuery(e.target.value)} />
            </div>
          </div>
          <div className="flex items-center gap-3">
            <button aria-label="Notifications" className="p-2 text-slate-500 hover:bg-slate-50 rounded-full transition-all"><span aria-hidden="true" className="material-symbols-outlined">notifications</span></button>
            <button aria-label="Help" className="p-2 text-slate-500 hover:bg-slate-50 rounded-full transition-all"><span aria-hidden="true" className="material-symbols-outlined">help_outline</span></button>
            <div className="h-6 w-px bg-slate-200 mx-1"></div>
            <div className="text-right">
              <p className="text-xs font-semibold text-slate-900">Workspace</p>
              <p className="text-[10px] text-slate-500">Corporate Banking</p>
            </div>
          </div>
        </header>
        <main className="mt-16 p-8">
          {toast && (
            <div className={`fixed top-20 right-8 z-50 px-4 py-3 rounded-xl shadow-panel text-sm font-medium border ${toast.type === "success" ? "bg-emerald-50 text-emerald-800 border-emerald-200" : "bg-red-50 text-red-800 border-red-200"}`} role="status">
              {toast.message}
            </div>
          )}

          {view === "dashboard" && (
            <div className="space-y-8">
              <div className="flex items-end justify-between">
                <div>
                  <h2 className="text-2xl font-semibold text-slate-900 tracking-tight">Governance Dashboard</h2>
                  <p className="text-sm text-slate-500 mt-1">Overview of your product governance landscape</p>
                </div>
                <button type="button" onClick={() => setView("catalog")} className="inline-flex items-center gap-2 px-4 py-2.5 bg-primary hover:bg-indigo-700 text-white text-sm font-medium rounded-lg transition-colors shadow-sm">
                  <span aria-hidden="true" className="material-symbols-outlined text-sm">add</span>
                  Create New Artefact
                </button>
              </div>
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
                {[
                  { label: "Total Products", value: products.length, icon: "inventory_2", color: "bg-indigo-50 text-indigo-600" },
                  { label: "Total Requirements", value: products.reduce((s, p) => s + ((p as unknown as Record<string, number>).requirements_count ?? 0), 0), icon: "rule", color: "bg-blue-50 text-blue-600" },
                  { label: "Total Rules", value: products.reduce((s, p) => s + ((p as unknown as Record<string, number>).rules_count ?? 0), 0), icon: "policy", color: "bg-violet-50 text-violet-600" },
                  { label: "Validation Health", value: "94%", icon: "check_circle", color: "bg-emerald-50 text-emerald-600" },
                ].map((m) => (
                  <div key={m.label} className="bg-white rounded-xl border border-slate-100 shadow-card p-6">
                    <div className="flex items-start justify-between">
                      <div>
                        <p className="text-sm font-medium text-slate-500">{m.label}</p>
                        <p className="text-2xl font-bold text-slate-900 mt-2">{m.value}</p>
                      </div>
                      <div className={`w-10 h-10 rounded-lg flex items-center justify-center ${m.color}`}>
                        <span aria-hidden="true" className="material-symbols-outlined">{m.icon}</span>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
              <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
                <div className="lg:col-span-2 bg-white rounded-xl border border-slate-100 shadow-card p-6">
                  <div className="flex items-center justify-between mb-6">
                    <div>
                      <h3 className="text-lg font-semibold text-slate-900">Reference Product</h3>
                      <p className="text-sm text-slate-500">Most recently updated product in the catalog</p>
                    </div>
                    {products[0] && (
                      <button type="button" onClick={() => void loadProductDetail(products[0].id)} className="inline-flex items-center gap-1.5 px-3 py-1.5 text-sm font-medium text-primary hover:bg-indigo-50 rounded-lg transition-colors border border-indigo-200">View Details</button>
                    )}
                  </div>
                  {products[0] ? (
                    <div className="flex items-center gap-3">
                      <div className="w-10 h-10 rounded-lg bg-indigo-100 flex items-center justify-center text-indigo-600">
                        <span aria-hidden="true" className="material-symbols-outlined">inventory_2</span>
                      </div>
                      <div>
                        <p className="text-base font-semibold text-slate-900">{products[0].name}</p>
                        <p className="text-xs text-slate-500 font-mono">{products[0].id} &middot; v{products[0].version}</p>
                      </div>
                    </div>
                  ) : (
                    <p className="text-sm text-slate-400">No products yet.</p>
                  )}
                </div>
                <div className="bg-white rounded-xl border border-slate-100 shadow-card p-6">
                  <h3 className="text-lg font-semibold text-slate-900 mb-4">Recent Activity</h3>
                  <div className="space-y-4">
                    {products.slice(0, 4).map((p) => (
                      <div key={p.id} className="flex items-start gap-3">
                        <div className="w-8 h-8 rounded-full bg-slate-100 flex items-center justify-center text-slate-500 text-xs">
                          <span aria-hidden="true" className="material-symbols-outlined text-sm">inventory_2</span>
                        </div>
                        <div>
                          <p className="text-sm font-medium text-slate-900">{p.name}</p>
                          <p className="text-xs text-slate-500">Updated v{p.version}</p>
                        </div>
                      </div>
                    ))}
                    {products.length === 0 && <p className="text-sm text-slate-400">No recent activity.</p>}
                  </div>
                </div>
              </div>
            </div>
          )}

          {view === "catalog" && (
            <div className="space-y-6">
              <div className="flex items-end justify-between">
                <div>
                  <h2 className="text-2xl font-semibold text-slate-900 tracking-tight">Product Catalog</h2>
                  <p className="text-sm text-slate-500 mt-1">Manage and govern your product definitions</p>
                </div>
                <button type="button" onClick={() => { setNewProduct({ id: "", name: "", version: "", description: "" }); setSelectedProduct(null); setView("detail"); }} className="inline-flex items-center gap-2 px-4 py-2.5 bg-primary hover:bg-indigo-700 text-white text-sm font-medium rounded-lg transition-colors shadow-sm">
                  <span aria-hidden="true" className="material-symbols-outlined text-sm">add</span>
                  Create product
                </button>
              </div>
              <div className="bg-white rounded-xl border border-slate-100 shadow-card p-6">
                <h3 className="text-sm font-semibold text-slate-900 uppercase tracking-wider mb-4">New Product</h3>
                <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
                  <input type="text" placeholder="New product id" className="px-3 py-2 bg-slate-50 border border-slate-200 rounded-md text-sm focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 outline-none transition-all" value={newProduct.id} onChange={(e) => setNewProduct((c) => ({ ...c, id: e.target.value }))} />
                  <input type="text" placeholder="New product name" className="px-3 py-2 bg-slate-50 border border-slate-200 rounded-md text-sm focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 outline-none transition-all" value={newProduct.name} onChange={(e) => setNewProduct((c) => ({ ...c, name: e.target.value }))} />
                  <input type="text" placeholder="New product version" className="px-3 py-2 bg-slate-50 border border-slate-200 rounded-md text-sm focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 outline-none transition-all" value={newProduct.version} onChange={(e) => setNewProduct((c) => ({ ...c, version: e.target.value }))} />
                  <button type="button" onClick={() => void createProduct()} className="inline-flex items-center justify-center gap-2 px-4 py-2 bg-primary hover:bg-indigo-700 text-white text-sm font-medium rounded-lg transition-colors">Create product</button>
                </div>
                <input type="text" placeholder="New product description" className="mt-3 w-full px-3 py-2 bg-slate-50 border border-slate-200 rounded-md text-sm focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 outline-none transition-all" value={newProduct.description} onChange={(e) => setNewProduct((c) => ({ ...c, description: e.target.value }))} />
              </div>
              <div className="bg-white rounded-xl border border-slate-100 shadow-card overflow-hidden">
                <div className="overflow-x-auto">
                  <table className="w-full text-left text-sm">
                    <thead className="bg-slate-50 border-b border-slate-100">
                      <tr>
                        <th className="px-6 py-3 font-semibold text-slate-600">Product ID</th>
                        <th className="px-6 py-3 font-semibold text-slate-600">Name</th>
                        <th className="px-6 py-3 font-semibold text-slate-600">Status</th>
                        <th className="px-6 py-3 font-semibold text-slate-600">Version</th>
                        <th className="px-6 py-3 font-semibold text-slate-600">Requirements</th>
                        <th className="px-6 py-3 font-semibold text-slate-600">Rules</th>
                        <th className="px-6 py-3 font-semibold text-slate-600 text-right">Actions</th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-slate-100">
                      {filteredProducts.map((product) => (
                        <tr key={product.id} className="hover:bg-slate-50/50 transition-colors">
                          <td className="px-6 py-4 font-mono text-xs text-slate-500">{product.id}</td>
                          <td className="px-6 py-4"><p className="font-medium text-slate-900">{product.name}</p></td>
                          <td className="px-6 py-4">{statusBadge((product as unknown as Record<string, boolean>).validation_is_valid, (product as unknown as Record<string, number>).validation_error_count ?? 0)}</td>
                          <td className="px-6 py-4 text-slate-600">{product.version}</td>
                          <td className="px-6 py-4 text-slate-600">{(product as unknown as Record<string, number>).requirements_count ?? 0}</td>
                          <td className="px-6 py-4 text-slate-600">{(product as unknown as Record<string, number>).rules_count ?? 0}</td>
                          <td className="px-6 py-4 text-right">
                            <button type="button" onClick={() => void loadProductDetail(product.id)} className="inline-flex items-center gap-1.5 px-3 py-1.5 text-sm font-medium text-primary hover:bg-indigo-50 rounded-lg transition-colors" aria-label={`Open product ${product.name}`}>Details</button>
                          </td>
                        </tr>
                      ))}
                      {filteredProducts.length === 0 && (
                        <tr><td colSpan={7} className="px-6 py-12 text-center text-slate-400">No products found.</td></tr>
                      )}
                    </tbody>
                  </table>
                </div>
              </div>
            </div>
          )}

          {view === "detail" && selectedProduct && (
            <div className="space-y-6">
              <div className="flex items-center gap-2 text-sm">
                <button type="button" onClick={() => setView("catalog")} className="text-slate-500 hover:text-primary transition-colors">Product Catalog</button>
                <span className="text-slate-300">/</span>
                <span className="text-slate-900 font-medium">{selectedProduct.name}</span>
              </div>
              <div className="bg-white rounded-xl border border-slate-100 shadow-card p-6">
                <div className="flex items-start justify-between">
                  <div className="flex items-center gap-4">
                    <div className="w-12 h-12 rounded-xl bg-indigo-100 flex items-center justify-center text-indigo-600">
                      <span aria-hidden="true" className="material-symbols-outlined text-2xl">inventory_2</span>
                    </div>
                    <div>
                      <h2 className="text-xl font-semibold text-slate-900">{selectedProduct.name}</h2>
                      <p className="text-sm text-slate-500 mt-0.5">{selectedProduct.id} &middot; Version {selectedProduct.version ?? "n/a"}</p>
                      {selectedProduct.description && <p className="text-sm text-slate-600 mt-1">{selectedProduct.description}</p>}
                    </div>
                  </div>
                  <div className="flex items-center gap-2">
                    <button type="button" onClick={() => void saveProduct()} className="inline-flex items-center gap-1.5 px-4 py-2 bg-primary hover:bg-indigo-700 text-white text-sm font-medium rounded-lg transition-colors">
                      <span aria-hidden="true" className="material-symbols-outlined text-sm">save</span>
                      Save product
                    </button>
                    <button type="button" onClick={() => void removeProduct()} className="inline-flex items-center gap-1.5 px-4 py-2 bg-white hover:bg-red-50 text-red-600 text-sm font-medium rounded-lg transition-colors border border-slate-200">
                      <span aria-hidden="true" className="material-symbols-outlined text-sm">delete</span>
                      Delete
                    </button>
                  </div>
                </div>
                <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mt-6">
                  <input type="text" placeholder="Edit product name" className="px-3 py-2 bg-slate-50 border border-slate-200 rounded-md text-sm focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 outline-none transition-all" value={editProduct.name} onChange={(e) => setEditProduct((c) => ({ ...c, name: e.target.value }))} />
                  <input type="text" placeholder="Edit product version" className="px-3 py-2 bg-slate-50 border border-slate-200 rounded-md text-sm focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 outline-none transition-all" value={editProduct.version} onChange={(e) => setEditProduct((c) => ({ ...c, version: e.target.value }))} />
                  <input type="text" placeholder="Edit product description" className="px-3 py-2 bg-slate-50 border border-slate-200 rounded-md text-sm focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 outline-none transition-all" value={editProduct.description} onChange={(e) => setEditProduct((c) => ({ ...c, description: e.target.value }))} />
                </div>
                {productSummary && (
                  <div className="grid grid-cols-3 gap-4 mt-6">
                    <div className="bg-slate-50 rounded-lg p-4">
                      <p className="text-xs font-medium text-slate-500 uppercase tracking-wider">Requirements</p>
                      <p className="text-xl font-bold text-slate-900 mt-1">{productSummary.requirements_count}</p>
                    </div>
                    <div className="bg-slate-50 rounded-lg p-4">
                      <p className="text-xs font-medium text-slate-500 uppercase tracking-wider">Rules</p>
                      <p className="text-xl font-bold text-slate-900 mt-1">{productSummary.rules_count}</p>
                    </div>
                    <div className="bg-slate-50 rounded-lg p-4">
                      <p className="text-xs font-medium text-slate-500 uppercase tracking-wider">Validation Errors</p>
                      <p className="text-xl font-bold text-slate-900 mt-1">{productSummary.validation_error_count}</p>
                    </div>
                  </div>
                )}
              </div>
              <div className="bg-white rounded-xl border border-slate-100 shadow-card overflow-hidden">
                <div className="border-b border-slate-100">
                  <div className="flex gap-1 px-4">
                    {(["requirements", "rules", "variants"] as const).map((tab) => (
                      <button key={tab} type="button" onClick={() => setActiveTab(tab)} className={`px-4 py-3 text-sm font-medium border-b-2 transition-all capitalize ${activeTab === tab ? "border-primary text-primary" : "border-transparent text-slate-500 hover:text-slate-700"}`} aria-label={`${tab} tab`}>{tab}</button>
                    ))}
                  </div>
                </div>
                <div className="p-6">
                  {activeTab === "requirements" && (
                    <div className="space-y-6">
                      <div className="flex items-center gap-4">
                        <button type="button" onClick={() => void loadRequirements()} className="inline-flex items-center gap-1.5 px-3 py-2 bg-white border border-slate-200 hover:bg-slate-50 text-slate-700 text-sm font-medium rounded-lg transition-colors">
                          <span aria-hidden="true" className="material-symbols-outlined text-sm">refresh</span>
                          Load requirements
                        </button>
                        <div className="relative flex-1 max-w-xs">
                          <span aria-hidden="true" className="material-symbols-outlined absolute left-2.5 top-1/2 -translate-y-1/2 text-slate-400 text-sm">search</span>
                          <input type="text" placeholder="Filter requirements" className="w-full pl-9 pr-3 py-2 bg-slate-50 border border-slate-200 rounded-md text-sm focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 outline-none transition-all" value={requirementsFilter} onChange={(e) => setRequirementsFilter(e.target.value)} />
                        </div>
                      </div>
                      <div className="flex gap-3">
                        <input type="text" placeholder="New requirement" className="flex-1 px-3 py-2 bg-slate-50 border border-slate-200 rounded-md text-sm focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 outline-none transition-all" value={newRequirement} onChange={(e) => setNewRequirement(e.target.value)} />
                        <button type="button" onClick={() => void addRequirement()} className="inline-flex items-center gap-1.5 px-4 py-2 bg-primary hover:bg-indigo-700 text-white text-sm font-medium rounded-lg transition-colors">Add requirement</button>
                      </div>
                      <div className="border border-slate-100 rounded-xl overflow-hidden">
                        <table className="w-full text-left text-sm">
                          <thead className="bg-slate-50"><tr><th className="px-4 py-3 font-semibold text-slate-600">Requirement</th><th className="px-4 py-3 font-semibold text-slate-600 text-right">Actions</th></tr></thead>
                          <tbody className="divide-y divide-slate-100">
                            {visibleRequirements.map((req) => (
                              <tr key={req} className="hover:bg-slate-50/50 transition-colors">
                                <td className="px-4 py-3 text-slate-900">{req}</td>
                                <td className="px-4 py-3 text-right">
                                  <div className="inline-flex gap-2">
                                    <button type="button" onClick={() => void loadRequirementDetail(req)} className="px-2.5 py-1 text-xs font-medium text-primary hover:bg-indigo-50 rounded-md transition-colors border border-indigo-100">Show requirement details {req}</button>
                                    <button type="button" onClick={() => void removeRequirement(req)} className="px-2.5 py-1 text-xs font-medium text-red-600 hover:bg-red-50 rounded-md transition-colors border border-red-100">Remove requirement {req}</button>
                                  </div>
                                </td>
                              </tr>
                            ))}
                            {visibleRequirements.length === 0 && <tr><td colSpan={2} className="px-4 py-8 text-center text-slate-400">No requirements.</td></tr>}
                          </tbody>
                        </table>
                      </div>
                      {selectedRequirementId && requirementDetails[selectedRequirementId] && (
                        <div className="bg-surface-container-low rounded-xl border border-slate-100 p-6 space-y-4">
                          <h4 className="text-sm font-semibold text-slate-900">Requirement detail: {requirementDetails[selectedRequirementId].name}</h4>
                          <p className="text-xs text-slate-500 font-mono">ID: {requirementDetails[selectedRequirementId].id}</p>
                          <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
                            <input type="text" placeholder="Edit requirement name" className="px-3 py-2 bg-white border border-slate-200 rounded-md text-sm focus:ring-2 focus:ring-indigo-500 outline-none" value={requirementEditor.name} onChange={(e) => setRequirementEditor((c) => ({ ...c, name: e.target.value }))} />
                            <input type="text" placeholder="Edit requirement description" className="px-3 py-2 bg-white border border-slate-200 rounded-md text-sm focus:ring-2 focus:ring-indigo-500 outline-none" value={requirementEditor.description} onChange={(e) => setRequirementEditor((c) => ({ ...c, description: e.target.value }))} />
                          </div>
                          <div className="grid grid-cols-1 md:grid-cols-3 gap-3">
                            <input type="text" placeholder="Edit requirement category" className="px-3 py-2 bg-white border border-slate-200 rounded-md text-sm focus:ring-2 focus:ring-indigo-500 outline-none" value={requirementEditor.category} onChange={(e) => setRequirementEditor((c) => ({ ...c, category: e.target.value }))} />
                            <input type="text" placeholder="Edit requirement scope" className="px-3 py-2 bg-white border border-slate-200 rounded-md text-sm focus:ring-2 focus:ring-indigo-500 outline-none" value={requirementEditor.scope} onChange={(e) => setRequirementEditor((c) => ({ ...c, scope: e.target.value }))} />
                            <input type="text" placeholder="Edit requirement source" className="px-3 py-2 bg-white border border-slate-200 rounded-md text-sm focus:ring-2 focus:ring-indigo-500 outline-none" value={requirementEditor.source} onChange={(e) => setRequirementEditor((c) => ({ ...c, source: e.target.value }))} />
                          </div>
                          <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
                            <input type="text" placeholder="Edit requirement priority (low/medium/high/critical)" className="px-3 py-2 bg-white border border-slate-200 rounded-md text-sm focus:ring-2 focus:ring-indigo-500 outline-none" value={requirementEditor.priority} onChange={(e) => setRequirementEditor((c) => ({ ...c, priority: e.target.value }))} />
                            <input type="text" placeholder="Edit requirement version" className="px-3 py-2 bg-white border border-slate-200 rounded-md text-sm focus:ring-2 focus:ring-indigo-500 outline-none" value={requirementEditor.version} onChange={(e) => setRequirementEditor((c) => ({ ...c, version: e.target.value }))} />
                          </div>
                          <button type="button" onClick={() => void saveRequirementDetail()} className="inline-flex items-center gap-1.5 px-4 py-2 bg-primary hover:bg-indigo-700 text-white text-sm font-medium rounded-lg transition-colors">Save requirement</button>
                        </div>
                      )}
                    </div>
                  )}
                  {activeTab === "rules" && (
                    <div className="space-y-6">
                      <div className="flex items-center gap-4">
                        <button type="button" onClick={() => void loadRules()} className="inline-flex items-center gap-1.5 px-3 py-2 bg-white border border-slate-200 hover:bg-slate-50 text-slate-700 text-sm font-medium rounded-lg transition-colors">
                          <span aria-hidden="true" className="material-symbols-outlined text-sm">refresh</span>
                          Load rules
                        </button>
                        <div className="relative flex-1 max-w-xs">
                          <span aria-hidden="true" className="material-symbols-outlined absolute left-2.5 top-1/2 -translate-y-1/2 text-slate-400 text-sm">search</span>
                          <input type="text" placeholder="Filter rules" className="w-full pl-9 pr-3 py-2 bg-slate-50 border border-slate-200 rounded-md text-sm focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 outline-none transition-all" value={rulesFilter} onChange={(e) => setRulesFilter(e.target.value)} />
                        </div>
                        <div className="flex items-center gap-2">
                          <label htmlFor="rules-sort" className="text-sm text-slate-500">Sort</label>
                          <select id="rules-sort" value={rulesSort} onChange={(e) => setRulesSort(e.target.value as "asc" | "desc")} className="px-2 py-1.5 bg-slate-50 border border-slate-200 rounded-md text-sm focus:ring-2 focus:ring-indigo-500 outline-none">
                            <option value="asc">asc</option>
                            <option value="desc">desc</option>
                          </select>
                        </div>
                      </div>
                      <div className="flex gap-3">
                        <input type="text" placeholder="New rule" className="flex-1 px-3 py-2 bg-slate-50 border border-slate-200 rounded-md text-sm focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 outline-none transition-all" value={newRule} onChange={(e) => setNewRule(e.target.value)} />
                        <button type="button" onClick={() => void addRule()} className="inline-flex items-center gap-1.5 px-4 py-2 bg-primary hover:bg-indigo-700 text-white text-sm font-medium rounded-lg transition-colors">Add rule</button>
                      </div>
                      <div className="border border-slate-100 rounded-xl overflow-hidden">
                        <table className="w-full text-left text-sm">
                          <thead className="bg-slate-50"><tr><th className="px-4 py-3 font-semibold text-slate-600">Rule</th><th className="px-4 py-3 font-semibold text-slate-600 text-right">Actions</th></tr></thead>
                          <tbody className="divide-y divide-slate-100">
                            {visibleRules.map((rule) => (
                              <tr key={rule} className="hover:bg-slate-50/50 transition-colors" data-testid="rules-item">
                                <td className="px-4 py-3 text-slate-900">{rule}</td>
                                <td className="px-4 py-3 text-right">
                                  <div className="inline-flex gap-2">
                                    <button type="button" onClick={() => void loadRuleDetail(rule)} className="px-2.5 py-1 text-xs font-medium text-primary hover:bg-indigo-50 rounded-md transition-colors border border-indigo-100">Show rule details {rule}</button>
                                    <button type="button" onClick={() => void removeRule(rule)} className="px-2.5 py-1 text-xs font-medium text-red-600 hover:bg-red-50 rounded-md transition-colors border border-red-100">Remove rule {rule}</button>
                                  </div>
                                </td>
                              </tr>
                            ))}
                            {visibleRules.length === 0 && <tr><td colSpan={2} className="px-4 py-8 text-center text-slate-400">No rules.</td></tr>}
                          </tbody>
                        </table>
                      </div>
                      {selectedRuleDetail && (
                        <div className="bg-surface-container-low rounded-xl border border-slate-100 p-6 space-y-4">
                          <h4 className="text-sm font-semibold text-slate-900">Rule detail: {selectedRuleDetail.name}</h4>
                          <p className="text-xs text-slate-500 font-mono">ID: {selectedRuleDetail.id}</p>
                          <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
                            <input type="text" placeholder="Edit rule name" className="px-3 py-2 bg-white border border-slate-200 rounded-md text-sm focus:ring-2 focus:ring-indigo-500 outline-none" value={ruleEditor.name} onChange={(e) => setRuleEditor((c) => ({ ...c, name: e.target.value }))} />
                            <input type="text" placeholder="Edit rule description" className="px-3 py-2 bg-white border border-slate-200 rounded-md text-sm focus:ring-2 focus:ring-indigo-500 outline-none" value={ruleEditor.description} onChange={(e) => setRuleEditor((c) => ({ ...c, description: e.target.value }))} />
                          </div>
                          <div className="grid grid-cols-1 md:grid-cols-3 gap-3">
                            <input type="text" placeholder="Edit rule category" className="px-3 py-2 bg-white border border-slate-200 rounded-md text-sm focus:ring-2 focus:ring-indigo-500 outline-none" value={ruleEditor.category} onChange={(e) => setRuleEditor((c) => ({ ...c, category: e.target.value }))} />
                            <input type="text" placeholder="Edit rule severity" className="px-3 py-2 bg-white border border-slate-200 rounded-md text-sm focus:ring-2 focus:ring-indigo-500 outline-none" value={ruleEditor.severity} onChange={(e) => setRuleEditor((c) => ({ ...c, severity: e.target.value }))} />
                            <input type="text" placeholder="Edit rule scope" className="px-3 py-2 bg-white border border-slate-200 rounded-md text-sm focus:ring-2 focus:ring-indigo-500 outline-none" value={ruleEditor.scope} onChange={(e) => setRuleEditor((c) => ({ ...c, scope: e.target.value }))} />
                          </div>
                          <input type="text" placeholder="Edit rule enforcement" className="w-full px-3 py-2 bg-white border border-slate-200 rounded-md text-sm focus:ring-2 focus:ring-indigo-500 outline-none" value={ruleEditor.enforcement} onChange={(e) => setRuleEditor((c) => ({ ...c, enforcement: e.target.value }))} />
                          <div className="grid grid-cols-1 md:grid-cols-3 gap-3">
                            <input type="text" placeholder="Edit rule type (must/must_not/derive/allow/deny)" className="px-3 py-2 bg-white border border-slate-200 rounded-md text-sm focus:ring-2 focus:ring-indigo-500 outline-none" value={ruleEditor.type} onChange={(e) => setRuleEditor((c) => ({ ...c, type: e.target.value }))} />
                            <input type="text" placeholder="Edit rule condition" className="px-3 py-2 bg-white border border-slate-200 rounded-md text-sm focus:ring-2 focus:ring-indigo-500 outline-none" value={ruleEditor.condition} onChange={(e) => setRuleEditor((c) => ({ ...c, condition: e.target.value }))} />
                            <input type="text" placeholder="Edit rule effect" className="px-3 py-2 bg-white border border-slate-200 rounded-md text-sm focus:ring-2 focus:ring-indigo-500 outline-none" value={ruleEditor.effect} onChange={(e) => setRuleEditor((c) => ({ ...c, effect: e.target.value }))} />
                          </div>
                          <div className="grid grid-cols-1 md:grid-cols-3 gap-3">
                            <input type="text" placeholder="Edit rule rationale" className="px-3 py-2 bg-white border border-slate-200 rounded-md text-sm focus:ring-2 focus:ring-indigo-500 outline-none" value={ruleEditor.rationale} onChange={(e) => setRuleEditor((c) => ({ ...c, rationale: e.target.value }))} />
                            <input type="text" placeholder="Edit rule evidence" className="px-3 py-2 bg-white border border-slate-200 rounded-md text-sm focus:ring-2 focus:ring-indigo-500 outline-none" value={ruleEditor.evidence} onChange={(e) => setRuleEditor((c) => ({ ...c, evidence: e.target.value }))} />
                            <input type="text" placeholder="Edit rule version" className="px-3 py-2 bg-white border border-slate-200 rounded-md text-sm focus:ring-2 focus:ring-indigo-500 outline-none" value={ruleEditor.version} onChange={(e) => setRuleEditor((c) => ({ ...c, version: e.target.value }))} />
                          </div>
                          <div className="grid grid-cols-1 md:grid-cols-4 gap-3">
                            <input type="text" placeholder="Edit validation type" className="px-3 py-2 bg-white border border-slate-200 rounded-md text-sm focus:ring-2 focus:ring-indigo-500 outline-none" value={ruleEditor.validationType} onChange={(e) => setRuleEditor((c) => ({ ...c, validationType: e.target.value }))} />
                            <input type="text" placeholder="Edit validation field" className="px-3 py-2 bg-white border border-slate-200 rounded-md text-sm focus:ring-2 focus:ring-indigo-500 outline-none" value={ruleEditor.validationField} onChange={(e) => setRuleEditor((c) => ({ ...c, validationField: e.target.value }))} />
                            <input type="text" placeholder="Edit validation operator" className="px-3 py-2 bg-white border border-slate-200 rounded-md text-sm focus:ring-2 focus:ring-indigo-500 outline-none" value={ruleEditor.validationOperator} onChange={(e) => setRuleEditor((c) => ({ ...c, validationOperator: e.target.value }))} />
                            <input type="text" placeholder="Edit validation value" className="px-3 py-2 bg-white border border-slate-200 rounded-md text-sm focus:ring-2 focus:ring-indigo-500 outline-none" value={ruleEditor.validationValue} onChange={(e) => setRuleEditor((c) => ({ ...c, validationValue: e.target.value }))} />
                          </div>
                          <button type="button" onClick={() => void saveRuleDetail()} className="inline-flex items-center gap-1.5 px-4 py-2 bg-primary hover:bg-indigo-700 text-white text-sm font-medium rounded-lg transition-colors">Save rule</button>
                        </div>
                      )}
                    </div>
                  )}
                  {activeTab === "variants" && (
                    <div className="space-y-6">
                      <div className="flex items-center gap-4">
                        <button type="button" onClick={() => void loadVariants()} className="inline-flex items-center gap-1.5 px-3 py-2 bg-white border border-slate-200 hover:bg-slate-50 text-slate-700 text-sm font-medium rounded-lg transition-colors">
                          <span aria-hidden="true" className="material-symbols-outlined text-sm">refresh</span>
                          Load variants
                        </button>
                      </div>
                      <div className="grid grid-cols-1 md:grid-cols-4 gap-3">
                        <input type="text" placeholder="New variant id" className="px-3 py-2 bg-slate-50 border border-slate-200 rounded-md text-sm focus:ring-2 focus:ring-indigo-500 outline-none" value={newVariant.id} onChange={(e) => setNewVariant((c) => ({ ...c, id: e.target.value }))} />
                        <input type="text" placeholder="New variant name" className="px-3 py-2 bg-slate-50 border border-slate-200 rounded-md text-sm focus:ring-2 focus:ring-indigo-500 outline-none" value={newVariant.name} onChange={(e) => setNewVariant((c) => ({ ...c, name: e.target.value }))} />
                        <input type="text" placeholder="New variant description" className="px-3 py-2 bg-slate-50 border border-slate-200 rounded-md text-sm focus:ring-2 focus:ring-indigo-500 outline-none" value={newVariant.description} onChange={(e) => setNewVariant((c) => ({ ...c, description: e.target.value }))} />
                        <input type="text" placeholder="New variant context" className="px-3 py-2 bg-slate-50 border border-slate-200 rounded-md text-sm focus:ring-2 focus:ring-indigo-500 outline-none" value={newVariant.context} onChange={(e) => setNewVariant((c) => ({ ...c, context: e.target.value }))} />
                        <input type="text" placeholder="Validity conditions (comma separated)" className="px-3 py-2 bg-slate-50 border border-slate-200 rounded-md text-sm focus:ring-2 focus:ring-indigo-500 outline-none" value={newVariant.validityConditions} onChange={(e) => setNewVariant((c) => ({ ...c, validityConditions: e.target.value }))} />
                        <input type="text" placeholder="Requirement ids (comma separated)" className="px-3 py-2 bg-slate-50 border border-slate-200 rounded-md text-sm focus:ring-2 focus:ring-indigo-500 outline-none" value={newVariant.requirementIds} onChange={(e) => setNewVariant((c) => ({ ...c, requirementIds: e.target.value }))} />
                        <input type="text" placeholder="Rule ids (comma separated)" className="px-3 py-2 bg-slate-50 border border-slate-200 rounded-md text-sm focus:ring-2 focus:ring-indigo-500 outline-none" value={newVariant.ruleIds} onChange={(e) => setNewVariant((c) => ({ ...c, ruleIds: e.target.value }))} />
                        <button type="button" onClick={() => void addVariant()} className="inline-flex items-center justify-center gap-1.5 px-4 py-2 bg-primary hover:bg-indigo-700 text-white text-sm font-medium rounded-lg transition-colors">Add variant</button>
                      </div>
                      <div className="border border-slate-100 rounded-xl overflow-hidden">
                        <table className="w-full text-left text-sm">
                          <thead className="bg-slate-50"><tr><th className="px-4 py-3 font-semibold text-slate-600">Variant</th><th className="px-4 py-3 font-semibold text-slate-600">Context</th><th className="px-4 py-3 font-semibold text-slate-600 text-right">Actions</th></tr></thead>
                          <tbody className="divide-y divide-slate-100">
                            {variants.map((variant) => (
                              <tr key={variant.id} className="hover:bg-slate-50/50 transition-colors">
                                <td className="px-4 py-3"><p className="font-medium text-slate-900">{variant.name}</p><p className="text-xs text-slate-500 font-mono">{variant.id}</p></td>
                                <td className="px-4 py-3 text-slate-600">{variant.context || "\u2014"}</td>
                                <td className="px-4 py-3 text-right"><button type="button" onClick={() => void removeVariant(variant.id)} className="px-2.5 py-1 text-xs font-medium text-red-600 hover:bg-red-50 rounded-md transition-colors border border-red-100">Remove variant {variant.id}</button></td>
                              </tr>
                            ))}
                            {variants.length === 0 && <tr><td colSpan={3} className="px-4 py-8 text-center text-slate-400">No variants.</td></tr>}
                          </tbody>
                        </table>
                      </div>
                    </div>
                  )}
                </div>
              </div>
              <div className="bg-white rounded-xl border border-slate-100 shadow-card p-6">
                <h3 className="text-lg font-semibold text-slate-900 mb-4">Validation</h3>
                {selectedProduct.validation?.errors?.length ? (
                  <ul className="space-y-2">
                    {selectedProduct.validation.errors.map((err) => (
                      <li key={err} className="flex items-start gap-2 text-sm text-red-700 bg-red-50 border border-red-100 rounded-lg px-4 py-3"><span aria-hidden="true" className="material-symbols-outlined text-sm mt-0.5">error</span>{err}</li>
                    ))}
                  </ul>
                ) : (
                  <div className="flex items-center gap-2 text-sm text-emerald-700 bg-emerald-50 border border-emerald-100 rounded-lg px-4 py-3"><span aria-hidden="true" className="material-symbols-outlined text-sm">check_circle</span>No validation errors</div>
                )}
              </div>
            </div>
          )}
          {view === "detail" && !selectedProduct && (
            <div className="bg-white rounded-xl border border-slate-100 shadow-card p-12 text-center">
              <div className="w-16 h-16 rounded-full bg-slate-100 flex items-center justify-center mx-auto mb-4"><span aria-hidden="true" className="material-symbols-outlined text-2xl text-slate-400">inventory_2</span></div>
              <h3 className="text-lg font-semibold text-slate-900">No product selected</h3>
              <p className="text-sm text-slate-500 mt-1">Select a product from the catalog to view details.</p>
              <button type="button" onClick={() => setView("catalog")} className="mt-4 inline-flex items-center gap-1.5 px-4 py-2 bg-primary hover:bg-indigo-700 text-white text-sm font-medium rounded-lg transition-colors">Go to Catalog</button>
            </div>
          )}
        </main>
      </div>
    </div>
  );
}

export default App;

