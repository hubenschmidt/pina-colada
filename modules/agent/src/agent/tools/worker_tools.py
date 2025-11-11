import logging
import re
import os
import smtplib
from email.mime.text import MIMEText
from email.mime.multipart import MIMEMultipart
from typing import List, Dict
from langchain_core.tools import Tool, StructuredTool
from langchain_community.agent_toolkits import FileManagementToolkit
from langchain_community.tools.wikipedia.tool import WikipediaQueryRun
from langchain_community.utilities import GoogleSerperAPIWrapper
from langchain_community.utilities.wikipedia import WikipediaAPIWrapper
from agent.tools.static_tools import push
from dotenv import load_dotenv

# Feature flag for choosing data source
USE_SUPABASE = os.getenv("USE_SUPABASE", "true").lower() == "true"
USE_LOCAL_POSTGRES = os.getenv("USE_LOCAL_POSTGRES", "false").lower() == "true"

if USE_LOCAL_POSTGRES:
    from agent.services.job_service import JobService
    from agent.repositories.job_repository import JobRepository
    
    def get_applied_jobs_tracker():
        """Get job service instance for local Postgres."""
        repository = JobRepository()
        return JobService(repository)
else:
    from agent.services.supabase_client import get_applied_jobs_tracker
from pydantic import BaseModel, Field

load_dotenv(override=True)
logger = logging.getLogger(__name__)


def get_file_tools():
    """Get file management tools for the 'me' directory"""
    try:
        toolkit = FileManagementToolkit(
            root_dir="me"
        )  # me directory contains resume info
        return toolkit.get_tools()
    except Exception as e:
        logger.warning(f"Could not initialize file tools: {e}")
        return []


def record_unknown_question(question: str):
    """Record questions that couldn't be answered"""
    message = f"Recording question that couldn't be answered: {question}"
    push(message)
    logger.info(message)
    return {"recorded": "ok", "question": question}


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
    url_match = re.search(r'https?://[^\s]+', line)
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
    
    lines = [line.strip() for line in job_listings.split('\n') if line.strip()]
    formatted_lines = [_format_line_with_url(line) for line in lines]
    
    return "".join(formatted_lines)


def _build_email_body(body: str, job_listings: str) -> str:
    """Build complete email body with optional job listings."""
    if not job_listings:
        return body
    
    formatted_listings = _format_job_listing_email(job_listings)
    return f"{body}\n\n{formatted_listings}"


def _create_email_message(to_email: str, subject: str, body: str, from_email: str) -> MIMEMultipart:
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


def send_email(
    to_email: str,
    subject: str,
    body: str,
    job_listings: str = ""
) -> str:
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
        import traceback
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
    return "No job applications have been tracked yet in the Supabase database. The applied_jobs table may be empty."


def _handle_specific_job_check(tracker, company: str, job_title: str) -> str:
    """Handle check for specific company and job title."""
    if tracker.is_job_applied(company, job_title):
        return f"Yes, you have already applied to {company} for the {job_title} position."
    return f"No record found for {company} - {job_title}. You have not applied to this position yet."


def _handle_company_filter(tracker, company: str, jobs_details: list) -> str:
    """Handle filtering by company."""
    company_lower = company.lower().strip()
    matching_jobs = [
        job for job in jobs_details 
        if company_lower in job.get("company", "").lower()
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
    logger.info(f"Searching for jobs with title containing '{title_lower}' in {len(jobs_details)} total jobs")
    
    # Log sample titles for debugging
    if jobs_details:
        sample_titles = [job.get("title", "")[:50] for job in jobs_details[:10]]
        logger.info(f"Sample job titles: {sample_titles}")
    
    # Use fuzzy matching
    matching_jobs = [
        job for job in jobs_details 
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
        return f"You have applied to {len(applied_jobs)} job(s){sheet_info}:\n{job_list}"
    
    job_list = "\n".join(_format_job_simple(job) for job in jobs_details[:20])
    return f"You have applied to {len(applied_jobs)} jobs total{sheet_info}. Here are the first 20:\n{job_list}\n\nAsk me about a specific company or job title to see more."


def get_data_source_info() -> str:
    """Get information about the connected data source."""
    try:
        return "Using Supabase database for job application tracking. Connected to applied_jobs table."
    except Exception as e:
        logger.error(f"Failed to get data source info: {e}")
        return f"Unable to get data source information: {e}"


def _parse_tool_input(query: str) -> tuple:
    """Parse tool input string into company and job_title parameters."""
    # LangChain Tool passes a single string, so we need to parse it
    # Try to extract company and job_title from the query
    query_lower = query.lower().strip()
    
    # If it's empty, return empty strings
    if not query:
        return "", ""
    
    # Try to detect if it's asking for a specific job title
    # Common patterns: "Software Engineer", "jobs with Software Engineer", "include Software Engineer"
    title_keywords = ["title", "role", "position", "job"]
    company_keywords = ["company", "at", "for"]
    
    # Simple heuristic: if query contains title-related words, treat as job_title
    # Otherwise, check if it looks like a company name or job title
    if any(kw in query_lower for kw in title_keywords):
        # Extract the job title part
        for kw in title_keywords:
            if kw in query_lower:
                parts = query_lower.split(kw)
                if len(parts) > 1:
                    potential_title = parts[-1].strip().strip('"').strip("'").strip()
                    if potential_title:
                        return "", potential_title
    
    # If it's a short phrase without keywords, assume it's a job title
    if len(query.split()) <= 5 and not any(kw in query_lower for kw in company_keywords):
        return "", query.strip()
    
    # Default: treat as job_title
    return "", query.strip()


def check_applied_jobs(query: str = "") -> str:
    """
    Check if a specific job has been applied to, or list all applied jobs.
    
    This function accepts a query string that can contain:
    - A job title (e.g., "Software Engineer")
    - A company name (e.g., "Google")
    - Both (e.g., "Google Software Engineer")
    - Empty string to list all jobs

    Args:
        query: Search query string (optional)

    Returns:
        Status message about applied jobs
    """
    try:
        # Parse the input - handle both string and dict formats
        company = ""
        job_title = ""
        
        if isinstance(query, dict):
            company = query.get("company", "")
            job_title = query.get("job_title", "")
        elif isinstance(query, str):
            if not query:
                company, job_title = "", ""
            if query:
                # Try to parse the string
                company, job_title = _parse_tool_input(query)
        
        logger.info(f"check_applied_jobs called with query='{query}', parsed as company='{company}', job_title='{job_title}'")
        
        tracker = get_applied_jobs_tracker()
        applied_jobs = tracker.fetch_applied_jobs(refresh=True)
        jobs_details = tracker.get_jobs_details(refresh=True)
        
        logger.info(f"Total applied_jobs count: {len(applied_jobs)}, Total jobs_details count: {len(jobs_details)}")

        if not applied_jobs:
            return _handle_no_applications(tracker)

        if company and job_title:
            return _handle_specific_job_check(tracker, company, job_title)

        if company:
            return _handle_company_filter(tracker, company, jobs_details)

        if job_title:
            return _handle_job_title_filter(tracker, job_title, jobs_details)

        return _handle_list_all(tracker, applied_jobs, jobs_details)

    except Exception as e:
        logger.error(f"Failed to check applied jobs: {e}")
        import traceback
        logger.error(traceback.format_exc())
        return f"Unable to check applied jobs: {e}. Please check that the Supabase integration is configured correctly."


def job_search_with_filter(query: str) -> str:
    """
    Search for jobs and filter out ones already applied to.

    This function:
    1. Performs a web search for jobs
    2. Fetches the list of already-applied jobs from Supabase
    3. Filters out duplicates
    4. Returns the filtered results
    """
    try:
        # Get applied jobs tracker first and refresh
        tracker = get_applied_jobs_tracker()
        applied_jobs = tracker.fetch_applied_jobs(refresh=True)
        jobs_details = tracker.get_jobs_details(refresh=True)
        
        logger.info(f"Loaded {len(applied_jobs)} applied jobs for filtering")

        # Perform the search
        serper = GoogleSerperAPIWrapper()
        raw_results = serper.run(query)

        if not applied_jobs:
            logger.warning("No applied jobs loaded - returning unfiltered results")
            return raw_results

        # Parse and filter results
        filtered_results = _filter_job_results(raw_results, tracker, jobs_details)

        if not filtered_results.strip():
            return "All job postings found have already been applied to. Try refining your search criteria or location."

        return filtered_results

    except Exception as e:
        logger.error(f"Job search with filter failed: {e}")
        import traceback
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
    search_clean = search_norm.replace(" inc", "").replace(" llc", "").replace(" corp", "").replace(" ltd", "").strip()
    applied_clean = applied_norm.replace(" inc", "").replace(" llc", "").replace(" corp", "").replace(" ltd", "").strip()
    
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
    search_words = [w for w in search_norm.split() if w not in ["senior", "lead", "staff", "principal", "junior", "mid", "level"]]
    applied_words = [w for w in applied_norm.split() if w not in ["senior", "lead", "staff", "principal", "junior", "mid", "level"]]
    
    # Require at least 2 core words to match
    if len(search_words) >= 2 and len(applied_words) >= 2:
        matching_words = [w for w in search_words if w in applied_words]
        if len(matching_words) >= 2:
            return True
    
    return False


def _matches_applied_job(company: str, title: str, jobs_details: list) -> bool:
    """Check if a job matches any applied job (strict matching)."""
    if not company and not title:
        return False
    
    # Require both company AND title to match (strict)
    if not company or not title:
        logger.debug(f"Skipping match check - missing company or title: company={bool(company)}, title={bool(title)}")
        return False
    
    def _check_job_match(applied_job: Dict[str, str]) -> bool:
        applied_company = applied_job.get("company", "")
        applied_title = applied_job.get("title", "")
        
        if not applied_company or not applied_title:
            return False
        
        company_match = _company_matches(company, applied_company)
        title_match = _title_matches(title, applied_title)
        
        if company_match and title_match:
            logger.info(f"Filtering out applied job: {company} - {title} (matches {applied_company} - {applied_title})")
            return True
        
        return False
    
    for applied_job in jobs_details:
        if _check_job_match(applied_job):
            return True
    
    return False


def _extract_company_from_line(line: str) -> str:
    """Extract company name from a line of text."""
    company_patterns = [
        r'(?:Company|At|@):\s*([A-Z][A-Za-z0-9\s&.,-]+?)(?:\s*[-–]|\s*$|\s*\n)',
        r'^([A-Z][A-Za-z0-9\s&.,-]+?)\s*[-–]\s*[A-Z]',
        r'\b([A-Z][A-Za-z0-9\s&.,-]{2,})\s+(?:is|hiring|seeking)',
    ]
    
    for pattern in company_patterns:
        match = re.search(pattern, line, re.IGNORECASE)
        if match:
            potential_company = match.group(1).strip()
            if potential_company.lower() not in ['the', 'a', 'an', 'for', 'with', 'and', 'or']:
                return potential_company
    
    return ""


def _extract_title_from_line(line: str) -> str:
    """Extract job title from a line of text."""
    title_patterns = [
        r'(?:Title|Position|Role):\s*([A-Z][A-Za-z\s/()-]+(?:Engineer|Developer|Manager|Analyst|Designer|Specialist|Lead|Senior|Junior|Architect|Scientist))',
        r'[-–]\s*([A-Z][A-Za-z\s/()-]+(?:Engineer|Developer|Manager|Analyst|Designer|Specialist|Lead|Senior|Junior|Architect|Scientist))',
        r'(?:looking for|seeking|hiring)\s+([A-Z][A-Za-z\s/()-]+(?:Engineer|Developer|Manager|Analyst|Designer|Specialist|Lead|Senior|Junior|Architect|Scientist))',
    ]
    
    for pattern in title_patterns:
        match = re.search(pattern, line, re.IGNORECASE)
        if match:
            return match.group(1).strip()
    
    return ""


def _should_skip_line(line: str, current_company: str, current_title: str, jobs_details: list) -> tuple[bool, str, str]:
    """Determine if line should be skipped and update current job info."""
    stripped = line.strip()
    if not stripped:
        return False, current_company, current_title
    
    company = _extract_company_from_line(line)
    title = _extract_title_from_line(line)
    
    new_company = company if company else current_company
    new_title = title if title else current_title
    
    if new_company and new_title:
        if _matches_applied_job(new_company, new_title, jobs_details):
            return True, "", ""
    
    return False, new_company, new_title


def _process_line_for_filtering(line: str, current_company: str, current_title: str, jobs_details: list) -> tuple[bool, str, str]:
    """Process a line and determine if it should be included."""
    skip, new_company, new_title = _should_skip_line(line, current_company, current_title, jobs_details)
    
    if skip:
        return False, "", ""
    
    return True, new_company, new_title


def _filter_job_results(raw_results: str, tracker, jobs_details: list) -> str:
    """Filter job search results to remove already-applied positions."""
    if not jobs_details:
        logger.warning("No applied jobs details available for filtering")
        return raw_results
    
    lines = raw_results.split('\n')
    filtered_lines = []
    current_company = ""
    current_title = ""

    for line in lines:
        include, new_company, new_title = _process_line_for_filtering(line, current_company, current_title, jobs_details)
        
        current_company = new_company
        current_title = new_title
        
        if include:
            filtered_lines.append(line)

    return '\n'.join(filtered_lines)


async def get_worker_tools():
    """Return all available tools for the Worker"""
    tools = []

    # Push notification tool
    push_tool = Tool(
        name="send_push_notification",
        func=push,
        description="Use this tool when you want to send a push notification to the owner",
    )
    tools.append(push_tool)

    # User details recording tool
    record_details_tool = Tool(
        name="record_user_details",
        func=record_user_details,
        description="Record user contact information. Requires email, optional name and notes. Use after collecting contact info from the conversation.",
    )
    tools.append(record_details_tool)
    
    # Email sending tool (using StructuredTool for multiple parameters)
    class SendEmailInput(BaseModel):
        to_email: str = Field(description="Recipient email address")
        subject: str = Field(description="Email subject line")
        body: str = Field(description="Email body text")
        job_listings: str = Field(default="", description="Optional job listings to include in email")
    
    send_email_tool = StructuredTool.from_function(
        func=send_email,
        name="send_email",
        description="Send an email to a user. Use when user requests to receive job listings or other information via email. Always include the job listings in the job_listings parameter if sending job search results.",
        args_schema=SendEmailInput,
    )
    tools.append(send_email_tool)

    # Unknown question recording tool
    record_question_tool = Tool(
        name="record_unknown_question",
        func=record_unknown_question,
        description="Record a question that you couldn't answer. This alerts the owner that additional information may be needed.",
    )
    tools.append(record_question_tool)

    # File management tools
    file_tools = get_file_tools()
    tools.extend(file_tools)  # extend since get_file_tools contains CRUD ops

    # Web search tool
    web_search_tool = Tool(
        name="web_search",
        func=serper_search,
        description="Search the web for current information (news, documentation, general queries). Input: a search query.",
    )
    tools.append(web_search_tool)

    # Job search tool with filtering
    job_search_tool = Tool(
        name="job_search",
        func=job_search_with_filter,
        description="Search for job postings and automatically filter out positions already applied to (tracked in Supabase). Use this instead of web_search for job-related queries. Input: a job search query (e.g., 'software engineer jobs in NYC').",
    )
    tools.append(job_search_tool)

    # Check applied jobs tool
    check_applied_tool = Tool(
        name="check_applied_jobs",
        func=check_applied_jobs,
        description="Check job applications tracked in Supabase database. ALWAYS use this tool when user asks about job applications. Input: a query string. Examples: '' (empty string) for all jobs, 'Software Engineer' to find jobs with that title, 'Google' to find jobs at that company, 'Google Software Engineer' for specific company+title. The tool performs fuzzy matching on job titles, so 'Software Engineer' will match 'Senior Software Engineer', 'Full Stack Software Engineer', etc.",
    )
    tools.append(check_applied_tool)
    
    # Get data source info tool
    data_source_info_tool = Tool(
        name="get_data_source_info",
        func=get_data_source_info,
        description="Get information about the connected data source (Supabase). Use when user asks about the data source being used for job tracking.",
    )
    tools.append(data_source_info_tool)

    # Wikipedia tool
    try:
        wikipedia = WikipediaAPIWrapper()
        wiki_tool = WikipediaQueryRun(api_wrapper=wikipedia)
        tools.append(wiki_tool)
    except Exception as e:
        logger.warning(f"Could not initialize Wikipedia tool: {e}")

    logger.info(f"Initialized {len(tools)} tools")
    return tools
