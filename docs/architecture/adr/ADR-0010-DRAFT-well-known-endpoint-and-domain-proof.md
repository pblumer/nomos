# ADR-0010 (DRAFT): Well-known-Endpoint-Spezifikation und Domain-Proof-Format

## Status

Draft

## Datum

2026-05-16

## Kontext

ADR-0009 legt fest, dass jede bestaetigte Domain einen Core-Knoten unter
`https://<domain>/.well-known/nomos` betreibt und einen kryptografischen Domain-Proof erbringt.
Dieses ADR spezifiziert das genaue Format beider Artefakte.

## Offene Fragen

- Welches genaue JSON-Schema hat die `.well-known/nomos`-Discovery-Antwort?
- Wie wird der Domain-Proof kryptografisch erstellt und verifiziert (DNS-TXT vs. X.509 vs. beides)?
- Wie lang ist ein Proof gueltig, und wie wird er erneuert?
- Wie verhalten sich Knoten, wenn der Proof eines Peers abgelaufen ist?
- Braucht es ein Revocation-Verfahren?

## Kandidaten-Optionen (noch nicht entschieden)

**Option A — DNS-TXT-Record:**
Der Domain-Inhaber hinterlegt einen TXT-Record `_nomos.<domain>` mit einem signierten Token.
Vorteil: Einfach, keine PKI noetig. Nachteil: DNS-Propagierung, TTL-Latenz.

**Option B — X.509-Zertifikat:**
Der Core-Knoten prasentiert ein TLS-Zertifikat, das die Domain beweist. Let's Encrypt liefert
dies automatisch. Vorteil: Bereits in HTTPS-Betrieb erzwungen. Nachteil: Certifikat-Rotation
muss gehandhabt werden.

**Option C — Signiertes cosmos.yaml:**
Das `cosmos.yaml` wird mit einem Domain-Schluessel signiert. Der oeffentliche Schluessel ist
per DNS-TXT oder via `.well-known/nomos` abrufbar. Vorteil: In Git versioniert, auditierbar.

## Naechste Schritte

- Proof-of-Concept mit Option B (X.509) als Baseline.
- Evaluierung ob Option C als ergaenzende Schicht noetig ist.
- Spezifikation des JSON-Schemas fuer `.well-known/nomos`.

## Referenz

- ADR-0009 — Nomos Cosmos Netzwerkarchitektur und Core Engine
