import logging
import re
from langchain_core.tools import Tool
from langchain_community.agent_toolkits import FileManagementToolkit
from langchain_community.tools.wikipedia.tool import WikipediaQueryRun
from langchain_community.utilities import GoogleSerperAPIWrapper
from langchain_community.utilities.wikipedia import WikipediaAPIWrapper
from agent.tools.static_tools import push
from agent.services.google_sheets import get_applied_jobs_tracker

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
    """Get sheet name info string."""
    sheet_name = tracker.get_sheet_name()
    if not sheet_name:
        return ""
    return f" (from Google Sheet: {sheet_name})"


def _handle_no_applications(tracker) -> str:
    """Handle case when no applications found."""
    sheet_name = tracker.get_sheet_name()
    if sheet_name:
        return f"No job applications have been tracked yet in the Google Sheet '{sheet_name}'. The sheet may be empty or the data hasn't been loaded."
    return "No job applications have been tracked yet. The Google Sheets integration may not be configured correctly."


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


def get_google_sheet_info() -> str:
    """Get information about the connected Google Sheet."""
    try:
        tracker = get_applied_jobs_tracker()
        sheet_name = tracker.get_sheet_name()
        if not sheet_name:
            return "Google Sheet name is not available. The integration may not be configured correctly."
        return f"The connected Google Sheet is named: '{sheet_name}'"
    except Exception as e:
        logger.error(f"Failed to get sheet info: {e}")
        return f"Unable to get Google Sheet information: {e}"


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
            else:
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
        return f"Unable to check applied jobs: {e}. Please check that the Google Sheets integration is configured correctly."


def job_search_with_filter(query: str) -> str:
    """
    Search for jobs and filter out ones already applied to.

    This function:
    1. Performs a web search for jobs
    2. Fetches the list of already-applied jobs from Google Sheets
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


def _matches_applied_job(company: str, title: str, jobs_details: list) -> bool:
    """Check if a job matches any applied job using fuzzy matching."""
    if not company and not title:
        return False
    
    company_norm = _normalize_for_matching(company)
    title_norm = _normalize_for_matching(title)
    
    for applied_job in jobs_details:
        applied_company = _normalize_for_matching(applied_job.get("company", ""))
        applied_title = _normalize_for_matching(applied_job.get("title", ""))
        
        # Check if company matches (fuzzy)
        company_matches = False
        if company_norm and applied_company:
            # Exact match or substring match
            if company_norm == applied_company or company_norm in applied_company or applied_company in company_norm:
                company_matches = True
        
        # Check if title matches (fuzzy)
        title_matches = False
        if title_norm and applied_title:
            # Check if all words in search title appear in applied title
            search_words = [w for w in title_norm.split() if len(w) > 2]
            if search_words:
                title_matches = all(any(word in applied_title for word in search_words) or any(applied_word in title_norm for applied_word in applied_title.split()))
            else:
                title_matches = title_norm in applied_title or applied_title in title_norm
        
        # If we have both company and title, both must match
        # If we only have one, that one must match
        if company_norm and title_norm:
            if company_matches and title_matches:
                logger.info(f"Filtering out applied job: {company} - {title} (matches {applied_company} - {applied_title})")
                return True
        elif company_norm and company_matches:
            logger.info(f"Filtering out applied job: {company} (matches {applied_company})")
            return True
        elif title_norm and title_matches:
            logger.info(f"Filtering out applied job: {title} (matches {applied_title})")
            return True
    
    return False


def _filter_job_results(raw_results: str, tracker, jobs_details: list) -> str:
    """
    Filter job search results to remove already-applied positions.

    Args:
        raw_results: Raw search results from Serper
        tracker: AppliedJobsTracker instance
        jobs_details: List of applied job details for matching

    Returns:
        Filtered results text
    """
    if not jobs_details:
        logger.warning("No applied jobs details available for filtering")
        return raw_results
    
    # Split results into lines
    lines = raw_results.split('\n')
    filtered_lines = []
    current_job_company = ""
    current_job_title = ""
    skip_current = False

    for line in lines:
        stripped = line.strip()
        if not stripped:
            if not skip_current:
                filtered_lines.append(line)
            continue

        # Detect company name patterns (more flexible)
        # Examples: "Company: Google", "Google - Software Engineer", "at Google", "Google Inc"
        company_patterns = [
            r'(?:Company|At|@):\s*([A-Z][A-Za-z0-9\s&.,-]+?)(?:\s*[-–]|\s*$|\s*\n)',
            r'^([A-Z][A-Za-z0-9\s&.,-]+?)\s*[-–]\s*[A-Z]',  # "Company - Title"
            r'\b([A-Z][A-Za-z0-9\s&.,-]{2,})\s+(?:is|hiring|seeking)',  # "Company is hiring"
        ]
        
        # Detect job title patterns (more flexible)
        title_patterns = [
            r'(?:Title|Position|Role):\s*([A-Z][A-Za-z\s/()-]+(?:Engineer|Developer|Manager|Analyst|Designer|Specialist|Lead|Senior|Junior|Architect|Scientist))',
            r'[-–]\s*([A-Z][A-Za-z\s/()-]+(?:Engineer|Developer|Manager|Analyst|Designer|Specialist|Lead|Senior|Junior|Architect|Scientist))',
            r'(?:looking for|seeking|hiring)\s+([A-Z][A-Za-z\s/()-]+(?:Engineer|Developer|Manager|Analyst|Designer|Specialist|Lead|Senior|Junior|Architect|Scientist))',
        ]

        # Try to extract company
        for pattern in company_patterns:
            match = re.search(pattern, line, re.IGNORECASE)
            if match:
                potential_company = match.group(1).strip()
                # Filter out common false positives
                if potential_company.lower() not in ['the', 'a', 'an', 'for', 'with', 'and', 'or']:
                    current_job_company = potential_company
                    break

        # Try to extract title
        for pattern in title_patterns:
            match = re.search(pattern, line, re.IGNORECASE)
            if match:
                current_job_title = match.group(1).strip()
                break

        # Check if we have enough info to match
        if current_job_company or current_job_title:
            if _matches_applied_job(current_job_company, current_job_title, jobs_details):
                skip_current = True
                # Reset for next job
                current_job_company = ""
                current_job_title = ""
                continue
            else:
                skip_current = False

        # Add line if we're not skipping current job
        if not skip_current:
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
        description="Search for job postings and automatically filter out positions already applied to (tracked in Google Sheets). Use this instead of web_search for job-related queries. Input: a job search query (e.g., 'software engineer jobs in NYC').",
    )
    tools.append(job_search_tool)

    # Check applied jobs tool
    check_applied_tool = Tool(
        name="check_applied_jobs",
        func=check_applied_jobs,
        description="Check job applications tracked in Google Sheets. ALWAYS use this tool when user asks about job applications. Input: a query string. Examples: '' (empty string) for all jobs, 'Software Engineer' to find jobs with that title, 'Google' to find jobs at that company, 'Google Software Engineer' for specific company+title. The tool performs fuzzy matching on job titles, so 'Software Engineer' will match 'Senior Software Engineer', 'Full Stack Software Engineer', etc. This tool reads directly from the connected Google Sheet.",
    )
    tools.append(check_applied_tool)
    
    # Get Google Sheet info tool
    sheet_info_tool = Tool(
        name="get_google_sheet_info",
        func=get_google_sheet_info,
        description="Get the name of the connected Google Sheet. Use when user asks 'what is the name of the google sheet?' or 'what google sheet are you using?'. Returns the sheet name.",
    )
    tools.append(sheet_info_tool)

    # Wikipedia tool
    try:
        wikipedia = WikipediaAPIWrapper()
        wiki_tool = WikipediaQueryRun(api_wrapper=wikipedia)
        tools.append(wiki_tool)
    except Exception as e:
        logger.warning(f"Could not initialize Wikipedia tool: {e}")

    logger.info(f"Initialized {len(tools)} tools")
    return tools
