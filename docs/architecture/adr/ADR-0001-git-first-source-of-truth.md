# ADR-0001 - Git-first als Quelle der Wahrheit

## Kontext
MVP 0.1 benötigt nachvollziehbare, reviewbare und versionierte Fachartefakte mit geringer Einfuehrungskomplexitaet.

## Entscheidung
Git ist die Quelle der Wahrheit für Business-Artefakte in MVP 0.1.

## Konsequenzen
- Artefakte werden dateibasiert und versioniert verwaltet.
- Review und Freigabe laufen über Pull Requests.
- API und UI muessen Git-Kontext berücksichtigen.

## Positive Effekte
- Native Historie und Auditierbarkeit.
- Klarer Governanceprozess.
- Hohe Transparenz für Fach- und Technikrollen.

## Trade-offs
- Konfliktauflösung bei parallelen Änderungen.
- Suche und Query brauchen spaeter ggf. Read Model.

## Betrachtete Alternativen
- PostgreSQL-first
- Dokumentdatenbank
- Headless CMS
- Dateispeicher ohne Git
- Hybrid Git plus Datenbank

## Kriterium für spaetere Neubewertung
Wenn Such- und Abfrageanforderungen die Git-direkte Nutzung signifikant übersteigen, kann ein abgeleitetes Read Model erweitert werden. Quelle der Wahrheit bleibt in MVP 0.1 Git.
