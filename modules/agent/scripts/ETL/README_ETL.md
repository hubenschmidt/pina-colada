# ETL README

## ETL Process

Three-step process to migrate production data:

### Step 1: Extract Production Data

Use `extract.py` to export production data from Supabase to CSV files.

```bash
python3 scripts/ETL/extract.py
```

**What It Does:**

- Connects to production Supabase database
- Exports business tables to CSV files in `./exports/`:
  - `Tenant.csv`
  - `User.csv`
  - `Organization.csv`
  - `Lead.csv`
  - `Job.csv`
  - `Deal.csv`

**Output:** CSV files written to `modules/agent/exports/` with headers included.

### Step 2: Transform CSV Files

Use `transform.py` to convert exported CSVs from old schema to new schema.

```bash
python3 scripts/ETL/transform.py
```

**What It Does:**

- Reads CSV files from `./exports/`
- Transforms schema (removes old fields, adds new ones, creates Account.csv)
- Writes transformed files to `./exports/transformed/`

**Output:** Transformed CSV files in `modules/agent/exports/transformed/`

### Step 3: Load Transformed Data

Use `load_*.py` to import transformed CSV files into target database.

```bash
python3 scripts/ETL/load_12012025.py
```

**What It Does:**

- Reads transformed CSV files from `./exports/transformed/`
- Imports tables in dependency order using `psql \COPY`
- Handles foreign key relationships

**Requirements:**

- `psql` (PostgreSQL client) must be installed
- Database credentials configured in load script
- Target database must have migrations applied

## Requirements

- `psql` (PostgreSQL client) must be installed
- Database credentials configured in scripts
