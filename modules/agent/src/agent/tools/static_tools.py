import os
import requests
import logging
import json
from typing import Dict, Any
from dotenv import load_dotenv

load_dotenv(override=True)
logger = logging.getLogger(__name__)

pushover_token = os.getenv("PUSHOVER_TOKEN")
pushover_user = os.getenv("PUSHOVER_USER")
pushover_url = "https://api.pushover.net/1/messages.json"


# --- Pretty push ---
def record_user_context(ctx: Dict[str, Any]):
    """Persist raw user context (testing: push as a notification)"""

    try:
        # Always log full JSON for debugging
        pretty_full = json.dumps(ctx, indent=2, ensure_ascii=False, sort_keys=True)

        # Compose body within practical push size (e.g., Pushover ~1024 chars)
        max_len = 1024
        body = pretty_full[:max_len]
        # Add ellipsis if we actually truncated
        body += "…" * (len(pretty_full) > max_len)

        push(body)
        return {"ok": True, "pushed": True}
    except Exception as e:
        logger.error(f"record_user_context failed: {e}")
        return {"ok": False, "error": str(e)}


def push(text: str):
    """Send a push notification to the user"""
    if not pushover_token or not pushover_user:
        logger.warning("Pushover credentials not configured")
        return "Pushover not configured"

    try:
        response = requests.post(
            pushover_url,
            data={"token": pushover_token, "user": pushover_user, "message": text},
            timeout=5,
        )
        logger.info(f"Push notification sent: {response.status_code}")
        return "success"
    except Exception as e:
        logger.error(f"Failed to send push notification: {e}")
        return f"failed: {e}"
