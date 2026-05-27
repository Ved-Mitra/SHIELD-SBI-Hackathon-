# thread-intel-database (prototype)

## Scope
- Store threat intelligence items for SHIELD.
- Provide searchable access for takedown and dashboard.

## Data model (minimal)
- `indicator` (domain, url, ip, handle)
- `source` (crawler, manual, partner)
- `confidence` (0-100)
- `severity` (low, medium, high)
- `evidence` (links, hashes, notes)
- `first_seen`, `last_seen`
- `status` (new, in_review, takedown_requested, resolved)

## Access patterns
- Insert new intel items.
- Query by indicator type/value.
- Filter by severity/status/time window.
- Join with takedown results for audit.

## Storage choice (prototype)
- Start with Postgres + JSONB for evidence.
- Add indexes on indicator type/value and status.

## Proposed structure
```
thread-intel-database/
  README.md
  schema/
    001_init.sql
  docs/
    data-dictionary.md
```

## References
- Postgres JSONB: https://www.postgresql.org/docs/current/datatype-json.html

