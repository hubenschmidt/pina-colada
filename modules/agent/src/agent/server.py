# server.py — FastAPI + WebSocket
# Flow:
#   - accept WS
#   - async-iterate frames
#   - handle each frame via a helper
#   - errors logged; socket closed on exit

__all__ = ["app"]

import json
import logging
from datetime import datetime
from fastapi import FastAPI, WebSocket
from fastapi.middleware.cors import CORSMiddleware
from agent.graph import invoke_graph
from agent.util.logging_config import configure_logging
from uuid import uuid4

# -----------------------------------------------------------------------------
# App + logging
# -----------------------------------------------------------------------------
configure_logging()  # plain logging to stdout (Docker captures it)
logger = logging.getLogger("app.server")
app = FastAPI()

# -----------------------------------------------------------------------------
# CORS middleware
# -----------------------------------------------------------------------------
app.add_middleware(
    CORSMiddleware,
    allow_origins=[
        "wss://api.pinacolada.co",  # Your production domain
        "https://pinacolada.co",  # Your production domain
        "https://www.pinacolada.co",  # www version
        "http://localhost:3000",  # Local development
        "http://localhost:3001",  # Local development (alternate port)
    ],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)


# -----------------------------------------------------------------------------
# Health check endpoint
# -----------------------------------------------------------------------------
@app.get("/")
async def health_check():
    return {"status": "ok", "message": "FastAPI server is running"}


@app.get("/health")
async def health():
    return {"status": "healthy"}


# -----------------------------------------------------------------------------
# Middleware to log all requests
# -----------------------------------------------------------------------------
from fastapi import Request


@app.middleware("http")
async def log_requests(request: Request, call_next):
    logger.info(f"Request: {request.method} {request.url.path}")
    response = await call_next(request)
    return response


# -----------------------------------------------------------------------------
# WebSocket endpoint (frontend connects here)
# -----------------------------------------------------------------------------
@app.websocket("/ws")
async def websocket_endpoint(websocket: WebSocket):
    # Log the attempted origin / ua before accept
    origin = websocket.headers.get("origin")
    ua = websocket.headers.get("user-agent")
    cf_ray = websocket.headers.get("cf-ray")
    logger.info(
        json.dumps(
            {
                "ts": datetime.now().isoformat(),
                "op": "ws.connect_attempt",
                "origin": origin,
                "user_agent": ua,
                "cf_ray": cf_ray,
                "client": websocket.client
                and f"{websocket.client.host}:{websocket.client.port}",
            }
        )
    )

    await websocket.accept()
    logger.info(
        json.dumps(
            {
                "ts": datetime.now().isoformat(),
                "op": "ws.accepted",
            }
        )
    )
    user_uuid: str | None = None

    async def handle_frame(raw: str, uid: str | None) -> str | None:
        # Parse JSON; on error, log and return
        try:
            payload = json.loads(raw)
        except json.JSONDecodeError as e:
            logger.error(
                json.dumps(
                    {
                        "timestamp": datetime.now().isoformat(),
                        "uuid": uid,
                        "op": f"JSON encoding error - {e}",
                    }
                )
            )
            return uid

        # Log what we received
        logger.info(
            json.dumps(
                {
                    "timestamp": datetime.now().isoformat(),
                    "uuid": uid,
                    "received": payload,
                }
            )
        )

        # Track conversation id if provided
        new_uid = payload.get("uuid") or uid or str(uuid4())

        # Init ping? Just log and return
        if payload.get("init"):
            logger.info(
                json.dumps(
                    {
                        "timestamp": datetime.now().isoformat(),
                        "uuid": new_uid,
                        "op": "Initializing ws with client.",
                    }
                )
            )
            return new_uid

        # Control/telemetry envelopes -> forward to graph (it will silently store & return)
        if payload.get("type") in ("user_context", "user_context_update"):
            await invoke_graph(websocket, payload, new_uid)
            return new_uid

        # No message? Nothing to do
        message = payload.get("message")
        if not message:
            return new_uid

        # We have a message: invoke the graph (it streams back over this WS)
        await invoke_graph(websocket, payload, new_uid)
        return new_uid

    try:
        async for data in websocket.iter_text():
            user_uuid = await handle_frame(data, user_uuid)
    except Exception as e:
        logger.error(
            json.dumps(
                {
                    "timestamp": datetime.now().isoformat(),
                    "uuid": user_uuid,
                    "op": f"Error: {e}",
                }
            )
        )
    finally:
        if user_uuid:
            logger.info(
                json.dumps(
                    {
                        "timestamp": datetime.now().isoformat(),
                        "uuid": user_uuid,
                        "op": "Closing connection.",
                    }
                )
            )
        try:
            await websocket.close()
        except RuntimeError as e:
            logger.error(
                json.dumps(
                    {
                        "timestamp": datetime.now().isoformat(),
                        "uuid": user_uuid,
                        "op": f"WebSocket close error: {e}",
                    }
                )
            )


# -----------------------------------------------------------------------------
# Local dev entrypoint (uvicorn)
# -----------------------------------------------------------------------------
if __name__ == "__main__":
    import uvicorn

    uvicorn.run(app, host="0.0.0.0", port=8000, log_level="warning")
