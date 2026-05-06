# ADR-0005: Blueprint-/Instance-/Assurance-Modell als Kernmodell

## Status

Accepted

## Kontext

Nomos hatte im MVP-Kontext bereits Produkt-, Regel-, Prozess- und Skill-Begriffe. Diese Begriffe sind nuetzlich, aber zu breit: Sie koennen leicht als reine Architektur- oder Dokumentationsartefakte missverstanden werden. Fuer den Git-first- und YAML-first-Ansatz braucht Nomos ein fachliches Kernmodell, das klar zwischen Soll-Zustand, Ist-Zustand und Assurance unterscheidet.

## Entscheidung

Nomos verwendet ein eigenes Blueprint-/Instance-/Assurance-Modell als Kern-Domain-Modell.

Ein Blueprint beschreibt den gewuenschten, provisionierbaren Soll-Zustand. Eine Instance beschreibt einen beobachteten oder provisionierten Ist-Zustand. Assurance vergleicht Instances gegen ihre Blueprints und dokumentiert Compliance-Status, Evidenz und Findings.

Nomos unterscheidet dabei explizit zwischen Namespace Service, Service Blueprint und Product Blueprint. Ein Namespace Service ist ein konkreter logischer Service im Cosmos-Baum, zum Beispiel `identity.blumer.cloud/user-account`. Ein Service Blueprint ist eine provisionierbare Vorlage fuer einen solchen Namespace Service und kann `namespace_service_ref` setzen. Ein Product Blueprint buendelt mehrere benoetigte Namespace Services und ordnet jeden Eintrag in `required_services` einem Service Blueprint zu.

## Konsequenzen

- Produktartefakte werden fachlich als Product Blueprints gerahmt.
- Service Blueprints und Service Instances werden First-Class-Konzepte.
- Product Blueprints behalten `required_service_blueprints` fuer Rueckwaertskompatibilitaet und verwenden `required_services` fuer die konkrete Abbildung Namespace Service -> Service Blueprint.
- Service Blueprints koennen mit `namespace_service_ref` zeigen, welchen Namespace Service sie beschreiben.
- Product Instances verweisen auf Product Blueprints und die provisionierten Service Instances.
- Assurance vergleicht tatsaechliche Instances mit Blueprints und dokumentiert Evidence, Findings und Compliance-Status.
- C4 und ArchiMate bleiben optionale View- und Dokumentationsnotationen; sie sind nicht das Nomos-Kernmodell.
- Bestehende `type: product`-Artefakte bleiben als rueckwaertskompatible Alias- bzw. Legacy-Form unterstuetzt.
- Die Quelle der Wahrheit bleibt Git-first und YAML-first; es wird keine Datenbank, BPMN-/DMN-Engine oder generische Skill Runtime eingefuehrt.
