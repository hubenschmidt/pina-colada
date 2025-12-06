from typing import List, Tuple

# Each entry: (keywords, criteria_text)
# First match wins, so order matters
_CRITERIA_MAP: List[Tuple[List[str], str]] = [
    # 1. CRM LOOKUPS
    (
        ["look up", "lookup", "find", "search for", "who is", "info on", "information about"],
        """Success criteria:
- Use appropriate lookup tool (lookup_individual, lookup_organization, lookup_account)
- Return relevant CRM data found
- Response is clear and organized
- Offer next steps (view related data, add notes, etc.)""",
    ),
    # 2. ACCOUNTS
    (
        ["account", "accounts", "customer", "customers", "client", "clients"],
        """Success criteria:
- Query accounts using lookup_account or list tools
- Return account details with linked orgs/individuals
- Clear, organized response
- Offer relevant follow-up actions""",
    ),
    # 3. CONTACTS/INDIVIDUALS
    (
        ["contact", "contacts", "individual", "individuals", "person", "people"],
        """Success criteria:
- Use lookup_individual for person searches
- Return name, email, title, and ID
- Organized list format
- Offer to show related data""",
    ),
    # 4. ORGANIZATIONS
    (
        ["organization", "organizations", "company", "companies", "org", "orgs"],
        """Success criteria:
- Use lookup_organization for company searches
- Return org name, website, and ID
- Organized response
- Offer related lookups""",
    ),
    # 5. DOCUMENTS
    (
        ["document", "documents", "file", "files", "resume", "attachment"],
        """Success criteria:
- Use search_documents to find files
- Use get_document_content to retrieve text
- Return relevant document info or content
- Share full text if requested""",
    ),
    # 6. GREETING
    (
        ["hi", "hello", "hey", "greetings", "good morning", "good afternoon"],
        """Success criteria:
- Warm, professional greeting
- Identify as CRM assistant (NOT a person)
- Offer to help with lookups, documents, etc.
- 2-3 sentences maximum""",
    ),
    # 7. HELP/CAPABILITIES
    (
        ["help", "what can you", "how do i", "capabilities", "features"],
        """Success criteria:
- Explain CRM assistant capabilities
- Mention: contacts, organizations, accounts, documents
- Offer example queries
- Brief and clear""",
    ),
]

_DEFAULT_CRITERIA = """Success criteria:
- Directly answers the user's question
- Uses CRM tools when appropriate for data lookups
- Response is appropriately concise
- Professional tone
- Plain text format (no markdown)
- If data isn't found, clearly state that"""


def get_success_criteria(message: str) -> str:
    """
    Generate context-aware success criteria based on the user's message.
    Returns the first matching criteria or a default.
    """
    msg = message.lower()

    for keywords, criteria in _CRITERIA_MAP:
        if any(kw in msg for kw in keywords):
            return criteria

    return _DEFAULT_CRITERIA
