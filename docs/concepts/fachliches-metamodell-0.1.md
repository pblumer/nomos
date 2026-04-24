<!--
Source: Fachliches Metamodell 0.1.pdf
Converted from PDF text extraction for use in the nomos repository.
Note: Diagrams embedded as images in the original PDF are represented by textual placeholders/notes where text extraction cannot preserve the visual layout.
-->

# Fachliches Metamodell 0.1
## 1. Ziel des Metamodells
Das Metamodell beschreibt, welche fachlichen Objekte die Plattform kennt und wie diese zusammenhängen.

Es soll verhindern, dass man später:

        Produktwissen irgendwo separat hält,
        Regeln anders strukturiert als Entscheidungen,
        Tasks losgelöst von Skills modelliert,
        oder Prozesse direkt mit technischer Logik vermischt.


## 2. Zentrale fachliche Artefakttypen
2.1 Produkt
Beschreibt den fachlichen Leistungsgegenstand.

Beispiel:
Benutzerkonto mit Mailbox

Typische Attribute

        Produkt-ID
        Name
        Beschreibung
        Zweck
        Varianten
        Zielsysteme
        verantwortliche Domäne
        Status
        Version


2.2 Produktvariante
Beschreibt eine fachliche Ausprägung eines Produkts.

Beispiele

        internes Benutzerkonto mit Mailbox
        externes Benutzerkonto mit Mailbox
        privilegiertes Benutzerkonto mit Mailbox

Typische Attribute

        Varianten-ID
        Name
        Beschreibung
        Gültigkeitsbedingungen
        zusätzliche Anforderungen
        zusätzliche Regeln


2.3 Anforderung
Beschreibt eine fachliche Erwartung oder Vorgabe.

Beispiele

        Für jedes Konto muss eine eindeutige Identität vorliegen.
        Für eine Mailbox muss ein gültiger Ziel-Tenant bestimmt sein.
        Externe Benutzer benötigen ein Ablaufdatum.


Typische Attribute

         Anforderungs-ID
         Titel
         Beschreibung
         Kategorie
         Geltungsbereich
         Quelle
         Priorität
         Version


2.4 Business Rule
Beschreibt eine prüfbare fachliche Regel.

Beispiele

         Ein Benutzerkonto darf nur erstellt werden, wenn Vorname und Nachname vorhanden sind.
         Externe Benutzer müssen ein Enddatum besitzen.
         Die Mailadresse muss gemäss Namenskonvention erzeugt werden.

Typische Attribute

         Rule-ID
         Titel
         Beschreibung
         Typ
         Geltungsbereich
         Bedingung
         Wirkung
         Severity
         Begründung
         Nachweise
         Version


2.5 Entscheidung
Beschreibt eine strukturierte fachliche Bewertung.

Beispiele

         Darf die Provisionierung gestartet werden?
         Welche Kontovariante ist zu wählen?
         Muss eine Freigabe eingeholt werden?
         Welche Mailbox-Konfiguration ist anzuwenden?

Typische Attribute

         Entscheidungs-ID
         Name
         Zweck
         Input
         Entscheidungslogik
         Output
         Erklärung
         referenzierte Regeln
         Version


2.6 Qualitätskriterium
Beschreibt, wann ein Antrag oder Ergebnis fachlich ausreichend ist.

Beispiele

         Alle Pflichtdaten sind vorhanden.
         Zielsysteme wurden erfolgreich aktualisiert.


         Mailbox wurde erstellt und ist erreichbar.
         Nachweise wurden abgelegt.

Typische Attribute

         Qualitätskriteriums-ID
         Name
         Beschreibung
         Prüfmethode
         Muss/Soll
         Geltungsbereich


2.7 Prozess
Beschreibt den fachlichen Ablauf der Provisionierung.

Beispiel
Provisionierung Benutzerkonto mit Mailbox

Typische Attribute

         Prozess-ID
         Name
         Ziel
         Trigger
         Startbedingung
         Endzustände
         referenzierte Entscheidungen
         referenzierte Tasks
         Version


2.8 Task
Beschreibt einen fachlichen Arbeitsschritt innerhalb eines Prozesses.

Beispiele

         Antrag prüfen
         Identität validieren
         Konto im IAM anlegen
         Mailbox in Microsoft 365 erstellen
         Qualitätskontrolle durchführen

Typische Attribute

         Task-ID
         Name
         Zweck
         Eingaben
         Ergebnisse
         Fehlerfälle
         Nachweise
         Reihenfolge / Position im Prozess


2.9 Skill
Beschreibt die ausführbare Bearbeitungslogik für einen Task.

Beispiele

         Skill zur Identitätsprüfung
         Skill zur Namensbildung
         Skill zur Kontoanlage
         Skill zur Mailbox-Erstellung
         Skill zur Qualitätsprüfung

Typische Attribute


        Skill-ID
        Name
        Zweck
        Trigger
        erwartete Eingaben
        verwendete Regeln
        referenzierte Entscheidungen
        Rückgabewerte
        Findings
        Nachweise
        Zielsysteme
        Version


2.10 Zielsystem
Beschreibt ein angebundenes System, in dem Aktionen ausgeführt oder Daten gelesen werden.

Beispiele

        IAM / Verzeichnisdienst
        Microsoft 365 / Exchange Online
        ServiceNow
        SAP

Typische Attribute

        Zielsystem-ID
        Name
        Typ
        Rolle im Prozess
        Schnittstellenart
        verantwortliche Einheit


2.11 Nachweis
Beschreibt einen Beleg, dass etwas geprüft, entschieden oder ausgeführt wurde.

Beispiele

        Entscheidungsprotokoll
        API-Response eines Zielsystems
        erzeugter Benutzername
        Mailbox-ID
        Freigabevermerk

Typische Attribute

        Nachweis-ID
        Typ
        Inhalt / Referenz
        Quelle
        Zeitstempel
        Bezug auf Objekt


2.12 Finding
Beschreibt eine festgestellte Abweichung, Warnung oder Sperrbedingung.

Beispiele

        Pflichtfeld fehlt
        Tenant nicht gültig
        Mailadresse kollidiert
        Freigabe fehlt
        Zielsystem nicht erreichbar

Typische Attribute


        Finding-ID
        Kategorie
        Beschreibung
        Severity
        betroffene Objekte
        Handlungsempfehlung
        Status


## 3. Beziehungen zwischen den Artefakten
Die wichtigsten Beziehungen sind:

        Ein Produkt hat eine oder mehrere Produktvarianten
        Ein Produkt hat mehrere Anforderungen
        Eine Anforderung wird durch eine oder mehrere Business Rules konkretisiert
        Mehrere Business Rules fliessen in eine Entscheidung ein
        Ein Produkt referenziert einen Prozess
        Ein Prozess besteht aus mehreren Tasks
        Jeder Task wird durch genau einen oder mehrere Skills ausgeführt
        Ein Skill nutzt Business Rules, Entscheidungen und Zielsysteme
        Ein Skill erzeugt Nachweise und gegebenenfalls Findings
        Qualitätskriterien prüfen Produkt, Prozessschritt oder Ergebnis
        Zielsysteme werden durch Skills oder Plattformservices angesprochen


## 4. Abgrenzung der Ebenen
Produkt
Was soll bereitgestellt werden?


Anforderung
Was muss dafür fachlich gelten?


Rule
Wie wird eine fachliche Aussage prüfbar formuliert?


Entscheidung
Wie wird aus Eingaben und Regeln ein fachliches Ergebnis bestimmt?


Prozess
In welcher Reihenfolge werden Schritte fachlich ausgeführt?


Task
Welcher konkrete Arbeitsschritt ist im Prozess notwendig?


Skill
Wie wird ein Task operativ bearbeitet?


Nachweis / Finding
Was ist dokumentiert worden und welche Abweichungen wurden festgestellt?


## 5. Referenzbeispiel
Produkt: Benutzerkonto mit Mailbox

5.1 Produktbeschreibung
Produkt-ID: PROD-ACC-MBX-001
Name: Benutzerkonto mit Mailbox
Zweck: Bereitstellung eines digitalen Benutzerkontos inklusive Mailbox für eine Person
Verantwortliche Domäne: Identity & Collaboration
Zielsysteme: IAM, Verzeichnisdienst, Microsoft 365 / Exchange Online
Trigger: Neueintrittt, Rollenwechsel, Anforderung eines externen Benutzers

Fachliche Beschreibung:
Das Produkt stellt einer Person ein Benutzerkonto mit Authentifizierungsfähigkeit sowie eine zugehörige Mailbox bereit. Je nach Benutzerart gelten
unterschiedliche Vorbedingungen, Freigaben und Qualitätskriterien.


5.2 Produktvarianten

Variante A: Internes Benutzerkonto mit Mailbox
         für interne Mitarbeitende
         benötigt Personalbezug
         Mailbox standardmässig aktiv
         keine Enddatumspflicht


Variante B: Externes Benutzerkonto mit Mailbox
         für externe Mitarbeitende oder Partner
         benötigt Sponsor / verantwortliche Stelle
         benötigt Enddatum
         ggf. zusätzliche Freigabe


Variante C: Privilegiertes Benutzerkonto mit Mailbox
         für administrative oder erhöhte Berechtigungen
         benötigt zusätzliche Genehmigung
         strengere Qualitäts- und Sicherheitsregeln


## 6. Anforderungen zum Referenzprodukt
ANF-001
Für jede Provisionierung muss eine eindeutig identifizierte Person vorliegen.


ANF-002
Für jede Provisionierung muss die Benutzerart bestimmt werden.


ANF-003
Für externe Benutzer muss ein Enddatum erfasst werden.


ANF-004
Für die Mailbox muss ein gültiger Ziel-Tenant vorliegen.


ANF-005
Für privilegierte Konten ist eine zusätzliche Freigabe erforderlich.


ANF-006
Die zu erzeugenden Identifikatoren müssen der Namenskonvention entsprechen.


ANF-007
Vor Abschluss der Provisionierung müssen fachliche und technische Qualitätskriterien erfüllt sein.


## 7. Beispielhafte Business Rules
BR-001 Pflichtidentität
Ein Konto darf nur provisioniert werden, wenn Vorname, Nachname und eindeutige Personenreferenz vorhanden sind.


BR-002 Benutzerart
Jeder Antrag muss genau eine Benutzerart besitzen: intern, extern oder privilegiert.


BR-003 Enddatum für Externe
Wenn Benutzerart = extern, dann muss ein Enddatum vorhanden und in der Zukunft liegend sein.


BR-004 Gültiger Tenant
Für die Mailbox-Provisionierung muss ein gültiger und aktiver Tenant ausgewählt sein.


BR-005 Freigabe privilegierter Konten
Wenn Benutzerart = privilegiert, dann muss eine zusätzliche Sicherheitsfreigabe vorliegen.


BR-006 Eindeutiger Benutzername
Der zu erzeugende Benutzername darf im Zielkontext nicht bereits vergeben sein.


BR-007 Mailadresskonvention
Die Mailadresse muss gemäss definierter Namenskonvention gebildet werden.


BR-008 Bereitstellungsreife
Die Provisionierung darf erst gestartet werden, wenn alle Muss-Anforderungen erfüllt sind.


## 8. Beispielhafte Entscheidungen
ENT-001 Bestimmung der Produktvariante
Input

         Benutzerart
         Rollenprofil
         Kontext der Anforderung

Output

         interne Variante
         externe Variante
         privilegierte Variante


ENT-002 Bereitstellungsfreigabe
Input

         Pflichtdaten vorhanden?
         Freigaben vorhanden?
         gültiger Tenant?
         Benutzername eindeutig?
         Sicherheitsauflagen erfüllt?

Output

         freigegeben
         nicht freigegeben
         manuelle Klärung erforderlich


ENT-003 Mailbox-Konfiguration
Input

         Tenant
         Benutzerart
         Organisationskontext

Output

         Standard-Mailbox
         eingeschränkte Mailbox
         Sonderkonfiguration


## 9. Qualitätskriterien
QK-001 Vollständigkeit
Alle Pflichtdaten zum Antrag sind vorhanden.


QK-002 Konsistenz
Die Daten widersprechen sich nicht.


QK-003 Entscheidungsfähigkeit
Alle benötigten Entscheidungen konnten ohne fachliche Lücke getroffen werden.


QK-004 Technische Ausführung
Konto und Mailbox wurden in den Zielsystemen erfolgreich angelegt.


QK-005 Nachweisbarkeit
Alle relevanten Nachweise und Protokolle sind vorhanden.


QK-006 Abschlussreife
Es existieren keine offenen Findings mit Severity Fehler oder Sperre.


## 10. Fachlicher Provisionierungsprozess
Prozess: PRC-ACC-MBX-001
Name: Provisionierung Benutzerkonto mit Mailbox


Prozessschritte
## 1. Antrag erfassen
## 2. Pflichtdaten und Kontext prüfen
## 3. Produktvariante bestimmen
## 4. Bereitstellungsfreigabe entscheiden
## 5. Benutzername und Mailadresse ableiten
## 6. Konto im IAM / Verzeichnis anlegen
## 7. Mailbox in Microsoft 365 anlegen
## 8. Qualitätsprüfung durchführen
## 9. Nachweise und Abschluss dokumentieren


## 11. BPMN-nahe Task-Struktur mit Skill-Zuordnung
Task 1: Antrag prüfen
Skill: SK-ACC-001 Antrag validieren
Zweck: Prüfung auf Vollständigkeit und Mindestkonsistenz
Eingaben: Personendaten, Benutzerart, Tenant, Freigaben, Kontext
Ergebnis: valider oder unvollständiger Antrag
Mögliche Findings: Pflichtfeld fehlt, Benutzerart unbekannt


Task 2: Produktvariante bestimmen
Skill: SK-ACC-002 Produktvariante bestimmen
Nutzt: ENT-001
Ergebnis: interne, externe oder privilegierte Variante


Task 3: Bereitstellungsfreigabe prüfen
Skill: SK-ACC-003 Provisionierungsfreigabe prüfen
Nutzt: BR-001 bis BR-008, ENT-002
Ergebnis: freigegeben / gesperrt / manuelle Klärung


Task 4: Identifikatoren bilden
Skill: SK-ACC-004 Benutzername und Mailadresse erzeugen
Nutzt: BR-006, BR-007
Ergebnis: Benutzername, Mailadresse
Mögliche Findings: Kollision, Konvention nicht anwendbar


Task 5: Konto anlegen
Skill: SK-ACC-005 Konto im IAM anlegen
Zielsystem: IAM / Verzeichnisdienst
Ergebnis: Konto-ID, technischer Nachweis


Task 6: Mailbox anlegen
Skill: SK-ACC-006 Mailbox provisionieren
Zielsystem: Microsoft 365 / Exchange
Nutzt: ENT-003
Ergebnis: Mailbox-ID, Routing-Adresse, technischer Nachweis


Task 7: Qualitätsprüfung
Skill: SK-ACC-007 Abschluss- und Qualitätspruefung
Nutzt: QK-001 bis QK-006
Ergebnis: erfolgreich / Nacharbeit erforderlich


Task 8: Dokumentation und Audit
Skill: SK-ACC-008 Nachweise und Audit-Eintrag erzeugen
Ergebnis: vollständige Nachweisakte


## 12. Beispiel für Findings
FIND-001
Kategorie: Pflichtdaten
Beschreibung: Enddatum fehlt für externen Benutzer
Severity: Fehler
Folge: Bereitstellung gesperrt


FIND-002
Kategorie: Identifikator
Beschreibung: Benutzername bereits vergeben
Severity: Warnung
Folge: Alternativbildung erforderlich


FIND-003
Kategorie: Freigabe
Beschreibung: Sicherheitsfreigabe für privilegiertes Konto fehlt
Severity: Sperre
Folge: manuelle Klärung


FIND-004
Kategorie: Zielsystem
Beschreibung: Mailbox konnte in Microsoft 365 nicht angelegt werden
Severity: Fehler
Folge: technische Nacharbeit


## 13. Beispiel für Nachweise
NACH-001 Entscheidungsprotokoll
         ermittelte Produktvariante
         Ergebnis Bereitstellungsfreigabe
         verwendete Regeln


NACH-002 Identifikator-Nachweis
         finaler Benutzername
         finale Mailadresse
         Kollisionsprüfung


NACH-003 IAM-Anlagenachweis
         Konto-ID
         Timestamp
         Zielsystemreferenz


NACH-004 Mailbox-Anlagenachweis
         Mailbox-ID
         Tenant
         Routing-Adresse


NACH-005 Abschlussnachweis
    Qualitätsprüfung bestanden
    keine offenen Sperr-Findings
    Vorgang abgeschlossen


## 14. Diagramm des Metamodells


## 15. Diagramm für das Referenzprodukt


## 16. Was daraus nun als nächster konkreter Projektschritt folgen sollte
Nach diesem Metamodell wäre der nächste wirklich sinnvolle Umsetzungsschritt:


Metamodell 0.1 in ein erstes Artefaktmodell überführen
Also konkret festlegen:

         wie ein Produktartefakt aussieht,
         wie eine Rule gespeichert wird,
         wie eine Entscheidung referenziert wird,
         wie ein Prozess mit Tasks auf Skills zeigt,
         und wie ein Skill-Vertrag fachlich beschrieben ist.


## Ergänzende Hinweise zu Diagrammen aus dem Original-PDF

Das Original-PDF enthält auf Seite 11 ein Diagramm des Metamodells. Es zeigt die Beziehungen zwischen Produkt, Produktvariante, Qualitätskriterium, Anforderung, Business Rule, Entscheidung, Prozess, Task, Skill, Zielsystem, Nachweis und Finding. Auf Seite 12 enthält das PDF ein Diagramm für das Referenzprodukt Benutzerkonto mit Mailbox. Dieses zeigt den Weg vom Produkt über Varianten, Anforderungen, Business Rules und Entscheidungen bis zum Provisionierungsprozess mit Tasks, Skills, Zielsystemen sowie Nachweisen und Findings.
