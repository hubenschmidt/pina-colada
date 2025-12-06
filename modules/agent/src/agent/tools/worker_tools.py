import logging
import re
import os
import smtplib
import traceback
from email.mime.text import MIMEText
from email.mime.multipart import MIMEMultipart
from typing import Dict, Optional
from langchain_core.tools import Tool, StructuredTool
from langchain_community.tools.wikipedia.tool import WikipediaQueryRun
from langchain_community.utilities import GoogleSerperAPIWrapper
from langchain_community.utilities.wikipedia import WikipediaAPIWrapper
from agent.tools.static_tools import push
from dotenv import load_dotenv
from services.job_service import (
    get_jobs_details,
    fetch_applied_jobs,
    is_job_applied,
    filter_jobs,
    add_job,
    get_all_jobs,
    update_job_by_company,
)
from pydantic import BaseModel, Field

load_dotenv(override=True)
logger = logging.getLogger(__name__)


# Pydantic models for tool inputs
class SendEmailInput(BaseModel):
    to_email: str = Field(description="Recipient email address")
    subject: str = Field(description="Email subject line")
    body: str = Field(description="Email body text")
    job_listings: str = Field(default="", description="Optional job listings to include in email")


class UpdateJobStatusInput(BaseModel):
    company: str = Field(description="Company name (fuzzy matching supported)")
    job_title: str = Field(description="Job title")
    status: str = Field(description="Status to set (applied, interviewing, rejected, offer, accepted, do_not_apply)")
    job_url: str = Field(default="", description="Optional URL for the job posting")
    notes: str = Field(default="", description="Optional notes about the application")


class JobCheckInput(BaseModel):
    query: str = Field(default="", description="Search query (company name, job title, or both). Empty string lists all jobs.")


def get_applied_jobs_tracker():
    """Get job service tracker (functional dict interface)."""
    return {
        "get_jobs_details": lambda refresh=False: get_jobs_details(refresh=refresh),
        "fetch_applied_jobs": lambda refresh=False: fetch_applied_jobs(refresh=refresh),
        "is_job_applied": lambda company, title: is_job_applied(company, title),
        "filter_jobs": lambda jobs: filter_jobs(jobs),
        "add_applied_job": lambda **kwargs: add_job(**kwargs),
    }


def _get_smtp_config():
    """Get SMTP configuration from environment variables."""
    return {
        "host": os.getenv("SMTP_HOST", "smtp.gmail.com"),
        "port": int(os.getenv("SMTP_PORT", "587")),
        "username": os.getenv("SMTP_USERNAME"),
        "password": os.getenv("SMTP_PASSWORD"),
        "from_email": os.getenv("SMTP_FROM_EMAIL", os.getenv("SMTP_USERNAME")),
    }


def _format_line_with_url(line: str) -> str:
    """Format a single line that contains a URL."""
    url_match = re.search(r"https?://[^\s]+", line)
    if not url_match:
        return f"{line.strip()}\n"

    url = url_match.group(0)
    parts = line.split(url)[0].strip()
    if not parts:
        return f"{url}\n"

    return f"{parts}\n{url}\n"


def _format_job_listing_email(job_listings: str) -> str:
    """Format job listings for email."""
    if not job_listings:
        return "No job listings to send."

    lines = [line.strip() for line in job_listings.split("\n") if line.strip()]
    formatted_lines = [_format_line_with_url(line) for line in lines]

    return "".join(formatted_lines)


def _build_email_body(body: str, job_listings: str) -> str:
    """Build complete email body with optional job listings."""
    if not job_listings:
        return body

    formatted_listings = _format_job_listing_email(job_listings)
    return f"{body}\n\n{formatted_listings}"


def _create_email_message(
    to_email: str, subject: str, body: str, from_email: str
) -> MIMEMultipart:
    """Create email message object."""
    msg = MIMEMultipart()
    msg["From"] = from_email
    msg["To"] = to_email
    msg["Subject"] = subject
    msg.attach(MIMEText(body, "plain"))
    return msg


def _send_via_smtp(config: dict, msg: MIMEMultipart) -> None:
    """Send email via SMTP."""
    with smtplib.SMTP(config["host"], config["port"]) as server:
        server.starttls()
        server.login(config["username"], config["password"])
        server.send_message(msg)


def send_email(to_email: str, subject: str, body: str, job_listings: str = "") -> str:
    """Send an email to the specified address."""
    config = _get_smtp_config()

    if not config["username"] or not config["password"]:
        logger.warning("SMTP credentials not configured")
        return "Email not configured. SMTP_USERNAME and SMTP_PASSWORD environment variables are required."

    if not to_email:
        return "Recipient email address is required."

    try:
        email_body = _build_email_body(body, job_listings)
        msg = _create_email_message(to_email, subject, email_body, config["from_email"])
        _send_via_smtp(config, msg)

        logger.info(f"Email sent successfully to {to_email}")
        return f"Email sent successfully to {to_email}"

    except Exception as e:
        logger.error(f"Failed to send email: {e}")

        logger.error(traceback.format_exc())
        return f"Failed to send email: {e}"


def record_user_details(
    email: str, name: str = "Name not provided", notes: str = "not provided"
):
    """Record user contact details"""
    message = f"Recording interest from {name} with email {email} and notes {notes}"
    push(message)
    logger.info(message)
    return {"recorded": "ok", "email": email, "name": name}


def serper_search(query: str) -> str:
    try:
        serper = GoogleSerperAPIWrapper()
        return serper.run(query)
    except Exception as e:
        logger.error(f"Serper search failed: {e}")
        return f"Web search failed: {e}"


def _format_job_with_date(job: dict) -> str:
    """Format job entry with company, title, date, and direct link."""
    company = job.get("company", "")
    title = job.get("title", "")
    date = job.get("date_applied", "")
    link = job.get("link", "")

    result = f"{company} - {title}"

    if date and date != "Not specified":
        result += f" (Applied: {date})"

    if link:
        result += f" - {link}"

    return result


def _format_job_title_only(job: dict) -> str:
    """Format job entry with title, date, and direct link."""
    title = job.get("title", "")
    date = job.get("date_applied", "")
    link = job.get("link", "")

    result = title

    if date and date != "Not specified":
        result += f" (Applied: {date})"

    if link:
        result += f" - {link}"

    return result


def _format_job_simple(job: dict) -> str:
    """Format job entry with company, title, and direct link."""
    company = job.get("company", "")
    title = job.get("title", "")
    link = job.get("link", "")

    result = f"{company} - {title}"

    if link:
        result += f" - {link}"

    return result


def _get_sheet_info(tracker) -> str:
    """Get data source info string."""
    return " (from Supabase)"


def _handle_no_applications(tracker) -> str:
    """Handle case when no applications found."""
    return "No job applications have been tracked yet in the Supabase database. The Job table may be empty."


def _handle_specific_job_check(tracker, company: str, job_title: str) -> str:
    """Handle check for specific company and job title."""
    if tracker["is_job_applied"](company, job_title):
        return (
            f"Yes, you have already applied to {company} for the {job_title} position."
        )
    return f"No record found for {company} - {job_title}. You have not applied to this position yet."


def _handle_company_filter(tracker, company: str, jobs_details: list) -> str:
    """Handle filtering by company."""
    company_lower = company.lower().strip()
    matching_jobs = [
        job for job in jobs_details if company_lower in job.get("company", "").lower()
    ]

    if not matching_jobs:
        return f"No applications found for {company}."

    job_list = "\n".join(_format_job_title_only(job) for job in matching_jobs)
    return f"You have applied to {len(matching_jobs)} position(s) at {company}:\n{job_list}"


def _fuzzy_match_title(search_term: str, job_title: str) -> bool:
    """Check if search term matches job title using fuzzy matching."""
    search_lower = search_term.lower().strip()
    title_lower = job_title.lower().strip()

    # Exact substring match
    if search_lower in title_lower:
        return True

    # Word-based matching: check if all words in search term appear in title
    search_words = [w.strip() for w in search_lower.split() if w.strip()]
    if not search_words:
        return False

    title_words = [w.strip() for w in title_lower.split() if w.strip()]

    # Check if all search words appear in title (in any order)
    for word in search_words:
        if not any(word in tw or tw in word for tw in title_words):
            return False

    return True


def _handle_job_title_filter(tracker, job_title: str, jobs_details: list) -> str:
    """Handle filtering by job title with fuzzy matching."""
    title_lower = job_title.lower().strip()
    logger.info(
        f"Searching for jobs with title containing '{title_lower}' in {len(jobs_details)} total jobs"
    )

    # Log sample titles for debugging
    if jobs_details:
        sample_titles = [job.get("title", "")[:50] for job in jobs_details[:10]]
        logger.info(f"Sample job titles: {sample_titles}")

    # Use fuzzy matching
    matching_jobs = [
        job
        for job in jobs_details
        if _fuzzy_match_title(title_lower, job.get("title", ""))
    ]
    logger.info(f"Found {len(matching_jobs)} matching jobs using fuzzy matching")

    # Log some matching titles for debugging
    if matching_jobs:
        matching_titles = [job.get("title", "") for job in matching_jobs[:10]]
        logger.info(f"Sample matching titles: {matching_titles}")

    if not matching_jobs:
        return f"No applications found for job title '{job_title}'."

    job_list = "\n".join(_format_job_with_date(job) for job in matching_jobs[:50])

    if len(matching_jobs) > 50:
        return f"You have applied to {len(matching_jobs)} position(s) with title containing '{job_title}':\n{job_list}\n... and {len(matching_jobs) - 50} more"

    return f"You have applied to {len(matching_jobs)} position(s) with title containing '{job_title}':\n{job_list}"


def _handle_list_all(tracker, applied_jobs: set, jobs_details: list) -> str:
    """Handle listing all jobs."""
    sheet_info = _get_sheet_info(tracker)

    if len(jobs_details) <= 20:
        job_list = "\n".join(_format_job_simple(job) for job in jobs_details)
        return (
            f"You have applied to {len(applied_jobs)} job(s){sheet_info}:\n{job_list}"
        )

    job_list = "\n".join(_format_job_simple(job) for job in jobs_details[:20])
    return f"You have applied to {len(applied_jobs)} jobs total{sheet_info}. Here are the first 20:\n{job_list}\n\nAsk me about a specific company or job title to see more."


def get_data_source_info() -> str:
    """Get information about the connected data source."""
    try:
        return "Using Supabase database for job application tracking. Connected to Job table."
    except Exception as e:
        logger.error(f"Failed to get data source info: {e}")
        return f"Unable to get data source information: {e}"


def _try_extract_title(query_lower: str, keyword: str) -> Optional[str]:
    """Try to extract a title after a specific keyword."""
    if keyword not in query_lower:
        return None
    parts = query_lower.split(keyword)
    if len(parts) <= 1:
        return None
    potential_title = parts[-1].strip().strip('"').strip("'").strip()
    return potential_title or None


def _extract_title_after_keyword(query_lower: str, keywords: list) -> Optional[str]:
    """Extract job title that follows a keyword."""
    for kw in keywords:
        title = _try_extract_title(query_lower, kw)
        if title:
            return title
    return None


def _parse_tool_input(query: str) -> tuple:
    """Parse tool input string into company and job_title parameters."""
    if not query:
        return "", ""

    query_lower = query.lower().strip()
    title_keywords = ["title", "role", "position", "job"]
    company_keywords = ["company", "at", "for"]

    # Try to extract title after keywords
    if any(kw in query_lower for kw in title_keywords):
        extracted = _extract_title_after_keyword(query_lower, title_keywords)
        if extracted:
            return "", extracted

    # Short phrase without company keywords -> assume job title
    is_short = len(query.split()) <= 5
    has_company_keyword = any(kw in query_lower for kw in company_keywords)
    if is_short and not has_company_keyword:
        return "", query.strip()

    return "", query.strip()


def _parse_query_input(query) -> tuple:
    """Parse query input into (company, job_title) tuple."""
    if isinstance(query, dict):
        return query.get("company", ""), query.get("job_title", "")
    if isinstance(query, str) and query:
        return _parse_tool_input(query)
    return "", ""


def _route_job_check(tracker, company: str, job_title: str, applied_jobs: list, jobs_details: list) -> str:
    """Route to appropriate handler based on query parameters."""
    if not applied_jobs:
        return _handle_no_applications(tracker)
    if company and job_title:
        return _handle_specific_job_check(tracker, company, job_title)
    if company:
        return _handle_company_filter(tracker, company, jobs_details)
    if job_title:
        return _handle_job_title_filter(tracker, job_title, jobs_details)
    return _handle_list_all(tracker, applied_jobs, jobs_details)


async def check_applied_jobs(query: str = "") -> str:
    """
    Check if a specific job has been applied to, or list all applied jobs.

    Args:
        query: Search query string (optional) - can be job title, company, or both

    Returns:
        Status message about applied jobs
    """
    try:
        company, job_title = _parse_query_input(query)
        logger.info(f"check_applied_jobs: query='{query}', company='{company}', job_title='{job_title}'")

        applied_jobs = await fetch_applied_jobs(refresh=True)
        jobs_details = await get_jobs_details(refresh=True)
        logger.info(f"Applied: {len(applied_jobs)}, Details: {len(jobs_details)}")

        tracker = get_applied_jobs_tracker()
        return _route_job_check(tracker, company, job_title, applied_jobs, jobs_details)

    except Exception as e:
        logger.error(f"Failed to check applied jobs: {e}")
        logger.error(traceback.format_exc())
        return f"Unable to check applied jobs: {e}. Please check that the Supabase integration is configured correctly."


_JOB_BOARD_EXCLUSIONS = [
    "linkedin.com", "indeed.com", "glassdoor.com", "ziprecruiter.com",
    "monster.com", "simplyhired.com", "careerbuilder.com", "dice.com",
]


def _enhance_job_query(query: str) -> str:
    """Enhance query to find direct company career pages."""
    # Add exclusions for major job boards
    exclusions = " ".join(f"-site:{site}" for site in _JOB_BOARD_EXCLUSIONS)
    # Focus on careers/jobs pages at company sites
    return f'{query} careers OR jobs site:*.com {exclusions}'


def _format_serper_results(results: dict) -> str:
    """Format Serper structured results into readable job listings."""
    organic = results.get("organic", [])
    if not organic:
        return "No job listings found for this search."

    lines = []
    for i, item in enumerate(organic[:10], 1):  # Top 10 results
        title = item.get("title", "Unknown")
        link = item.get("link", "")
        snippet = item.get("snippet", "")

        # Try to extract company name from title or snippet
        # Common pattern: "Job Title at Company" or "Company - Job Title"
        company = "Unknown Company"
        if " at " in title:
            parts = title.split(" at ")
            if len(parts) >= 2:
                company = parts[-1].strip()
                title = parts[0].strip()
        elif " - " in title:
            parts = title.split(" - ")
            if len(parts) >= 2:
                # Could be "Company - Title" or "Title - Company"
                company = parts[0].strip()

        lines.append(f"{i}. {company} - {title}")
        if link:
            lines.append(f"   URL: {link}")
        if snippet:
            lines.append(f"   {snippet[:150]}...")
        lines.append("")

    return "\n".join(lines)


async def job_search_with_filter(query: str) -> str:
    """
    Search for jobs and filter out ones already applied to or marked as 'do not apply'.

    This function:
    1. Performs a web search for jobs (excluding job boards)
    2. Fetches all jobs from the database
    3. Filters out duplicates by company and title (for jobs with status='applied' or 'do_not_apply')
    4. Returns the filtered results
    """
    try:
        # Get job service to check applied jobs
        tracker = get_applied_jobs_tracker()

        # Get all jobs (not just status='applied') to include 'do_not_apply' in filtering
        all_jobs = await get_all_jobs(refresh=True)

        logger.info(
            f"Loaded {len(all_jobs)} jobs from database (will filter duplicates)"
        )

        # Perform the search with query enhancements
        enhanced_query = _enhance_job_query(query)
        logger.info(f"Job search query: {enhanced_query}")

        # Use Serper with organic results - returns structured data
        serper = GoogleSerperAPIWrapper(type="search")
        raw_results = serper.results(enhanced_query)

        # Format results as structured list with URLs
        formatted = _format_serper_results(raw_results)
        logger.info(f"Formatted {len(raw_results.get('organic', []))} search results")

        if not all_jobs:
            logger.info("No jobs found in database - returning unfiltered results")
            return formatted

        # Parse and filter results (filter jobs with status='applied' or 'do_not_apply')
        filtered_results = _filter_job_results(formatted, tracker, all_jobs)

        if not filtered_results.strip():
            return "All job postings found have already been applied to or marked as 'do not apply'. Try refining your search criteria or location."

        return filtered_results

    except Exception as e:
        logger.error(f"Job search with filter failed: {e}")
        logger.error(traceback.format_exc())
        return f"Job search failed: {e}"


def _normalize_for_matching(text: str) -> str:
    """Normalize text for fuzzy matching."""
    if not text:
        return ""
    return text.lower().strip()


def _company_matches(search_company: str, applied_company: str) -> bool:
    """Check if company names match (strict)."""
    if not search_company or not applied_company:
        return False

    search_norm = _normalize_for_matching(search_company)
    applied_norm = _normalize_for_matching(applied_company)

    # Exact match
    if search_norm == applied_norm:
        return True

    # Remove common suffixes/prefixes for comparison
    search_clean = (
        search_norm.replace(" inc", "")
        .replace(" llc", "")
        .replace(" corp", "")
        .replace(" ltd", "")
        .strip()
    )
    applied_clean = (
        applied_norm.replace(" inc", "")
        .replace(" llc", "")
        .replace(" corp", "")
        .replace(" ltd", "")
        .strip()
    )

    if search_clean == applied_clean:
        return True

    # One is substring of the other (but require at least 4 chars to avoid false matches)
    if len(search_clean) >= 4 and len(applied_clean) >= 4:
        if search_clean in applied_clean or applied_clean in search_clean:
            return True

    return False


def _title_matches(search_title: str, applied_title: str) -> bool:
    """Check if job titles match (strict - requires exact or very close match)."""
    if not search_title or not applied_title:
        return False

    search_norm = _normalize_for_matching(search_title)
    applied_norm = _normalize_for_matching(applied_title)

    # Exact match
    if search_norm == applied_norm:
        return True

    # One contains the other (for variations like "Senior Software Engineer" vs "Software Engineer")
    if search_norm in applied_norm or applied_norm in search_norm:
        return True

    # Check if core title words match (ignore modifiers like "Senior", "Lead", etc.)
    search_words = [
        w
        for w in search_norm.split()
        if w not in ["senior", "lead", "staff", "principal", "junior", "mid", "level"]
    ]
    applied_words = [
        w
        for w in applied_norm.split()
        if w not in ["senior", "lead", "staff", "principal", "junior", "mid", "level"]
    ]

    # Require at least 2 core words to match
    if len(search_words) >= 2 and len(applied_words) >= 2:
        matching_words = [w for w in search_words if w in applied_words]
        if len(matching_words) >= 2:
            return True

    return False


def _is_tracked_job(job: Dict[str, str]) -> bool:
    """Check if job exists in database (any status means we've seen it before)."""
    has_required_fields = bool(job.get("company")) and bool(job.get("title"))
    return has_required_fields


def _job_matches_criteria(job: Dict[str, str], company: str, title: str) -> bool:
    """Check if a single job matches the search criteria."""
    company_match = _company_matches(company, job.get("company", ""))
    title_match = _title_matches(title, job.get("title", ""))
    if company_match and title_match:
        logger.info(
            f"Filtering out job: {company} - {title} "
            f"(matches {job.get('company')} - {job.get('title')}, status={job.get('status')})"
        )
        return True
    return False


def _matches_tracked_job(company: str, title: str, jobs_details: list) -> bool:
    """Check if a job matches any job in the database (filter duplicates)."""
    # Require both company AND title for strict matching
    if not company or not title:
        logger.debug(f"Skipping match check - company={bool(company)}, title={bool(title)}")
        return False

    tracked_jobs = [j for j in jobs_details if _is_tracked_job(j)]
    return any(_job_matches_criteria(j, company, title) for j in tracked_jobs)


def _extract_company_from_line(line: str) -> str:
    """Extract company name from a line of text."""
    company_patterns = [
        r"(?:Company|At|@):\s*([A-Z][A-Za-z0-9\s&.,-]+?)(?:\s*[-–]|\s*$|\s*\n)",
        r"^([A-Z][A-Za-z0-9\s&.,-]+?)\s*[-–]\s*[A-Z]",
        r"\b([A-Z][A-Za-z0-9\s&.,-]{2,})\s+(?:is|hiring|seeking)",
    ]

    for pattern in company_patterns:
        match = re.search(pattern, line, re.IGNORECASE)
        if match:
            potential_company = match.group(1).strip()
            if potential_company.lower() not in [
                "the",
                "a",
                "an",
                "for",
                "with",
                "and",
                "or",
            ]:
                return potential_company

    return ""


def _extract_title_from_line(line: str) -> str:
    """Extract job title from a line of text."""
    title_patterns = [
        r"(?:Title|Position|Role):\s*([A-Z][A-Za-z\s/()-]+(?:Engineer|Developer|Manager|Analyst|Designer|Specialist|Lead|Senior|Junior|Architect|Scientist))",
        r"[-–]\s*([A-Z][A-Za-z\s/()-]+(?:Engineer|Developer|Manager|Analyst|Designer|Specialist|Lead|Senior|Junior|Architect|Scientist))",
        r"(?:looking for|seeking|hiring)\s+([A-Z][A-Za-z\s/()-]+(?:Engineer|Developer|Manager|Analyst|Designer|Specialist|Lead|Senior|Junior|Architect|Scientist))",
    ]

    for pattern in title_patterns:
        match = re.search(pattern, line, re.IGNORECASE)
        if match:
            return match.group(1).strip()

    return ""


def _should_skip_line(
    line: str, current_company: str, current_title: str, jobs_details: list
) -> tuple[bool, str, str]:
    """Determine if line should be skipped and update current job info."""
    stripped = line.strip()
    if not stripped:
        return False, current_company, current_title

    company = _extract_company_from_line(line)
    title = _extract_title_from_line(line)

    new_company = company if company else current_company
    new_title = title if title else current_title

    if new_company and new_title:
        if _matches_tracked_job(new_company, new_title, jobs_details):
            return True, "", ""

    return False, new_company, new_title


def _process_line_for_filtering(
    line: str, current_company: str, current_title: str, jobs_details: list
) -> tuple[bool, str, str]:
    """Process a line and determine if it should be included."""
    skip, new_company, new_title = _should_skip_line(
        line, current_company, current_title, jobs_details
    )

    if skip:
        return False, "", ""

    return True, new_company, new_title


def _filter_job_results(raw_results: str, tracker, jobs_details: list) -> str:
    """Filter job search results to remove already-applied positions."""
    if not jobs_details:
        logger.warning("No applied jobs details available for filtering")
        return raw_results

    lines = raw_results.split("\n")
    filtered_lines = []
    current_company = ""
    current_title = ""

    for line in lines:
        include, new_company, new_title = _process_line_for_filtering(
            line, current_company, current_title, jobs_details
        )

        current_company = new_company
        current_title = new_title

        if include:
            filtered_lines.append(line)

    return "\n".join(filtered_lines)


VALID_JOB_STATUSES = ["applied", "interviewing", "rejected", "offer", "accepted", "do_not_apply"]


def _try_update_existing_job(company: str, job_title: str, status: str, job_url: str, notes: str) -> Optional[str]:
    """Try to update an existing job. Returns success message or None if not found."""
    updated = update_job_by_company(
        company=company,
        job_title=job_title,
        status=status,
        lead_status_id=None,
        job_url=job_url or None,
        notes=notes or None,
    )
    if updated:
        logger.info(f"Updated existing job: {company} - {job_title} to status {status}")
        return f"Successfully marked {company} - {job_title} as {status}"
    return None


def _create_new_job(company: str, job_title: str, status: str, job_url: str, notes: str) -> str:
    """Create a new job entry. Returns success/failure message."""
    logger.info(f"Creating new job: {company} - {job_title} with status {status}")
    new_job = add_job(
        company=company,
        job_title=job_title,
        status=status,
        job_url=job_url or "",
        notes=notes or "",
        source="agent",
    )
    if not new_job:
        return f"Failed to create job entry for {company} - {job_title}"
    return f"Successfully marked {company} - {job_title} as {status}"


def update_job_status(
    company: str, job_title: str, status: str, job_url: str = "", notes: str = ""
) -> str:
    """
    Update the status of a job application.

    Args:
        company: The company name (fuzzy matching supported)
        job_title: The job title
        status: The status to set (applied, interviewing, rejected, offer, accepted, do_not_apply)
        job_url: Optional URL for the job posting
        notes: Optional notes about the application
    """
    status_lower = status.lower()
    if status_lower not in VALID_JOB_STATUSES:
        return f"Invalid status '{status}'. Valid: {', '.join(VALID_JOB_STATUSES)}"

    try:
        result = _try_update_existing_job(company, job_title, status_lower, job_url, notes)
        if result:
            return result
        return _create_new_job(company, job_title, status_lower, job_url, notes)
    except Exception as e:
        logger.error(f"Failed to update job status: {e}")
        logger.error(traceback.format_exc())
        return f"Failed to update job status: {e}"


async def get_general_tools():
    """Return general-purpose tools (web search, wikipedia) - NO job search tools."""
    tools = []

    # Web search tool
    web_search_tool = Tool(
        name="web_search",
        func=serper_search,
        description="Search the web for current information (news, documentation, general queries). Input: a search query.",
    )
    tools.append(web_search_tool)

    # Wikipedia tool
    try:
        wikipedia = WikipediaAPIWrapper()
        wiki_tool = WikipediaQueryRun(api_wrapper=wikipedia)
        tools.append(wiki_tool)
    except Exception as e:
        logger.warning(f"Could not initialize Wikipedia tool: {e}")

    logger.info(f"Initialized {len(tools)} general tools")
    return tools


async def get_job_search_tools():
    """Return job search specific tools - only for job_search worker."""
    tools = []

    # Job search tool with filtering (async)
    job_search_tool = Tool(
        name="job_search",
        func=lambda q: None,  # Placeholder for sync
        coroutine=job_search_with_filter,
        description="Search for job postings and automatically filter out positions already applied to (tracked in Supabase). Input: a job search query (e.g., 'software engineer jobs in NYC').",
    )
    tools.append(job_search_tool)

    # Check applied jobs tool (async)
    check_applied_tool = StructuredTool.from_function(
        func=lambda **kwargs: None,
        coroutine=check_applied_jobs,
        name="check_applied_jobs",
        description="Check job applications tracked in Supabase database. Input: a query string. Examples: '' (empty string) for all jobs, 'Software Engineer' to find jobs with that title, 'Google' to find jobs at that company.",
        args_schema=JobCheckInput,
    )
    tools.append(check_applied_tool)

    # Get data source info tool
    data_source_info_tool = Tool(
        name="get_data_source_info",
        func=get_data_source_info,
        description="Get information about the connected data source (Supabase).",
    )
    tools.append(data_source_info_tool)

    # Update job status tool
    update_job_status_tool = StructuredTool.from_function(
        func=update_job_status,
        name="update_job_status",
        description="Update the status of a job application. Use when user says they applied to a company or wants to mark a job as 'do not apply'.",
        args_schema=UpdateJobStatusInput,
    )
    tools.append(update_job_status_tool)

    # Email tool for sending job listings
    send_email_tool = StructuredTool.from_function(
        func=send_email,
        name="send_email",
        description=(
            "Send an email to one or more recipients. You CAN send emails - use this tool! "
            "Parameters: to_email (recipient address), subject (email subject line), "
            "body (main email text), job_listings (optional formatted job list to append)."
        ),
        args_schema=SendEmailInput,
    )
    tools.append(send_email_tool)

    logger.info(f"Initialized {len(tools)} job search tools")
    return tools


async def get_worker_tools():
    """Return all tools (for backwards compatibility) - DEPRECATED, use specific getters."""
    general = await get_general_tools()
    job_search = await get_job_search_tools()
    return general + job_search
