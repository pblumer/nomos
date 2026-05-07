<!--
Source: Fachliches Metamodell_Servicegraph.pdf
Converted from PDF for use in the nomos repository.
-->

# Fachliches Metamodell: Servicegraph 0.2

## 1. Ziel des Metamodells

Dieses Metamodell beschreibt die fachlichen Bausteine, aus denen ein graphbasiertes Servicemodell aufgebaut werden soll. Es definiert:

- welche Objekttypen modelliert werden dürfen
- welche Beziehungstypen zwischen diesen Objekten zulässig sind
- welche Eigenschaften diese Objekte und Beziehungen besitzen
- welche fachlichen Regeln für die Modellierung gelten

Das Metamodell schafft eine einheitliche Grundlage für die Beschreibung komplexer Services, ihrer Teilservices, Abhängigkeiten, Varianten, Vorbedingungen und Provisionierungslogiken.

Es beschreibt nicht die konkrete technische Speicherung, sondern die fachliche Modellierungssprache.


## 2. Zweck

| Ziel | Beschreibung |
|---|---|
| Einheitlichkeit | Alle Services nach denselben Modellierungsprinzipien beschreiben |
| Auswertbarkeit | Modell nicht nur lesbar, sondern fachlich auswertbar (Abhängigkeiten, Reihenfolgen, Blockaden) |
| Wiederverwendbarkeit | Servicebausteine in mehreren Kompositionen verwendbar |
| Trennung der Sichten | Struktur, Abhängigkeiten, Regeln und Ausführung sauber getrennt, aber verknüpfbar |


## 3. Grundprinzipien

1. **Hierarchie und Graph sind zwei verschiedene Dinge.** Zugehörigkeit ≠ Abhängigkeit.
2. **Knoten und Beziehungen sind gleichwertige Modellbestandteile.**
3. **Fachliche Modellierung vor technischer Modellierung.**
4. **Versionierbarkeit.** Jedes Modell und jeder Baustein muss versionierbar sein.
5. **Erweiterbarkeit.** Weitere Knoten- oder Beziehungstypen können später ergänzt werden.


## 4. Überblick über die Metamodell-Elemente

Das Metamodell besteht aus vier Ebenen:

1. Modellkontext
2. Knotentypen
3. Beziehungstypen
4. Regeln und Attribute


## 5. Modellkontext

Das Modell ist die Gesamtheit aller Knoten, Beziehungen und Regeln, die eine fachliche Service-Komposition beschreiben.

### Attribute des Modells

| Attribut | Beschreibung |
|---|---|
| Modell-ID | Eindeutige Identifikation |
| Name | Name des Modells |
| Beschreibung | Fachliche Beschreibung |
| Version | Versionsstand |
| Status | Entwurf, freigegeben, ausser Betrieb |
| Gültig ab | Beginn der Gültigkeit |
| Gültig bis | Ende der Gültigkeit |
| Verantwortliche Rolle | Fachliche Verantwortung |
| Bezugsservice | Welcher Service durch dieses Modell beschrieben wird |


## 6. Knotentypen

### 6.1 Gemeinsame Basisattribute aller Knoten

| Attribut | Beschreibung |
|---|---|
| Knoten-ID | Eindeutige Identifikation |
| Name | Fachlicher Name |
| Beschreibung | Fachliche Erläuterung |
| Knotentyp | Typ gemäss Metamodell |
| Status | Modellierungsstatus |
| Pflicht/Optional | Immer erforderlich oder nur in bestimmten Fällen |
| Aktivierungsregel | Unter welchen Bedingungen aktiv |
| Variante | Für welche Varianten gilt der Knoten |
| Verantwortliche Rolle | Fachlich zuständige Rolle |
| Wiederverwendbar | Ja/Nein |
| Version | Versionsstand |
| Tags/Kategorien | Zur fachlichen Gruppierung |

### 6.2 Service-Knoten

Repräsentiert einen fachlich sichtbaren oder bestellbaren Service.

Zusätzliche Attribute: Service-Code, Serviceziel, Zielgruppe.

### 6.3 Sub-Service-Knoten

Repräsentiert einen fachlichen Bestandteil eines Service. Wiederverwendbar.

Zusätzliche Attribute: Parent-Service-ID, Fachlicher Zweck.

### 6.4 Aktivitäts-Knoten

Repräsentiert einen ausführbaren oder fachlich auszulösenden Schritt.

Zusätzliche Attribute: Ausführungsart (automatisch/manuell/hybrid), Ergebnisart, Idempotenz erwartet (Ja/Nein), Retry-fähig (Ja/Nein).

### 6.5 Vorbedingungs-Knoten

Repräsentiert eine Bedingung, die erfüllt sein muss, bevor ein anderer Knoten aktiviert werden darf.

Zusätzliche Attribute: Bedingungstyp (fachlich/technisch/organisatorisch/zeitlich), Prüfart (manuell/automatisch), Muss-Kriterium (Ja/Nein).

### 6.6 Entscheidungs-Knoten

Repräsentiert eine Regel oder Auswahl, die über den weiteren Verlauf entscheidet.

Zusätzliche Attribute: Entscheidungslogik, Entscheidungsbasis, Ergebniswerte.

### 6.7 Validierungs-Knoten

Prüft, ob ein gewünschter fachlicher oder technischer Zustand erreicht wurde.

Zusätzliche Attribute: Validierungsziel, Validierungsart (automatisch/manuell), Blockierend (Ja/Nein).

### 6.8 Manueller Knoten

Repräsentiert eine Tätigkeit durch eine Person oder Rolle (Freigaben, Klärungen, Nachbearbeitung).

Zusätzliche Attribute: Bearbeitungsrolle, Frist relevant (Ja/Nein), Eskalationsregel.

### 6.9 Kompensations-Knoten

Beschreibt eine fachliche Gegenmassnahme bei Fehlschlag oder Rückbau.

Zusätzliche Attribute: Kompensationsziel, Kompensationsart (rückgängig machen/deaktivieren/kennzeichnen/manuell bereinigen), Pflicht bei Fehler (Ja/Nein).

### 6.10 Zustands-Knoten

Repräsentiert einen fachlich relevanten Zustand (z.B. „Konto vorhanden", „Mailbox aktiv").

Zusätzliche Attribute: Zustandstyp (fachlich/technisch/organisatorisch), Zielzustand.

### 6.11 Referenz-Knoten

Repräsentiert einen wiederverwendbaren oder extern referenzierten Modellbaustein. Reduziert Duplikate.

Zusätzliche Attribute: Referenzziel (ID), Referenzart (intern/extern), Bindungsart (statisch/versioniert).


## 7. Beziehungstypen

### 7.1 Gemeinsame Basisattribute aller Beziehungen

| Attribut | Beschreibung |
|---|---|
| Beziehungs-ID | Eindeutige Identifikation |
| Quell-Knoten-ID | Ursprung |
| Ziel-Knoten-ID | Ziel |
| Beziehungstyp | Typ gemäss Metamodell |
| Verbindlichkeit | hart / weich |
| Bedingung | Optionaler Aktivierungskontext |
| Beschreibung | Fachliche Erläuterung |
| Version | Versionsstand |

### 7.2 Beziehungstyp-Katalog

| Typ | Bedeutung | Beispiel |
|---|---|---|
| besteht aus | Hierarchische Zerlegung | Service besteht aus Sub-Service |
| depends on | Voraussetzung für Aktivierung | Mailbox depends on Konto |
| enables | Freischaltung | Genehmigung enables Bereitstellung |
| blocks | Verhinderung der Ausführung | Fehlende Stammdaten blocks Konto anlegen |
| validates | Prüfung des Ergebnisses | Mailbox validieren validates Mailbox bereitstellen |
| compensates | Fachliche Kompensation | Mailbox deaktivieren compensates Mailbox bereitstellen |
| alternative to | Alternativbeziehung | U-Account alternative to X-Account |
| requires condition | Bedingte Aktivierung | Zusatzschritt requires condition „externer Mitarbeiter" |
| produces state | Erzeugt einen Zustand | Konto anlegen produces state „Konto vorhanden" |
| consumes state | Benötigt einen vorhandenen Zustand | Mailbox anlegen consumes state „Konto vorhanden" |


## 8. Regeln als eigenes Metamodell-Element

### 8.1 Basisattribute einer Regel

| Attribut | Beschreibung |
|---|---|
| Regel-ID | Eindeutige ID |
| Name | Fachlicher Name |
| Beschreibung | Erläuterung |
| Regeltyp | Aktivierung / Validierung / Variantenregel / Konsistenzregel / Kompensationsregel |
| Ausdruck | Fachlicher oder strukturierter Regelausdruck |
| Geltungsbereich | Für welche Knoten/Beziehungen die Regel gilt |
| Verbindlichkeit | Muss / Soll / Kann |

### 8.2 Regeltypen

| Typ | Zweck |
|---|---|
| Aktivierungsregel | Steuert, wann ein Knoten oder eine Beziehung aktiv ist |
| Variantenregel | Legt fest, welche Knoten für welche Variante gelten |
| Validierungsregel | Definiert, wann ein Ergebnis als erfolgreich gilt |
| Konsistenzregel | Prüft, ob das Modell fachlich gültig ist |
| Kompensationsregel | Definiert, wann und wie Kompensationen anzuwenden sind |


## 9. Varianten im Metamodell

Varianten sollen nicht durch Duplikation ganzer Modelle, sondern durch gezielte Unterschiede beschrieben werden.

### Attribute einer Variante

| Attribut | Beschreibung |
|---|---|
| Varianten-ID | Eindeutige ID |
| Name | Variantenname |
| Beschreibung | Fachliche Erläuterung |
| Kontext | Wann die Variante gilt |
| Regelbezug | Welche Regeln die Variante steuern |

Varianten können auf drei Ebenen wirken: auf Knoten, auf Beziehungen, auf Regeln.


## 10. Fachliche Konsistenzregeln

| Regel | Beschreibung |
|---|---|
| Eindeutigkeit | Jeder Knoten, jede Beziehung und jede Regel muss eindeutig identifizierbar sein |
| Typisierung | Jeder Knoten und jede Beziehung muss genau einen Typ besitzen |
| Keine impliziten Abhängigkeiten | Abhängigkeiten müssen explizit modelliert werden |
| Zyklusfreiheit | depends-on mit Verbindlichkeit „hart" darf keinen Zyklus bilden |
| Referenzierbarkeit von Regeln | Regelgesteuerte Aktivierung muss explizit referenziert sein |
| Trennung Struktur/Ausführung | besteht-aus allein ≠ Ausführungsreihenfolge |
| Verantwortlichkeit | Modellrelevante Knoten brauchen eine verantwortliche Rolle |


## 11. Empfohlenes Minimalset für die erste Einführung

### Minimale Knotentypen

- Service-Knoten
- Sub-Service-Knoten
- Aktivitäts-Knoten
- Vorbedingungs-Knoten
- Validierungs-Knoten
- Manueller Knoten

### Minimale Beziehungstypen

- besteht aus
- depends on
- validates
- requires condition

### Minimale Regeltypen

- Aktivierungsregel
- Variantenregel
- Konsistenzregel


## 12. Metamodell-Struktur (Übersicht)

    Modell
    ├── enthält Knoten
    │     ├── Service-Knoten
    │     ├── Sub-Service-Knoten
    │     ├── Aktivitäts-Knoten
    │     ├── Vorbedingungs-Knoten
    │     ├── Entscheidungs-Knoten
    │     ├── Validierungs-Knoten
    │     ├── Manueller Knoten
    │     ├── Kompensations-Knoten
    │     ├── Zustands-Knoten
    │     └── Referenz-Knoten
    ├── enthält Beziehungen
    │     ├── besteht aus
    │     ├── depends on
    │     ├── enables
    │     ├── blocks
    │     ├── validates
    │     ├── compensates
    │     ├── alternative to
    │     ├── requires condition
    │     ├── produces state
    │     └── consumes state
    └── enthält Regeln
          ├── Aktivierungsregel
          ├── Variantenregel
          ├── Validierungsregel
          ├── Konsistenzregel
          └── Kompensationsregel


## 13. Zusammenfassung

Das fachliche Metamodell definiert die Modellierungssprache für graphbasierte Service-Kompositionen. Im Zentrum stehen:

- **Knoten** als fachliche Bausteine
- **Beziehungen** als explizite Verknüpfungen
- **Regeln** zur Steuerung von Aktivierung, Varianten, Validierung und Konsistenz

Dadurch wird eine Modellierung möglich, die sowohl verständlich als auch auswertbar ist und deutlich über eine rein hierarchische Zerlegung hinausgeht.
