# syntax=docker/dockerfile:1.6
FROM python:3.12-slim AS base
WORKDIR /app

# OS deps (cached unless this changes)
RUN apt-get update && apt-get install -y --no-install-recommends \
    curl git build-essential ca-certificates \
    # Node.js for MCP servers
    nodejs npm \
 && rm -rf /var/lib/apt/lists/*

# uv (cached unless this changes)
RUN curl -LsSf https://astral.sh/uv/install.sh | sh && ln -s /root/.local/bin/uv /usr/local/bin/uv

# Dev image
FROM base AS dev
ENV PYTHONUNBUFFERED=1 \
    PYTHONDONTWRITEBYTECODE=1 \
    UV_SYSTEM_PYTHON=1 \
    PYO3_USE_ABI3_FORWARD_COMPATIBILITY=1

# Copy dependency files first (better caching)
COPY pyproject.toml README.md ./

# Copy source code (needed for editable install)
COPY src ./src

# Install Python deps with cache mount (faster rebuilds)
# This layer is cached unless pyproject.toml or src/ changes
RUN --mount=type=cache,target=/root/.cache/uv \
    uv pip install -e . --no-cache-dir && \
    uv pip install watchfiles --no-cache-dir

# Copy remaining application files (changes frequently)
COPY scripts ./scripts
COPY migrations ./migrations
COPY seeders ./seeders
COPY start.sh /usr/src/app/start.sh

# Give execution permissions to the start.sh script
RUN chmod +x /usr/src/app/start.sh

EXPOSE 2024 8000
CMD ["/usr/src/app/start.sh"]
