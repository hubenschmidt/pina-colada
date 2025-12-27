# Agent Go Refactor

Migration from Python/LangGraph agent to Go agent.

## Status: In Progress

## Summary

Removed the Python-based LangGraph agent (`modules/agent`) and promoted the Go agent (`modules/agent-go`) to be the primary agent implementation, renamed to `modules/agent`.

## Changes

### Removed

| Item | Description |
|------|-------------|
| `modules/agent/` | Entire Python LangGraph agent directory |
| `agent:` service (old) | Python agent service from docker-compose.yml |
| `scripts/apply_migrations.py` | Python migration script |
| `scripts/check_migration_status.py` | Python migration status checker |
| `scripts/run_seeders.py` | Python seeder runner |
| `scripts/seed_documents.py` | Python document seeder |
| `scripts/ETL/extract.py` | Python DB export script |
| `scripts/ETL/transform.py` | Python CSV transform script |
| `scripts/import_jobs.py` | Python job import script |

### Renamed

| From | To |
|------|-----|
| `modules/agent-go/` | `modules/agent/` |
| `agent-go:` service | `agent:` service |

### Created Go Commands

| Command | Location | Replaces | Description |
|---------|----------|----------|-------------|
| `migrate` | `cmd/migrate/` | `apply_migrations.py`, `check_migration_status.py` | Database migrations via golang-migrate |
| `seed` | `cmd/seed/` | `run_seeders.py` | Run SQL seeder files |
| `seed-documents` | `cmd/seed-documents/` | `seed_documents.py` | Upload seed documents to storage |
| `etl` | `cmd/etl/` | `extract.py`, `transform.py`, `import_jobs.py` | ETL operations |

### ETL Subcommands

```bash
# Export tables to CSV
etl extract --dir ./exports

# Transform CSVs to new schema
etl transform --input ./exports --output ./transformed

# Import jobs from CSV
etl import-jobs --file ./imports/jobs.csv --dry-run
etl import-jobs --file ./imports/jobs.csv --tenant-id 1 --deal-id 1 --user-id 1
```

### Updated Files

| File | Changes |
|------|---------|
| `docker-compose.yml` | Removed old `agent:`, renamed `agent-go:` to `agent:`, updated paths |
| `start.sh` | Rewritten for Go: runs migrate, seed, seed-documents, then air |
| `start-prod.sh` | Rewritten for Go: runs /migrate, then /agent-go |
| `Dockerfile` | Added builds for migrate, seed, seed-documents, etl |
| `Dockerfile.prod` | Converted from Python to Go multi-stage build |
| `go.mod` | Added golang-migrate dependency |

### New Internal Package

Created `internal/etl/` package with:
- `extract.go` - Export database tables to CSV
- `transform.go` - Transform CSVs between schema versions
- `import_jobs.go` - Import jobs from CSV file

## Migration Commands

### Local Development

```bash
# Migrations run automatically on container start via start.sh
# Or run manually:
docker compose exec agent go run ./cmd/migrate up
docker compose exec agent go run ./cmd/migrate status
docker compose exec agent go run ./cmd/seed
docker compose exec agent go run ./cmd/seed-documents
```

### Production (Supabase)

```bash
# Set DATABASE_URL to Supabase connection string
export DATABASE_URL="postgres://..."

# Run migrations
/migrate up
/migrate status

# The start-prod.sh script runs migrations automatically before starting the agent
```

## Dependencies Added

- `github.com/golang-migrate/migrate/v4` - Database migration library
- `github.com/lib/pq` - PostgreSQL driver for golang-migrate

## Docker Images

Both `Dockerfile` and `Dockerfile.prod` now build these binaries:
- `/agent-go` - Main agent server
- `/migrate` - Migration tool
- `/seed` - SQL seeder
- `/seed-documents` - Document seeder
- `/etl` - ETL tool

## TODO

- [ ] Test full docker-compose up with new agent
- [ ] Verify migrations work against Supabase
- [ ] Update Azure pipelines if needed
- [ ] Remove any remaining Python-specific CI/CD steps
