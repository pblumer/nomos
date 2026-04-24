<!--
Source: Einführungsmodell und MVP-Scope 0.1.pdf
Converted from PDF text extraction for use in the nomos repository.
Note: Diagrams embedded as images in the original PDF are represented by textual placeholders/notes where text extraction cannot preserve the visual layout.
-->

# Einführungsmodell und MVP-Scope 0.1
Schrittweise Umsetzung einer Plattform für Produktanforderungen, Business Rules und
Provisionierung

## 1. Ziel und Zweck
Das fachliche Zielbild der geplanten Plattform ist tragfähig, jedoch in einem ersten Umsetzungsschritt zu umfangreich. Das Review hat gezeigt, dass
insbesondere die gleichzeitige Einführung von Produktmodellierung, Regelmodellierung, Entscheidungen, Prozessen, Skills, KI-Nutzung und
Integrationsarchitektur ein hohes fachliches und technisches Risiko mit sich bringt.

Dieses Einführungsmodell beschreibt daher eine schrittweise und steuerbare Einführung der Plattform. Es reduziert die anfängliche Komplexität,
schärft zentrale Begriffe wie den Skill-Begriff, grenzt die Rolle der KI klar ein und definiert, welche Bausteine im MVP tatsächlich enthalten sind.

Ziel des Dokuments ist es, aus dem fachlichen Zielbild ein realistisch umsetzbares Vorgehensmodell abzuleiten.


## 2. Ausgangslage
Die Plattform soll langfristig eine zentrale Grundlage für folgende Aufgaben schaffen:

         strukturierte Beschreibung von Produkten und Produktanforderungen
         fachliche Modellierung und Auswertung von Business Rules
         Abbildung von Entscheidungen
         Unterstützung von Provisionierungsprozessen
         skillbasierte Ausführung einzelner Tasks
         Nutzung durch Orchestrierungsplattformen wie Pega oder Power Platform
         spätere Nutzung durch KI-Agenten
         Validierung, Nachvollziehbarkeit und Auditierbarkeit

Das Review hat jedoch deutlich gemacht, dass diese Zielsetzung in einer ersten Umsetzungsphase zu breit ist. Es besteht das Risiko, dass die
Organisation sowohl fachlich als auch technisch überfordert wird, wenn alle Bausteine gleichzeitig eingeführt werden.


## 3. Zielbild und Einführungsansatz
3.1 Langfristiges Zielbild
Langfristig soll die Plattform fünf fachliche Ebenen verbinden:

         Produkte
         Regeln
         Entscheidungen
         Prozesse
         Skills

Ergänzt wird dieses Zielbild durch Integrations-, Ausführungs- und Qualitätssicherungsfunktionen.


3.2 Einführungsansatz
Für die Umsetzung wird jedoch nicht das vollständige Zielbild auf einmal realisiert. Stattdessen wird ein stufenweises Einführungsmodell festgelegt.

Das Grundprinzip lautet:

         zuerst die fachliche Basis stabilisieren
         danach Entscheidungen und Referenzprozesse ergänzen
         erst später Skills, Orchestrierungsintegration und KI-Unterstützung ausbauen

Dadurch wird vermieden, dass die Plattform bereits zu Beginn zu viele Modellierungsarten, Abhängigkeiten und organisatorische Verantwortlichkeiten
vereint.


## 4. Zentrale Leitplanken für die Einführung
4.1 Reduktion der Komplexität
Nicht alle fachlichen Artefakttypen werden von Beginn an verpflichtend eingeführt. In der ersten Phase werden nur diejenigen Bausteine verbindlich
modelliert, die für einen ersten fachlich belastbaren Nutzen notwendig sind.


4.2 Fokus auf strukturiertes Fachwissen
Die erste Ausbaustufe konzentriert sich auf Produktbeschreibung, Anforderungen, Regeln und Validierung. Entscheidungen, Prozesse und Skills werden
vorbereitet, aber nicht flächendeckend umgesetzt.


4.3 Trennung von Zielbild und MVP
Das Zielbild bleibt erhalten, wird aber ausdrücklich vom MVP abgegrenzt. Dadurch kann die Plattform langfristig wachsen, ohne den Projektstart zu
überladen.


4.4 Governance von Anfang an
Auch im MVP müssen Rollen, Freigaben, Versionierung und Nachvollziehbarkeit klar geregelt sein. Governance wird nicht erst in einer späteren Phase
ergänzt, sondern von Beginn an berücksichtigt.


## 5. Ausbaustufen des Einführungsmodells
5.1 Phase 1 – MVP: Produktkatalog, Anforderungen, Regeln und Validierung

Zielsetzung
In der ersten Phase wird die fachliche Basis geschaffen. Die Organisation soll lernen, Produkte, Anforderungen und Regeln strukturiert zu modellieren und
Daten gegen diese Regeln zu prüfen.


Enthaltene Bausteine
        Produktartefakte
        Anforderungsartefakte
        Regelartefakte
        Validierungsservice
        erste REST-Schnittstellen für Lesen und Prüfen
        Versionierung
        Audit-Grundlagen
        Findings und Validierungsergebnisse


Noch nicht enthalten
        vollwertige BPMN-Ausführung
        DMN als eigener Modellierungsstandard
        generische Skill-Runtime
        KI als Laufzeitentscheider
        Event-basierte Vollintegration
        komplexe Orchestrierungslogik


Nutzen dieser Phase
Diese Phase erzeugt bereits einen konkreten Mehrwert:

        Produkte und Regeln werden strukturiert und versioniert abgelegt
        Daten können konsistent gegen Regeln geprüft werden
        erste Governance-Mechanismen werden etabliert
        die Basis für spätere Entscheidungen, Prozesse und Skills wird geschaffen


5.2 Phase 2 – Entscheidungen und Referenzprozess

Zielsetzung
In der zweiten Phase werden strukturierte Entscheidungen und ein erster fachlicher Referenzprozess eingeführt.


Enthaltene Bausteine
         Entscheidungsartefakte
         Priorisierung und Konfliktbehandlung von Regeln
         erster Referenzprozess für ein Produkt
         Zuordnung von Prozessschritten zu fachlichen Bearbeitungsbausteinen
         erste Nachweise und erweiterte Findings


Fokus
Diese Phase dient dazu, das Zusammenspiel zwischen Regeln, Entscheidungen und Prozessschritten zu konkretisieren, ohne bereits eine vollständige
Workflow- oder Skill-Plattform aufzubauen.


5.3 Phase 3 – Skills und Orchestrierungsintegration

Zielsetzung
In der dritten Phase werden standardisierte Skills und erste Integrationen mit Orchestrierungsplattformen eingeführt.


Enthaltene Bausteine
         standardisierte Skill-Definition
         Skill-Registry
         Skill-Runtime oder standardisierte Ausführungsumgebung
         Anbindung erster Orchestratoren wie Pega oder Power Platform
         Nutzung von REST für synchrone Integrationen
         erste asynchrone Integrationsmuster bei Bedarf


Fokus
Diese Phase operationalisiert die Plattform. Der Schwerpunkt liegt auf standardisierten Ausführungsbausteinen und kontrollierter Integrationsfähigkeit.


5.4 Phase 4 – Erweiterte KI-Unterstützung und Skalierung

Zielsetzung
In einer späteren Phase wird die Plattform um erweiterte KI-Funktionen und skalierende Integrationsmuster ergänzt.


Mögliche Bausteine
         KI für Regelvorschläge und Modellanalyse
         KI für Inkonsistenz- und Lückenprüfung
         erklärbare KI-Ausgaben
         Event-basierte Kommunikation
         Queue-basierte Verarbeitung
         Caching häufig genutzter Regeln und Entscheidungen
         hochverfügbare Ausführung einzelner Bausteine


Fokus
Diese Phase dient der Optimierung und Skalierung. Sie ist nicht Voraussetzung für den fachlichen Start.


## 6. MVP-Scope 0.1
6.1 Inhalt des MVP
Der MVP umfasst die minimal notwendige fachliche und technische Struktur, um die Plattform mit einem ersten Referenzprodukt nutzbar zu machen.


Im MVP enthalten
          strukturierte Produktbeschreibung
          strukturierte Anforderungen
          strukturierte Regeln
          Versionierung der Artefakte
          Validierungsservice
          REST-Schnittstellen für Lesen und Prüfen
          Findings und Validierungsergebnisse
          Nachvollziehbare Protokollierung
          Governance für Änderungen an Produkten und Regeln


Im MVP nicht enthalten
          vollständige BPMN-Modellierung und Ausführung
          DMN-Modellierungswerkzeuge
          generische Skill-Laufzeit
          autonome KI-Entscheidungen
          umfassende Event-Architektur
          automatische Konfliktauflösung über alle Artefakttypen
          vollständige Hochverfügbarkeitsarchitektur


6.2 Referenz-Use-Case für den MVP
Als Referenz-Use-Case für den MVP wird das Produkt Benutzerkonto mit Mailbox verwendet.

Dieses Beispiel eignet sich besonders, weil sich daran folgende Aspekte gut demonstrieren lassen:

          Produktbeschreibung
          Variantenbildung
          Vorbedingungen
          Pflichtattribute
          Regelprüfung
          Konfliktpotenziale
          Validierung von Eingabedaten
          spätere Erweiterbarkeit in Richtung Entscheidung, Prozess und Skill

Der MVP soll an diesem Referenzprodukt zeigen, dass die Plattform fachlich nutzbar und technisch anschlussfähig ist.


## 7. Verbindliche Definition des Skill-Begriffs
Das Review hat gezeigt, dass der bisherige Skill-Begriff zu offen ist. Für die weitere Arbeit wird daher eine verbindlichere Definition festgelegt.


7.1 Definition
Ein Skill ist ein standardisierter, versionierter und eindeutig adressierbarer Ausführungsbaustein zur Bearbeitung eines einzelnen fachlichen Tasks
mit definiertem Input, definiertem Output, klaren Vorbedingungen und standardisiertem Nachweisverhalten.


7.2 Abgrenzung
Ein Skill ist:

          kein beliebiges loses Skript
          keine eigenständige Rule Engine
          keine autonome Entscheidungsinstanz
          kein unstrukturierter Wissenscontainer

Ein Skill kann technisch unterschiedlich umgesetzt sein, zum Beispiel als:


         Service
         Container
         Worker
         Adapter
         Bot

Unabhängig von der technischen Umsetzung muss ein Skill jedoch nach außen einem einheitlichen fachlichen Vertrag folgen.


7.3 Fachliche Leitregel
Skills dürfen ausführen, aber nicht eigenständig fachlich entscheiden.

Fachliche Entscheidungen müssen in:

         Regeln
         Entscheidungsartefakten
         oder Plattformservices

verankert sein.

Skills dürfen daher beispielsweise:

         Daten prüfen
         Zielsysteme aufrufen
         Identifikatoren erzeugen
         Nachweise erstellen

Skills dürfen jedoch nicht selbst definieren:

         ob eine Provisionierung erlaubt ist
         welche Produktvariante gilt
         welche Regel überstimmt wird
         welche fachliche Priorisierung anzuwenden ist


7.4 Konsequenz für die Einführung
Da der Skill-Begriff noch nicht organisatorisch und technisch etabliert ist, werden Skills im MVP noch nicht als verpflichtender Plattformbestandteil
umgesetzt. Sie werden konzeptionell vorbereitet, aber erst in einer späteren Phase verbindlich eingeführt.


## 8. Rolle der KI
8.1 Grundsatz
Die Plattform soll grundsätzlich so gestaltet werden, dass ihre Artefakte später von KI-Systemen genutzt werden können. In den ersten Ausbaustufen hat
KI jedoch eine unterstützende, nicht aber eine autoritative Rolle.


8.2 KI im MVP
Im MVP ist KI nicht für finale fachliche Entscheidungen zuständig.

Mögliche KI-Nutzungen in späteren oder begleitenden Szenarien sind:

         Vorschläge für Regeln
         Analyse von Inkonsistenzen
         Identifikation fehlender Artefakte
         Unterstützung bei Dokumentation und Erklärung
         Unterstützung bei der Suche nach passenden Regeln oder Modellen


8.3 Nicht zulässige KI-Rolle im MVP
Im MVP soll KI insbesondere nicht:

         verbindliche Provisionierungsentscheidungen treffen
         Regeln stillschweigend interpretieren oder verändern
         Governance-Freigaben ersetzen
         als Black Box fachliche Logik bestimmen


8.4 Konsequenz
Die Rolle der KI wird im Einführungsmodell bewusst eingegrenzt. KI-Nutzbarkeit bleibt ein Ziel, aber keine tragende Säule des ersten Umsetzungsschritts.


## 9. Governance-Modell
Das Review hat gezeigt, dass Verantwortlichkeiten operativer beschrieben werden müssen. Es genügt nicht, die Verantwortung grob zwischen Fachseite,
IT und Delivery aufzuteilen. Stattdessen wird pro Artefakttyp eine konkrete Verantwortungslogik benötigt.


9.1 Verantwortlichkeiten nach Artefakttyp

Produkt
         fachliche Verantwortung: Produktverantwortung oder Fachdomäne
         technische Mitwirkung: Architektur
         Freigabe: Fachverantwortung


Anforderung
         fachliche Verantwortung: Fachdomäne
         methodische Unterstützung: Business Analyse oder Architektur
         Freigabe: Fachverantwortung


Regel
         fachliche Verantwortung: Fachdomäne
         technische Prüfung: Plattform- oder Architekturteam
         Freigabe: Fachseite, bei kritischen Regeln zusätzlich Plattformverantwortung


Entscheidung
         fachliche Verantwortung: Fachdomäne
         methodische Prüfung: Architektur oder Governance
         Freigabe: gemeinsame Freigabe von Fachseite und Plattformverantwortung


Prozess
         fachliche Verantwortung: Prozessverantwortung
         technische Operationalisierung: Delivery oder Orchestrierungsteam


Skill
         fachliche Anforderungen: Fachdomäne oder Prozessverantwortung
         technische Umsetzung: Delivery-Team
         Betriebsverantwortung: Plattform- oder Service-Team


9.2 Änderungs- und Freigabeprozess
Für kritische Änderungen ist ein verbindlicher Review- und Freigabeprozess notwendig. Dazu zählen insbesondere Änderungen an:

         Regeln
         Entscheidungslogiken
         Konfliktprioritäten
         Zielsystemreferenzen
         Skills
         Prozesspfaden

Je nach Kritikalität kann dies als leichtgewichtiger Review-Prozess oder als formalisierte Freigabekette umgesetzt werden. Entscheidend ist, dass
Änderungen nachvollziehbar, versioniert und freigegeben erfolgen.


9.3 Versionierung und Historisierung
Historische Entscheidungen und Prüfergebnisse dürfen nicht unbrauchbar werden, wenn sich Regeln oder Modelle ändern. Deshalb ist festzulegen, dass
für relevante Ausführungen und Entscheidungen immer nachvollziehbar bleibt:

         welche Regelversion galt
         welche Produktversion galt
         welche Eingaben verwendet wurden
         welches Ergebnis erzeugt wurde

Empfohlen wird dafür ein Snapshot-Prinzip, bei dem bei jeder relevanten Auswertung die zu diesem Zeitpunkt gültigen Artefaktversionen referenziert oder
gesichert werden.


## 10. Umgang mit Regelkonflikten
Das Review hat zu Recht darauf hingewiesen, dass Konflikte zwischen Regeln im bisherigen Modell noch nicht ausreichend behandelt sind. Für die
Plattform ist deshalb eine explizite Konfliktstrategie erforderlich.


10.1 Zielsetzung
Die Plattform soll nicht nur Regeln auswerten, sondern auch Regelkonflikte erkennen und sichtbar machen.


10.2 Fachliche Leitlogik
Für die erste Ausbaustufe wird folgende Grundlogik empfohlen:

         sperrende Regeln haben Vorrang vor erlaubenden Regeln
         Muss-Regeln haben Vorrang vor Soll-Regeln
         produktspezifische Regeln dürfen generische Regeln nur explizit übersteuern
         Konflikte müssen als Findings sichtbar werden


10.3 Erweiterung des Regelschemas
Für spätere Ausbaustufen sollten Regeln zusätzliche Attribute erhalten, zum Beispiel:

         priority
         policy_effect
         conflict_group
         override_mode

Beispielhafte Werte für policy_effect:

         allow
         deny
         require
         derive

Diese Erweiterung ist für das MVP noch nicht zwingend vollständig umzusetzen, sollte aber bereits konzeptionell vorgesehen werden.


## 11. Nichtfunktionale Leitplanken
11.1 Integrationsstil
REST wird im MVP als primärer Integrationsstil vorgesehen. Dies ist für synchrone Anwendungsfälle wie Lesen, Prüfen und spätere
Entscheidungsabfragen ausreichend und praktikabel.


11.2 Erweiterbarkeit
Für spätere Phasen wird eine Architektur angestrebt, die auch asynchrone oder event-basierte Kommunikation unterstützen kann, wenn Last, Entkopplung
oder Prozesscharakter dies erforderlich machen.


11.3 Performance
Im MVP stehen fachliche Strukturierung und Nachvollziehbarkeit im Vordergrund. Performance-Optimierungen wie Caching, Queue-Verarbeitung oder
Eventing werden erst eingeführt, wenn konkrete Lastanforderungen dies notwendig machen.


11.4 Ausfallsicherheit
Auch wenn im MVP noch keine vollständige Hochverfügbarkeitsarchitektur vorgesehen ist, sollen zentrale Bausteine bereits so geschnitten werden, dass
eine spätere redundante Ausführung möglich bleibt.


## 12. Konsequenzen für die weitere Arbeit
Aus dem Review ergeben sich mehrere unmittelbare Konsequenzen für die weitere Projektarbeit.


12.1 Das Zielbild bleibt erhalten
Die fachliche Vision einer Plattform für Produkte, Regeln, Entscheidungen, Prozesse und Skills bleibt bestehen.


12.2 Der Einstieg wird reduziert
Für den Projektstart wird der Scope bewusst auf Produktmodellierung, Anforderungen, Regeln und Validierung eingegrenzt.


12.3 Skills werden vorbereitet, aber nicht sofort operationalisiert
Der Skill-Begriff wird geschärft, Skills werden jedoch erst in einer späteren Ausbaustufe verbindlich eingeführt.


12.4 KI bleibt zunächst Assistenz
KI wird zunächst nicht als entscheidende Laufzeitkomponente eingeführt, sondern nur als unterstützende Funktion betrachtet.


12.5 Governance wird konkretisiert
Änderungsrechte, Freigaben, Versionierung und Historisierung werden von Anfang an berücksichtigt.


## 13. Zusammenfassung
Das Review hat gezeigt, dass das fachliche Zielbild richtig ist, jedoch für einen ersten Umsetzungsschritt zu breit angelegt war. Die sinnvolle Reaktion
besteht nicht darin, das Zielbild zu verwerfen, sondern es in ein realistisches Einführungsmodell zu überführen.

Das Einführungsmodell 0.1 verfolgt daher folgende Grundlogik:

         langfristig umfassendes Zielbild
         kurzfristig reduzierter MVP
         klare Abgrenzung von Scope und Nicht-Scope
         verbindliche Definition des Skill-Begriffs
         enge und kontrollierte Rolle der KI
         konkrete Governance
         vorbereitete, aber nicht überfrachtete Architektur

Damit entsteht eine belastbare Grundlage, um die Plattform schrittweise aufzubauen und gleichzeitig organisatorisch beherrschbar zu halten.
