# 001 - MVP 0.1 Scope

## Zweck
Dieses Dokument definiert den Umfang von MVP 0.1 inklusive Ziele, Grenzen, Risiken und Erfolgskriterien.

## Ziele
- Strukturierte Verwaltung von Produktartefakten für das Produktmanagement.
- Trennung von Anforderung und Business Rule.
- Deterministische Validierung gegen definierte Regeln.
- Nachvollziehbarkeit über Versionen, Findings und Git Historie.
- Git-basierter Änderungs- und Reviewprozess.

## Enthaltene Fähigkeiten (In Scope)
- Verwaltung von Produktartefakten.
- Verwaltung von Anforderungsartefakten.
- Verwaltung von Business Rule Artefakten.
- Verwaltung von Validierungsszenarien.
- Validierung von Beispieldaten gegen Business Rules.
- Erzeugung von Findings.
- Traceability zu Produktversion und Regelversion.
- Konzeptionelle REST API Endpunkte.
- UI Konzept für Produktmanagement.
- Git-basierte Änderungsvorschläge inkl. Review/Freigabe.
- Optionale LLM Assistenz für Drafts von Anforderungen und Regeln.

## Ausgeschlossene Fähigkeiten (Out of Scope)
- BPMN Runtime.
- DMN Editor.
- DMN Runtime.
- Generische Skill Runtime.
- Autonome LLM Entscheidungen.
- Zielsystem-Provisionierung.
- Event-driven Integrationsarchitektur.
- Vollstaendiges Multi-Tenant Betriebsmodell.
- Produktionsreife SSO- und RBAC-Ausgestaltung.
- Grossskalige Sucharchitektur.

## Annahmen
- Fachliche Artefakte liegen versionierbar in Git vor.
- YAML ist für Stakeholder lesbar und pflegbar.
- Review-Regeln werden organisatorisch getragen.
- Deterministische Validierung ist für MVP ausreichend.

## Randbedingungen
- Keine Einfuehrung einer primären Artefaktdatenbank.
- Kein Umsetzen von Runtime Services im MVP-Dokumentationsschritt.
- Keine Umgehung von Governance über direkte Hauptbranch-Änderungen.

## Risiken
- Zu breite Erstumsetzung gefährdet Lieferfähigkeit.
- Unklare Rollen können Reviews verzoegern.
- Fehlende Konventionen für IDs und Status erschweren Skalierung.

## Erwartetes MVP-Ergebnis
Ein steuerbarer, nachvollziehbarer und reviewbarer Artefaktprozess für Produktkatalog, Anforderungen, Regeln, Validierung und Findings auf Git-Basis.

## Erfolgskriterien
- Fachliche Artefakte sind strukturiert und versioniert.
- Änderungsvorschläge sind über Branch und Pull Request nachvollziehbar.
- Validierungsergebnisse referenzieren Produkt- und Regelversionen.
- Governance-Regeln sind dokumentiert und im Prozess verankert.

## Warum dieser Scope bewusst reduziert ist
Die Reduktion minimiert technische und organisatorische Komplexitaet in der Einfuehrung. Nur wenn Artefaktmodell, Git-first Governance und deterministische Validierung stabil funktionieren, kann das System kontrolliert in Richtung Orchestrierung, erweiterte Integrationen und KI-Assistenz wachsen.
