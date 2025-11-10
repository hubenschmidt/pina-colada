# Google Sheets Job Tracking Integration

This guide explains how to set up Google Sheets integration to track job applications and prevent the agent from suggesting duplicate jobs.

## Overview

The agent now integrates with your Google Sheets job tracker to:
- Read the list of jobs you've already applied to
- Automatically filter out these jobs from search results
- Prevent duplicate recommendations

## Prerequisites

- A Google Cloud Platform account (free tier works fine)
- Your job tracker Google Sheet (already at: `https://docs.google.com/spreadsheets/d/1booMeM9GMAbrF6DgV9hsKo5_dxwYcD5kr049kAULAIA`)

## Setup Steps

### 1. Create a Google Cloud Project

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project (or use an existing one)
3. Note your project ID

### 2. Enable Google Sheets API

1. In your Google Cloud project, go to **APIs & Services > Library**
2. Search for "Google Sheets API"
3. Click **Enable**

### 3. Create a Service Account

1. Go to **APIs & Services > Credentials**
2. Click **Create Credentials** > **Service Account**
3. Fill in the details:
   - **Service account name**: `job-tracker-agent` (or any name you prefer)
   - **Service account ID**: Will auto-generate
   - **Description**: "Service account for job search agent to read job tracker"
4. Click **Create and Continue**
5. Skip the optional steps (no roles needed for Sheets access)
6. Click **Done**

### 4. Create and Download Service Account Key

1. Find your newly created service account in the list
2. Click on it to open details
3. Go to the **Keys** tab
4. Click **Add Key** > **Create new key**
5. Choose **JSON** format
6. Click **Create**
7. The key file will download automatically
8. **IMPORTANT**: Save this file securely - it provides access to your Google account

### 5. Share Your Google Sheet with the Service Account

1. Open the downloaded JSON key file
2. Find the `client_email` field (looks like `job-tracker-agent@your-project.iam.gserviceaccount.com`)
3. Copy this email address
4. Open your [job tracker Google Sheet](https://docs.google.com/spreadsheets/d/1booMeM9GMAbrF6DgV9hsKo5_dxwYcD5kr049kAULAIA)
5. Click **Share** button
6. Paste the service account email
7. Set permission to **Viewer** (read-only)
8. Uncheck "Notify people"
9. Click **Share**

### 6. Configure the Agent

1. Convert the JSON key file to a single-line string:
```bash
# Navigate to where you downloaded the credentials file
cd /path/to/your/credentials

# Convert JSON to single line using Python
python3 -c "import json; print(json.dumps(json.load(open('google-credentials.json'))))"
```

2. Copy the entire output (it will be one long line of JSON)

3. Update your `.env` file:
```bash
# Google Sheets Integration
GOOGLE_SHEETS_ID="1booMeM9GMAbrF6DgV9hsKo5_dxwYcD5kr049kAULAIA"
GOOGLE_SHEETS_CREDENTIALS_JSON='{"type":"service_account","project_id":"your-project",...}'
```

**Important Notes:**
- Use single quotes around the JSON to avoid shell escaping issues
- The entire JSON must be on one line with no newlines
- The `credentials/` directory is already in `.gitignore`, so your local credentials file won't be committed
- You can delete the `google-credentials.json` file after adding it to `.env` (optional)

### 7. Verify Google Sheet Structure

Your Google Sheet should have these columns (first row):
- **Company**: Company name
- **Job Title**: Job title or role

Example:
| Company | Job Title |
|---------|-----------|
| Google | Software Engineer |
| Meta | Frontend Developer |
| Amazon | Senior Backend Engineer |

The tracker will match jobs based on normalized company name + job title.

## Testing

1. Add some test jobs to your Google Sheet
2. **Rebuild your Docker images** to install the new dependencies:
   ```bash
   cd /home/hubenschmidt/pina-colada-co/modules/agent
   docker build -t your-agent-image .
   ```
3. Start the agent server
4. Ask the agent to search for jobs
5. Verify that jobs in your sheet are filtered out from results

You should see log messages like:
```
✓ Google Sheets client initialized successfully
✓ Loaded 15 applied jobs from Google Sheets
Filtering out: Google - Software Engineer
```

## Troubleshooting

### "GOOGLE_SHEETS_CREDENTIALS_PATH not set"
- Make sure you've set the environment variable in `.env`
- Use an absolute path, not a relative path

### "Credentials file not found"
- Verify the path in `.env` points to the correct location
- Check file permissions (should be readable by the user running the agent)

### "Permission denied" errors
- Verify you've shared the Sheet with the service account email
- Check that the service account has at least Viewer permissions

### No jobs being filtered
- Check that column names in your Sheet match exactly: "Company" and "Job Title"
- Verify data is in the first sheet (Sheet1)
- Check logs for parsing errors

### Jobs not matching correctly
- The matching is case-insensitive and normalized
- Make sure company names and job titles in your sheet match how they appear in search results
- Try adding variations if needed (e.g., "Google" and "Google LLC")

## Security Notes

- **RECOMMENDED**: Use `GOOGLE_SHEETS_CREDENTIALS_JSON` environment variable (already in your `.env`, which is gitignored)
- **Never commit credentials to git** - both `.env` and `credentials/` are already in `.gitignore`
- The service account only needs read access to the specific Sheet (set as "Viewer")
- Consider rotating service account keys periodically
- If using file-based credentials locally, restrict permissions: `chmod 600 google-credentials.json`
- Environment variables are more secure for Docker deployments (no file mounting needed)

## Architecture

### Job Search Filtering
```
User: "Find me jobs in NYC"
  ↓
Worker LLM detects job search
  ↓
Calls job_search tool (NOT web_search)
  ↓
job_search_with_filter():
  1. Fetches jobs from Google Serper API
  2. Reads applied jobs from Google Sheets
  3. Filters out duplicates (company + title match)
  4. Returns only new jobs
  ↓
Results streamed to user
```

### Checking Applied Jobs
```
User: "Did I apply to Google for Software Engineer?"
  ↓
Worker LLM calls check_applied_jobs tool
  ↓
check_applied_jobs(company="Google", job_title="Software Engineer")
  ↓
Reads from Google Sheets cache
  ↓
Returns: "Yes, you have already applied to Google for the Software Engineer position."
```

The agent can now:
- Answer questions about specific applications
- List jobs applied to at a company
- Report total application count

## Files Modified/Added

- **New**: `/modules/agent/src/agent/services/google_sheets.py` - Google Sheets integration
- **Modified**: `/modules/agent/src/agent/tools/worker_tools.py` - Added job_search_with_filter and check_applied_jobs tools
- **Modified**: `/modules/agent/src/agent/workers/worker.py` - Updated prompt to use job_search tool
- **Modified**: `/modules/agent/pyproject.toml` - Added gspread and google-auth dependencies

### Tools Available to Agent

1. **`job_search`** - Searches for jobs and automatically filters out applied positions
2. **`check_applied_jobs`** - Answers questions about what jobs have been applied to

## Optional: Automate Sheet Updates

In the future, you could add a tool for the agent to automatically update the Sheet when you apply to a job. For now, manually update the Sheet when you submit applications.
