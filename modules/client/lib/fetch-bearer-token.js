import axios from "axios";

let cachedToken = null;
let tokenExpiry = 0;
let pendingRequest = null;

const TOKEN_BUFFER_MS = 60000; // Refresh 1 min before expiry

const fetchToken = async () => {
  const res = await axios.get("/api/auth/token");
  cachedToken = res.data.accessToken;
  // Auth0 tokens typically expire in 1 hour, cache for 55 mins
  tokenExpiry = Date.now() + 55 * 60000;
  return cachedToken;
};

export const fetchBearerToken = async () => {
  // Return cached token if still valid
  if (cachedToken && Date.now() < tokenExpiry - TOKEN_BUFFER_MS) {
    return { headers: { Authorization: `Bearer ${cachedToken}` } };
  }

  // Dedupe concurrent requests
  if (!pendingRequest) {
    pendingRequest = fetchToken().finally(() => {
      pendingRequest = null;
    });
  }

  const token = await pendingRequest;
  return { headers: { Authorization: `Bearer ${token}` } };
};
