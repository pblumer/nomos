<!--
Source: Fachliche_Modellierungsvorlage_Servicegraph_nach_metamodell_erfassen.pdf
Converted from PDF for use in the nomos repository.
-->

# Fachliche Modellierungsvorlage: Servicegraph nach Metamodell erfassen

## 1. Zweck

Diese Vorlage dient dazu, einen Service nach dem fachlichen Metamodell der graphbasierten Service-Modellierung strukturiert zu dokumentieren. Pro Service sollen mindestens folgende Aspekte erfasst werden:

- fachliche Struktur
- Knoten
- Beziehungen
- Regeln
- Varianten
- Abhängigkeiten
- Validierungen
- manuelle Schritte
- Kompensationen


## 2. Empfohlene Seitenstruktur

### 2.1 Kopfbereich

Service-Name, Kurzbeschreibung, Version, Status, verantwortliche Rolle, Gültig ab/bis, Modell-ID.

### 2.2 Überblick

Fachliches Ziel, Varianten, wichtigste Vorbedingungen, wichtigste Abhängigkeiten, Besonderheiten.

### 2.3 Knotenliste

Alle Knoten des Servicegraphen.

### 2.4 Beziehungsliste

Alle Kanten des Servicegraphen.

### 2.5 Regelliste

Aktivierungs-, Varianten-, Validierungs- und Konsistenzregeln.

### 2.6 Sichten

Hierarchiesicht, Abhängigkeitssicht, Ausführungssicht, Ausnahme-/Kompensationssicht.


## 3. Vorlage: Kopfbereich

| Feld | Beschreibung | Beispiel |
|---|---|---|
| Modell-ID | Eindeutige Kennung | SG-ACC-001 |
| Modellname | Fachlicher Name | Benutzerkonto mit Exchange-Mailbox bereitstellen |
| Kurzbeschreibung | Kurze Beschreibung | Modell zur Bereitstellung eines Kontos inkl. Mailbox |
| Version | Versionsstand | 0.1 |
| Status | Entwurf / in Review / freigegeben / ausser Betrieb | Entwurf |
| Gültig ab | Beginn der Gültigkeit | 2026-04-16 |
| Gültig bis | Ende der Gültigkeit | offen |
| Verantwortliche Rolle | Fachliche Modellverantwortung | Serviceverantwortung IAM |
| Bezugsservice | Bestellbarer Service | Benutzerkonto mit Exchange-Mailbox bereitstellen |


## 4. Vorlage: Knotenliste

### Standardvorlage

| Knoten-ID | Name | Knotentyp | Kurzbeschreibung | Pflicht/Optional | Variante | Aktivierungsregel | Verantwortliche Rolle | Wiederverwendbar | Bemerkung |
|---|---|---|---|---|---|---|---|---|---|
| | | | | | | | | | |

### Erläuterung

| Feld | Bedeutung |
|---|---|
| Knoten-ID | Eindeutige Identifikation |
| Name | Fachliche Bezeichnung |
| Knotentyp | Service, Sub-Service, Aktivität, Vorbedingung, Entscheidung, Validierung, Manuell, Kompensation, Zustand, Referenz |
| Kurzbeschreibung | Wofür der Knoten steht |
| Pflicht/Optional | Immer aktiv oder nur in bestimmten Fällen |
| Variante | Für welche Variante der Knoten gilt |
| Aktivierungsregel | Regelreferenz oder Bedingung |
| Verantwortliche Rolle | Fachlich zuständige Rolle |
| Wiederverwendbar | Kennzeichnung für Mehrfachverwendung |
| Bemerkung | Zusätzliche Hinweise |

### Erweiterte Knotenliste (optional)

Zusätzlich: Parent-Knoten, Ausführungsart, Retry-fähig, Idempotenz erwartet.


## 5. Vorlage: Beziehungsliste

### Standardvorlage

| Beziehungs-ID | Quell-Knoten-ID | Ziel-Knoten-ID | Beziehungstyp | Verbindlichkeit | Bedingung | Kurzbeschreibung |
|---|---|---|---|---|---|---|
| | | | | | | |

### Empfohlene Beziehungstypen für den Start

| Beziehungstyp | Bedeutung |
|---|---|
| besteht aus | strukturelle Zugehörigkeit |
| depends on | fachliche oder technische Voraussetzung |
| validates | ein Knoten prüft einen anderen |
| requires condition | Aktivierung nur unter Bedingung |
| compensates | fachliche Gegenmassnahme |
| alternative to | alternative Pfade oder Varianten |


## 6. Vorlage: Regelliste

| Regel-ID | Regelname | Regeltyp | Beschreibung | Geltungsbereich | Verbindlichkeit | Ausdruck / Fachlogik |
|---|---|---|---|---|---|---|
| | | | | | | |

### Empfohlene Regeltypen

| Regeltyp | Bedeutung |
|---|---|
| Aktivierungsregel | Wann wird ein Knoten aktiv |
| Variantenregel | Für welche Variante gilt ein Knoten/Beziehung |
| Validierungsregel | Wann gilt ein Schritt als erfolgreich |
| Konsistenzregel | Prüft Modellqualität |
| Kompensationsregel | Wann und wie wird kompensiert |


## 7. Vorlage: Variantensicht

| Varianten-ID | Variantenname | Beschreibung | Aktivierte Knoten | Deaktivierte Knoten | Spezielle Regeln | Bemerkung |
|---|---|---|---|---|---|---|
| | | | | | | |


## 8. Vorlage: Hierarchiesicht

| Parent-Knoten-ID | Parent-Name | Child-Knoten-ID | Child-Name | Beziehung |
|---|---|---|---|---|
| | | | | besteht aus |


## 9. Vorlage: Abhängigkeitssicht

| Ziel-Knoten-ID | Ziel-Knoten | Hängt ab von Knoten-ID | Hängt ab von | Verbindlichkeit | Bedingung | Bemerkung |
|---|---|---|---|---|---|---|
| | | | | | | |


## 10. Vorlage: Ausführungssicht

| Ausführungsschritt | Knoten-ID | Knotenname | Ausführungsart | Vorbedingungen erfüllt durch | Validierung | Fehlerbehandlung |
|---|---|---|---|---|---|---|
| | | | | | | |


## 11. Vorlage: Ausnahme- und Kompensationssicht

| Fehler-/Ausnahmesituation | Betroffener Knoten | Reaktion | Manueller Schritt | Kompensationsknoten | Bemerkung |
|---|---|---|---|---|---|
| | | | | | |


## 12. Konsistenzprüfung (Checkliste)

| Prüffrage | Ja/Nein | Bemerkung |
|---|---|---|
| Haben alle Knoten eine eindeutige ID? | | |
| Haben alle Knoten einen Typ? | | |
| Sind alle harten Abhängigkeiten explizit modelliert? | | |
| Gibt es Zyklen in harten depends-on-Beziehungen? | | |
| Sind Varianten explizit beschrieben? | | |
| Sind Validierungen modelliert? | | |
| Sind manuelle Schritte modelliert? | | |
| Sind Fehler- und Kompensationspfade berücksichtigt? | | |
| Ist die Verantwortlichkeit pro Knoten nachvollziehbar? | | |


## 13. Minimale Pflichtinhalte (Phase 1)

**Pflicht:** Kopfbereich, Knotenliste, Beziehungsliste, Regelliste, Variantensicht, Abhängigkeitssicht.

**Optional in Phase 1:** Detaillierte Ausführungssicht, separate Kompensationssicht, Zustandsknoten, Referenzknoten.


## 14. Modellierungskonventionen

### Knoten-IDs

Format: `<Servicekürzel>-N-<laufende Nummer>` (z.B. ACC-N-010)

### Beziehungs-IDs

Format: `<Servicekürzel>-R-<laufende Nummer>` (z.B. ACC-R-025)

### Regel-IDs

Format: `<Servicekürzel>-G-<laufende Nummer>` (z.B. ACC-G-003)

### Benennung

- Knoten mit klaren Substantiven oder fachlichen Tätigkeiten benennen
- Beziehungstypen standardisiert verwenden
- keine versteckten Regeln im Freitext
