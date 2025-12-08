"""Tests for intent classifier routing logic."""

import pytest
from agent.routers.intent_classifier import route_from_classifier, FastPathIntent


class TestRouteFromClassifier:
    """Test the routing logic based on classifier output."""

    def test_simple_lookup_fast_path(self):
        """Simple lookup with entity and query should fast-path."""
        state = {
            "fast_path_intent": "lookup",
            "lookup_entity_type": "individual",
            "lookup_query": "William Hubenschmidt",
        }
        assert route_from_classifier(state) == "fast_lookup"

    def test_count_fast_path(self):
        """Count queries should fast-path."""
        state = {
            "fast_path_intent": "count",
            "lookup_entity_type": "contact",
            "lookup_query": None,
        }
        assert route_from_classifier(state) == "fast_count"

    def test_list_fast_path(self):
        """List queries should fast-path."""
        state = {
            "fast_path_intent": "list",
            "lookup_entity_type": "account",
            "lookup_query": None,
        }
        assert route_from_classifier(state) == "fast_list"

    def test_document_always_full_flow(self):
        """Document operations should always go to full flow."""
        state = {
            "fast_path_intent": "lookup",
            "lookup_entity_type": "document",
            "lookup_query": "resume",
        }
        assert route_from_classifier(state) == "router"

    def test_other_intent_full_flow(self):
        """Other/complex intents should go to full flow."""
        state = {
            "fast_path_intent": "other",
            "lookup_entity_type": None,
            "lookup_query": None,
        }
        assert route_from_classifier(state) == "router"

    def test_lookup_without_query_full_flow(self):
        """Lookup without query should go to full flow."""
        state = {
            "fast_path_intent": "lookup",
            "lookup_entity_type": "individual",
            "lookup_query": None,
        }
        assert route_from_classifier(state) == "router"

    def test_lookup_without_entity_full_flow(self):
        """Lookup without entity type should go to full flow."""
        state = {
            "fast_path_intent": "lookup",
            "lookup_entity_type": None,
            "lookup_query": "John Smith",
        }
        assert route_from_classifier(state) == "router"


# These prompts SHOULD be classified as "other" (full flow)
FULL_FLOW_PROMPTS = [
    # Multi-step queries
    "look up William Hubenschmidt's Individual account and tell me the contents of his resume file",
    "find John Smith and also search for jobs matching his skills",
    "look up Acme Corp and analyze their recent performance",

    # Job search queries
    "can you conduct a job search on William Hubenschmidt's behalf",
    "search for software engineering jobs for John",
    "find job openings that match my resume",

    # Document operations
    "what's in William's resume",
    "read the contents of the contract document",
    "summarize the uploaded file",

    # Analysis/comparison queries
    "compare all accounts by revenue",
    "which contacts have the highest engagement",
    "analyze the trends in our data",

    # Cover letter / content generation
    "write a cover letter for John Smith",
    "draft an email to follow up with this contact",
    "generate a report on this organization",
]

# These prompts SHOULD be classified for fast-path
FAST_PATH_PROMPTS = [
    # Simple lookups
    ("look up William Hubenschmidt", "lookup", "individual"),
    ("find John Smith's individual account", "lookup", "individual"),
    ("show me Acme Corp", "lookup", "organization"),
    ("look up contact Jane Doe", "lookup", "contact"),

    # Count queries
    ("how many contacts do we have", "count", "contact"),
    ("count all individuals", "count", "individual"),
    ("how many organizations are in the system", "count", "organization"),

    # List queries
    ("list all accounts", "list", "account"),
    ("show all organizations", "list", "organization"),
    ("list contacts", "list", "contact"),
]


class TestClassifierExpectations:
    """Document expected classifier behavior for various prompts.

    These tests serve as a specification for what the classifier SHOULD do.
    They can be run as integration tests with the actual LLM.
    """

    @pytest.mark.parametrize("prompt", FULL_FLOW_PROMPTS)
    def test_full_flow_prompts_should_not_fast_path(self, prompt):
        """These prompts should be classified as 'other' for full flow."""
        # This is a specification test - actual LLM integration test below
        # For now, just document the expectation
        assert prompt is not None  # Placeholder

    @pytest.mark.parametrize("prompt,expected_intent,expected_entity", FAST_PATH_PROMPTS)
    def test_fast_path_prompts_should_fast_path(self, prompt, expected_intent, expected_entity):
        """These prompts should be classified for fast-path."""
        # This is a specification test - actual LLM integration test below
        assert prompt is not None  # Placeholder


@pytest.mark.integration
@pytest.mark.asyncio
class TestClassifierIntegration:
    """Integration tests that actually call the classifier LLM.

    Run with: pytest -m integration tests/test_intent_classifier.py
    """

    @pytest.fixture
    async def classifier_node(self):
        """Create the classifier node."""
        from agent.routers.intent_classifier import create_intent_classifier_node
        return await create_intent_classifier_node()

    @pytest.mark.parametrize("prompt", FULL_FLOW_PROMPTS)
    async def test_full_flow_classification(self, classifier_node, prompt):
        """Verify prompts that should go to full flow."""
        from langchain_core.messages import HumanMessage

        state = {"messages": [HumanMessage(content=prompt)]}
        result = await classifier_node(state)

        # Should be classified as "other" OR have document entity
        intent = result.get("fast_path_intent")
        entity = result.get("lookup_entity_type")

        # Build the state for routing
        route = route_from_classifier(result)

        assert route == "router", (
            f"Prompt should route to full flow:\n"
            f"  Prompt: {prompt}\n"
            f"  Got: intent={intent}, entity={entity}, route={route}"
        )

    @pytest.mark.parametrize("prompt,expected_intent,expected_entity", FAST_PATH_PROMPTS)
    async def test_fast_path_classification(self, classifier_node, prompt, expected_intent, expected_entity):
        """Verify prompts that should fast-path."""
        from langchain_core.messages import HumanMessage

        state = {"messages": [HumanMessage(content=prompt)]}
        result = await classifier_node(state)

        intent = result.get("fast_path_intent")
        entity = result.get("lookup_entity_type")
        route = route_from_classifier(result)

        # Should route to fast path
        expected_route = f"fast_{expected_intent}"
        assert route == expected_route, (
            f"Prompt should fast-path:\n"
            f"  Prompt: {prompt}\n"
            f"  Expected: intent={expected_intent}, entity={expected_entity}, route={expected_route}\n"
            f"  Got: intent={intent}, entity={entity}, route={route}"
        )
