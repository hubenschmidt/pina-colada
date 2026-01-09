import { NextResponse } from "next/server";

import { auth0 } from "./lib/auth0";

export const proxy = async (request) => {
  const { pathname } = request.nextUrl;

  // Proxy /api/v1/* to backend
  if (pathname.startsWith("/api/v1/")) {
    const apiUrl = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8000";
    const backendPath = pathname.replace("/api/v1", "");
    const url = new URL(backendPath + request.nextUrl.search, apiUrl);
    return NextResponse.rewrite(url);
  }

  // Allow public routes
  const publicRoutes = ["/", "/login", "/about", "/terms", "/privacy"];

  // Allow auth routes (both /auth/* and /api/auth/*)
  if (pathname.startsWith("/auth/")) {
    return await auth0.middleware(request);
  }
  if (pathname.startsWith("/api/auth/")) {
    return NextResponse.next();
  }

  // Allow public routes and hash routes (handled client-side)
  if (publicRoutes.includes(pathname)) {
    return NextResponse.next();
  }

  // For protected routes, check authentication
  const session = await auth0.getSession(request);

  if (!session?.user) {
    // Redirect to root if not authenticated
    return NextResponse.redirect(new URL("/", request.url));
  }

  return await auth0.middleware(request);
};

export const config = {
  matcher: [
    "/((?!_next/static|_next/image|favicon.ico|.*\\.png|.*\\.jpg|.*\\.jpeg|.*\\.svg|.*\\.ico|logo.png|icon.png).*)",
  ],
};
