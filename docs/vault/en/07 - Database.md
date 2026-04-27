# 07 — Database / Veritabanı

[[00 - PomeloHook Index|← Index]]

> Pure-Go SQLite via `modernc.org/sqlite`. No CGO, no system sqlite3 required. Single file: `pomelodata.db`.

---

## Schema

```sql
organizations
  id          TEXT PRIMARY KEY          -- "org_1", manual seed
  name        TEXT NOT NULL
  created_at  DATETIME DEFAULT CURRENT_TIMESTAMP

users
  id          TEXT PRIMARY KEY          -- "usr_1", manual seed; uuid for new
  org_id      TEXT REFERENCES organizations(id)
  email       TEXT UNIQUE NOT NULL
  name        TEXT NOT NULL
  api_key     TEXT UNIQUE NOT NULL      -- "ph_" + 48 hex chars
  role        TEXT NOT NULL DEFAULT 'member'   -- 'admin' | 'member'
  created_at  DATETIME DEFAULT CURRENT_TIMESTAMP

tunnels
  id             TEXT PRIMARY KEY       -- uuid
  type           TEXT NOT NULL          -- 'personal' | 'org'
  user_id        TEXT REFERENCES users(id)   -- null if org tunnel
  org_id         TEXT REFERENCES organizations(id)  -- null if personal tunnel
  subdomain      TEXT UNIQUE NOT NULL   -- random hex(4) or org tunnel name
  active_user_id TEXT REFERENCES users(id)   -- who is currently connected
  status         TEXT NOT NULL DEFAULT 'inactive'  -- 'active' | 'inactive'
  created_at     DATETIME DEFAULT CURRENT_TIMESTAMP

webhook_events
  id              TEXT PRIMARY KEY      -- uuid
  tunnel_id       TEXT REFERENCES tunnels(id)
  received_at     DATETIME NOT NULL     -- UTC, RFC3339
  method          TEXT NOT NULL
  path            TEXT NOT NULL         -- full path including /webhook/{subdomain}/...
  headers         TEXT NOT NULL         -- JSON: {"Content-Type": ["application/json"]}
  request_body    TEXT
  response_status INTEGER               -- 0 if forward failed
  response_body   TEXT
  response_ms     INTEGER
  forwarded       BOOLEAN NOT NULL DEFAULT FALSE
  replayed_at     DATETIME              -- null until first replay

INDEX: idx_events_tunnel_received ON webhook_events(tunnel_id, received_at)
```

---

## Design Decisions

### Why SQLite?

**Considered:** PostgreSQL, MySQL, embedded key-value (bbolt, badger)

**Chose SQLite because:**
- Zero ops — no separate database process, no connection string, no credentials
- Single file backup: `cp pomelodata.db pomelodata.db.bak`
- Full relational queries — retention cleanup, event listing by tunnel, replay lookups
- `modernc.org/sqlite` compiles to pure Go (no CGO), so the binary works anywhere Go does

**Trade-off:** Single writer. `db.SetMaxOpenConns(1)` is enforced. This is fine — PomeloHook is not a high-write system. Events arrive one-by-one; concurrent writes are rare.

### Why TEXT PRIMARY Keys, Not Auto-Increment?

UUIDs for programmatically created rows (`uuid.NewString()`), manual strings for seeded rows (`org_1`, `usr_1`). UUIDs are safe to expose in URLs and API responses without revealing row counts or sequential IDs.

### Why Store Headers as JSON TEXT?

`http.Header` is `map[string][]string`. SQLite has no native map type. JSON serialization is straightforward, and the dashboard needs to display headers as-is. No need to query individual header values at the DB level.

### Why `COALESCE` in Column Lists?

```go
const tunnelColumns = `id, type, COALESCE(user_id,''), ...`
```

Nullable columns (user_id, org_id, active_user_id) would return `sql.NullString` if scanned normally. `COALESCE(col,'')` lets us scan directly into `string`, simplifying the scan functions. The trade-off: you can't distinguish NULL from empty string, but that distinction is never needed here.

### Why `received_at` as TEXT (RFC3339), Not DATETIME?

SQLite stores datetimes as text internally regardless. Storing RFC3339 explicitly means:
- Portable across SQLite drivers
- `time.Parse(time.RFC3339, ...)` always works
- The index on `(tunnel_id, received_at)` works because RFC3339 sorts lexicographically

### Foreign Keys + WAL Mode

```go
dsn = dsn + "?_pragma=foreign_keys(1)"
```

Foreign keys are OFF by default in SQLite and must be enabled per-connection. Enforced at the DSN level so every connection gets it automatically.

WAL mode is not explicitly set — the default journal mode is used. For a single-writer setup this is fine.

---

## Migration Strategy

`store.Open()` calls `migrate(db)` on every startup:

```go
_, err = tx.Exec(`
    CREATE TABLE IF NOT EXISTS organizations (...);
    CREATE TABLE IF NOT EXISTS users (...);
    ...
`)
```

`IF NOT EXISTS` means this is idempotent — safe to run on an existing database. **There is no versioned migration system.** Adding columns requires `ALTER TABLE` statements added to the migrate function, or manual SQL.

**Implication:** If you add a column, you must also handle existing databases that don't have it. The current approach works because the schema has been stable since v1.0.

---

## Retention

```go
// Runs every 24 hours
db.DeleteEventsOlderThan(cfg.RetentionDays)

// SQL:
DELETE FROM webhook_events WHERE received_at < ?
-- cutoff = time.Now().UTC().AddDate(0, 0, -RetentionDays)
```

Default: 30 days. Set `POMELO_RETENTION_DAYS` to change.

The ticker starts 24 hours after boot — no immediate cleanup on startup.

---

## Admin DB Panel

The admin panel exposes direct DB access:

```
GET  /api/admin/db/tables         → list all table names
GET  /api/admin/db/tables/{name}  → paginated rows (limit, offset)
POST /api/admin/db/query          → raw SQL
```

`handleRunQuery` executes any SQL. Write queries (not pure SELECT) trigger a confirmation dialog in the UI. There is **no server-side restriction** on query type — the confirmation is UI-only.

---

## Related Notes

- Store layer code → [[04 - Server]]
- Admin panel → [[06 - Dashboard]]
- API endpoints → [[09 - API Reference]]
