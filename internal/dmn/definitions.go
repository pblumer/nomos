package dmn

import (
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/nomos/nomos/internal/model"
)

// ParseDefinitions parses a DMN 1.x XML document into the full
// model.DMNDefinitions representation (DRG + DRD).
//
// It is the round-trip counterpart to dmn-js: dmn-js authors the XML,
// Nomos persists it as decision.dmn, and ParseDefinitions yields the
// structured DRG that the rest of Nomos reads.
func ParseDefinitions(xmlData []byte) (*model.DMNDefinitions, error) {
	var raw xmlDefinitionsFull
	if err := xml.Unmarshal(xmlData, &raw); err != nil {
		return nil, fmt.Errorf("dmn xml parse: %w", err)
	}
	out := &model.DMNDefinitions{
		ID:              raw.ID,
		Name:            raw.Name,
		Namespace:       raw.Namespace,
		ExpressionLang:  raw.ExpressionLang,
		TypeLang:        raw.TypeLang,
		ExporterName:    raw.Exporter,
		ExporterVersion: raw.ExporterVersion,
		Description:     strings.TrimSpace(raw.Description),
	}
	for _, it := range raw.ItemDefinitions {
		out.ItemDefinitions = append(out.ItemDefinitions, mapItemDefinition(it))
	}
	for _, imp := range raw.Imports {
		out.Imports = append(out.Imports, model.DMNImport{
			Name: imp.Name, Namespace: imp.Namespace, ImportType: imp.ImportType, LocationURI: imp.LocationURI,
		})
	}
	for _, in := range raw.InputData {
		out.InputData = append(out.InputData, model.DMNInputData{
			ID:          in.ID,
			Name:        in.Name,
			Label:       in.Label,
			Description: strings.TrimSpace(in.Description),
			Variable:    mapInformationItem(in.Variable),
		})
	}
	for _, ks := range raw.KnowledgeSources {
		out.KnowledgeSource = append(out.KnowledgeSource, model.DMNKnowledgeSource{
			ID:                    ks.ID,
			Name:                  ks.Name,
			Label:                 ks.Label,
			Description:           strings.TrimSpace(ks.Description),
			Type:                  strings.TrimSpace(ks.Type),
			LocationURI:           strings.TrimSpace(ks.LocationURI),
			Owner:                 ks.Owner.HRef,
			AuthorityRequirements: mapAuthorityReqs(ks.AuthorityRequirements),
		})
	}
	for _, bkm := range raw.BKMs {
		out.BKMs = append(out.BKMs, model.DMNBusinessKnowledgeModel{
			ID:                    bkm.ID,
			Name:                  bkm.Name,
			Label:                 bkm.Label,
			Description:           strings.TrimSpace(bkm.Description),
			Variable:              mapInformationItem(bkm.Variable),
			EncapsulatedLogic:     mapFunctionDefinition(bkm.EncapsulatedLogic),
			KnowledgeRequirements: mapKnowledgeReqs(bkm.KnowledgeRequirements),
			AuthorityRequirements: mapAuthorityReqs(bkm.AuthorityRequirements),
		})
	}
	for _, ds := range raw.DecisionServices {
		out.DecisionService = append(out.DecisionService, model.DMNDecisionService{
			ID:                    ds.ID,
			Name:                  ds.Name,
			Label:                 ds.Label,
			Description:           strings.TrimSpace(ds.Description),
			Variable:              mapInformationItem(ds.Variable),
			OutputDecisions:       hrefs(ds.OutputDecisions),
			EncapsulatedDecisions: hrefs(ds.EncapsulatedDecisions),
			InputDecisions:        hrefs(ds.InputDecisions),
			InputData:             hrefs(ds.InputData),
		})
	}
	for _, d := range raw.Decisions {
		dec := model.DMNDecision{
			ID:                    d.ID,
			Name:                  d.Name,
			Label:                 d.Label,
			Description:           strings.TrimSpace(d.Description),
			Question:              strings.TrimSpace(d.Question),
			AllowedAnswers:        strings.TrimSpace(d.AllowedAnswers),
			Variable:              mapInformationItem(d.Variable),
			KnowledgeRequirements: mapKnowledgeReqs(d.KnowledgeRequirements),
			AuthorityRequirements: mapAuthorityReqs(d.AuthorityRequirements),
		}
		for _, ir := range d.InformationRequirements {
			dec.InformationRequirements = append(dec.InformationRequirements, model.DMNInformationRequirement{
				ID:               ir.ID,
				RequiredDecision: ir.RequiredDecision.HRef,
				RequiredInput:    ir.RequiredInput.HRef,
			})
		}
		dec.Logic = mapLogicChildren(&d.xmlLogicChildren)
		out.Decisions = append(out.Decisions, dec)
	}
	for _, dia := range raw.DMNDI.Diagrams {
		md := model.DMNDiagram{ID: dia.ID, Name: dia.Name}
		for _, s := range dia.Shapes {
			md.Shapes = append(md.Shapes, model.DMNShape{
				ID: s.ID, DMNElementRef: s.DMNElementRef,
				X: s.Bounds.X, Y: s.Bounds.Y, Width: s.Bounds.Width, Height: s.Bounds.Height,
			})
		}
		for _, e := range dia.Edges {
			edge := model.DMNEdge{ID: e.ID, DMNElementRef: e.DMNElementRef}
			for _, w := range e.Waypoints {
				edge.Waypoints = append(edge.Waypoints, model.DMNWaypoint{X: w.X, Y: w.Y})
			}
			md.Edges = append(md.Edges, edge)
		}
		out.Diagrams = append(out.Diagrams, md)
	}
	return out, nil
}

func mapItemDefinition(in xmlItemDefinition) model.DMNItemDefinition {
	out := model.DMNItemDefinition{
		ID:           in.ID,
		Name:         in.Name,
		Label:        in.Label,
		Description:  strings.TrimSpace(in.Description),
		TypeRef:      in.typeRef(),
		TypeLanguage: in.TypeLanguage,
		IsCollection: in.IsCollection,
	}
	if in.AllowedValues != nil {
		out.AllowedValues = textsOf(in.AllowedValues.Texts)
	}
	if in.TypeConstraint != nil {
		out.TypeConstraint = textsOf(in.TypeConstraint.Texts)
	}
	for _, c := range in.ItemComponents {
		out.ItemComponents = append(out.ItemComponents, mapItemDefinition(c))
	}
	if in.FunctionItem != nil {
		fi := &model.DMNFunctionItem{OutputTypeRef: in.FunctionItem.OutputTypeRef}
		for _, p := range in.FunctionItem.Parameters {
			fi.Parameters = append(fi.Parameters, mapInformationItem(p))
		}
		out.FunctionItem = fi
	}
	return out
}

func mapInformationItem(in xmlInformationItem) model.DMNInformationItem {
	return model.DMNInformationItem{
		ID: in.ID, Name: in.Name, TypeRef: in.typeRef(), Label: in.Label,
	}
}

func mapKnowledgeReqs(reqs []xmlKnowledgeRequirement) []model.DMNKnowledgeRequirement {
	out := make([]model.DMNKnowledgeRequirement, 0, len(reqs))
	for _, r := range reqs {
		out = append(out, model.DMNKnowledgeRequirement{ID: r.ID, RequiredKnowledge: r.RequiredKnowledge.HRef})
	}
	return out
}

func mapAuthorityReqs(reqs []xmlAuthorityRequirement) []model.DMNAuthorityRequirement {
	out := make([]model.DMNAuthorityRequirement, 0, len(reqs))
	for _, r := range reqs {
		out = append(out, model.DMNAuthorityRequirement{
			ID:                r.ID,
			RequiredAuthority: r.RequiredAuthority.HRef,
			RequiredDecision:  r.RequiredDecision.HRef,
			RequiredInput:     r.RequiredInput.HRef,
		})
	}
	return out
}

func hrefs(items []xmlHRef) []string {
	out := make([]string, 0, len(items))
	for _, it := range items {
		out = append(out, it.HRef)
	}
	return out
}

func textsOf(items []xmlText) []string {
	out := make([]string, 0, len(items))
	for _, it := range items {
		out = append(out, strings.TrimSpace(it.Text))
	}
	return out
}

// mapLogicChildren picks the first non-nil boxed expression child from a
// container and builds the matching *model.DMNLogic variant.
func mapLogicChildren(c *xmlLogicChildren) *model.DMNLogic {
	if c == nil {
		return nil
	}
	switch {
	case c.DecisionTable != nil:
		return &model.DMNLogic{Kind: "decisionTable", ID: c.DecisionTable.ID, TypeRef: c.DecisionTable.typeRef(), Label: c.DecisionTable.Label, DecisionTable: mapDecisionTable(*c.DecisionTable)}
	case c.LiteralExpression != nil:
		return &model.DMNLogic{Kind: "literalExpression", ID: c.LiteralExpression.ID, TypeRef: c.LiteralExpression.typeRef(), Label: c.LiteralExpression.Label, LiteralExpression: mapLiteralExpression(*c.LiteralExpression)}
	case c.Context != nil:
		return &model.DMNLogic{Kind: "context", ID: c.Context.ID, TypeRef: c.Context.TypeRef, Label: c.Context.Label, Context: mapContext(*c.Context)}
	case c.Invocation != nil:
		return &model.DMNLogic{Kind: "invocation", ID: c.Invocation.ID, TypeRef: c.Invocation.TypeRef, Label: c.Invocation.Label, Invocation: mapInvocation(*c.Invocation)}
	case c.FunctionDefinition != nil:
		return &model.DMNLogic{Kind: "functionDefinition", ID: c.FunctionDefinition.ID, TypeRef: c.FunctionDefinition.TypeRef, Label: c.FunctionDefinition.Label, FunctionDefinition: mapFunctionDefinition(c.FunctionDefinition)}
	case c.List != nil:
		return &model.DMNLogic{Kind: "list", ID: c.List.ID, TypeRef: c.List.TypeRef, Label: c.List.Label, List: mapList(*c.List)}
	case c.Relation != nil:
		return &model.DMNLogic{Kind: "relation", ID: c.Relation.ID, TypeRef: c.Relation.TypeRef, Label: c.Relation.Label, Relation: mapRelation(*c.Relation)}
	case c.Conditional != nil:
		return &model.DMNLogic{Kind: "conditional", ID: c.Conditional.ID, TypeRef: c.Conditional.TypeRef, Conditional: mapConditional(*c.Conditional)}
	case c.For != nil:
		return &model.DMNLogic{Kind: "for", ID: c.For.ID, TypeRef: c.For.TypeRef, For: mapFor(*c.For)}
	case c.Every != nil:
		return &model.DMNLogic{Kind: "every", ID: c.Every.ID, TypeRef: c.Every.TypeRef, Every: mapQuantEvery(*c.Every)}
	case c.Some != nil:
		return &model.DMNLogic{Kind: "some", ID: c.Some.ID, TypeRef: c.Some.TypeRef, Some: mapQuantSome(*c.Some)}
	case c.Filter != nil:
		return &model.DMNLogic{Kind: "filter", ID: c.Filter.ID, TypeRef: c.Filter.TypeRef, Filter: mapFilter(*c.Filter)}
	}
	return nil
}

func mapDecisionTable(dt xmlDecisionTableFull) *model.DMNDecisionTable {
	out := &model.DMNDecisionTable{
		HitPolicy:       normaliseHitPolicy(dt.HitPolicy),
		Aggregation:     normaliseAggregation(dt.Aggregation),
		PreferredOrient: dt.PreferredOrientation,
		OutputLabel:     dt.OutputLabel,
	}
	for _, in := range dt.Inputs {
		col := model.DMNDecisionTableInput{
			ID:         in.ID,
			Label:      in.Label,
			Expression: strings.TrimSpace(in.InputExpression.Text),
			TypeRef:    in.InputExpression.typeRef(),
		}
		if col.Expression == "" {
			col.Expression = in.Label
		}
		if in.InputValues != nil {
			col.InputValues = textsOf(in.InputValues.Texts)
		}
		out.Inputs = append(out.Inputs, col)
	}
	for _, o := range dt.Outputs {
		col := model.DMNDecisionTableOutput{
			ID:           o.ID,
			Name:         o.Name,
			Label:        o.Label,
			TypeRef:      o.typeRef(),
			DefaultValue: strings.TrimSpace(o.DefaultOutputEntry.Text),
		}
		if o.OutputValues != nil {
			col.OutputValues = textsOf(o.OutputValues.Texts)
		}
		out.Outputs = append(out.Outputs, col)
	}
	for _, a := range dt.AnnotationClauses {
		out.Annotations = append(out.Annotations, model.DMNRuleAnnotationClause{Name: a.Name})
	}
	for _, r := range dt.Rules {
		row := model.DMNDecisionRule{ID: r.ID, Description: strings.TrimSpace(r.Description)}
		for _, ie := range r.InputEntries {
			row.InputEntries = append(row.InputEntries, strings.TrimSpace(ie.Text))
		}
		for _, oe := range r.OutputEntries {
			row.OutputEntries = append(row.OutputEntries, strings.TrimSpace(oe.Text))
		}
		for _, ae := range r.AnnotationEntries {
			row.Annotations = append(row.Annotations, strings.TrimSpace(ae.Text))
		}
		out.Rules = append(out.Rules, row)
	}
	return out
}

func mapLiteralExpression(le xmlLiteralExpression) *model.DMNLiteralExpression {
	return &model.DMNLiteralExpression{
		Text:               strings.TrimSpace(le.Text.Text),
		ExpressionLanguage: le.ExpressionLanguage,
		ImportedValuesURI:  le.ImportedValues.HRef,
	}
}

func mapContext(c xmlContext) *model.DMNContext {
	out := &model.DMNContext{}
	for _, e := range c.Entries {
		entry := model.DMNContextEntry{ID: e.ID, Value: mapLogicChildren(&e.xmlLogicChildren)}
		if e.Variable != nil {
			v := mapInformationItem(*e.Variable)
			entry.Variable = &v
		}
		out.Entries = append(out.Entries, entry)
	}
	return out
}

func mapInvocation(inv xmlInvocation) *model.DMNInvocation {
	out := &model.DMNInvocation{CalledFunction: mapLogicChildren(&inv.xmlLogicChildren)}
	for _, b := range inv.Bindings {
		out.Bindings = append(out.Bindings, model.DMNBinding{Parameter: mapInformationItem(b.Parameter), Value: mapLogicChildren(&b.xmlLogicChildren)})
	}
	return out
}

func mapFunctionDefinition(fd *xmlFunctionDefinition) *model.DMNFunctionDefinition {
	if fd == nil {
		return nil
	}
	out := &model.DMNFunctionDefinition{Kind: fd.Kind, Body: mapLogicChildren(&fd.xmlLogicChildren)}
	for _, p := range fd.FormalParameters {
		out.Parameters = append(out.Parameters, mapInformationItem(p))
	}
	return out
}

// mapList collects every boxed-expression variant inside a <list> and
// returns them in a stable order (decisionTable, literalExpression, context,
// invocation, functionDefinition, list, relation, conditional, for, every,
// some, filter). DMN documents virtually always use a single homogeneous
// variant for list items, so this ordering matches the natural reading
// order for any realistic file.
func mapList(l xmlList) *model.DMNList {
	out := &model.DMNList{}
	for i := range l.DecisionTableItems {
		out.Items = append(out.Items, model.DMNLogic{Kind: "decisionTable", DecisionTable: mapDecisionTable(l.DecisionTableItems[i])})
	}
	for i := range l.LiteralExpressionItems {
		out.Items = append(out.Items, model.DMNLogic{Kind: "literalExpression", LiteralExpression: mapLiteralExpression(l.LiteralExpressionItems[i])})
	}
	for i := range l.ContextItems {
		out.Items = append(out.Items, model.DMNLogic{Kind: "context", Context: mapContext(l.ContextItems[i])})
	}
	for i := range l.InvocationItems {
		out.Items = append(out.Items, model.DMNLogic{Kind: "invocation", Invocation: mapInvocation(l.InvocationItems[i])})
	}
	for i := range l.FunctionDefinitionItems {
		fd := l.FunctionDefinitionItems[i]
		out.Items = append(out.Items, model.DMNLogic{Kind: "functionDefinition", FunctionDefinition: mapFunctionDefinition(&fd)})
	}
	for i := range l.ListItems {
		out.Items = append(out.Items, model.DMNLogic{Kind: "list", List: mapList(l.ListItems[i])})
	}
	for i := range l.RelationItems {
		out.Items = append(out.Items, model.DMNLogic{Kind: "relation", Relation: mapRelation(l.RelationItems[i])})
	}
	for i := range l.ConditionalItems {
		out.Items = append(out.Items, model.DMNLogic{Kind: "conditional", Conditional: mapConditional(l.ConditionalItems[i])})
	}
	for i := range l.ForItems {
		out.Items = append(out.Items, model.DMNLogic{Kind: "for", For: mapFor(l.ForItems[i])})
	}
	for i := range l.EveryItems {
		out.Items = append(out.Items, model.DMNLogic{Kind: "every", Every: mapQuantEvery(l.EveryItems[i])})
	}
	for i := range l.SomeItems {
		out.Items = append(out.Items, model.DMNLogic{Kind: "some", Some: mapQuantSome(l.SomeItems[i])})
	}
	for i := range l.FilterItems {
		out.Items = append(out.Items, model.DMNLogic{Kind: "filter", Filter: mapFilter(l.FilterItems[i])})
	}
	return out
}

func mapRelation(r xmlRelation) *model.DMNRelation {
	out := &model.DMNRelation{}
	for _, c := range r.Columns {
		out.Columns = append(out.Columns, mapInformationItem(c))
	}
	for _, row := range r.Rows {
		var cells []model.DMNLogic
		if lg := mapLogicChildren(&row.xmlLogicChildren); lg != nil {
			cells = append(cells, *lg)
		}
		out.Rows = append(out.Rows, cells)
	}
	return out
}

func mapConditional(c xmlConditional) *model.DMNConditional {
	return &model.DMNConditional{If: mapLogicChildren(&c.If.xmlLogicChildren), Then: mapLogicChildren(&c.Then.xmlLogicChildren), Else: mapLogicChildren(&c.Else.xmlLogicChildren)}
}

func mapFor(f xmlIteratorExpr) *model.DMNFor {
	return &model.DMNFor{Iterator: f.IteratorVariable, In: mapLogicChildren(&f.In.xmlLogicChildren), Return: mapLogicChildren(&f.Return.xmlLogicChildren)}
}

func mapQuantEvery(f xmlIteratorExpr) *model.DMNEvery {
	return &model.DMNEvery{Iterator: f.IteratorVariable, In: mapLogicChildren(&f.In.xmlLogicChildren), Satisfies: mapLogicChildren(&f.Satisfies.xmlLogicChildren)}
}

func mapQuantSome(f xmlIteratorExpr) *model.DMNSome {
	return &model.DMNSome{Iterator: f.IteratorVariable, In: mapLogicChildren(&f.In.xmlLogicChildren), Satisfies: mapLogicChildren(&f.Satisfies.xmlLogicChildren)}
}

func mapFilter(f xmlFilter) *model.DMNFilter {
	return &model.DMNFilter{In: mapLogicChildren(&f.In.xmlLogicChildren), Match: mapLogicChildren(&f.Match.xmlLogicChildren)}
}
