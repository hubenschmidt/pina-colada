import { initAuth0 } from '@auth0/nextjs-auth0';

export const auth0 = initAuth0({
  baseURL: process.env.AUTH0_BASE_URL || process.env.APP_BASE_URL,
  issuerBaseURL: `https://${process.env.AUTH0_DOMAIN}`,
  clientID: process.env.AUTH0_CLIENT_ID,
  clientSecret: process.env.AUTH0_CLIENT_SECRET,
  secret: process.env.AUTH0_SECRET,
  authorizationParams: {
    scope: process.env.AUTH0_SCOPE || 'openid profile email',
    audience: process.env.AUTH0_AUDIENCE,
  },
});
