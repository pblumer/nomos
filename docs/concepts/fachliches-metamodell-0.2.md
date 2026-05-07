# Fachliches Metamodell 0.2 – Integration Servicegraph

## Zusammenfassung der Änderungen gegenüber 0.1

Das Metamodell 0.2 erweitert Version 0.1 um die graphbasierte Service-Modellierung (Servicegraph).
Die bestehenden Artefakttypen (Product, Rule, Decision, Process, Task, Skill) bleiben erhalten.
Neu hinzu kommt eine **Servicegraph-Schicht**, die Hierarchie und Abhängigkeiten in einem
fachlich auswertbaren Graphmodell vereint.

### Wesentliche Neuerungen

- Servicegraph als eigenständiger Artefakttyp (neben Product, Rule, Decision, Process, Skill)
- Knotentypen: 10 fachliche Knotentypen mit standardisierten Attributen
- Beziehungstypen: 10 explizit typisierte Kantenarten
- Graphregeln als eigenständige Modellelemente
- Varianten ohne Modellduplikation
- Zustandsmodellierung via produces/consumes state
- Kompensationsmodellierung als Erstklasselement
- Mapping zwischen Servicegraph-Knoten und bestehenden Nomos-Artefakten


## 1. Gesamtmodell: Nomos 0.2

### 1.1 Artefakttypen

| Artefakttyp | Rolle im Modell | Seit |
|---|---|---|
| Product | Fachlicher Leistungsgegenstand | 0.1 |
| Rule | Prüfbare fachliche Aussage | 0.1 |
| Decision | Strukturierte fachliche Bewertung | 0.1 |
| Process | Fachlicher Ablauf (BPMN-nah) | 0.1 |
| Task | Arbeitsschritt im Prozess | 0.1 |
| Skill | Operative Ausführungseinheit | 0.1 |
| **Servicegraph** | **Graphbasiertes Modell einer Service-Komposition** | **0.2** |
| **Graph-Node** | **Knoten im Servicegraph** | **0.2** |
| **Graph-Edge** | **Beziehung im Servicegraph** | **0.2** |
| **Graph-Rule** | **Regel zur Steuerung des Servicegraphen** | **0.2** |

### 1.2 Schichtenmodell

```
┌─────────────────────────────────────────────────────────────────┐
│  Product Layer                                                   │
│  (Was soll bereitgestellt werden?)                              │
├─────────────────────────────────────────────────────────────────┤
│  Servicegraph Layer  [NEU 0.2]                                  │
│  (Woraus besteht der Service? Welche Abhängigkeiten bestehen?)  │
│  Knoten ─── Beziehungen ─── Regeln ─── Varianten               │
├─────────────────────────────────────────────────────────────────┤
│  Rule & Decision Layer                                          │
│  (Welche Regeln gelten? Welche Entscheidungen werden getroffen?)│
├─────────────────────────────────────────────────────────────────┤
│  Process & Orchestration Layer                                  │
│  (In welcher Reihenfolge? Welche Pfade?)                        │
├─────────────────────────────────────────────────────────────────┤
│  Skill & Execution Layer                                        │
│  (Wie wird ein konkreter Schritt operativ ausgeführt?)          │
└─────────────────────────────────────────────────────────────────┘
```

### 1.3 Positionierung des Servicegraphen

Der Servicegraph sitzt **zwischen Product und Process**:

- Das **Product** definiert, was fachlich bereitgestellt wird.
- Der **Servicegraph** modelliert, **woraus** ein Service besteht und **welche Abhängigkeiten** gelten.
- Der **Process** leitet aus dem Graphen eine **ausführbare Reihenfolge** ab.
- **Rules/Decisions** steuern das Verhalten im Graphen.
- **Skills** führen die Aktivitätsknoten operativ aus.


## 2. Mapping: Servicegraph-Knoten auf Nomos-Artefakte

| Graph-Knotentyp | Nomos-Artefakt 0.1 | Beziehung |
|---|---|---|
| Service-Knoten | Product | 1:1 – Einstiegspunkt |
| Sub-Service-Knoten | – (neu) | Fachliche Zerlegung, kein eigenes Product |
| Aktivitäts-Knoten | Task + Skill | Task beschreibt WAS, Skill beschreibt WIE |
| Vorbedingungs-Knoten | Rule (precondition) | Regel prüft die Vorbedingung |
| Entscheidungs-Knoten | Decision | Entscheidungsartefakt bewertet |
| Validierungs-Knoten | Rule (quality_gate) | Qualitätskriterium prüft Ergebnis |
| Manueller Knoten | Task (manual_review) | Task mit menschlicher Bearbeitungsrolle |
| Kompensations-Knoten | – (neu) | Eigener Typ, kein Äquivalent in 0.1 |
| Zustands-Knoten | – (neu) | Explizites Zustandsmanagement |
| Referenz-Knoten | – (Querverweis) | Wiederverwendungsreferenz |


## 3. Servicegraph-Artefakt

### 3.1 Schema

```yaml
id: SG-ACC-001
type: servicegraph
name: Benutzerkonto mit Exchange-Mailbox bereitstellen
version: 0.2.0
status: draft
owner: Identity & Collaboration
summary: >
  Graphbasiertes Modell für die Bereitstellung eines
  Benutzerkontos inkl. Exchange-Mailbox.

related_product: PROD-ACC-MBX-001
related_process: PROC-ACC-MBX-001

variants:
  - id: ACC-V-001
    name: U-Account
    context: interner Mitarbeitender
  - id: ACC-V-002
    name: X-Account
    context: externer Mitarbeitender

nodes:
  - id: ACC-N-001
    type: service
    name: Benutzerkonto mit Exchange-Mailbox bereitstellen
    mandatory: true
    reusable: false
  - id: ACC-N-040
    type: activity
    name: Benutzerkonto anlegen
    mandatory: true
    reusable: true
    activation_rule: ACC-G-001
    skill_ref: SKILL-ACC-005
  # ... (weitere Knoten)

edges:
  - id: ACC-R-009
    source: ACC-N-020
    target: ACC-N-040
    type: depends_on
    binding: hard
    description: Kontoanlage braucht bekannte Person
  - id: ACC-R-015
    source: ACC-N-040
    target: ACC-N-100
    type: produces_state
    binding: hard
  # ... (weitere Kanten)

rules:
  - id: ACC-G-001
    name: Kontotyp muss bestimmt sein
    type: activation
    scope: ACC-N-040
    binding: must
    expression: kontotyp IN ['U', 'X']
  # ... (weitere Regeln)
```

### 3.2 Basisattribute des Servicegraph-Artefakts

| Feld | Typ | Beschreibung |
|---|---|---|
| id | string | Eindeutige ID (Prefix SG-) |
| type | string | Immer `servicegraph` |
| name | string | Fachlicher Name |
| version | semver | Versionsstand |
| status | enum | draft / in_review / approved / deprecated |
| owner | string | Verantwortliche Rolle |
| related_product | ref | Verweis auf Product-Artefakt |
| related_process | ref | Verweis auf Process-Artefakt |
| variants | list | Variantendefinitionen |
| nodes | list | Knotenliste |
| edges | list | Beziehungsliste |
| rules | list | Graphregeln |


## 4. Knotentypen (node.type)

| Typ-Schlüssel | Knotentyp | Mapping |
|---|---|---|
| `service` | Service-Knoten | → Product |
| `sub_service` | Sub-Service-Knoten | fachliche Zerlegung |
| `activity` | Aktivitäts-Knoten | → Task + Skill |
| `precondition` | Vorbedingungs-Knoten | → Rule |
| `decision` | Entscheidungs-Knoten | → Decision |
| `validation` | Validierungs-Knoten | → Rule (quality_gate) |
| `manual` | Manueller Knoten | → Task (manual_review) |
| `compensation` | Kompensations-Knoten | eigener Typ |
| `state` | Zustands-Knoten | eigener Typ |
| `reference` | Referenz-Knoten | Querverweis |

### Knotenattribute

```yaml
id: string           # Eindeutig (Format: <PREFIX>-N-<NNN>)
type: enum           # Knotentyp (siehe oben)
name: string         # Fachlicher Name
description: string  # optional
mandatory: bool      # Pflicht (true) oder Optional (false)
variant: string      # optional – Varianten-ID
activation_rule: ref # optional – Verweis auf Graph-Rule
owner: string        # Verantwortliche Rolle
reusable: bool       # Kann in anderen Graphen referenziert werden
skill_ref: ref       # optional – Verweis auf Skill (bei activity)
decision_ref: ref    # optional – Verweis auf Decision (bei decision)
rule_ref: ref        # optional – Verweis auf Rule (bei precondition/validation)
```


## 5. Beziehungstypen (edge.type)

| Typ-Schlüssel | Semantik |
|---|---|
| `composed_of` | Hierarchische Zugehörigkeit (besteht aus) |
| `depends_on` | Voraussetzung für Aktivierung (hart = DAG-Pflicht) |
| `enables` | Freischaltung |
| `blocks` | Verhinderung |
| `validates` | Prüfung des Ergebnisses |
| `compensates` | Fachliche Rückbehandlung |
| `alternative_to` | Alternativbeziehung |
| `requires_condition` | Bedingte Aktivierung |
| `produces_state` | Erzeugt einen Zustandsknoten |
| `consumes_state` | Benötigt einen vorhandenen Zustand |

### Kantenattribute

```yaml
id: string           # Eindeutig (Format: <PREFIX>-R-<NNN>)
source: ref          # Quell-Knoten-ID
target: ref          # Ziel-Knoten-ID
type: enum           # Beziehungstyp (siehe oben)
binding: enum        # hard / soft
condition: string    # optional – Aktivierungsbedingung
description: string  # Fachliche Beschreibung
```


## 6. Graphregeln (rule.type)

| Typ-Schlüssel | Zweck |
|---|---|
| `activation` | Steuert, wann ein Knoten/Beziehung aktiv ist |
| `variant` | Definiert variantenspezifische Knoten-/Kantenaktivierung |
| `validation` | Erfolgskriterium für einen Knoten |
| `consistency` | Modellqualitätsprüfung (z.B. Zyklusfreiheit) |
| `compensation` | Wann und wie Kompensation greift |

### Regelattribute

```yaml
id: string           # Eindeutig (Format: <PREFIX>-G-<NNN>)
name: string         # Fachlicher Name
type: enum           # Regeltyp (siehe oben)
scope: ref | list    # Geltungsbereich (Knoten-IDs oder Modell)
binding: enum        # must / should / may
expression: string   # Fachlicher oder strukturierter Ausdruck
description: string  # optional
```


## 7. Konsistenzregeln für Servicegraphen

1. **Eindeutigkeit**: Jede Knoten-ID, Kanten-ID und Regel-ID muss im Graphen eindeutig sein.
2. **Typisierung**: Jeder Knoten und jede Beziehung muss einen definierten Typ besitzen.
3. **DAG bei harten Abhängigkeiten**: `depends_on` mit `binding: hard` darf keinen Zyklus bilden.
4. **Erreichbarkeit**: Jeder Pflichtknoten muss über den Abhängigkeitsgraphen erreichbar sein.
5. **Referenzielle Integrität**: Jede Kante muss auf existente Knoten-IDs verweisen.
6. **Trennung Struktur/Ausführung**: `composed_of` impliziert keine Ausführungsreihenfolge.
7. **Variantenkonsistenz**: Varianten dürfen keine Widersprüche in Pflichtknoten erzeugen.
8. **Kompensations-Coverage**: Jeder kritische Aktivitätsknoten sollte einen Kompensationsknoten haben.
9. **Zustandskonsistenz**: Jeder `consumes_state`-Knoten muss einen korrespondierenden `produces_state` haben.


## 8. Zusammenspiel mit bestehenden Artefakten

### 8.1 Product → Servicegraph

```
Product (PROD-ACC-MBX-001)
  └── servicegraph_ref: SG-ACC-001
```

Ein Product referenziert seinen Servicegraph. Der Graph beschreibt die innere Komposition.

### 8.2 Servicegraph → Process

Der Process kann aus dem Servicegraph abgeleitet werden:

- **Topologische Sortierung** der harten `depends_on`-Kanten ergibt die Ausführungsreihenfolge.
- **Parallelisierbare Knoten** haben keine gegenseitigen Abhängigkeiten.
- **Entscheidungsknoten** erzeugen Gateways im BPMN.
- **Validierungsknoten** erzeugen Quality Gates.

### 8.3 Aktivitäts-Knoten → Task + Skill

```
Graph-Node (ACC-N-040, type: activity)
  ├── task_ref: TASK-ACC-005
  └── skill_ref: SKILL-ACC-005
```

Der Aktivitätsknoten im Graph referenziert einen Task (WAS) und einen Skill (WIE).

### 8.4 Vorbedingungs-Knoten → Rule

```
Graph-Node (ACC-N-020, type: precondition)
  └── rule_ref: RULE-ACC-002
```

Die Business Rule aus dem Regelkatalog formalisiert die Vorbedingung.

### 8.5 Entscheidungs-Knoten → Decision

```
Graph-Node (ACC-N-010, type: decision)
  └── decision_ref: DEC-ACC-001
```

Das Decision-Artefakt bewertet die Entscheidung im Graphen.


## 9. Ablagestruktur (erweitert)

```
catalog/
  products/
    benutzerkonto-mit-mailbox.yaml
  servicegraphs/                          # [NEU 0.2]
    benutzerkonto-mit-mailbox.yaml
  rules/
    konto/
      pflichtidentitaet.yaml
      ...
  decisions/
    produktvariante-bestimmen.yaml
    ...
  skills/
    konto-im-iam-anlegen.yaml
    ...
  processes/
    provisionierung-benutzerkonto-mit-mailbox.yaml
```


## 10. Validierungsregeln für Servicegraph-Artefakte

| Regel | Beschreibung |
|---|---|
| `id` gesetzt und eindeutig | Pflicht |
| `type` = servicegraph | Pflicht |
| `related_product` referenziert existentes Product | Pflicht |
| Mindestens 1 Knoten mit type=service | Pflicht |
| Mindestens 1 Kante mit type=composed_of | Pflicht |
| Keine Zyklen in hard depends_on | Pflicht |
| Alle Knoten-IDs in Kanten existieren als Knoten | Pflicht |
| Alle rule_ref / skill_ref / decision_ref zeigen auf existente Artefakte | Soll (Warnung) |


## 11. Migration 0.1 → 0.2

Bestehende Artefakte aus 0.1 bleiben unverändert gültig. Der Servicegraph ist ein **optionales Zusatzartefakt**, das schrittweise eingeführt werden kann:

1. **Phase 1**: Servicegraph für bestehende Referenzprodukte erstellen (z.B. Benutzerkonto mit Mailbox).
2. **Phase 2**: Process-Artefakte mit `servicegraph_ref` verknüpfen.
3. **Phase 3**: Automatische Reihenfolge-Ableitung aus Graph testen.
4. **Phase 4**: Graph als primäre Modellierungsform verbindlich machen; Process wird generiert.


## 12. Zusammenfassung

Das Metamodell 0.2 erweitert Nomos um ein graphbasiertes Service-Modell, das:

- **Hierarchie und Abhängigkeiten trennt** (composed_of ≠ Ausführungsreihenfolge)
- **explizite Beziehungstypen** statt impliziter Annahmen bietet
- **Zustandsmanagement** als Erstklasskonzept einführt (produces/consumes state)
- **Kompensation** fachlich modellierbar macht
- **Varianten** ohne Graphduplikation abbildet
- **Parallelisierung** durch topologische Analyse ermöglicht
- **Konsistenzprüfung** durch formale Graphregeln erlaubt
- vollständig **kompatibel mit bestehenden Artefakten** (Product, Rule, Decision, Process, Skill) bleibt

Die graphbasierte Modellierung wird damit zur fachlichen Klammer zwischen Produktdefinition
und Provisionierungsorchestrierung.
