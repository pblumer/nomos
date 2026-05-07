<!--
Source: Instanzbeispiel Servicegraph - Benutzerkonto mit Exchange-Mailbox bereitstellen.pdf
Converted from PDF for use in the nomos repository.
-->

# Instanzbeispiel: Servicegraph – Benutzerkonto mit Exchange-Mailbox bereitstellen

## 1. Kopfbereich

| Feld | Wert |
|---|---|
| Modell-ID | ACC-001 |
| Modellname | Benutzerkonto mit Exchange-Mailbox bereitstellen |
| Kurzbeschreibung | Bereitstellung eines Benutzerkontos inklusive Exchange-Mailbox für neue Mitarbeitende |
| Version | 0.1 |
| Status | Entwurf |
| Gültig ab | 2026-04-16 |
| Gültig bis | offen |
| Verantwortliche Rolle | Serviceverantwortung Benutzerkonto |
| Bezugsservice | Benutzerkonto mit Exchange-Mailbox bereitstellen |


## 2. Überblick

**Fachliches Ziel:** Für einen neuen Mitarbeiter soll ein nutzbares Benutzerkonto mit Exchange-Mailbox bereitgestellt werden.

**Varianten:**
- U-Account für interne Mitarbeitende
- X-Account für externe Mitarbeitende

**Wichtigste Vorbedingungen:**
- Person ist bekannt
- Kontotyp ist bestimmt
- organisatorische Zuordnung ist bekannt
- Startdatum ist bekannt

**Wichtigste fachliche Abhängigkeit:** Die Mailbox darf erst bereitgestellt werden, wenn das Benutzerkonto vorhanden und fachlich ausreichend ausgesteuert ist.


## 3. Knotenliste

| Knoten-ID | Name | Knotentyp | Kurzbeschreibung | Pflicht/Optional | Variante | Aktivierungsregel | Verantwortliche Rolle | Wiederverwendbar | Bemerkung |
|---|---|---|---|---|---|---|---|---|---|
| ACC-N-001 | Benutzerkonto mit Exchange-Mailbox bereitstellen | Service-Knoten | Gesamter bestellbarer Service | Pflicht | alle | immer | Serviceverantwortung | Nein | Einstiegspunkt |
| ACC-N-010 | Kontotyp bestimmen | Entscheidungs-Knoten | Ermittelt U oder X | Pflicht | alle | immer | Fachverantwortung | Ja | steuert Variante |
| ACC-N-020 | Person bekannt | Vorbedingungs-Knoten | Personendaten liegen vor | Pflicht | alle | immer | anfordernde Rolle / Quellsystemverantwortung | Ja | Startvoraussetzung |
| ACC-N-030 | Organisation bekannt | Vorbedingungs-Knoten | OU bzw. organisatorischer Kontext ist bekannt | Pflicht | alle | immer | Fachverantwortung | Ja | Startvoraussetzung |
| ACC-N-040 | Benutzerkonto anlegen | Aktivitäts-Knoten | Anlage des Kontos | Pflicht | alle | wenn Person bekannt und Kontotyp bestimmt | Betriebs-/Integrationsverantwortung | Ja | Kernaktivität |
| ACC-N-050 | Basisattribute setzen | Aktivitäts-Knoten | Setzen von Sprache, Gültigkeit, organisatorischem Kontext | Pflicht | alle | nach Kontoanlage | Betriebs-/Integrationsverantwortung | Ja | fachliche Grundausprägung |
| ACC-N-060 | Mailbox bereitstellen | Aktivitäts-Knoten | Bereitstellung der Exchange-Mailbox | Pflicht | alle | nach Kontoanlage und Grundausprägung | Betriebs-/Integrationsverantwortung | Ja | zweiter Kernschritt |
| ACC-N-070 | Benutzerkonto validieren | Validierungs-Knoten | Prüft, ob das Konto vorhanden und korrekt ausgesteuert ist | Pflicht | alle | nach Kontoanlage und Attributsetzung | Betriebsverantwortung | Ja | blockierend |
| ACC-N-080 | Mailbox validieren | Validierungs-Knoten | Prüft, ob die Mailbox korrekt bereitgestellt wurde | Pflicht | alle | nach Mailboxbereitstellung | Betriebsverantwortung | Ja | blockierend |
| ACC-N-090 | Stammdaten klären | Manueller Knoten | Manuelle Klärung bei unvollständigen Daten | Optional | alle | bei Datenfehler | operative Bearbeitung | Nein | Ausnahmefall |
| ACC-N-100 | Konto vorhanden | Zustands-Knoten | Fachlicher Zustand „Konto vorhanden" | Pflicht | alle | entsteht nach erfolgreicher Kontoanlage | System | Ja | Zustandsreferenz |
| ACC-N-110 | Grundausprägung vollständig | Zustands-Knoten | Pflichtattribute sind gesetzt | Pflicht | alle | entsteht nach erfolgreicher Attributsetzung | System | Ja | Zustandsreferenz |
| ACC-N-120 | Mailbox vorhanden | Zustands-Knoten | Mailbox ist vorhanden | Pflicht | alle | entsteht nach erfolgreicher Mailboxbereitstellung | System | Ja | Zustandsreferenz |
| ACC-N-130 | Mailbox kompensieren | Kompensations-Knoten | Fachliche Rückbehandlung der Mailbox | Optional | alle | bei definiertem Fehlerfall | Betriebsverantwortung | Ja | nur falls nötig |
| ACC-N-140 | Konto kompensieren | Kompensations-Knoten | Fachliche Rückbehandlung des Kontos | Optional | alle | bei definiertem Fehlerfall | Betriebsverantwortung | Ja | nur in definierten Fällen |


## 4. Beziehungsliste

| Beziehungs-ID | Quell-Knoten-ID | Ziel-Knoten-ID | Beziehungstyp | Verbindlichkeit | Bedingung | Kurzbeschreibung |
|---|---|---|---|---|---|---|
| ACC-R-001 | ACC-N-001 | ACC-N-010 | besteht aus | hart | | Service enthält Kontotypbestimmung |
| ACC-R-002 | ACC-N-001 | ACC-N-020 | besteht aus | hart | | Service enthält Vorbedingung Person bekannt |
| ACC-R-003 | ACC-N-001 | ACC-N-030 | besteht aus | hart | | Service enthält Vorbedingung Organisation bekannt |
| ACC-R-004 | ACC-N-001 | ACC-N-040 | besteht aus | hart | | Service enthält Kontoanlage |
| ACC-R-005 | ACC-N-001 | ACC-N-050 | besteht aus | hart | | Service enthält Grundausprägung |
| ACC-R-006 | ACC-N-001 | ACC-N-060 | besteht aus | hart | | Service enthält Mailboxbereitstellung |
| ACC-R-007 | ACC-N-001 | ACC-N-070 | besteht aus | hart | | Service enthält Kontovalidierung |
| ACC-R-008 | ACC-N-001 | ACC-N-080 | besteht aus | hart | | Service enthält Mailboxvalidierung |
| ACC-R-009 | ACC-N-020 | ACC-N-040 | depends on | hart | | Kontoanlage braucht bekannte Person |
| ACC-R-010 | ACC-N-010 | ACC-N-040 | depends on | hart | | Kontoanlage braucht bestimmten Kontotyp |
| ACC-R-011 | ACC-N-030 | ACC-N-050 | depends on | hart | | Attributsetzung braucht organisatorischen Kontext |
| ACC-R-012 | ACC-N-040 | ACC-N-050 | depends on | hart | | Attribute werden auf bestehendem Konto gesetzt |
| ACC-R-013 | ACC-N-040 | ACC-N-070 | validates | hart | | Konto wird validiert |
| ACC-R-014 | ACC-N-050 | ACC-N-070 | depends on | hart | | Kontovalidierung erst nach Attributsetzung |
| ACC-R-015 | ACC-N-040 | ACC-N-100 | produces state | hart | | Kontoanlage erzeugt Zustand Konto vorhanden |
| ACC-R-016 | ACC-N-050 | ACC-N-110 | produces state | hart | | Attributsetzung erzeugt Zustand Grundausprägung vollständig |
| ACC-R-017 | ACC-N-100 | ACC-N-060 | consumes state | hart | | Mailbox braucht vorhandenes Konto |
| ACC-R-018 | ACC-N-110 | ACC-N-060 | consumes state | hart | | Mailbox braucht vollständige Grundausprägung |
| ACC-R-019 | ACC-N-060 | ACC-N-080 | validates | hart | | Mailbox wird validiert |
| ACC-R-020 | ACC-N-060 | ACC-N-120 | produces state | hart | | Mailboxbereitstellung erzeugt Zustand Mailbox vorhanden |
| ACC-R-021 | ACC-N-090 | ACC-N-040 | enables | weich | nach Klärung | Manuelle Datenklärung ermöglicht Kontoanlage |
| ACC-R-022 | ACC-N-130 | ACC-N-060 | compensates | weich | bei definiertem Fehlerfall | Mailbox kann fachlich kompensiert werden |
| ACC-R-023 | ACC-N-140 | ACC-N-040 | compensates | weich | nur falls fachlich verlangt | Konto kann fachlich kompensiert werden |


## 5. Regelliste

| Regel-ID | Regelname | Regeltyp | Beschreibung | Geltungsbereich | Verbindlichkeit | Ausdruck / Fachlogik |
|---|---|---|---|---|---|---|
| ACC-G-001 | Kontotyp muss bestimmt sein | Aktivierungsregel | Konto darf erst angelegt werden, wenn U oder X bestimmt ist | ACC-N-040 | Muss | Kontoanlage nur bei eindeutigem Kontotyp |
| ACC-G-002 | Person muss bekannt sein | Aktivierungsregel | Kontoanlage nur bei vorhandenen Personendaten | ACC-N-040 | Muss | Personenkontext vorhanden |
| ACC-G-003 | Mailbox nur bei vorhandenem Konto | Aktivierungsregel | Mailbox erst nach erfolgreicher Kontoanlage | ACC-N-060 | Muss | Zustand Konto vorhanden |
| ACC-G-004 | Mailbox nur bei vollständiger Grundausprägung | Aktivierungsregel | Mailbox erst, wenn Pflichtattribute gesetzt sind | ACC-N-060 | Muss | Zustand Grundausprägung vollständig |
| ACC-G-005 | U- und X-Account sind Varianten | Variantenregel | Es gilt genau eine der beiden Varianten | Modell | Muss | exklusiv entweder U oder X |
| ACC-G-006 | Kontovalidierung ist blockierend | Validierungsregel | Ohne erfolgreiche Kontovalidierung keine fachliche Erfolgsmeldung | ACC-N-070 | Muss | Validierung = erfolgreich |
| ACC-G-007 | Mailboxvalidierung ist blockierend | Validierungsregel | Ohne erfolgreiche Mailboxvalidierung keine fachliche Erfolgsmeldung | ACC-N-080 | Muss | Validierung = erfolgreich |
| ACC-G-008 | Harte Abhängigkeiten müssen azyklisch sein | Konsistenzregel | Das Modell darf keine Zyklen in harten depends-on-Kanten haben | Modell | Muss | azyklischer Abhängigkeitsgraph |
| ACC-G-009 | Konto wird nicht automatisch kompensiert | Kompensationsregel | Bei Mailboxfehler bleibt Konto grundsätzlich bestehen | ACC-N-140 | Soll | nur bei expliziter fachlicher Regel kompensieren |


## 6. Variantensicht

| Varianten-ID | Variantenname | Beschreibung | Aktivierte Knoten | Deaktivierte Knoten | Spezielle Regeln | Bemerkung |
|---|---|---|---|---|---|---|
| ACC-V-001 | U-Account | Konto für interne Mitarbeitende | alle Basisknoten | keine in dieser ersten Fassung | ACC-G-005 | Unterschiede später verfeinerbar |
| ACC-V-002 | X-Account | Konto für externe Mitarbeitende | alle Basisknoten | keine in dieser ersten Fassung | ACC-G-005 | Unterschiede später verfeinerbar |


## 7. Hierarchiesicht

| Parent-Knoten-ID | Parent-Name | Child-Knoten-ID | Child-Name | Beziehung |
|---|---|---|---|---|
| ACC-N-001 | Benutzerkonto mit Exchange-Mailbox bereitstellen | ACC-N-010 | Kontotyp bestimmen | besteht aus |
| ACC-N-001 | Benutzerkonto mit Exchange-Mailbox bereitstellen | ACC-N-020 | Person bekannt | besteht aus |
| ACC-N-001 | Benutzerkonto mit Exchange-Mailbox bereitstellen | ACC-N-030 | Organisation bekannt | besteht aus |
| ACC-N-001 | Benutzerkonto mit Exchange-Mailbox bereitstellen | ACC-N-040 | Benutzerkonto anlegen | besteht aus |
| ACC-N-001 | Benutzerkonto mit Exchange-Mailbox bereitstellen | ACC-N-050 | Basisattribute setzen | besteht aus |
| ACC-N-001 | Benutzerkonto mit Exchange-Mailbox bereitstellen | ACC-N-060 | Mailbox bereitstellen | besteht aus |
| ACC-N-001 | Benutzerkonto mit Exchange-Mailbox bereitstellen | ACC-N-070 | Benutzerkonto validieren | besteht aus |
| ACC-N-001 | Benutzerkonto mit Exchange-Mailbox bereitstellen | ACC-N-080 | Mailbox validieren | besteht aus |


## 8. Abhängigkeitssicht

| Ziel-Knoten-ID | Ziel-Knoten | Hängt ab von Knoten-ID | Hängt ab von | Verbindlichkeit | Bedingung | Bemerkung |
|---|---|---|---|---|---|---|
| ACC-N-040 | Benutzerkonto anlegen | ACC-N-020 | Person bekannt | hart | | fachliche Mindestvoraussetzung |
| ACC-N-040 | Benutzerkonto anlegen | ACC-N-010 | Kontotyp bestimmen | hart | | Variante muss bestimmt sein |
| ACC-N-050 | Basisattribute setzen | ACC-N-040 | Benutzerkonto anlegen | hart | | Attribute brauchen vorhandenes Konto |
| ACC-N-050 | Basisattribute setzen | ACC-N-030 | Organisation bekannt | hart | | organisatorischer Kontext erforderlich |
| ACC-N-060 | Mailbox bereitstellen | ACC-N-100 | Konto vorhanden | hart | | ohne Konto keine Mailbox |
| ACC-N-060 | Mailbox bereitstellen | ACC-N-110 | Grundausprägung vollständig | hart | | Pflichtattribute müssen gesetzt sein |
| ACC-N-070 | Benutzerkonto validieren | ACC-N-040 | Benutzerkonto anlegen | hart | | Konto muss existieren |
| ACC-N-070 | Benutzerkonto validieren | ACC-N-050 | Basisattribute setzen | hart | | Validierung nach Grundausprägung |
| ACC-N-080 | Mailbox validieren | ACC-N-060 | Mailbox bereitstellen | hart | | Prüfung nach Bereitstellung |


## 9. Ausführungssicht

| Schritt | Knoten-ID | Knotenname | Ausführungsart | Vorbedingungen erfüllt durch | Validierung | Fehlerbehandlung |
|---|---|---|---|---|---|---|
| 1 | ACC-N-020 | Person bekannt | automatisch / prüfend | Eingangsdaten | Plausibilisierung | bei Fehler manueller Knoten |
| 2 | ACC-N-010 | Kontotyp bestimmen | automatisch / regelbasiert | Eingangsdaten | Regelprüfung | bei Unklarheit manuell |
| 3 | ACC-N-040 | Benutzerkonto anlegen | automatisch | Person bekannt, Kontotyp bestimmt | ACC-N-070 | stoppt bei Fehler |
| 4 | ACC-N-050 | Basisattribute setzen | automatisch | Konto vorhanden, Organisation bekannt | ACC-N-070 | stoppt oder manuell |
| 5 | ACC-N-060 | Mailbox bereitstellen | automatisch | Konto vorhanden, Grundausprägung vollständig | ACC-N-080 | Teilfehler möglich |
| 6 | ACC-N-070 | Benutzerkonto validieren | automatisch | Kontoanlage und Attributsetzung | erfolgreich/fehlgeschlagen | blockierend |
| 7 | ACC-N-080 | Mailbox validieren | automatisch | Mailboxbereitstellung | erfolgreich/fehlgeschlagen | blockierend |


## 10. Ausnahme- und Kompensationssicht

| Fehler-/Ausnahmesituation | Betroffener Knoten | Reaktion | Manueller Schritt | Kompensationsknoten | Bemerkung |
|---|---|---|---|---|---|
| Personendaten unvollständig | ACC-N-020 | Auftrag blockieren | ACC-N-090 | keiner | Vor Ausführung klären |
| Kontotyp nicht eindeutig | ACC-N-010 | Auftrag blockieren | ACC-N-090 | keiner | fachliche Klärung nötig |
| Kontoanlage fehlgeschlagen | ACC-N-040 | Prozess stoppen | optional manuell | keiner | keine Mailboxbereitstellung |
| Attributsetzung fehlgeschlagen | ACC-N-050 | Prozess stoppen oder manuell | ACC-N-090 | keiner | Mailbox nicht starten |
| Mailboxbereitstellung fehlgeschlagen | ACC-N-060 | Teilfehler speichern | optional manuell | ACC-N-130 | Konto bleibt bestehen |
| Gesamtkompensation gefordert | ACC-N-040/060 | definierte Rückbehandlung | manuell möglich | ACC-N-130, ACC-N-140 | nur nach klarer Regel |


## 11. Fachliche Interpretation

Dieses Beispiel zeigt, warum die hierarchische Sicht allein nicht reicht:

- Hierarchisch gehören Kontoanlage, Attributsetzung, Mailboxbereitstellung und Validierung alle zum gleichen Service.
- Graphbasiert wird sichtbar:
  - Die Mailbox hängt nicht direkt nur vom Service, sondern konkret von den Zuständen „Konto vorhanden" und „Grundausprägung vollständig" ab.
  - Validierung ist ein eigener modellierter Schritt.
  - Ausnahmebehandlung ist modellierbar (nicht nur Fliesstext).
  - Kompensation ist als eigener Knoten und eigene Beziehung beschreibbar.

Genau dadurch wird das Modell später fachlich auswertbar.


## 12. Graph-Diagramm (textuell)

```
    [ACC-N-020 Person bekannt] ──depends on──┐
                                              ▼
    [ACC-N-010 Kontotyp bestimmen] ──depends on──▶ [ACC-N-040 Konto anlegen]
                                                           │
                                              ┌────────────┼────────────┐
                                              │ produces    │ depends on │
                                              ▼ state      ▼            │
                                    [ACC-N-100 Konto    [ACC-N-050      │
                                     vorhanden]         Basisattribute] │
                                              │            │            │
                                              │ consumes   │ produces   │
                                              │ state      │ state      │
                                              ▼            ▼            │
    [ACC-N-030 Organisation] ──depends on──▶ ...   [ACC-N-110 Grund-   │
                                                    ausprägung]         │
                                              │            │            │
                                              │ consumes   │            │
                                              │ state      │            │
                                              ▼            ▼            │
                                    [ACC-N-060 Mailbox bereitstellen]   │
                                              │                         │
                                              │ validates               │ validates
                                              ▼                         ▼
                                    [ACC-N-080 Mailbox      [ACC-N-070 Konto
                                     validieren]             validieren]
```
