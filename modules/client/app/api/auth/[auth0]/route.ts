import { auth0 } from '../../../../lib/auth0';
import { NextRequest, NextResponse } from 'next/server';

export async function GET(
  request: NextRequest,
  { params }: { params: Promise<{ auth0: string }> }
) {
  const { auth0: route } = await params;

  if (route === 'login') {
    return auth0.startInteractiveLogin();
  }

  if (route === 'logout') {
    const session = await auth0.getSession();
    const returnTo = process.env.APP_BASE_URL || 'http://localhost:3000';
    const logoutUrl = `https://${process.env.AUTH0_DOMAIN}/v2/logout?client_id=${process.env.AUTH0_CLIENT_ID}&returnTo=${encodeURIComponent(returnTo)}`;

    // Clear the session cookie
    const response = NextResponse.redirect(logoutUrl);
    response.cookies.delete('appSession');

    return response;
  }

  return new Response('Not Found', { status: 404 });
}
