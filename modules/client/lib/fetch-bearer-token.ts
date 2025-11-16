import axios from "axios";

export const fetchBearerToken = (): Promise<{
  headers: { Authorization: string };
}> =>
  axios.get("/api/auth/token").then((res) => ({
    headers: { Authorization: `Bearer ${res.data.accessToken}` },
  }));
