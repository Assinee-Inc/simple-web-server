# Contract: Client CSV Export

This document defines the contract for the client data CSV export.

## CSV File Format

The exported CSV file will have the following structure:

**File Name**: `clients_export.csv`

**Header Row**:

```csv
name,email,phone
```

**Data Rows**:

Each row in the CSV file will correspond to a single client, with the data in the same order as the header row.

**Example**:

```csv
"John Doe","john.doe@example.com","+15551234567"
"Jane Smith","jane.smith@example.com","+15557654321"
```

## Changes from Previous Version

- The `birthdate` column has been removed from the CSV export.
