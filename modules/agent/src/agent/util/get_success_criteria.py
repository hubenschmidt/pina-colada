def get_success_criteria(message: str) -> str:
    """
    Generate context-aware success criteria based on the user's message.
    This gives the evaluator clear, measurable criteria to check.
    """
    message_lower = message.lower()

    # 1. CONTACT/EMAIL QUERIES
    if any(word in message_lower for word in ["email", "contact", "reach", "touch"]):
        return """Success criteria:
- Provide William's email address or contact method
- Response is brief and direct
- No unnecessary elaboration"""

    # 2. COVER LETTER REQUESTS
    if any(
        phrase in message_lower
        for phrase in ["cover letter", "write a letter", "application letter"]
    ):
        return """Success criteria:
- Cover letter is properly formatted (greeting, body paragraphs, closing)
- References specific job details from the posting
- Uses William's actual experience from resume data
- Matches professional tone from sample cover letters
- Between 250-400 words
- No markdown formatting (plain text only)"""

    # 3. JOB SEARCH REQUESTS
    if (
        any(
            phrase in message_lower
            for phrase in ["job search", "find jobs", "job postings", "job openings"]
        )
        or ("find" in message_lower and "jobs" in message_lower)
        or ("search" in message_lower and "jobs" in message_lower)
        or ("looking for" in message_lower and "jobs" in message_lower)
    ):
        return """Success criteria:
- Provides at least 3-5 specific job postings
- Each posting includes: Company name, Job title, and DIRECT URL to application page
- URLs are actual job posting links (not general career pages)
- Jobs match William's skills and experience level
- Jobs are recent (posted within last 30 days if possible)
- Jobs are in requested location (default: NYC)"""

    # 4. "TELL ME ABOUT" QUESTIONS (check this before background questions)
    if "tell me about" in message_lower:
        return """Success criteria:
- Directly answers what was asked about
- Provides specific details from resume data
- Structured response (3-5 sentences)
- Stays on topic
- Professional and conversational tone"""

    # 5. TECHNICAL/EXPERIENCE QUESTIONS
    if any(
        word in message_lower
        for word in [
            "experience",
            "skills",
            "worked with",
            "know about",
            "projects",
            "built",
        ]
    ):
        return """Success criteria:
- Answer directly addresses the specific skill/experience asked about
- Cites specific examples or projects from William's resume
- Response is concise (2-4 sentences typically)
- Accurate information from resume data only
- No invented or assumed information"""

    # 5. TECHNICAL/EXPERIENCE QUESTIONS
    if any(
        word in message_lower
        for word in [
            "experience",
            "skills",
            "worked with",
            "know about",
            "projects",
            "built",
        ]
    ):
        return """Success criteria:
- Answer directly addresses the specific skill/experience asked about
- Cites specific examples or projects from William's resume
- Response is concise (2-4 sentences typically)
- Accurate information from resume data only
- No invented or assumed information"""

    # 6. BACKGROUND/CAREER QUESTIONS
    if any(
        word in message_lower
        for word in [
            "background",
            "career",
            "education",
            "worked at",
            "previous",
            "history",
            "about you",
        ]
    ):
        return """Success criteria:
- Provides accurate information from resume/summary
- Organized and easy to understand
- Focuses on most relevant information
- 3-5 sentences maximum unless more detail explicitly requested
- Professional tone"""

    # 7. COMPARISON QUESTIONS (X vs Y, better at, prefer)
    if any(
        word in message_lower
        for word in [
            " vs ",
            " versus ",
            "compared to",
            "better at",
            "prefer",
            "difference between",
        ]
    ):
        return """Success criteria:
- Addresses both items being compared
- Provides specific examples or experience with each
- Clear statement of preference/strength if applicable
- Based on actual resume information
- Concise comparison (2-4 sentences)"""

    # 8. GREETING/CASUAL
    if any(
        word in message_lower
        for word in [
            "hi",
            "hello",
            "hey",
            "greetings",
            "good morning",
            "good afternoon",
        ]
    ):
        return """Success criteria:
- Warm, professional greeting response
- Brief introduction of who William is
- Offers to help with questions
- 2-3 sentences maximum
- No repeated greeting if already established in conversation"""

    # 9. AVAILABILITY/HIRING QUESTIONS
    if any(
        phrase in message_lower
        for phrase in [
            "available",
            "looking for work",
            "open to",
            "hiring",
            "can you start",
            "when can you",
        ]
    ):
        return """Success criteria:
- Clear statement about current availability status
- Information about ideal role type if applicable
- Brief and direct (1-3 sentences)
- Professional and positive tone"""

    # 10. SALARY/COMPENSATION QUESTIONS
    if any(
        word in message_lower
        for word in ["salary", "compensation", "pay", "rate", "wage"]
    ):
        return """Success criteria:
- Professional deflection to discuss during interview process
- OR provide salary range if explicitly available in resume data
- Brief response (1-2 sentences)
- Maintains negotiating position"""

    # DEFAULT: Generic but more specific than before
    return """Success criteria:
- Directly answers the user's question
- Uses accurate information from resume data only
- Response is appropriately concise (typically 2-5 sentences)
- Professional and conversational tone
- Plain text format (no markdown)
- If information isn't in resume data, uses record_unknown_question tool"""
