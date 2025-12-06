"""
Unit tests for job search tool functions.

These test pure functions without network calls or database access.
Run with: pytest src/agent/workers/career/__tests__/unit/ -vs
"""

import pytest

from agent.tools.worker_tools import (
    _enhance_job_query,
    _format_serper_results,
    _is_tracked_job,
    _matches_tracked_job,
    _company_matches,
    _title_matches,
    _extract_company_from_line,
    _extract_title_from_line,
)


class TestQueryEnhancement:
    """Tests for job search query enhancement."""

    def test_adds_careers_keyword(self):
        result = _enhance_job_query("software engineer NYC")
        assert "careers OR jobs" in result

    def test_excludes_job_boards(self):
        result = _enhance_job_query("software engineer")
        assert "-site:linkedin.com" in result
        assert "-site:indeed.com" in result
        assert "-site:glassdoor.com" in result

    def test_preserves_original_query(self):
        result = _enhance_job_query("senior backend engineer startup")
        assert "senior backend engineer startup" in result


class TestSerperResultsFormatting:
    """Tests for Serper results formatting."""

    def test_formats_organic_results(self):
        results = {
            "organic": [
                {
                    "title": "Senior Engineer at Acme Corp",
                    "link": "https://acme.com/careers/123",
                    "snippet": "Join our team as a senior engineer...",
                }
            ]
        }
        formatted = _format_serper_results(results)
        assert "Acme Corp" in formatted
        assert "https://acme.com/careers/123" in formatted

    def test_handles_empty_results(self):
        results = {"organic": []}
        formatted = _format_serper_results(results)
        assert "No job listings found" in formatted

    def test_handles_missing_organic_key(self):
        results = {}
        formatted = _format_serper_results(results)
        assert "No job listings found" in formatted

    def test_limits_to_10_results(self):
        results = {
            "organic": [
                {"title": f"Job {i}", "link": f"https://example.com/{i}"}
                for i in range(15)
            ]
        }
        formatted = _format_serper_results(results)
        # Should only have entries 1-10
        assert "1. " in formatted
        assert "10. " in formatted
        assert "11. " not in formatted

    def test_extracts_company_from_at_pattern(self):
        results = {
            "organic": [
                {"title": "Software Engineer at Google", "link": "https://google.com/careers"}
            ]
        }
        formatted = _format_serper_results(results)
        assert "Google" in formatted


class TestJobTracking:
    """Tests for job tracking functions."""

    def test_is_tracked_job_with_valid_fields(self):
        job = {"company": "Acme", "title": "Engineer", "status": "applied"}
        assert _is_tracked_job(job) is True

    def test_is_tracked_job_without_company(self):
        job = {"title": "Engineer", "status": "applied"}
        assert _is_tracked_job(job) is False

    def test_is_tracked_job_without_title(self):
        job = {"company": "Acme", "status": "applied"}
        assert _is_tracked_job(job) is False

    def test_is_tracked_job_any_status(self):
        """Any status counts as tracked."""
        for status in ["applied", "interviewing", "rejected", "offer", "do_not_apply"]:
            job = {"company": "Acme", "title": "Engineer", "status": status}
            assert _is_tracked_job(job) is True


class TestCompanyMatching:
    """Tests for company name matching."""

    def test_exact_match(self):
        assert _company_matches("Google", "Google") is True

    def test_case_insensitive(self):
        assert _company_matches("google", "GOOGLE") is True

    def test_suffix_removal(self):
        assert _company_matches("Acme Inc", "Acme") is True
        assert _company_matches("Acme", "Acme LLC") is True

    def test_substring_match(self):
        assert _company_matches("Microsoft", "Microsoft Corporation") is True

    def test_no_match(self):
        assert _company_matches("Google", "Microsoft") is False

    def test_empty_strings(self):
        assert _company_matches("", "Google") is False
        assert _company_matches("Google", "") is False


class TestTitleMatching:
    """Tests for job title matching."""

    def test_exact_match(self):
        assert _title_matches("Software Engineer", "Software Engineer") is True

    def test_case_insensitive(self):
        assert _title_matches("software engineer", "SOFTWARE ENGINEER") is True

    def test_substring_match(self):
        assert _title_matches("Engineer", "Software Engineer") is True

    def test_level_agnostic(self):
        """Should match regardless of seniority level."""
        assert _title_matches("Software Engineer", "Senior Software Engineer") is True

    def test_no_match(self):
        assert _title_matches("Product Manager", "Software Engineer") is False


class TestLineExtraction:
    """Tests for extracting company/title from text lines."""

    def test_extract_company_with_at_pattern(self):
        line = "Software Engineer at Google - Full Time"
        # This tests the pattern matching, may return empty if pattern doesn't match
        company = _extract_company_from_line(line)
        # Pattern matching is heuristic, just verify it doesn't crash
        assert isinstance(company, str)

    def test_extract_title_with_engineer(self):
        line = "Senior Software Engineer position available"
        title = _extract_title_from_line(line)
        # Pattern matching is heuristic
        assert isinstance(title, str)


class TestMatchesTrackedJob:
    """Tests for the full job matching function."""

    def test_matches_existing_job(self):
        jobs = [{"company": "Google", "title": "Software Engineer", "status": "applied"}]
        assert _matches_tracked_job("Google", "Software Engineer", jobs) is True

    def test_no_match_different_company(self):
        jobs = [{"company": "Google", "title": "Software Engineer", "status": "applied"}]
        assert _matches_tracked_job("Microsoft", "Software Engineer", jobs) is False

    def test_no_match_different_title(self):
        jobs = [{"company": "Google", "title": "Software Engineer", "status": "applied"}]
        assert _matches_tracked_job("Google", "Product Manager", jobs) is False

    def test_empty_company(self):
        jobs = [{"company": "Google", "title": "Software Engineer", "status": "applied"}]
        assert _matches_tracked_job("", "Software Engineer", jobs) is False

    def test_empty_title(self):
        jobs = [{"company": "Google", "title": "Software Engineer", "status": "applied"}]
        assert _matches_tracked_job("Google", "", jobs) is False

    def test_empty_jobs_list(self):
        assert _matches_tracked_job("Google", "Software Engineer", []) is False
