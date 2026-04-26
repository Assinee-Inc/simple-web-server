# Data Model for "Remove Client ↔ Creator Relationship"

This document describes the data entities after the refactoring, showing the simplified model.

## Key Entities

### Client (modified)

Represents a client in the system. After removing the `client_creators` relationship, a Client is associated with a Creator only through their purchases.

**Attributes**:

| Name        | Type     | Description                                      |
|-------------|----------|--------------------------------------------------|
| `id`        | `uint`   | Primary key                                      |
| `public_id` | `string` | Semantic ID with `cli_` prefix                   |
| `name`      | `string` | Full name of the client                          |
| `cpf`       | `string` | CPF number (unique, used for deduplication)       |
| `birthdate` | `string` | Date of birth (DD-MM-YYYY format)                |
| `email`     | `string` | Email address                                    |
| `phone`     | `string` | Phone number                                     |
| `validated` | `bool`   | Whether the client has been validated            |

**Relationships (AFTER refactoring)**:

```text
Client (1) ────── (N) Purchase
                         │
                         ▼
                    Ebook (N) ────── (1) Creator
```

**Derived relationship** (via purchases):
- A Client belongs to a Creator IF they have at least one Purchase for an Ebook created by that Creator
- This is NOT stored as a separate field — it is computed on query

**Constraints**:
- `cpf` must be unique across all clients (GORM unique index)
- `public_id` must be unique (used in URLs)

### Purchase (unchanged)

Links a Client to an Ebook. The Creator is derived via Ebook.CreatorID.

| Name        | Type     | Description                              |
|-------------|----------|------------------------------------------|
| `id`        | `uint`   | Primary key                              |
| `client_id` | `uint`   | Foreign key to Client                    |
| `ebook_id`  | `uint`   | Foreign key to Ebook                     |
| `status`    | `string` | Purchase status                          |
| ...         | ...      | Other purchase fields                    |

**Relationship to Creator**: Computed via `Purchase.Ebook.Creator`

### Ebook (unchanged)

| Name        | Type     | Description                              |
|-------------|----------|------------------------------------------|
| `id`        | `uint`   | Primary key                              |
| `creator_id` | `uint`  | Foreign key to Creator                  |
| ...         | ...      | Other ebook fields                       |

### Creator (unchanged)

| Name     | Type     | Description                       |
|----------|----------|-----------------------------------|
| `id`     | `uint`   | Primary key                       |
| `email`  | `string` | Email (used to identify creator)  |
| ...      | ...      | Other creator fields              |

## Removed Entities

### ClientCreator (DELETED)

The join table `client_creators` and the `ClientCreator` model are completely removed.

**Before**:
```text
Client (N) ────── (N) ClientCreator ────── (N) Creator
```

**After**: No direct Client-Creator relationship. Only via:
```text
Client (1) ────── (N) Purchase ────── (N) Ebook ────── (1) Creator
```

## Query Patterns

### Find all clients for a Creator (REWRITTEN)

```sql
-- BEFORE (using client_creators):
SELECT DISTINCT clients.*
FROM clients
JOIN client_creators ON client_creators.client_id = clients.id
WHERE client_creators.creator_id = ?

-- AFTER (using purchases):
SELECT DISTINCT clients.*
FROM clients
JOIN purchases ON purchases.client_id = clients.id
JOIN ebooks ON ebooks.id = purchases.ebook_id
WHERE ebooks.creator_id = ?
```

### Find clients who purchased an ebook (UNCHANGED)

```sql
SELECT clients.*
FROM clients
JOIN purchases ON purchases.client_id = clients.id
WHERE purchases.ebook_id = ?
```

### Find clients who did NOT purchase an ebook from a creator

```sql
-- BEFORE (using client_creators):
SELECT clients.*
FROM clients
JOIN client_creators ON client_creators.client_id = clients.id
WHERE client_creators.creator_id = ?
  AND clients.id NOT IN (SELECT client_id FROM purchases WHERE ebook_id = ?)

-- AFTER (using purchases):
SELECT clients.*
FROM clients
JOIN purchases ON purchases.client_id = clients.id
JOIN ebooks ON ebooks.id = purchases.ebook_id
WHERE ebooks.creator_id = ?
  AND clients.id NOT IN (SELECT client_id FROM purchases WHERE ebook_id = ?)
```

## Migration Notes

- **No data migration needed**: The `client_creators` table is a cache/shortcut. Removing it doesn't lose information — the same relationships are derivable from `purchases + ebooks`.
- **Existing purchases**: All existing purchases already link clients to creators via ebooks. No purchase is orphaned by this change.
- **Application behavior**: Client listing/export features will work exactly the same — just with a different SQL query path.