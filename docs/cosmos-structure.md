# cosmos-structure

## DNS-like namespace and domain tree

Nomos presents domains as a DNS-like tree. The `Cosmos` root contains a `Namespaces` branch. Each namespace is a top-level zone such as `com`, `net`, `ch`, or `cloud`. Domains are label nodes below that zone.

```text
Cosmos
└── Namespaces
    └── <namespace>
        └── <domain-label>
            └── <child-domain-label>
```

The canonical DNS name is derived from the path by reading the labels from the leaf back to the namespace:

- `com/blumer` becomes `blumer.com`.
- `com/blumer/identity` becomes `identity.blumer.com`.
- `cloud/blumer/home` becomes `home.blumer.cloud`.
- `ch/beispiel` becomes `beispiel.ch`.

The tree displays labels only, for example `com → blumer → identity`. Detail panels and APIs continue to use the canonical name such as `identity.blumer.com` as the source of truth. Existing storage remains under `domains/<canonical>`.

Services are attached to domain nodes, but the UI separates them below a `Services` pseudo-node so that service names are not confused with child domain labels.
