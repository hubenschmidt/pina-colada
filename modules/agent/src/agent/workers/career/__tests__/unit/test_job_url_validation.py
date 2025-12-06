"""
Unit tests for job URL validation.

Quick tests to verify URLs from job search are valid and reachable.
Run with: pytest src/agent/workers/career/__tests__/unit/test_job_url_validation.py -vs
"""

import os

import pytest
import httpx
from urllib.parse import urlparse

from agent.tools.worker_tools import _format_serper_results, _enhance_job_query


def extract_urls_from_text(text: str) -> list[str]:
    """Extract URLs from formatted job search results."""
    import re
    return re.findall(r'https?://[^\s<>"{}|\\^`\[\]]+', text)


class TestUrlStructure:
    """Tests for URL structure validation."""

    def test_formatted_results_contain_urls(self):
        """Formatted results should contain valid URLs."""
        results = {
            "organic": [
                {
                    "title": "Senior Engineer at Acme",
                    "link": "https://acme.com/careers/123",
                    "snippet": "Great job",
                }
            ]
        }
        formatted = _format_serper_results(results)
        urls = extract_urls_from_text(formatted)

        assert len(urls) == 1
        assert urls[0] == "https://acme.com/careers/123"

    def test_urls_have_valid_structure(self):
        """URLs should have valid scheme and netloc."""
        results = {
            "organic": [
                {"title": "Job 1", "link": "https://company1.com/jobs/1"},
                {"title": "Job 2", "link": "https://company2.io/careers"},
                {"title": "Job 3", "link": "https://jobs.lever.co/company/abc123"},
            ]
        }
        formatted = _format_serper_results(results)
        urls = extract_urls_from_text(formatted)

        for url in urls:
            parsed = urlparse(url)
            assert parsed.scheme in ("http", "https"), f"Invalid scheme: {url}"
            assert parsed.netloc, f"Missing netloc: {url}"
            assert "." in parsed.netloc, f"Invalid domain: {url}"

    def test_no_job_board_urls_in_query(self):
        """Enhanced query should exclude job board domains."""
        query = _enhance_job_query("software engineer NYC")

        excluded = ["linkedin.com", "indeed.com", "glassdoor.com", "monster.com"]
        for domain in excluded:
            assert f"-site:{domain}" in query


class TestUrlReachability:
    """Tests that verify URLs are actually reachable (requires network)."""

    @pytest.fixture
    def sample_job_urls(self):
        """Sample URLs that should be valid career pages."""
        return [
            "https://www.ycombinator.com/jobs",
            "https://www.builtinnyc.com/jobs",
            "https://boards.greenhouse.io",
        ]

    @pytest.mark.asyncio
    async def test_known_career_sites_reachable(self, sample_job_urls):
        """Known career sites should be reachable."""
        async with httpx.AsyncClient(timeout=10.0, follow_redirects=True) as client:
            for url in sample_job_urls:
                try:
                    resp = await client.head(url)
                    assert resp.status_code < 400, f"{url} returned {resp.status_code}"
                    print(f"✓ {url} -> {resp.status_code}")
                except httpx.RequestError as e:
                    pytest.fail(f"Failed to reach {url}: {e}")


class TestLiveJobSearch:
    """Live test of job search (requires Serper API key and network).

    Run manually with: RUN_LIVE=1 pytest -vs -k test_live_job_search
    """

    @pytest.mark.asyncio
    @pytest.mark.skipif(
        not os.getenv("RUN_LIVE"),
        reason="Live test - run with RUN_LIVE=1"
    )
    async def test_live_job_search_returns_valid_urls(self):
        """Live job search should return valid, reachable URLs."""
        from agent.tools.worker_tools import job_search_with_filter

        results = await job_search_with_filter("software engineer NYC startup")
        urls = extract_urls_from_text(results)

        assert len(urls) > 0, "Should return at least one URL"
        print(f"\nFound {len(urls)} URLs in results")

        # Check all URLs and report on validity
        valid = []
        broken = []
        async with httpx.AsyncClient(timeout=10.0, follow_redirects=True) as client:
            for url in urls[:5]:  # Check first 5
                try:
                    resp = await client.head(url)
                    if resp.status_code < 400:
                        valid.append(url)
                        print(f"✓ {url} -> {resp.status_code}")
                    else:
                        broken.append((url, resp.status_code))
                        print(f"✗ {url} -> {resp.status_code}")
                except httpx.RequestError as e:
                    broken.append((url, str(e)))
                    print(f"✗ {url} -> {e}")

        print(f"\nSummary: {len(valid)} valid, {len(broken)} broken")

        # At least 50% should be valid
        validity_rate = len(valid) / (len(valid) + len(broken)) if (valid or broken) else 0
        assert validity_rate >= 0.5, f"Too many broken links ({len(broken)}/{len(valid)+len(broken)})"
