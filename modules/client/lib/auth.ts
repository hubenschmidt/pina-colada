import { getAuthConfig } from './get-auth-config'

const AUTH_TOKEN_KEY = "job_tracker_auth_token";

export async function login(password: string): Promise<boolean> {
  if (typeof window === "undefined") return false;

  const { isDev, password: devPassword } = getAuthConfig();

  // Development: Use client-side check for convenience
  if (isDev) {
    if (password === devPassword) {
      const token = `dev_${Date.now()}_${Math.random().toString(36).substring(7)}`;
      localStorage.setItem(AUTH_TOKEN_KEY, token);
      return true;
    }
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

export function logout(): void {
  if (typeof window !== "undefined") {
    localStorage.removeItem(AUTH_TOKEN_KEY);
  }
}

export function isAuthenticated(): boolean {
  if (typeof window === "undefined") return false;
  const token = localStorage.getItem(AUTH_TOKEN_KEY);
  return !!token;
}

export function getAuthToken(): string | null {
  if (typeof window === "undefined") return null;
  return localStorage.getItem(AUTH_TOKEN_KEY);
}
