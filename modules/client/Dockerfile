# syntax=docker/dockerfile:1.6
FROM node:24

RUN apt-get update && apt-get install -y dumb-init && rm -rf /var/lib/apt/lists/*

# Create user BEFORE copying files
RUN groupadd --gid 1001 nodejs \
    && useradd --uid 1001 --gid nodejs --home-dir /home/appuser --create-home appuser

WORKDIR /app

# Copy with ownership already set
COPY --chown=appuser:nodejs package*.json ./

# Install as root (for npm cache), then fix ownership of node_modules
RUN --mount=type=cache,target=/root/.npm \
    npm ci --prefer-offline --no-audit \
    && chown -R appuser:nodejs node_modules

COPY --chown=appuser:nodejs . .

RUN npm run build

USER appuser
ENV HOME=/home/appuser

EXPOSE 3001
ENTRYPOINT ["dumb-init", "--"]
CMD ["npm", "start"]
