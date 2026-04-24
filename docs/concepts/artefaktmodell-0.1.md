<!--
Source: Artefaktmodell 0.1.pdf
Converted from PDF text extraction for use in the nomos repository.
Note: Diagrams embedded as images in the original PDF are represented by textual placeholders/notes where text extraction cannot preserve the visual layout.
-->

# Artefaktmodell 0.1
Strukturmodell für die fachliche Abbildung von Produkten, Regeln, Entscheidungen, Prozessen
und Skills

## 1. Zielsetzung
Das Artefaktschema 0.1 beschreibt die erste fachliche Struktur für die geplante Provisionierungsplattform. Es definiert, welche Artefakttypen benötigt
werden, welche Mindestinformationen diese enthalten sollen und wie sie zueinander in Beziehung stehen.

Das Ziel des Schemas besteht darin, eine einheitliche Grundlage für folgende Aspekte zu schaffen:

         strukturierte Abbildung von Produktanforderungen
         fachliche Beschreibung von Business Rules
         Abbildung von Entscheidungen
         Beschreibung von Provisionierungsprozessen
         Zuordnung operativer Skills zu fachlichen Tasks
         spätere technische Nutzung durch Services, APIs, KI-Agenten und Orchestrierungsplattformen

Das Schema ist bewusst als erste Version gehalten. Es soll fachlich klar, technisch verarbeitbar und für eine Ablage in Git oder in einem strukturierten
Repository geeignet sein.


## 2. Leitprinzipien des Artefaktmodells
Das Artefaktmodell folgt mehreren Grundprinzipien.


2.1 Trennung der fachlichen Ebenen
Produkte, Regeln, Entscheidungen, Prozesse und Skills werden als eigene Artefakte geführt. Dadurch wird vermieden, dass unterschiedliche fachliche
Sichten in einem einzigen Dokument oder Datenobjekt vermischt werden.


2.2 Referenzierung statt Vollständigkeitsduplikation
Ein Produkt referenziert die zugehörigen Regeln, Entscheidungen, Prozesse und Skills, anstatt deren gesamte Logik mehrfach inline zu definieren.
Dadurch bleibt die Struktur wartbar und wiederverwendbar.


2.3 Fachliche Lesbarkeit
Die Artefakte sollen nicht nur maschinell verarbeitbar, sondern auch für Fachpersonen, Architektur, Governance und Betrieb lesbar sein.


2.4 KI-Nutzbarkeit
Die Struktur soll so gestaltet sein, dass KI-Agenten die Artefakte später gezielt lesen, interpretieren und in Provisionierungsprozessen nutzen können.


2.5 Erweiterbarkeit
Das Modell ist so aufgebaut, dass spätere Erweiterungen wie BPMN, DMN, Connectoren, Event-Modelle oder Governance-Regeln schrittweise ergänzt
werden können.


## 3. Empfohlene logische Ablagestruktur
Für die erste Version wird eine logische Struktur empfohlen, die Produkte, Regeln, Entscheidungen, Skills und Prozesse sauber trennt.


 catalog/
   products/
     benutzerkonto-mit-mailbox.yaml

      rules/
        konto/
          pflichtidentitaet.yaml
          benutzerart.yaml
          externe-enddatum.yaml
          gueltiger-tenant.yaml
          privilegierte-freigabe.yaml
          eindeutiger-benutzername.yaml
          mailadresskonvention.yaml
          bereitstellungsreife.yaml

      decisions/
        produktvariante-bestimmen.yaml
        bereitstellungsfreigabe.yaml
        mailbox-konfiguration.yaml

      skills/
        antrag-validieren.yaml
        produktvariante-bestimmen.yaml
        provisionierungsfreigabe-pruefen.yaml
        benutzername-mailadresse-erzeugen.yaml
        konto-im-iam-anlegen.yaml
        mailbox-provisionieren.yaml
        qualitaetspruefung.yaml
        audit-dokumentieren.yaml

      processes/
        provisionierung-benutzerkonto-mit-mailbox.yaml


Diese Struktur ist als fachlich-logische Zielstruktur zu verstehen. Sie kann später technisch angepasst oder erweitert werden.


## 4. Gemeinsame Basisfelder für alle Artefakte
Alle Artefakte sollen einige gemeinsame Basisfelder aufweisen, damit eine einheitliche Verarbeitung und Versionierung möglich wird.


4.1 Basisattribute
 id: Eindeutige-ID
 type: artefakttyp
 name: Lesbarer Name
 version: 0.1.0
 status: draft
 owner: Verantwortliche Domäne oder Rolle
 summary: Kurzbeschreibung
 tags:
   - fachlich
   - produkt


4.2 Bedeutung der Basisattribute
  Feld        Bedeutung

 id          Eindeutige Identifikation des Artefakts

 type        Typ des Artefakts, z. B. product, rule, decision

 name        Fachlich lesbarer Name


 version    Versionsstand des Artefakts

 status     Bearbeitungs- oder Freigabestatus

 owner      Verantwortliche Rolle, Domäne oder Organisationseinheit

 summary    Kurze fachliche Beschreibung

 tags       Verschlagwortung für Suche, Kategorisierung und KI-Nutzung


4.3 Empfohlene Statuswerte
 status:
   - draft
   - in_review
   - approved
   - deprecated


## 5. ID-Konvention
Fuer eine spätere eindeutige Referenzierung wird eine konsistente ID-Systematik empfohlen.

  Praefix     Bedeutung

 PROD-*      Produkt

 RULE-*      Regel

 DEC-*       Entscheidung

 SKILL-*     Skill

 PROC-*      Prozess

 QK-*        Qualitaetskriterium

 SYS-*       Zielsystem


Beispiele
         PROD-ACC-MBX-001
         RULE-ACC-003
         DEC-ACC-002
         SKILL-ACC-006
         PROC-ACC-MBX-001


## 6. Produktschema 0.1
6.1 Zweck des Produktschemas
Das Produktschema beschreibt den fachlichen Leistungsgegenstand. Es bildet ab, was das Produkt ist, welche Varianten existieren, welche Eingaben
benötigt werden und welche Regeln, Entscheidungen, Prozesse und Skills dem Produkt zugeordnet sind.


6.2 Produktschema
 id: PROD-ACC-MBX-001


type: product
name: Benutzerkonto mit Mailbox
version: 0.1.0
status: draft
owner: Identity & Collaboration
summary: >
  Bereitstellung eines Benutzerkontos mit zugehoeriger Mailbox
  fuer interne, externe oder privilegierte Benutzer.

tags:
  - identity
  - mailbox
  - provisionierung

description: >
  Das Produkt stellt einer Person ein Benutzerkonto und eine Mailbox bereit.
  Je nach Benutzerart gelten unterschiedliche Vorbedingungen, Freigaben
  und Qualitaetskriterien.

business_domain: Identity & Collaboration
trigger_events:
  - Neueintrittt
  - Rollenwechsel
  - Antrag fuer externen Benutzer

variants:
  - id: VAR-ACC-MBX-INT
    name: Internes Benutzerkonto mit Mailbox
    summary: Standardvariante fuer interne Mitarbeitende
  - id: VAR-ACC-MBX-EXT
    name: Externes Benutzerkonto mit Mailbox
    summary: Variante fuer externe Mitarbeitende oder Partner
  - id: VAR-ACC-MBX-PRIV
    name: Privilegiertes Benutzerkonto mit Mailbox
    summary: Variante fuer administrative oder besonders schutzbeduerftige Konten

inputs:
  required:
    - person_reference
    - first_name
    - last_name
    - user_type
    - tenant_id
  optional:
    - end_date
    - sponsor
    - org_unit
    - cost_center
    - approval_reference

input_definitions:
  person_reference:
    type: string
    description: Eindeutige Referenz auf die Person
  first_name:
    type: string
  last_name:
    type: string
  user_type:
    type: string
    allowed_values: [internal, external, privileged]
  tenant_id:
    type: string
  end_date:
    type: date
  sponsor:
    type: string
  approval_reference:
    type: string

target_systems:


    - SYS-IAM-001
    - SYS-M365-001

 requirements:
   - ANF-001
   - ANF-002
   - ANF-003
   - ANF-004
   - ANF-005
   - ANF-006
   - ANF-007

 rules:
   - RULE-ACC-001
   - RULE-ACC-002
   - RULE-ACC-003
   - RULE-ACC-004
   - RULE-ACC-005
   - RULE-ACC-006
   - RULE-ACC-007
   - RULE-ACC-008

 decisions:
   - DEC-ACC-001
   - DEC-ACC-002
   - DEC-ACC-003

 processes:
   - PROC-ACC-MBX-001

 quality_criteria:
   - QK-001
   - QK-002
   - QK-003
   - QK-004
   - QK-005
   - QK-006

 skills:
   - SKILL-ACC-001
   - SKILL-ACC-002
   - SKILL-ACC-003
   - SKILL-ACC-004
   - SKILL-ACC-005
   - SKILL-ACC-006
   - SKILL-ACC-007
   - SKILL-ACC-008


6.3 Fachliche Bedeutung
Das Produktartefakt ist das zentrale fachliche Objekt. Es beschreibt den Leistungsgegenstand und dient als Einstiegspunkt für alle weiteren Artefakte.


## 7. Ruleschema 0.1
7.1 Zweck des Ruleschemas
Eine Regel beschreibt eine fachliche Aussage in prüfbarer Form. Die Regel soll fachlich lesbar, aber gleichzeitig maschinell auswertbar sein.

Es wird empfohlen, in jeder Regel drei Ebenen zu unterscheiden:

         fachliche Beschreibung
         Anwendungsbereich
         auswertbare Logik


7.2 Regelschema
id: RULE-ACC-003
type: rule
name: Enddatum für externe Benutzer
version: 0.1.0
status: draft
owner: Identity Governance
summary: Externe Benutzer muessen ein gültiges Enddatum besitzen.

tags:
  - identity
  - external
  - precondition

description: >
  Wenn die Benutzerart extern ist, muss ein Enddatum vorhanden sein.
  Das Enddatum muss in der Zukunft liegen.

rule_category: precondition
severity: error
scope:
  products:
    - PROD-ACC-MBX-001
  variants:
    - VAR-ACC-MBX-EXT

applies_when:
  expression_language: simple
  expression: "input.user_type == 'external'"

logic:
  expression_language: simple
  expression: "input.end_date != null and input.end_date > today()"

result:
  on_success: passed
  on_failure: failed

message:
  success: Enddatum für externen Benutzer vorhanden und gültig.
  failure: Enddatum fehlt oder liegt nicht in der Zukunft.

evidence:
  - input.user_type
  - input.end_date

related_requirements:
  - ANF-003

related_findings:
  - FIND-001


7.3 Empfohlene Regelkategorien
rule_category:
  - precondition
  - validation
  - derivation
  - compliance
  - quality_gate
  - naming
  - approval


7.4 Empfohlene Severity-Werte
 severity:
   - info
   - warning
   - error
   - blocking


7.5 Fachliche Bedeutung
Regeln konkretisieren Anforderungen. Sie beschreiben, wie fachliche Erwartungen in eine nachvollziehbare und prüfbare Logik ueberführt werden.


## 8. Entscheidungsschema 0.1
8.1 Zweck des Entscheidungsschemas
Eine Entscheidung beschreibt eine strukturierte fachliche Bewertung. Sie nutzt Eingaben und Regeln, um ein fachliches Ergebnis zu bestimmen.

Eine Entscheidung soll nicht die Einzelregeln duplizieren, sondern den Entscheidungszweck, die benötigten Inputs, die verwendeten Regeln und die
möglichen Outputs explizit machen.


8.2 Entscheidungsschema


id: DEC-ACC-002
type: decision
name: Bereitstellungsfreigabe
version: 0.1.0
status: draft
owner: Identity Governance
summary: >
  Entscheidung, ob die Provisionierung gestartet werden darf,
  gesperrt ist oder manuell geklaert werden muss.

tags:
  - decision
  - readiness
  - provisioning

description: >
  Diese Entscheidung bewertet, ob alle relevanten Muss-Bedingungen
  für die Bereitstellung eines Benutzerkontos mit Mailbox erfüllt sind.

decision_type: eligibility
inputs:
  - person_reference
  - first_name
  - last_name
  - user_type
  - tenant_id
  - end_date
  - approval_reference

input_context:
  source: product_request

rules_used:
  - RULE-ACC-001
  - RULE-ACC-002
  - RULE-ACC-003
  - RULE-ACC-004
  - RULE-ACC-005
  - RULE-ACC-006
  - RULE-ACC-008

decision_logic:
  mode: aggregated_rules
  policy: all_blocking_rules_must_pass

outputs:
  decision_status:
    type: string
    allowed_values:
      - approved
      - rejected
      - manual_clarification
  reason_codes:
    type: list
  findings:
    type: list

explanations:
  enabled: true
  include_rules: true
  include_evidence: true

related_products:
  - PROD-ACC-MBX-001

related_processes:
  - PROC-ACC-MBX-001


8.3 Empfohlene Entscheidungstypen
 decision_type:
   - eligibility
   - classification
   - routing
   - configuration
   - approval
   - quality_assessment


8.4 Fachliche Bedeutung
Entscheidungen bilden die Brücke zwischen einzelnen Regeln und dem weiteren Prozessverlauf. Sie liefern ein fachlich relevantes Ergebnis, das später in
BPMN, APIs oder Services genutzt werden kann.


## 9. Skillschema 0.1
9.1 Zweck des Skillschemas
Ein Skill beschreibt einen operativen Bearbeitungsbaustein. Er repräsentiert die ausführbare Logik für einen fachlichen Task. Ein Skill ist damit mehr als
ein Skript oder technischer Connector. Er enthält auch fachlichen Zweck, Eingaben, Regeln, Entscheidungen, erwartete Ergebnisse und erzeugte
Nachweise.


9.2 Skillschema
 id: SKILL-ACC-006
 type: skill
 name: Mailbox provisionieren
 version: 0.1.0
 status: draft
 owner: Messaging Operations
 summary: Provisioniert eine Mailbox im Zielsystem Microsoft 365.

 tags:
   - mailbox
   - exchange
   - provisioning

 description: >
   Dieser Skill legt für ein bereits freigegebenes Benutzerkonto
   die zugehörige Mailbox im vorgesehenen Tenant an.

 skill_type: execution
 trigger_mode:
   - process_task
   - api_call

 supported_products:
   - PROD-ACC-MBX-001

 process_tasks:
   - TASK-ACC-006

 preconditions:
   decisions_required:
     - DEC-ACC-002
     - DEC-ACC-003
   rules_required:
     - RULE-ACC-004
     - RULE-ACC-008

 inputs:


  required:
    - account_id
    - mailbox_address
    - tenant_id
    - user_type
  optional:
    - mailbox_profile
    - license_plan

input_contract:
  account_id:
    type: string
  mailbox_address:
    type: string
  tenant_id:
    type: string
  user_type:
    type: string

actions:
  - validate_inputs
  - resolve_mailbox_configuration
  - call_target_system
  - collect_response
  - create_evidence

target_systems:
  - SYS-M365-001

rules_used:
  - RULE-ACC-004
  - RULE-ACC-008

decisions_used:
  - DEC-ACC-003

outputs:
  mailbox_id:
    type: string
  primary_smtp_address:
    type: string
  provisioning_status:
    type: string

possible_findings:
  - FIND-004

evidence_generated:
  - mailbox_creation_response
  - mailbox_identifier
  - timestamp

error_handling:
  retryable_errors:
    - target_system_temporarily_unavailable
  non_retryable_errors:
    - invalid_tenant
    - invalid_mailbox_configuration

related_processes:
  - PROC-ACC-MBX-001


9.3 Empfohlene Skill-Typen


 skill_type:
   - validation
   - decision_support
   - derivation
   - execution
   - quality_check
   - documentation


9.4 Fachliche Bedeutung
Skills bilden die operative Ebene der Plattform. Sie setzen Tasks um, greifen auf Regeln und Entscheidungen zurück und interagieren bei Bedarf mit
Zielsystemen.


## 10. Prozessschema 0.1
10.1 Zweck des Prozessschemas
Das Prozessartefakt beschreibt den fachlichen Ablauf eines Provisionierungsvorgangs. In der ersten Version wird empfohlen, den Prozess zunächst als
lesbares Fachartefakt zu modellieren und erst später auf ein konkretes BPMN-Format zu überführen.

Das Prozessartefakt beschreibt:

        Trigger
        Tasks
        erwartete Ergebnisse
        Verzweigungen
        Endzustände
        Referenzen auf Skills und Entscheidungen


10.2 Prozessschema
 id: PROC-ACC-MBX-001
 type: process
 name: Provisionierung Benutzerkonto mit Mailbox
 version: 0.1.0
 status: draft
 owner: Identity & Collaboration
 summary: >
   Fachlicher End-to-End-Prozess für die Bereitstellung eines
   Benutzerkontos mit Mailbox.

 tags:
   - provisioning
   - identity
   - mailbox

 description: >
   Der Prozess beschreibt die fachlichen Schritte von der Antragsprüfung
   bis zum dokumentierten Abschluss der Provisionierung.

 process_type: provisioning
 trigger_events:
   - new_joiner
   - role_change
   - external_user_request

 related_products:
   - PROD-ACC-MBX-001

 related_decisions:
   - DEC-ACC-001
   - DEC-ACC-002
   - DEC-ACC-003


tasks:
  - id: TASK-ACC-001
    name: Antrag prüfen
    task_type: validation
    skill_ref: SKILL-ACC-001
    expected_outcomes:
       - valid_request
       - incomplete_request

  - id: TASK-ACC-002
    name: Produktvariante bestimmen
    task_type: decision
    skill_ref: SKILL-ACC-002
    decision_ref: DEC-ACC-001
    expected_outcomes:
      - internal_variant
      - external_variant
      - privileged_variant

  - id: TASK-ACC-003
    name: Bereitstellungsfreigabe prüfen
    task_type: decision
    skill_ref: SKILL-ACC-003
    decision_ref: DEC-ACC-002
    expected_outcomes:
      - approved
      - rejected
      - manual_clarification

  - id: TASK-ACC-004
    name: Benutzername und Mailadresse erzeugen
    task_type: derivation
    skill_ref: SKILL-ACC-004
    expected_outcomes:
      - identifiers_generated
      - collision_detected

  - id: TASK-ACC-005
    name: Konto im IAM anlegen
    task_type: execution
    skill_ref: SKILL-ACC-005
    expected_outcomes:
      - account_created
      - account_creation_failed

  - id: TASK-ACC-006
    name: Mailbox provisionieren
    task_type: execution
    skill_ref: SKILL-ACC-006
    decision_ref: DEC-ACC-003
    expected_outcomes:
      - mailbox_created
      - mailbox_creation_failed

  - id: TASK-ACC-007
    name: Qualitaetsprüfung durchführen
    task_type: quality_check
    skill_ref: SKILL-ACC-007
    expected_outcomes:
      - quality_passed
      - rework_required

  - id: TASK-ACC-008
    name: Nachweise und Audit dokumentieren
    task_type: documentation
    skill_ref: SKILL-ACC-008
    expected_outcomes:
      - case_completed

flow:


  - from: TASK-ACC-001
    to: TASK-ACC-002

  - from: TASK-ACC-002
    to: TASK-ACC-003

  - from: TASK-ACC-003
    to: TASK-ACC-004
    when: approved

  - from: TASK-ACC-003
    to: END
    when: rejected

  - from: TASK-ACC-003
    to: MANUAL_REVIEW
    when: manual_clarification

  - from: TASK-ACC-004
    to: TASK-ACC-005
    when: identifiers_generated

  - from: TASK-ACC-004
    to: MANUAL_REVIEW
    when: collision_detected

  - from: TASK-ACC-005
    to: TASK-ACC-006
    when: account_created

  - from: TASK-ACC-006
    to: TASK-ACC-007
    when: mailbox_created

  - from: TASK-ACC-007
    to: TASK-ACC-008
    when: quality_passed

  - from: TASK-ACC-007
    to: REWORK
    when: rework_required

end_states:
  - completed
  - rejected
  - manual_review
  - rework

bpmn:
  planned: true
  bpmn_process_key: account_mailbox_provisioning


10.3 Empfohlene Task-Typen
task_type:
  - validation
  - decision
  - derivation
  - execution
  - quality_check
  - documentation
  - manual_review


10.4 Fachliche Bedeutung


Das Prozessartefakt bildet die fachliche Orchestrierung ab. Es zeigt, welche Schritte notwendig sind, in welcher Reihenfolge diese erfolgen und wie
Entscheidungen den weiteren Ablauf beeinflussen.


## 11. Inhalte, die in Version 0.1 bewusst noch nicht modelliert werden
Um das Modell in der ersten Version schlank und steuerbar zu halten, werden bestimmte Aspekte noch nicht oder nur am Rand berücksichtigt.

Dazu gehören insbesondere:

         vollständige BPMN-XML-Artefakte
         vollständige DMN-Tabellen
         tiefe Rollen- und Berechtigungsmodelle
         technische Connector-Definitionen
         Event-Sourcing-Strukturen
         Migrations- und Vererbungslogiken
         ausformulierte API-Spezifikationen
         umfangreiche Mandantenlogiken

Diese Punkte können in späteren Versionen ergänzt werden.


## 12. Mindestvalidierungen für das Artefaktmodell selbst
Bereits in der ersten Version sollen einige Meta-Regeln für die Qualität der Artefakte gelten.


12.1 Allgemeine Regeln für alle Artefakte
         id muss vorhanden und eindeutig sein
         type muss gesetzt sein
         version muss gesetzt sein
         status muss gesetzt sein
         owner muss gesetzt sein


12.2 Regeln für Produktartefakte
         mindestens ein Zielsystem
         mindestens ein referenzierter Prozess
         mindestens eine referenzierte Regel oder Entscheidung


12.3 Regeln für Regelartefakte
         rule_category muss gesetzt sein
         severity muss gesetzt sein
         logic.expression muss gesetzt sein


12.4 Regeln für Entscheidungsartefakte
         mindestens ein definierter Output
         mindestens eine referenzierte Regel oder eine explizite Entscheidungslogik


12.5 Regeln für Skill-Artefakte
         mindestens ein Input oder ein klarer Trigger
         mindestens ein Output, Finding oder Nachweis
         bei skill_type: execution mindestens ein Zielsystem


12.6 Regeln für Prozessartefakte
         mindestens ein Task
         jeder Task benötigt skill_ref
         jeder Flow darf nur auf existente Task-IDs oder definierte Endpunkte verweisen


## 13. Fachliches Zusammenspiel der Artefakte
Das Zusammenspiel der Artefakte lässt sich wie folgt beschreiben:


Produkt
Das Produkt ist das fachliche Leitobjekt. Es referenziert Regeln, Entscheidungen, Prozesse, Skills, Zielsysteme und Qualitätskriterien.


Regel
Die Regel bildet eine prüfbare fachliche Aussage ab und konkretisiert Anforderungen.


Entscheidung
Die Entscheidung nutzt Regeln und Eingaben, um ein fachliches Ergebnis zu bestimmen.


Prozess
Der Prozess beschreibt die fachliche Reihenfolge der Bearbeitung und die möglichen Pfade.


Skill
Der Skill führt einen Task operativ aus und nutzt dabei bei Bedarf Regeln, Entscheidungen und Zielsysteme.


## 14. Empfehlung für den MVP-Schnitt
Für einen ersten umsetzbaren MVP wird empfohlen, zunächst nur die folgenden Artefakttypen verbindlich einzuführen:

         product
         rule
         decision
         skill
         process

Mit diesen fünf Artefakttypen kann bereits ein fachlich konsistenter End-to-End-Durchstich realisiert werden.

Als erstes Referenzprodukt bietet sich das Produkt Benutzerkonto mit Mailbox an, da sich daran Produktanforderungen, Vorbedingungen,
Entscheidungen, Prozessschritte, Skills und Zielsystemanbindungen gut demonstrieren lassen.
