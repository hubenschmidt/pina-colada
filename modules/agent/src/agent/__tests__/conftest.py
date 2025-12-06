"""
Pytest fixtures for AI chat integration tests.
"""

import asyncio
import json
from dataclasses import dataclass
from typing import Any
from uuid import uuid4

import pytest
import websockets


WS_URL = "ws://localhost:8000/ws"
QUERY_TIMEOUT = 60  # seconds


@dataclass
class QueryResult:
    """Result from a query including response and token usage."""
    response: str
    total_tokens: int
    input_tokens: int
    output_tokens: int

    def __str__(self) -> str:
        return self.response


@pytest.fixture
async def ws_connection():
    """Create and initialize a WebSocket connection per test."""
    ws = await websockets.connect(WS_URL)
    test_id = str(uuid4())[:8]
    await ws.send(json.dumps({"uuid": f"pytest-{test_id}", "init": True}))
    await asyncio.sleep(0.2)
    yield ws, test_id
    await ws.close()


async def send_query(ws_tuple: tuple, message: str) -> QueryResult:
    """Send a query and collect the response with token usage."""
    ws, test_id = ws_tuple
    await ws.send(json.dumps({"uuid": f"pytest-{test_id}", "message": message}))

    response = ""
    cumulative_tokens = {"input": 0, "output": 0, "total": 0}

    while True:
        try:
            raw = await asyncio.wait_for(ws.recv(), timeout=QUERY_TIMEOUT)
            data = json.loads(raw)

            # Streaming response content
            if "on_chat_model_stream" in data:
                response += data["on_chat_model_stream"]
            # Token usage from graph.py adapter
            if "on_token_cumulative" in data:
                cumulative_tokens = data["on_token_cumulative"]
            # End of stream
            if "on_chat_model_end" in data:
                break
        except asyncio.TimeoutError:
            break

    return QueryResult(
        response=response,
        total_tokens=cumulative_tokens.get("total", 0),
        input_tokens=cumulative_tokens.get("input", 0),
        output_tokens=cumulative_tokens.get("output", 0),
    )
