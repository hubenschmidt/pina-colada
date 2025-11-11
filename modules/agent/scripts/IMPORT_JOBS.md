# Import Jobs Script - Production Guide

## Overview

The `import_jobs.py` script imports job applications from a CSV file into the database. It includes duplicate detection to prevent accidentally importing the same jobs twice.

## Local Development

**Note**: Since the application runs in Docker locally, you need to run the script inside the Docker container.

```bash
# From project root directory

# First, do a dry run to preview what will be imported
docker compose exec agent python scripts/import_jobs.py imports/jobs.csv --dry-run

# If everything looks good, run the actual import
docker compose exec agent python scripts/import_jobs.py imports/jobs.csv
```

**Alternative**: If you want to run it directly (without Docker), you'll need to fix the virtual environment permissions first, but Docker is recommended for consistency.

## Production Deployment

**Important**: This script should NOT be run as part of the CI/CD pipeline to avoid accidental duplicate runs.

### Option 1: Manual Execution via SSH (Recommended)

#### Step 1: Get Your DigitalOcean Server Details

1. Log into [DigitalOcean Dashboard](https://cloud.digitalocean.com/)
2. Go to **Droplets** → Select your server
3. Note your server's **IP address** (e.g., `123.45.67.89`)
4. Check your **SSH key** setup (usually in Settings → Security)

#### Step 2: SSH Into Your Server

**If you have SSH key set up:**
```bash
ssh root@YOUR_SERVER_IP
# or if you have a non-root user:
ssh your-username@YOUR_SERVER_IP
```

**If you need to use password:**
```bash
ssh root@YOUR_SERVER_IP
# Enter password when prompted
```

**If you have a custom SSH key:**
```bash
ssh -i /path/to/your/private/key root@YOUR_SERVER_IP
```

#### Step 3: Find Your Application Directory

Once connected, locate where your application is deployed:

```bash
# Common locations:
cd /var/www/pina-colada-co
# or
cd /home/your-user/pina-colada-co
# or
cd /opt/pina-colada-co
# or if using Docker:
docker ps  # Find your agent container name
```

**If using Docker:**
```bash
# List running containers
docker ps

# Find the agent container (might be named something like "agent" or "pina-colada-agent")
# You'll need to exec into it later
```

#### Step 4: Copy CSV File to Server

**From your local machine** (in a new terminal, while SSH'd in another):

```bash
# Replace with your actual server IP and path
scp modules/agent/imports/jobs.csv root@YOUR_SERVER_IP:/path/to/app/modules/agent/imports/jobs.csv
```

**Or if you're already SSH'd in, you can create the file directly:**
```bash
# On the server, create the imports directory if needed
mkdir -p modules/agent/imports

# Then use nano/vim to create the file, or use scp from another terminal
```

#### Step 5: Set Up Environment Variables

**If NOT using Docker** (application runs directly on server):
```bash
# Check if .env file exists
cat modules/agent/.env

# Or set environment variables manually
export POSTGRES_HOST=your-supabase-host.db.supabase.co
export POSTGRES_PORT=5432
export POSTGRES_USER=postgres
export POSTGRES_PASSWORD=your-password
export POSTGRES_DB=postgres
```

**If using Docker** (most common):
```bash
# Check environment variables in container
docker exec your-agent-container env | grep POSTGRES

# Or check docker-compose.yml or .env file
cat docker-compose.yml | grep POSTGRES
```

#### Step 6: Run the Import Script

**If NOT using Docker:**
```bash
cd modules/agent

# Dry run first (always recommended!)
python scripts/import_jobs.py imports/jobs.csv --dry-run

# If everything looks good, run for real
python scripts/import_jobs.py imports/jobs.csv
```

**If using Docker:**
```bash
# Option A: Exec into the container and run script
docker exec -it your-agent-container bash
cd /app
python scripts/import_jobs.py imports/jobs.csv --dry-run
python scripts/import_jobs.py imports/jobs.csv
exit

# Option B: Run command directly without entering container
docker exec -it your-agent-container python scripts/import_jobs.py imports/jobs.csv --dry-run
docker exec -it your-agent-container python scripts/import_jobs.py imports/jobs.csv

# Option C: If using docker-compose
docker compose exec agent python scripts/import_jobs.py imports/jobs.csv --dry-run
docker compose exec agent python scripts/import_jobs.py imports/jobs.csv
```

#### Step 7: Verify Import

Check that jobs were imported:
```bash
# If you have a way to query the database or check via API
# Or check the script output - it will show how many were imported
```

### Quick Reference: Common DigitalOcean Commands

```bash
# 1. Connect to server
ssh root@YOUR_SERVER_IP

# 2. Find your app (if using Docker)
docker ps
docker compose ps

# 3. Copy file from local to server (from your local machine)
scp modules/agent/imports/jobs.csv root@YOUR_SERVER_IP:/root/pina-colada-co/modules/agent/imports/

# 4. Run import in Docker container
docker compose exec agent python scripts/import_jobs.py imports/jobs.csv --dry-run
docker compose exec agent python scripts/import_jobs.py imports/jobs.csv
```

### Option 2: One-Time Azure Pipeline Job (Alternative)

If you prefer to use Azure Pipelines but want to ensure it only runs once:

1. Create a separate pipeline file: `azure-pipelines-import.yml`
2. Add a manual approval gate
3. Use a pipeline variable to track if import has been run
4. Only trigger manually, not on commits

**Example pipeline structure:**
```yaml
trigger: none  # Manual only

stages:
- stage: ImportJobs
  jobs:
  - job: Import
    steps:
    - script: |
        python scripts/import_jobs.py imports/jobs.csv --dry-run
      displayName: "Dry Run Import"
    
    - script: |
        python scripts/import_jobs.py imports/jobs.csv
      displayName: "Import Jobs"
      condition: eq(variables['Build.SourceBranch'], 'refs/heads/master')
```

## Safety Features

The script includes several safety features:

1. **Duplicate Detection**: Automatically skips jobs that already exist (matches on company + job_title + application_date)
2. **Dry Run Mode**: Preview imports without saving (`--dry-run` flag)
3. **Error Handling**: Continues processing even if individual rows fail
4. **Detailed Output**: Shows exactly what was imported, skipped, or failed

## Troubleshooting

### "File not found" error
- Make sure the CSV file path is correct
- The script looks for files relative to `modules/agent/` directory
- Use absolute paths if needed: `/full/path/to/jobs.csv`

### "Could not load existing jobs" warning
- Check database connection settings (POSTGRES_* environment variables)
- Verify database is accessible from the server
- The script will continue but won't check for duplicates

### Duplicate detection too strict/loose
- Use `--no-skip-duplicates` to disable duplicate checking (not recommended)
- Modify `check_duplicate()` function in the script to adjust matching logic

## CSV Format

Expected CSV columns:
- `Description` - Company name (required)
- `Job title` - Job title (required)
- `date applied` - Application date (MM/DD/YYYY format)
- `link` - Job posting URL (optional)
- `note` - Notes (optional)
- `job board` - Source (optional, defaults to "manual")
- `Resume` - Resume version used (optional, added to notes)

Rows are skipped if:
- Missing company or job title
- Date field contains "did not apply", "to apply", "jen to apply", or "need to finish"

