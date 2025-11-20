import { NextResponse } from 'next/server';
import type { NextRequest } from 'next/server';
import { auth0 } from './lib/auth0';

export const middleware = async (request: NextRequest) => {
  const { pathname } = request.nextUrl;

  // Allow public routes
  const publicRoutes = [
    '/',
    '/login',
    '/about',
    '/terms',
    '/privacy',
  ];

  // Allow auth routes
  if (pathname.startsWith('/auth/')) {
    return await auth0.middleware(request);
  }

  // Allow public routes and hash routes (handled client-side)
  if (publicRoutes.includes(pathname)) {
    return NextResponse.next();
  }

  // For protected routes, check authentication
  const session = await auth0.getSession(request);

  if (!session?.user) {
    // Redirect to root if not authenticated
    return NextResponse.redirect(new URL('/', request.url));
  }

  return await auth0.middleware(request);
};

export const config = {
  matcher: [
    '/((?!_next/static|_next/image|favicon.ico|.*\\.png|.*\\.jpg|.*\\.jpeg|.*\\.svg|.*\\.ico|logo.png|icon.png).*)',
  ],
};
