from typing import List, Tuple

# Each entry: (keywords, criteria_text)
# First match wins, so order matters
_CRITERIA_MAP: List[Tuple[List[str], str]] = [
    # 1. CONTACT/EMAIL
    (
        ["email", "contact", "reach", "touch"],
        """Success criteria:
- Provide William's email address or contact method
- Response is brief and direct
- No unnecessary elaboration""",
    ),
    # 2. COVER LETTER
    (
        ["cover letter", "write a letter", "application letter"],
        """Success criteria:
- Cover letter is properly formatted (greeting, body paragraphs, closing)
- References specific job details from the posting
- Uses William's actual experience from resume data
- Matches professional tone from sample cover letters
- Between 250-400 words
- No markdown formatting (plain text only)""",
    ),
    # 3. JOB SEARCH
    (
        ["job search", "find jobs", "job postings", "job openings"],
        """Success criteria:
- Provides at least 3-5 specific job postings
- Each posting includes: Company name, Job title, and DIRECT URL to application page
- URLs are actual job posting links (not general career pages)
- Jobs match William's skills and experience level
- Jobs are recent (posted within last 30 days if possible)
- Jobs are in requested location (default: NYC)""",
    ),
    # 4. TELL ME ABOUT
    (
        ["tell me about"],
        """Success criteria:
- Directly answers what was asked about
- Provides specific details from resume data
- Structured response (3-5 sentences)
- Stays on topic
- Professional and conversational tone""",
    ),
    # 5. TECHNICAL/EXPERIENCE
    (
        ["experience", "skills", "worked with", "know about", "projects", "built"],
        """Success criteria:
- Answer directly addresses the specific skill/experience asked about
- Cites specific examples or projects from William's resume
- Response is concise (2-4 sentences typically)
- Accurate information from resume data only
- No invented or assumed information""",
    ),
    # 6. BACKGROUND/CAREER
    (
        ["background", "career", "education", "worked at", "previous", "history", "about you"],
        """Success criteria:
- Provides accurate information from resume/summary
- Organized and easy to understand
- Focuses on most relevant information
- 3-5 sentences maximum unless more detail explicitly requested
- Professional tone""",
    ),
    # 7. COMPARISON
    (
        [" vs ", " versus ", "compared to", "better at", "prefer", "difference between"],
        """Success criteria:
- Addresses both items being compared
- Provides specific examples or experience with each
- Clear statement of preference/strength if applicable
- Based on actual resume information
- Concise comparison (2-4 sentences)""",
    ),
    # 8. GREETING
    (
        ["hi", "hello", "hey", "greetings", "good morning", "good afternoon"],
        """Success criteria:
- Warm, professional greeting response
- Brief introduction of who William is
- Offers to help with questions
- 2-3 sentences maximum
- No repeated greeting if already established in conversation""",
    ),
    # 9. AVAILABILITY
    (
        ["available", "looking for work", "open to", "hiring", "can you start", "when can you"],
        """Success criteria:
- Clear statement about current availability status
- Information about ideal role type if applicable
- Brief and direct (1-3 sentences)
- Professional and positive tone""",
    ),
    # 10. SALARY
    (
        ["salary", "compensation", "pay", "rate", "wage"],
        """Success criteria:
- Professional deflection to discuss during interview process
- OR provide salary range if explicitly available in resume data
- Brief response (1-2 sentences)
- Maintains negotiating position""",
    ),
]

_DEFAULT_CRITERIA = """Success criteria:
- Directly answers the user's question
- Uses accurate information from resume data only
- Response is appropriately concise (typically 2-5 sentences)
- Professional and conversational tone
- Plain text format (no markdown)
- If information isn't in resume data, uses record_unknown_question tool"""


def _matches_job_search(msg: str) -> bool:
    """Check compound job search patterns."""
    return (
        ("find" in msg and "jobs" in msg)
        or ("search" in msg and "jobs" in msg)
        or ("looking for" in msg and "jobs" in msg)
    )


def get_success_criteria(message: str) -> str:
    """
    Generate context-aware success criteria based on the user's message.
    Returns the first matching criteria or a default.
    """
    msg = message.lower()

    # Check compound job search first
    if _matches_job_search(msg):
        return _CRITERIA_MAP[2][1]  # job search criteria

    for keywords, criteria in _CRITERIA_MAP:
        if any(kw in msg for kw in keywords):
            return criteria

    return _DEFAULT_CRITERIA
