# syntax=docker/dockerfile:1.6
FROM node:24

# Install dumb-init for proper signal handling (cached unless this changes)
RUN apt-get update && apt-get install -y dumb-init && rm -rf /var/lib/apt/lists/*

# Set the working directory in the Docker container to /app
WORKDIR /app

# Copy package files first (better caching)
# This layer is cached unless package*.json changes
COPY package*.json ./

# Install dependencies with cache mount (faster rebuilds)
# Use npm ci for faster, more reliable builds in production
RUN --mount=type=cache,target=/root/.npm \
    npm ci --prefer-offline --no-audit

# Copy the rest of your application's source files
# This changes frequently, so copy after deps are installed
COPY . .

# Build the application (only runs if source code changes)
RUN npm run build

# Make port 3001 available in development and production
EXPOSE 3001

# Use dumb-init to run your process
ENTRYPOINT ["dumb-init", "--"]

# Define the command to run the app
CMD ["npm", "start"]