# Data Model for "Omit Client's Date of Birth from CSV Export"

This document describes the data entities relevant to this feature, based on the feature specification.

## Key Entities

### Client

Represents a client in the system.

**Attributes**:

| Name        | Type   | Description                               |
|-------------|--------|-------------------------------------------|
| `id`        | `int`  | Unique identifier for the client.         |
| `name`      | `string` | The full name of the client.              |
| `email`     | `string` | The email address of the client.          |
| `phone`     | `string` | The phone number of the client.           |
| `birthdate` | `date`   | The client's date of birth. **(Omitted from export)** |

**Relationships**:

- A `Client` can have multiple `Purchases`.
- A `Client` belongs to a `Creator`.

**State Transitions**:

- Not applicable for this feature.
