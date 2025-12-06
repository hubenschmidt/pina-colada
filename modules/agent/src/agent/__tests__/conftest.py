"""
Pytest fixtures for AI chat integration tests.
"""

import asyncio
import json
from typing import Any
from uuid import uuid4

import pytest
import websockets


WS_URL = "ws://localhost:8000/ws"
QUERY_TIMEOUT = 60  # seconds


@pytest.fixture
async def ws_connection():
    """Create and initialize a WebSocket connection per test."""
    ws = await websockets.connect(WS_URL)
    test_id = str(uuid4())[:8]
    await ws.send(json.dumps({"uuid": f"pytest-{test_id}", "init": True}))
    await asyncio.sleep(0.2)
    yield ws, test_id
    await ws.close()


async def send_query(ws_tuple: tuple, message: str) -> str:
    """Send a query and collect the response."""
    ws, test_id = ws_tuple
    await ws.send(json.dumps({"uuid": f"pytest-{test_id}", "message": message}))

    response = ""
    while True:
        try:
            raw = await asyncio.wait_for(ws.recv(), timeout=QUERY_TIMEOUT)
            data = json.loads(raw)

            if "on_chat_model_stream" in data:
                response = data["on_chat_model_stream"]
            if "on_chat_model_end" in data:
                break
        except asyncio.TimeoutError:
            break

    return response
