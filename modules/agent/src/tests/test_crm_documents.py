"""Test CRM lookup with document access and job search."""

import asyncio
from agent.tools.crm_tools import crm_lookup
from agent.tools.document_tools import list_documents, read_document
from agent.tools.worker_tools import job_search_with_filter


async def test_crm_and_documents():
    """Test looking up an individual and listing their documents."""
    print("\n=== Testing CRM Lookup ===")
    result = await crm_lookup("individual", "William Hubenschmidt")
    print(f"CRM Lookup Result:\n{result}\n")

    print("\n=== Testing List Documents ===")
    docs = await list_documents()
    print(f"Available Documents:\n{docs}\n")

    print("\n=== Reading Resume (id=6) ===")
    resume_content = await read_document("william_hubenschmidt_resume.pdf")
    print(f"Resume Content:\n{resume_content[:2000]}...")


async def test_job_search():
    """Test job search for William based on resume criteria."""
    print("\n" + "=" * 60)
    print("=== Testing Job Search ===")
    print("=" * 60)

    # Simple query - let the tool handle enhancements
    query = 'Senior Software Engineer AI Engineer NYC startup "series A" OR "series B" OR "seed" -ML -"machine learning"'

    print(f"Search Query:\n{query}\n")
    print("Searching...")

    results = await job_search_with_filter(query)
    print(f"\nJob Search Results:\n{results}")


if __name__ == "__main__":
    # asyncio.run(test_crm_and_documents())
    asyncio.run(test_job_search())
