# ADR-0004 - KI Assistenz nur für Drafts, keine autonomen Entscheidungen

## Kontext
LLM kann Produktteams bei Modellierung unterstuetzen, bringt aber Risiken für Nachvollziehbarkeit und Governance.

## Entscheidung
KI darf in MVP 0.1 nur assistiv für Drafts genutzt werden und keine finalen Entscheidungen treffen oder freigegebene Artefakte ohne Review ändern.

## Erlaubte Rollen
- Draft-Erstellung für Anforderungen und Regeln.
- Formulierungshilfe, Konsistenzhinweise, Szenariovorschläege.

## Verbotene Rollen
- Autoritative Rule-Auslegung zur Laufzeit.
- Freigabeersatz.
- Stille Änderung an freigegebenen Artefakten.

## Pflichtprinzipien
- Human Review bleibt verpflichtend.
- Deterministische Validierung fuehrt freigegebene Regeln aus.
- Governanceprozess darf nicht umgangen werden.

## Governancewirkung
Jede akzeptierte KI-Unterstuetzung wird als normaler branchbasierter Änderungsvorschlag behandelt und reviewt.
