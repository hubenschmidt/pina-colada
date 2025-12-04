 function _optionalChain(ops) { let lastAccessLHS = undefined; let value = ops[0]; let i = 1; while (i < ops.length) { const op = ops[i]; const fn = ops[i + 1]; i += 2; if ((op === 'optionalAccess' || op === 'optionalCall') && value == null) { return undefined; } if (op === 'access' || op === 'optionalAccess') { lastAccessLHS = value; value = fn(value); } else if (op === 'call' || op === 'optionalCall') { value = fn((...args) => value.call(lastAccessLHS, ...args)); lastAccessLHS = undefined; } } return value; }import { NextResponse } from 'next/server';

import { auth0 } from './lib/auth0';

export const middleware = async (request) => {
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

  if (!_optionalChain([session, 'optionalAccess', _ => _.user])) {
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
