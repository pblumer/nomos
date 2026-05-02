# Nomos Idea Paper: Kosmos, Domaenen, Services, Faehigkeiten und verifizierbare fachliche Hoheit

## 1. Ausgangslage

Nomos fokussiert heute vor allem produktbezogenes Wissen: Produkte, Produktvarianten, Anforderungen, Regeln, Entscheidungen, Prozesse, Validierung sowie Skills als ausfuehrbare Bausteine. Diese Grundlage ist bereits sehr leistungsfaehig, weil sie fachliche Aussagen strukturiert, versionierbar macht und fuer Pruefung sowie Nachvollziehbarkeit vorbereitet.

Mit wachsender Reife und mit mehreren Organisationseinheiten entsteht jedoch eine neue Ebene von Fragen, die im bisherigen Fokus nur indirekt beantwortet werden:

- Wo gehoert ein Produkt, eine Regel, ein Prozess oder ein Skill fachlich hin?
- Wer darf ein Artefakt erstellen, aendern, freigeben oder ablehnen?
- Wie kann ein anderes System oder eine andere Organisation verifizieren, dass ein Artefakt tatsaechlich zum behaupteten Besitzer gehoert?
- Wie kann Nomos foederierte, verteilte und Git-basierte Wissensraeume unterstuetzen, ohne die fachliche Klarheit zu verlieren?

Die bestehende Modellierung adressiert die **inhaltliche Wahrheit** sehr gut. Fuer einen organisationsuebergreifenden Einsatz braucht es zusaetzlich eine robuste Struktur fuer **Herkunft, Zustaendigkeit und Vertrauen**.

## 2. Grundidee: Der Kosmos als oberster fachlicher Vertrauensraum

Die zentrale Idee ist die Einfuehrung eines **Kosmos** als oberster Namespace, Besitzraum und Vertrauensgrenze in Nomos.

Ein Kosmos kann eine Organisation, eine Organisationseinheit, eine Plattform, eine oeffentliche Institution, ein Unternehmen, einen Mandanten, ein Environment oder eine Community repraesentieren. Damit ist ein Kosmos:

- nicht nur ein Ordner,
- sondern eine semantische und governance-relevante Grenze,
- sowie ein verifizierbarer Vertrauensraum.

Ein Kosmos besitzt oder enthaelt Domaenen. Er kann an DNS-aehnliche Namen gekoppelt werden, etwa:

```text
kosmos: admin.ch
kosmos: blumer.cloud
kosmos: identity.example.org
```

Damit erweitert Nomos seinen Fokus: nicht nur fachliche Inhalte werden modelliert, sondern auch deren **Herkunft**, **Zustaendigkeit** und **Vertrauenswuerdigkeit**.

## 3. Domaenen innerhalb eines Kosmos

Eine **Domaene** ist ein fachlicher Verantwortungsbereich innerhalb eines Kosmos. Typische Beispiele:

- Identity
- Messaging
- Workplace
- HR
- Finance
- IAM
- Collaboration
- Document Management

Domaenen koennen DNS-aehnlich benannt werden:

```text
identity.admin.ch
messaging.admin.ch
workplace.admin.ch
document.blumer.cloud
```

Domaenen tragen Verantwortung fuer Services, Regeln, Entscheidungen, Prozesse, Skills und weitere governance-relevante Artefakte. Das bestehende Owner-Konzept in Nomos-Artefakten kann dadurch gestaerkt werden, indem Owner klar auf Domaenen referenziert werden.

## 4. Services als fachliche Leistungseinheiten

Ein **Service** ist eine fachliche Leistungseinheit innerhalb einer Domaene.

Wichtig ist die Abgrenzung:

- Ein Service ist nicht zwingend ein technischer Microservice.
- Ein Service repraesentiert eine stabile fachliche Leistung bzw. einen Leistungscontainer.
- Produkte sind bestellbare oder provisionierbare Auspraegungen eines Services.

Ein Service kann Produkte, Varianten, Faehigkeiten, Prozesse, Regeln, Entscheidungen und Skills enthalten oder auf diese verweisen.

Beispielstruktur:

```text
kosmos: admin.ch
  domain: identity.admin.ch
    service: user-account
    service: privileged-account
    service: group-membership
    service: mailbox
```

Beziehung Service zu Produkt:

- **Service** = breitere fachliche Leistung.
- **Produkt** = konkretes bestellbares/provisionierbares Angebot.

Beispiel: Der Service `user-account` kann das Produkt `Benutzerkonto mit Mailbox` anbieten.

## 5. Faehigkeiten als fachliche Beschreibung dessen, was ein Service kann

Der Begriff **Faehigkeit** (Capability) beschreibt, was ein Service fachlich leisten kann.

Die Unterscheidung ist zentral:

- **Capability/Faehigkeit** = fachliche Faehigkeit
- **Skill** = ausfuehrbarer, versionierter Baustein

Faehigkeiten sind semantisch stabiler als Skills. Ein Service kann zum Beispiel folgende Faehigkeiten bereitstellen:

- Identitaet validieren
- Benutzerkonto erstellen
- Benutzerkonto deaktivieren
- Benutzername ableiten
- Mailbox provisionieren
- Qualitaetspruefung durchfuehren
- Nachweise erzeugen

Eine Faehigkeit kann durch einen oder mehrere Skills umgesetzt oder unterstuetzt werden.

```text
Service: Benutzerkonto
  Capability: Identitaet validieren
    Requirements:
      - ANF-IDENTITY-001
    Rules:
      - RULE-IDENTITY-001
    Decisions:
      - DEC-IDENTITY-001
    Skills:
      - SKILL-IDENTITY-VALIDATE-PERSON
```

Warum diese Trennung wichtig ist:

- Skills koennen sich technisch aendern, ohne die fachliche Faehigkeit zu aendern.
- Mehrere Skills koennen dieselbe Faehigkeit stuetzen.
- Faehigkeiten helfen Architektur, Governance, Wiederverwendung und KI-Verstaendnis.
- Skills helfen der operativen Ausfuehrung.

## 6. Prozesse als fachliche Orchestrierung nicht vergessen

Prozesse bleiben ein Kernelement von Nomos. Sie verbinden fachliche Absicht mit operativer Ausfuehrung.

Ein Prozess beschreibt unter anderem:

- Trigger
- Abfolge von Tasks
- Entscheidungen
- Gateways bzw. Verzweigungen
- manuelle Arbeitsschritte
- skill-basierte Ausfuehrung
- Qualitaetspruefungen
- Nachweiserzeugung
- Endzustaende

Beziehungsmodell:

```text
Process
  contains Tasks

Task
  requires Capability
  invokes Skill

Skill
  executes operational logic
  produces Evidence
  produces Findings
```

Wesentliche Leitgedanken:

- Prozesse sollen nicht mit technischer Detail-Logik ueberladen werden.
- BPMN kann spaeter als Modellierungssprache genutzt werden.
- Zunaechst kann Nomos Prozesse als strukturierte YAML/Markdown-Artefakte beschreiben und spaeter auf BPMN abbilden.
- Der Prozess ist die verbindende Schicht zwischen Service, Produkt, Faehigkeit, Entscheidung und Skill.

## 7. Der fachliche Tree von Nomos

Vorgeschlagenes Tree-Modell:

```text
Kosmos
 └── Domaene
      └── Service
           └── Faehigkeit
                ├── Anforderungen
                ├── Regeln
                ├── Entscheidungen
                ├── Prozesse
                ├── Tasks
                ├── Skills
                ├── Nachweise
                └── Findings
```

Dasselbe als Beziehungsmodell:

```text
Cosmos
  owns Domains

Domain
  owns Services

Service
  offers Capabilities

Capability
  is governed by Requirements
  is constrained by Rules
  is evaluated by Decisions
  is realized through Processes
  is executed or supported by Skills

Process
  contains Tasks

Task
  requires Capabilities
  invokes Skills

Skill
  produces Evidence
  produces Findings
```

Diese Struktur sollte nicht als starre UI-Hierarchie missverstanden werden. Sie ist gleichzeitig:

- konzeptionelles Modell,
- moegliche Repository-Struktur,
- Navigationsstruktur,
- Governance- und Vertrauensstruktur.

## 8. Git als Quelle der Wahrheit

Git kann als Source of Truth fuer den Nomos-Tree genutzt werden. Das betrifft nicht nur Speicherung, sondern auch Governance und Trust.

Vorteile von Git in diesem Kontext:

| Aspekt | Nutzen |
|---|---|
| Versionierung | Historie fachlicher Aenderungen bleibt nachvollziehbar |
| Branches | Parallele Ausarbeitung von Varianten und Entwuerfen |
| Pull Requests | Formale Review- und Diskussionspunkte |
| Reviews/Approvals | Governance-Prozesse werden operationalisiert |
| Signierte Commits | Kryptografische Integritaet von Aenderungen |
| Reproduzierbare Historie | Vergleichbarkeit von Staenden |
| Auditierbarkeit | Nachweis ueber Entscheidungen und Freigaben |
| Rollback | Sichere Ruecknahme fehlerhafter Aenderungen |
| Zusammenarbeit | Bruecke zwischen Fachseite, Architektur, Plattform und Delivery |

Moegliche Repository-Struktur:

```text
nomos/
  cosmos/
    admin.ch/
      cosmos.yaml
      ownership.yaml
      trust.yaml

      domains/
        identity.admin.ch/
          domain.yaml

          services/
            user-account/
              service.yaml

              capabilities/
                identity-validation.yaml
                account-creation.yaml
                mailbox-provisioning.yaml

              products/
                benutzerkonto-mit-mailbox.yaml

              requirements/
                anf-identity-001.yaml

              rules/
                rule-identity-001.yaml

              decisions/
                dec-provisioning-readiness.yaml

              processes/
                proc-account-mailbox-provisioning.yaml

              skills/
                skill-identity-validate-person/
                  SKILL.md
                  skill.yaml
```

Git ist damit nicht nur Speicher, sondern Teil des fachlichen Governance- und Vertrauensmodells.

## 9. Verifizierbare Hoheit ueber Kosmos und Domaene

Ein Kosmos oder eine Domaene kann Besitz an einem Namespace behaupten. Diese Behauptung sollte verifizierbar sein.

Moegliche Verifikationsmethoden:

- DNS TXT Record
- TLS/Domain Ownership
- signierte Git-Commits
- signierte Git-Tags/Releases
- GPG-Signaturen
- Sigstore
- interne PKI
- organisationsspezifisches Trust-Registry-Modell

Beispiel fuer DNS-TXT-basierte Verifikation:

```yaml
id: COSMOS-ADMIN-CH
type: cosmos
name: Schweizer Bundesverwaltung
namespace: admin.ch

ownership:
  verification_method: dns_txt
  dns_name: _nomos.admin.ch
  expected_value: nomos-verify=sha256:abc123...
```

Passender DNS-Record:

```text
TXT _nomos.admin.ch = "nomos-verify=sha256:abc123..."
```

Bedeutung:

- Existiert der Record, hat der Controller der DNS-Zone diesen Nomos-Kosmos autorisiert.
- Das beweist technische Domain-Kontrolle, aber nicht automatisch fachliche Legitimation.
- Organisatorische Governance bleibt weiterhin notwendig.

Nutzen dieser Verifizierbarkeit:

- vertrauenswuerdige Foederation,
- externe Ueberpruefbarkeit,
- maschinenlesbare Ownership,
- KI-Agenten koennen Vertrauensquellen einschaetzen,
- Orchestrierungssysteme koennen entscheiden, ob Regel/Skill konsumiert werden darf,
- Organisationen koennen autoritative Service- und Regelmodelle publizieren.

## 10. Vertrauensstufen

Vertrauen sollte nicht binaer modelliert werden, sondern abgestuft.

```yaml
trust_level: unverified
```

Bedeutung: lokal oder experimentell, ohne externe Verifikation.

```yaml
trust_level: dns_verified
```

Bedeutung: Namespace/Domain Ownership ueber DNS verifiziert.

```yaml
trust_level: signed
```

Bedeutung: Artefakte, Commits oder Releases kryptografisch signiert.

```yaml
trust_level: approved
```

Bedeutung: intern gemass Governance-Prozess geprueft und freigegeben.

```yaml
trust_level: authoritative
```

Bedeutung: verifiziert, signiert, freigegeben und als offizielle Quelle publiziert.

## 11. Artefakt-Erweiterungen

Die folgenden Artefakt-Typen sind als minimale konzeptionelle Erweiterung gedacht.

### cosmos.yaml

```yaml
id: COSMOS-BLUMER-CLOUD
type: cosmos
name: Blumer Cloud
namespace: blumer.cloud
version: 0.1.0
status: draft
owner: Patrick Blumer
summary: Fachlicher Nomos-Kosmos fuer Blumer Cloud Services.

verification:
  status: unverified
  methods: []

domains:
  - DOM-HOME-BLUMER-CLOUD
  - DOM-DOCUMENT-BLUMER-CLOUD
```

### domain.yaml

```yaml
id: DOM-IDENTITY-ADMIN-CH
type: domain
name: Identity Domain
namespace: identity.admin.ch
version: 0.1.0
status: draft
owner: Identity Governance
summary: Verantwortungsbereich fuer Identitaet, Benutzerkonten und Berechtigungen.

services:
  - SVC-USER-ACCOUNT
  - SVC-GROUP-MEMBERSHIP
```

### service.yaml

```yaml
id: SVC-USER-ACCOUNT
type: service
name: Benutzerkonto-Service
version: 0.1.0
status: draft
owner: Identity Operations
summary: Fachlicher Service fuer die Bereitstellung und Verwaltung von Benutzerkonten.

products:
  - PROD-ACC-MBX-001

capabilities:
  - CAP-IDENTITY-VALIDATION
  - CAP-ACCOUNT-CREATION
  - CAP-ACCOUNT-DEACTIVATION
```

### capability.yaml

```yaml
id: CAP-IDENTITY-VALIDATION
type: capability
name: Identitaet validieren
version: 0.1.0
status: draft
owner: Identity Governance
summary: Faehigkeit zur Pruefung, ob eine Person eindeutig identifiziert werden kann.

requirements:
  - ANF-IDENTITY-001

rules:
  - RULE-IDENTITY-001

decisions:
  - DEC-IDENTITY-VALIDITY

processes:
  - PROC-ACC-MBX-001

skills:
  - SKILL-IDENTITY-VALIDATE-PERSON
```

Diese Beispiele sind bewusst konzeptionell und definieren noch kein finales Schema.

## 12. Bezug zum bestehenden Nomos-Metamodell

Das Kosmos-Modell erweitert das bestehende Nomos-Metamodell, ersetzt es aber nicht.

Bestehende Kernartefakte:

- product
- product variant
- requirement
- rule
- decision
- process
- task
- skill
- target system
- evidence
- finding

Neue Artefakte auf uebergeordneter Ebene:

- cosmos
- domain
- service
- capability

Mehrwert der neuen Ebene:

- Namespace,
- Ownership,
- Governance-Kontext,
- Trust-Kontext,
- Navigationsstruktur,
- Wiederverwendungsgrenzen,
- Foederationsgrenzen.

## 13. Governance-Idee

Eine moegliche Governance in diesem Modell kann sich an klaren Verantwortlichkeiten orientieren.

Leitfragen:

- Wer darf einen Kosmos anlegen?
- Wer darf einen Kosmos verifizieren?
- Wem gehoert eine Domaene?
- Wer gibt eine Service-Definition frei?
- Wer besitzt eine Faehigkeit?
- Wer gibt Regeln frei?
- Wer gibt Entscheidungen frei?
- Wer darf einen Skill releasen?
- Wer darf Artefakte als authoritative publizieren?

Moegliche Rollen:

- Kosmos Owner
- Domain Owner
- Service Owner
- Capability Owner
- Rule Owner
- Process Owner
- Skill Owner
- Platform Governance
- Security / Compliance Reviewer

Git-basierte Reviews unterstuetzen diese Governance:

- Pull Requests,
- CODEOWNERS,
- Required Reviews,
- signierte Commits,
- Protected Branches,
- Release Tags,
- Audit Trail.

## 14. KI- und Agenten-Nutzbarkeit

Das Modell ist fuer KI-Agenten und Automatisierungssysteme besonders nuetzlich, weil es Kontext, Zustaendigkeiten und Vertrauenssignale strukturiert bereitstellt.

Ein Agent kann unter anderem beantworten:

- In welchem Kosmos befinde ich mich?
- Welche Domaene ist zustaendig?
- Welcher Service ist relevant?
- Welche Faehigkeit wird benoetigt?
- Welche Regeln gelten?
- Welche Entscheidungen muessen getroffen werden?
- Welcher Prozess ist vorgesehen?
- Welcher Skill darf ausgefuehrt werden?
- Ist die Quelle vertrauenswuerdig?
- Ist das Artefakt geprueft, signiert oder nur experimentell?

Wichtig bleibt:

- KI sollte nicht Source of Truth sein.
- KI kann navigieren, erklaeren, vorschlagen, validieren, vergleichen und Luecken aufzeigen.
- Finale Autoritaet bleibt bei versionierten, reviewten und verifizierbaren Artefakten.

## 15. MVP-Abgrenzung

Die Idee darf MVP 0.1 nicht ueberladen. Fuer den Einstieg sollte nur eine minimale Struktur vorbereitet werden.

Fokus fuer MVP 0.1:

- cosmos-Metadaten erlauben,
- domain-Metadaten erlauben,
- service-Metadaten erlauben,
- capability optional als leichtgewichtiges Artefakt einfuehren,
- verification status vorerst nur als Metadatum fuehren,
- keine vollstaendige DNS-Verifikation implementieren,
- keine vollstaendige Signatur-Infrastruktur implementieren,
- kein komplexes Trust Registry Modell implementieren.

Vorgeschlagene MVP-Felder:

- cosmos_id
- domain_id
- service_id
- capability_id
- owner
- namespace
- verification.status

Moegliche verification.status-Werte:

- unverified
- planned
- dns_verified
- signed
- approved

Die eigentliche Verifikation kann in spaeteren Phasen umgesetzt werden.

## 16. Risiken und offene Fragen

### Risiken

- konzeptionelle Komplexitaet,
- Gefahr von Over-Engineering,
- unklare Ownership in realen Organisationen,
- Verwechslung zwischen Service, Produkt, Faehigkeit und Skill,
- DNS beweist technische Kontrolle, nicht automatisch organisatorische Legitimation,
- Git-Workflows koennen fuer manche Fachbereiche zu technisch sein,
- Trust- und Signaturmodelle koennen schnell komplex werden,
- Foederation braucht sorgfaeltiges Konfliktmanagement.

### Offene Fragen

- Ist ein Kosmos immer genau einem DNS-Namespace zugeordnet?
- Kann ein Kosmos mehrere DNS-Namespaces enthalten?
- Kann eine Domaene zu mehreren Kosmosse gehoeren?
- Wie werden delegierte Domaenen modelliert?
- Wie werden widerspruechliche Regeln zwischen Kosmosse behandelt?
- Was ist der genaue Unterschied zwischen official, approved, trusted und authoritative?
- Soll capability bereits in MVP 0.1 ein First-Class-Artefakt sein?
- Wie sollen Git-Branches auf draft/review/approved-Zustaende abgebildet werden?
- Wie koennen nicht-technische Nutzer mit Git-basierten Artefakten ueber UI arbeiten?

## 17. Fazit

Nomos sollte sich von einem Produkt/Regel/Skill-Katalog hin zu einem verifizierbaren semantischen Wissens- und Ausfuehrungsraum weiterentwickeln.

Das Kosmos-Modell ergaenzt:

- Namespace,
- Ownership,
- Trust,
- Foederation,
- Git-basierte Governance,
- DNS-aehnliche Verifizierbarkeit,
- strukturierte Navigation von Kosmos zu Domaene zu Service zu Faehigkeit zu Regeln, Entscheidungen, Prozessen und Skills.

> Nomos modelliert damit nicht nur fachliche Wahrheit, sondern auch deren Herkunft, Zustaendigkeit und Vertrauenswuerdigkeit.

## 18. Weiterfuehrende Arbeit

Moegliche naechste Schritte:

- Terminologie schaerfen,
- ersten Schema-Entwurf fuer cosmos/domain/service/capability erstellen,
- Beispiele fuer `blumer.cloud` oder `admin.ch` ergaenzen,
- minimale MVP-Metadaten festlegen,
- entscheiden, ob capability ein First-Class-Artefakt wird,
- Git-Repository-Konventionen definieren,
- initiales Verifikationsmodell definieren,
- in einem Folgepapier ein visuelles Mermaid-Diagramm erarbeiten.

---

Dieses Dokument ist bewusst als Idea Paper formuliert. Es beschreibt eine konzeptionelle Richtung und dient als Grundlage fuer spaetere Architekturentscheidungen, nicht als finale Spezifikation.
