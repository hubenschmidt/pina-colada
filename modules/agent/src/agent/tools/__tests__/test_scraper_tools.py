"""Test scraper tools against mock 401k endpoint."""
import pytest
from agent.tools.scraper_tools import (
    analyze_page_structure,
    scrape_static_page,
    fill_form,
    automate_browser,
)


@pytest.mark.asyncio
async def test_analyze_page_structure():
    """Test analyze_page_structure tool."""
    base_url = "http://localhost:8000/mocks/401k-rollover"
    result = await analyze_page_structure.ainvoke({"url": base_url + "/"})

    assert result.get('page_type') == 'form'
    assert result.get('requires_js') is True
    assert result.get('recommended_strategy') == 'browser'
    assert len(result.get('form_fields', [])) == 2


@pytest.mark.asyncio
async def test_scrape_static_page():
    """Test scrape_static_page tool."""
    base_url = "http://localhost:8000/mocks/401k-rollover"
    result = await scrape_static_page.ainvoke({
        "url": base_url + "/",
        "selectors": {
            "title": "h1",
            "subtitle": ".subtitle"
        }
    })

    assert result['success'] is True
    assert result['data']['title'] == '401(k) Account Portal'
    assert result['data']['subtitle'] == 'Secure access to your retirement savings'


@pytest.mark.asyncio
async def test_fill_form():
    """Test fill_form tool with login."""
    base_url = "http://localhost:8000/mocks/401k-rollover"
    result = await fill_form.ainvoke({
        "url": base_url + "/",
        "form_data": {
            "username": "demo",
            "password": "demo123"
        },
        "submit": True
    })

    assert result['success'] is True
    assert result['submitted'] is True
    assert base_url in result.get('final_url', '')


@pytest.mark.asyncio
async def test_automate_browser():
    """Test automate_browser with full 401k flow."""
    base_url = "http://localhost:8000/mocks/401k-rollover"
    result = await automate_browser.ainvoke({
        "url": base_url + "/",
        "actions": [
            {"type": "fill", "selector": "input#username", "value": "demo"},
            {"type": "fill", "selector": "input#password", "value": "demo123"},
            {"type": "click", "selector": "button[type='submit']"},
            {"type": "wait", "selector": ".account-card", "timeout": 5000},
            {"type": "extract", "selector": ".provider", "key": "providers"},
            {"type": "extract", "selector": ".balance", "key": "balances"},
        ]
    })

    assert result['success'] is True
    assert result.get('final_url') == f"{base_url}/accounts"
    assert 'providers' in result.get('data', {})
    assert 'balances' in result.get('data', {})
    assert len(result['data']['providers']) == 2
    assert len(result['data']['balances']) == 2
