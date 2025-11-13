import { withMiddlewareAuthRequired } from '@auth0/nextjs-auth0/edge';

export default withMiddlewareAuthRequired();

export const config = {
  matcher: [
    '/((?!_next/static|_next/image|favicon.ico|login|api/auth|about).*)',
  ],
};
