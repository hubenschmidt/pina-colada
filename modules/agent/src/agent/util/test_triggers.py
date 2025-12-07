"""
Test triggers for simulating various error scenarios during development.

To use these triggers, import and call them in graph.py before processing messages.
"""

import json
import asyncio
import logging
import os
import traceback
from fastapi import WebSocket

logger = logging.getLogger(__name__)

VERBOSE_ERROR_LOGS = os.getenv("VERBOSE_ERROR_LOGS", "").lower() == "true"


async def handle_test_error_trigger(
    websocket: WebSocket, message: str
) -> bool:
    """
    Test trigger: simulate error for testing error display formatting.
    
    When message is "test error" or "trigger error", sends a partial response
    followed by an error message to test the error display formatting.
    
    Returns True if trigger was handled, False otherwise.
    """
    if message.lower().strip() not in {"test error", "trigger error"}:
        return False

    logger.info("Test error triggered - simulating error for testing")
    try:
        # Send initial partial response
        await websocket.send_text(
            json.dumps(
                {
                    "on_chat_model_stream": "Absolutely — I can write a tailored, professional cover letter for you. Could you please share the **job description or a link to the posting**?\n\nThat helps me align your experience and skills (from your resume) with the specific requirements and tone of the role. Once I have that, I'll craft a complete, ready-to-send cover letter."
                }
            )
        )
        # Small delay to simulate streaming
        await asyncio.sleep(0.1)
        # Send error message on new line
        await websocket.send_text(
            json.dumps(
                {
                    "on_chat_model_stream": "\n\nSorry, there was an error generating the response."
                }
            )
        )
        await websocket.send_text(json.dumps({"on_chat_model_end": True}))

        # Send structured error for error banner
        try:
            raise ValueError("Test error: 'dict' object has no attribute 'id' (simulated)")
        except Exception as e:
            error_payload = {
                "type": "error",
                "message": str(e),
            }
            if VERBOSE_ERROR_LOGS:
                error_payload["details"] = traceback.format_exc()
            await websocket.send_text(json.dumps(error_payload))
    except Exception as send_err:
        send_error_name = type(send_err).__name__
        send_error_msg = str(send_err).lower()
        is_disconnect_error = (
            "disconnect" in send_error_name.lower() 
            or "close" in send_error_name.lower()
            or "disconnect" in send_error_msg
            or "close" in send_error_msg
        )
        
        if is_disconnect_error:
            logger.debug("Could not send test error message, client already disconnected")
            return True
        
        logger.debug(f"Could not send test error message: {send_error_name}")
    
    return True

