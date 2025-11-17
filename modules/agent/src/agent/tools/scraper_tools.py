"""Scraper tools for web automation and data extraction."""
import asyncio
import logging
from typing import Any, Dict, List, Optional

import requests
from bs4 import BeautifulSoup
from langchain_core.tools import tool
from langchain_openai import ChatOpenAI
from playwright.async_api import async_playwright, Browser, Page

logger = logging.getLogger(__name__)


@tool
async def analyze_page_structure(url: str) -> dict:
    """
    Analyze page HTML to determine optimal scraping strategy.

    Uses LLM to examine page structure and recommend whether to use
    static scraping (requests + BeautifulSoup) or browser automation (Playwright).

    Args:
        url: Target URL to analyze

    Returns:
        {
            "page_type": "static" | "dynamic" | "form",
            "requires_js": bool,
            "form_fields": [...],
            "recommended_strategy": "static" | "browser",
            "analysis": "detailed explanation"
        }
    """
    try:
        logger.info(f"Analyzing page structure for: {url}")

        # Fetch the page HTML
        response = requests.get(url, timeout=10)
        response.raise_for_status()
        html_content = response.text

        # Parse with BeautifulSoup to get structure
        soup = BeautifulSoup(html_content, "lxml")

        # Extract key indicators
        has_forms = len(soup.find_all("form")) > 0
        has_scripts = len(soup.find_all("script")) > 0
        form_fields = []

        if has_forms:
            for form in soup.find_all("form"):
                for input_field in form.find_all(["input", "select", "textarea"]):
                    field_info = {
                        "type": input_field.get("type", "text"),
                        "name": input_field.get("name"),
                        "id": input_field.get("id"),
                        "required": input_field.has_attr("required"),
                    }
                    form_fields.append(field_info)

        # Use LLM to analyze (GPT-4o mini for speed)
        llm = ChatOpenAI(model="gpt-4o-mini", temperature=0)

        # Get just the visible text and structure
        page_text = soup.get_text(separator="\n", strip=True)[:2000]

        prompt = f"""Analyze this webpage and determine the best scraping strategy.

URL: {url}

Page Statistics:
- Has forms: {has_forms}
- Has JavaScript: {has_scripts}
- Form fields found: {len(form_fields)}

Visible content preview:
{page_text}

Based on this, determine:
1. Page type (static/dynamic/form)
2. Whether JavaScript execution is required
3. Recommended scraping strategy (static or browser automation)

Respond in JSON format:
{{
    "page_type": "static|dynamic|form",
    "requires_js": true|false,
    "recommended_strategy": "static|browser",
    "analysis": "brief explanation"
}}
"""

        response = llm.invoke(prompt)

        # Parse LLM response
        import json
        try:
            analysis = json.loads(response.content)
        except json.JSONDecodeError:
            # Fallback to heuristic-based decision
            if has_forms and has_scripts:
                analysis = {
                    "page_type": "form",
                    "requires_js": True,
                    "recommended_strategy": "browser",
                    "analysis": "Page has forms with JavaScript - browser automation recommended"
                }
            elif has_scripts:
                analysis = {
                    "page_type": "dynamic",
                    "requires_js": True,
                    "recommended_strategy": "browser",
                    "analysis": "Page has JavaScript - may need browser automation"
                }
            else:
                analysis = {
                    "page_type": "static",
                    "requires_js": False,
                    "recommended_strategy": "static",
                    "analysis": "Simple static page - use requests + BeautifulSoup"
                }

        # Add form fields to response
        analysis["form_fields"] = form_fields

        logger.info(f"Analysis complete: {analysis['recommended_strategy']} strategy")
        return analysis

    except Exception as e:
        logger.error(f"Error analyzing page structure: {e}")
        return {
            "error": str(e),
            "page_type": "unknown",
            "requires_js": False,
            "recommended_strategy": "browser",
            "form_fields": [],
            "analysis": f"Error during analysis: {str(e)}"
        }


@tool
async def scrape_static_page(url: str, selectors: Optional[Dict[str, str]] = None) -> dict:
    """
    Traditional web scraping using requests + BeautifulSoup.

    Best for static HTML pages without JavaScript rendering.

    Args:
        url: Target URL to scrape
        selectors: Optional dict of {label: css_selector} to extract specific data

    Returns:
        {
            "success": bool,
            "data": {extracted data},
            "html_preview": "first 500 chars of HTML",
            "error": str | None
        }
    """
    try:
        logger.info(f"Scraping static page: {url}")

        # Fetch the page
        response = requests.get(url, timeout=10)
        response.raise_for_status()

        # Parse HTML
        soup = BeautifulSoup(response.text, "lxml")

        # Extract data based on selectors
        extracted_data = {}

        if selectors:
            for label, selector in selectors.items():
                elements = soup.select(selector)
                if elements:
                    if len(elements) == 1:
                        extracted_data[label] = elements[0].get_text(strip=True)
                    else:
                        extracted_data[label] = [el.get_text(strip=True) for el in elements]
                else:
                    extracted_data[label] = None
        else:
            # Extract common elements if no selectors provided
            extracted_data = {
                "title": soup.title.string if soup.title else None,
                "headings": [h.get_text(strip=True) for h in soup.find_all(["h1", "h2", "h3"])[:5]],
                "paragraphs": [p.get_text(strip=True) for p in soup.find_all("p")[:3]],
            }

        logger.info(f"Successfully scraped {len(extracted_data)} data points")

        return {
            "success": True,
            "data": extracted_data,
            "html_preview": response.text[:500],
            "error": None
        }

    except Exception as e:
        logger.error(f"Error scraping static page: {e}")
        return {
            "success": False,
            "data": {},
            "html_preview": None,
            "error": str(e)
        }


@tool
async def automate_browser(
    url: str,
    actions: List[Dict[str, Any]],
    headless: bool = True
) -> dict:
    """
    Browser automation using Playwright for JavaScript-heavy sites.

    Supports complex interactions like clicking, filling forms, waiting for elements.

    Args:
        url: Starting URL
        actions: List of action dicts:
            - {"type": "click", "selector": "button#login"}
            - {"type": "fill", "selector": "input#username", "value": "demo"}
            - {"type": "wait", "selector": ".dashboard"}
            - {"type": "extract", "selector": "h1", "attribute": "text"}
            - {"type": "screenshot", "path": "/tmp/screenshot.png"}

    Returns:
        {
            "success": bool,
            "data": {extracted data},
            "screenshots": [paths],
            "final_url": "url after actions",
            "error": str | None
        }
    """
    browser: Optional[Browser] = None
    page: Optional[Page] = None

    try:
        logger.info(f"Starting browser automation for: {url}")

        async with async_playwright() as p:
            browser = await p.chromium.launch(headless=headless)
            page = await browser.new_page()

            await page.goto(url)
            logger.info(f"Navigated to {url}")

            extracted_data = {}
            screenshots = []

            # Execute actions sequentially
            for i, action in enumerate(actions):
                action_type = action.get("type")
                selector = action.get("selector")

                logger.info(f"Executing action {i+1}/{len(actions)}: {action_type}")

                if action_type == "click":
                    await page.click(selector)
                    await page.wait_for_timeout(1000)  # Wait 1s after click

                elif action_type == "fill":
                    value = action.get("value", "")
                    await page.fill(selector, value)

                elif action_type == "wait":
                    timeout = action.get("timeout", 5000)
                    await page.wait_for_selector(selector, timeout=timeout)

                elif action_type == "extract":
                    attribute = action.get("attribute", "text")
                    elements = await page.query_selector_all(selector)

                    if attribute == "text":
                        values = [await el.inner_text() for el in elements]
                    else:
                        values = [await el.get_attribute(attribute) for el in elements]

                    key = action.get("key", f"extracted_{i}")
                    extracted_data[key] = values[0] if len(values) == 1 else values

                elif action_type == "screenshot":
                    path = action.get("path", f"/tmp/screenshot_{i}.png")
                    await page.screenshot(path=path)
                    screenshots.append(path)

                elif action_type == "navigate":
                    new_url = action.get("url")
                    await page.goto(new_url)
                    await page.wait_for_load_state("networkidle")

            final_url = page.url

            await browser.close()

            logger.info(f"Browser automation completed successfully")

            return {
                "success": True,
                "data": extracted_data,
                "screenshots": screenshots,
                "final_url": final_url,
                "error": None
            }

    except Exception as e:
        logger.error(f"Error in browser automation: {e}")

        # Cleanup
        if page:
            await page.close()
        if browser:
            await browser.close()

        return {
            "success": False,
            "data": {},
            "screenshots": [],
            "final_url": None,
            "error": str(e)
        }


@tool
async def fill_form(
    url: str,
    form_data: Dict[str, str],
    submit: bool = True,
    multi_step: bool = False
) -> dict:
    """
    Intelligent form filling with automatic field detection.

    Automatically detects form fields and fills them based on provided data.
    Can handle multi-step forms with navigation.

    Args:
        url: Form URL
        form_data: Dict of {field_name: value} - will match by name, id, or label
        submit: Whether to submit the form after filling
        multi_step: Whether to handle multi-page forms

    Returns:
        {
            "success": bool,
            "submitted": bool,
            "confirmation_data": {extracted from confirmation page},
            "final_url": "url after submission",
            "error": str | None
        }
    """
    browser: Optional[Browser] = None
    page: Optional[Page] = None

    try:
        logger.info(f"Starting form fill for: {url}")

        async with async_playwright() as p:
            browser = await p.chromium.launch(headless=True)
            page = await browser.new_page()

            await page.goto(url)
            await page.wait_for_load_state("networkidle")

            # Fill form fields
            for field_name, value in form_data.items():
                # Try multiple selector strategies
                selectors = [
                    f'input[name="{field_name}"]',
                    f'input[id="{field_name}"]',
                    f'select[name="{field_name}"]',
                    f'select[id="{field_name}"]',
                    f'textarea[name="{field_name}"]',
                    f'textarea[id="{field_name}"]',
                ]

                filled = False
                for selector in selectors:
                    try:
                        element = await page.query_selector(selector)
                        if element:
                            element_type = await element.get_attribute("type")
                            tag_name = await element.evaluate("el => el.tagName.toLowerCase()")

                            if element_type == "checkbox":
                                if value in [True, "true", "True", "1", "yes"]:
                                    await element.check()
                            elif tag_name == "select":
                                await element.select_option(value=value)
                            else:
                                await element.fill(str(value))

                            logger.info(f"Filled field: {field_name}")
                            filled = True
                            break
                    except Exception:
                        continue

                if not filled:
                    logger.warning(f"Could not find field: {field_name}")

            # Submit form if requested
            submitted = False
            confirmation_data = {}

            if submit:
                # Try to find and click submit button
                submit_selectors = [
                    'button[type="submit"]',
                    'input[type="submit"]',
                    'button:has-text("Submit")',
                    'button:has-text("Sign In")',
                    'button:has-text("Login")',
                ]

                for selector in submit_selectors:
                    try:
                        button = await page.query_selector(selector)
                        if button:
                            await button.click()
                            await page.wait_for_load_state("networkidle", timeout=10000)
                            submitted = True
                            logger.info("Form submitted successfully")
                            break
                    except Exception:
                        continue

            # Extract confirmation data if on new page
            if submitted:
                # Try to extract confirmation information
                try:
                    confirmation_data = {
                        "title": await page.title(),
                        "url": page.url,
                        "content_preview": (await page.content())[:500]
                    }
                except Exception:
                    pass

            final_url = page.url

            await browser.close()

            logger.info(f"Form fill completed")

            return {
                "success": True,
                "submitted": submitted,
                "confirmation_data": confirmation_data,
                "final_url": final_url,
                "error": None
            }

    except Exception as e:
        logger.error(f"Error filling form: {e}")

        # Cleanup
        if page:
            await page.close()
        if browser:
            await browser.close()

        return {
            "success": False,
            "submitted": False,
            "confirmation_data": {},
            "final_url": None,
            "error": str(e)
        }


# Export all tools
SCRAPER_TOOLS = [
    analyze_page_structure,
    scrape_static_page,
    automate_browser,
    fill_form,
]
