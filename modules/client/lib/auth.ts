import { getAuthConfig } from './get-auth-config'

const AUTH_TOKEN_KEY = "job_tracker_auth_token";

export const login = async (password: string): Promise<boolean> => {
  if (typeof window === "undefined") return false;

  const { isDev, password: devPassword } = getAuthConfig();

  // Development: Use client-side check for convenience
  if (isDev) {
    if (!devPassword) {
      console.error("NEXT_PUBLIC_JOBS_PASSWORD not set in .env.local");
      return false;
    }
    if (password === devPassword) {
      const token = `dev_${Date.now()}_${Math.random().toString(36).substring(7)}`;
      localStorage.setItem(AUTH_TOKEN_KEY, token);
      return true;
    }
    console.log("Password mismatch. Expected:", devPassword ? "***" : "NOT SET", "Got:", password);
    return false;
  }

  // Production: Use server-side API route
  try {
    const response = await fetch("/api/auth/login", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ password }),
    });

    if (!response.ok) return false;

    const data = await response.json();

    if (data.success && data.token) {
      localStorage.setItem(AUTH_TOKEN_KEY, data.token);
      return true;
    }

    return false;
  } catch (error) {
    console.error("Login error:", error);
    return false;
  }
}

export const logout = (): void => {
  if (typeof window !== "undefined") {
    localStorage.removeItem(AUTH_TOKEN_KEY);
  }
}

export const isAuthenticated = (): boolean => {
  if (typeof window === "undefined") return false;
  const token = localStorage.getItem(AUTH_TOKEN_KEY);
  return !!token;
}

export const getAuthToken = (): string | null => {
  if (typeof window === "undefined") return null;
  return localStorage.getItem(AUTH_TOKEN_KEY);
}
