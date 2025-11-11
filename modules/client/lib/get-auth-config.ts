/**
 * Runtime environment detection for authentication
 * Follows the pattern from discogs-player-2024
 */

// Server-side: use JOBS_PASSWORD (no NEXT_PUBLIC prefix)
// Client-side: use NEXT_PUBLIC_JOBS_PASSWORD (with NEXT_PUBLIC prefix)
const DEV_PASSWORD = typeof window === "undefined" 
  ? process.env.JOBS_PASSWORD 
  : process.env.NEXT_PUBLIC_JOBS_PASSWORD;

export const getAuthConfig = () => {
  if (typeof window === "undefined") {
    return { isDev: false, password: DEV_PASSWORD };
  }

  const { host } = window.location;

  // Development environment
  if (host === "localhost:3001" || host === "127.0.0.1:3001") {
    return { isDev: true, password: DEV_PASSWORD };
  }

  // Production environment
  return { isDev: false, password: null };
};
