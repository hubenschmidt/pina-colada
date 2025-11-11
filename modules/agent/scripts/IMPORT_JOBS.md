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

### Option 1: DigitalOcean App Platform (Recommended)

DigitalOcean App Platform provides a console/SSH feature to access your running containers.

#### Step 1: Access App Platform Console

**Method A: Via DigitalOcean Dashboard (Easiest)**

1. Log into [DigitalOcean Dashboard](https://cloud.digitalocean.com/)
2. Go to **Apps** → Select your `pina-colada-agent` app
3. Click on the **Runtime Logs** tab
4. Look for the **"Console"** button or **"Open Console"** link (usually in the top right)
5. This opens a web-based terminal into your running container

**Method B: Via `doctl` CLI Tool**

1. Install `doctl` if you haven't already:
   ```bash
   # macOS
   brew install doctl
   
   # Linux
   wget https://github.com/digitalocean/doctl/releases/download/v1.104.0/doctl-1.104.0-linux-amd64.tar.gz
   tar xf doctl-1.104.0-linux-amd64.tar.gz
   sudo mv doctl /usr/local/bin
   ```

2. Authenticate:
   ```bash
   doctl auth init
   # Enter your DigitalOcean API token when prompted
   ```

3. List your apps to find the app ID:
   ```bash
   doctl apps list
   ```

4. Get a console session:
   ```bash
   # Get the component ID (usually the service name like "agent")
   doctl apps get YOUR_APP_ID
   
   # Open console (this will give you instructions)
   doctl apps logs YOUR_APP_ID --component=agent --type=run
   ```

#### Step 2: Upload CSV File to App Platform

**Method A: Via App Platform Console**

1. Once in the console, navigate to the app directory:
   ```bash
   cd /app
   ls -la  # Verify you're in the right place
   ```

2. Create the imports directory:
   ```bash
   mkdir -p imports
   ```

3. Use `cat` to create the file (paste CSV content):
   ```bash
   cat > imports/jobs.csv << 'EOF'
   # Paste your CSV content here
   # Press Ctrl+D when done
   EOF
   ```

**Method B: Via App Spec (if you have file storage configured)**

If your app has persistent storage, you can mount it and copy files there.

**Method C: Base64 Encode/Decode (for binary-safe transfer)**

1. On your local machine, encode the CSV:
   ```bash
   base64 modules/agent/imports/jobs.csv > jobs.csv.b64
   ```

2. Copy the base64 content and paste it in the console:
   ```bash
   # In App Platform console
   cat > imports/jobs.csv.b64 << 'EOF'
   # Paste base64 content here
   EOF
   
   # Decode it
   base64 -d imports/jobs.csv.b64 > imports/jobs.csv
   rm imports/jobs.csv.b64
   ```

#### Step 3: Verify Environment Variables

The import script requires these PostgreSQL environment variables to connect to Supabase:

**Required Environment Variables:**
- `POSTGRES_HOST` - Your Supabase database host (e.g., `db.xxxxx.supabase.co` or `xxxxx.supabase.co`)
- `POSTGRES_PORT` - Database port (usually `5432` or `6543` for connection pooling)
- `POSTGRES_USER` - Database user (usually `postgres`)
- `POSTGRES_PASSWORD` - Your Supabase database password
- `POSTGRES_DB` - Database name (usually `postgres`)

**In App Platform Console:**

1. Check if environment variables are set:
   ```bash
   env | grep POSTGRES
   ```

2. If they're missing, you need to add them in App Platform:
   - Go to your App Platform app settings
   - Navigate to **Settings** → **App-Level Environment Variables**
   - Add each variable listed above
   - **Important**: You'll need to restart/redeploy the app for changes to take effect

3. To find your Supabase connection details:
   - Go to [Supabase Dashboard](https://app.supabase.com/)
   - Select your project
   - Go to **Settings** → **Database**
   - Find the **Connection string** section
   - Extract the host, port, user, password, and database name from the connection string

**Example connection string format:**
```
postgresql://postgres:[YOUR-PASSWORD]@db.xxxxx.supabase.co:5432/postgres
```

**Quick Test:**
```bash
# In App Platform console, test the connection
python -c "
import os
from agent.repositories.job_repository import _get_connection_string
print('Connection string:', _get_connection_string())
"
```

#### Step 4: Run the Import Script

```bash
# Navigate to app directory (if not already there)
cd /app

# Dry run first (ALWAYS recommended!)
python scripts/import_jobs.py imports/jobs.csv --dry-run

# If everything looks good, run the actual import
python scripts/import_jobs.py imports/jobs.csv
```

#### Step 5: Verify Import

The script will output how many jobs were imported. You can also verify via your application's API or database.

### Option 2: Traditional VPS/Droplet (If Not Using App Platform)

If you're using a traditional Droplet instead of App Platform:

#### Step 1: SSH Into Your Server

```bash
# Using SSH key (recommended)
ssh root@YOUR_DROPLET_IP

# Or with custom key
ssh -i /path/to/your/key root@YOUR_DROPLET_IP
```

#### Step 2: Find Your Application

```bash
# If using Docker Compose
docker compose ps
cd /path/to/your/app

# If using plain Docker
docker ps
```

#### Step 3: Copy CSV File

From your local machine:
```bash
scp modules/agent/imports/jobs.csv root@YOUR_DROPLET_IP:/tmp/jobs.csv
```

Then on the server:
```bash
# If using Docker Compose
docker compose cp /tmp/jobs.csv agent:/app/imports/jobs.csv

# Or copy into container
docker cp /tmp/jobs.csv CONTAINER_NAME:/app/imports/jobs.csv
```

#### Step 4: Run Import

```bash
# If using Docker Compose
docker compose exec agent python scripts/import_jobs.py imports/jobs.csv --dry-run
docker compose exec agent python scripts/import_jobs.py imports/jobs.csv

# If using plain Docker
docker exec -it CONTAINER_NAME python scripts/import_jobs.py imports/jobs.csv --dry-run
docker exec -it CONTAINER_NAME python scripts/import_jobs.py imports/jobs.csv
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

