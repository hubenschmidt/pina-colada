# server.py — FastAPI + WebSocket
# Flow:
#   - accept WS
#   - async-iterate frames
#   - handle each frame via a helper
#   - errors logged; socket closed on exit

__all__ = ["app"]

import json
import logging
import os
from datetime import datetime
import time
from fastapi import FastAPI, WebSocket, Request
from fastapi.middleware.cors import CORSMiddleware
from agent.graph import invoke_graph
from agent.orchestrator import cancel_streaming
from agent.util.logging_config import configure_logging
from middleware import AuthMiddleware, ErrorLoggingMiddleware
from api.routes import jobs_routes, leads_routes, auth_routes, users_routes, preferences_routes, organizations_routes, individuals_routes, industries_routes, salary_ranges_routes, employee_count_ranges_routes, funding_stages_routes, notes_routes, contacts_routes, accounts_routes, revenue_ranges_routes, technologies_routes, provenance_routes, reports_routes, projects_routes, opportunities_routes, partnerships_routes, tasks_routes, comments_routes, notifications_routes, documents_routes, tags_routes, conversations_routes, usage_routes, costs_routes
from api.routes.mocks.k401_rollover import router as mock_401k_router
from uuid import uuid4
import uvicorn

# -----------------------------------------------------------------------------
# App + logging
# -----------------------------------------------------------------------------
configure_logging()  # plain logging to stdout (Docker captures it)
logger = logging.getLogger("app.server")
app = FastAPI(redirect_slashes=False)

# -----------------------------------------------------------------------------
# CORS middleware (MUST be added before routers)
# -----------------------------------------------------------------------------
app.add_middleware(
    CORSMiddleware,
    allow_origins=[
        "wss://api.pinacolada.co",  # production domain
        "https://pinacolada.co",  # production domain
        "https://www.pinacolada.co",  # www version
        "wss://test.api.pinacolada.co",  # test domain
        "https://test.pinacolada.co",  # test domain
        "https://www.test.pinacolada.co",  # test www version
        "http://localhost:3000",  # Local development
        "http://localhost:3001",  # Local development (alternate port)
    ],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Auth and error logging middleware
app.add_middleware(AuthMiddleware)
app.add_middleware(ErrorLoggingMiddleware)

# Include routers (AFTER middleware)
app.include_router(jobs_routes)
app.include_router(leads_routes)
app.include_router(auth_routes)
app.include_router(users_routes)
app.include_router(preferences_routes)
app.include_router(organizations_routes)
app.include_router(individuals_routes)
app.include_router(industries_routes)
app.include_router(salary_ranges_routes)
app.include_router(employee_count_ranges_routes)
app.include_router(funding_stages_routes)
app.include_router(notes_routes)
app.include_router(contacts_routes)
app.include_router(accounts_routes)
app.include_router(revenue_ranges_routes)
app.include_router(technologies_routes)
app.include_router(provenance_routes)
app.include_router(reports_routes)
app.include_router(projects_routes)
app.include_router(opportunities_routes)
app.include_router(partnerships_routes)
app.include_router(tasks_routes)
app.include_router(comments_routes)
app.include_router(notifications_routes)
app.include_router(documents_routes)
app.include_router(tags_routes)
app.include_router(conversations_routes)
app.include_router(usage_routes)
app.include_router(costs_routes)
app.include_router(mock_401k_router)


# -----------------------------------------------------------------------------
# Health check endpoint
# -----------------------------------------------------------------------------
@app.get("/")
async def health_check():
    return {"status": "ok", "message": "FastAPI server is running"}


@app.get("/health")
async def health():
    return {"status": "healthy"}


@app.get("/version")
async def version():
    return {"build_id": os.getenv("BUILD_ID", "local")}


# -----------------------------------------------------------------------------
# Middleware to log all requests
# -----------------------------------------------------------------------------
@app.middleware("http")
async def log_requests(request: Request, call_next):
    start_time = time.time()

    # Log incoming request
    auth_header = request.headers.get("authorization", "")[:50]  # Truncate for security
    logger.info(f"→ {request.method} {request.url.path} | Auth: {auth_header}...")

    # Process request
    response = await call_next(request)

    # Log response
    duration_ms = int((time.time() - start_time) * 1000)
    logger.info(
        f"← {request.method} {request.url.path} | {response.status_code} | {duration_ms}ms"
    )

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

        # Cancel request? Stop the running graph execution
        if payload.get("type") == "cancel":
            logger.info(
                json.dumps(
                    {
                        "timestamp": datetime.now().isoformat(),
                        "uuid": new_uid,
                        "op": "Cancel request received",
                    }
                )
            )
            await cancel_streaming(new_uid)
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
    uvicorn.run(app, host="0.0.0.0", port=8000, log_level="warning")
