'use client';

import { useUser } from '@auth0/nextjs-auth0/client';
import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { ChevronRight, ChevronLeft } from 'lucide-react';
import Image from 'next/image';
import { useUserContext } from '../../context/userContext';
import { SET_TENANT_NAME } from '../../reducers/userReducer';

const LoginPage = () => {
  const { user, isLoading } = useUser();
  const router = useRouter();
  const { dispatchUser } = useUserContext();
  const [showSignupForm, setShowSignupForm] = useState(false);
  const [tenantName, setTenantName] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);

  // If already authenticated, redirect to tenant selection or home
  useEffect(() => {
    if (!isLoading && user) {
      router.push('/tenant/select');
    }
  }, [user, isLoading, router]);

  // Show loading state while checking auth
  if (isLoading) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <Image src="/icon.png" alt="Loading" width={200} height={200} />
      </div>
    );
  }

  // Already authenticated
  if (user) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <p className="text-blue-700">Redirecting...</p>
      </div>
    );
  }

  const handleSignup = () => {
    if (!tenantName.trim()) return;
    
    setIsSubmitting(true);
    // Store tenant name in context
    dispatchUser({
      type: SET_TENANT_NAME,
      payload: tenantName.trim(),
    });
    
    // Redirect to Auth0 signup
    router.push('/auth/login?screen_hint=signup');
  };

  // Unauthenticated: show login UI
  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 to-blue-100 flex items-center justify-center p-4">
      <div className="flex flex-col items-center gap-4 max-w-md w-full">
        <div className="text-center mb-4">
          <h1 className="text-4xl font-bold text-transparent bg-clip-text bg-gradient-to-r from-lime-500 via-yellow-400 to-lime-500 mb-2">
            PinaColada
          </h1>
          <p className="text-lg text-blue-700">
            Software and AI Solutions
          </p>
        </div>

        <div className="w-full bg-white rounded-2xl shadow-xl p-8 border border-blue-100">
          <div className="text-center mb-6">
            <h2 className="text-2xl font-semibold text-blue-900 mb-2">
              Welcome Back
            </h2>
            <p className="text-blue-600">Sign in to continue to your dashboard</p>
          </div>

          {showSignupForm ? (
            <div className="flex flex-col gap-4">
              <div>
                <label htmlFor="orgName" className="block text-sm font-medium text-zinc-700 mb-2">
                  Organization Name
                </label>
                <input
                  id="orgName"
                  type="text"
                  placeholder="Enter your organization name"
                  value={tenantName}
                  onChange={(e) => setTenantName(e.target.value)}
                  required
                  className="w-full px-4 py-3 border border-zinc-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-lime-500 focus:border-transparent"
                />
              </div>
              <div className="flex gap-4">
                <button
                  onClick={() => {
                    setShowSignupForm(false);
                    setTenantName('');
                  }}
                  className="flex-1 px-6 py-3 text-sm font-semibold text-blue-700 bg-blue-50 border border-blue-200 rounded-lg hover:bg-blue-100 hover:border-blue-300 transition-colors"
                >
                  Cancel
                </button>
                <button
                  onClick={handleSignup}
                  disabled={!tenantName.trim() || isSubmitting}
                  className="flex-1 inline-flex items-center justify-center rounded-lg bg-gradient-to-r from-lime-500 to-yellow-400 px-6 py-3 text-sm font-semibold text-blue-900 hover:brightness-95 disabled:opacity-50 disabled:cursor-not-allowed transition-all"
                >
                  {isSubmitting ? 'Loading...' : 'Continue'}
                </button>
              </div>
            </div>
          ) : (
            <div className="flex flex-col gap-4">
              {/* Login button */}
              <a
                href="/auth/login"
                className="inline-flex items-center justify-center gap-2 rounded-lg bg-gradient-to-r from-lime-500 to-yellow-400 px-6 py-3 text-sm font-semibold text-blue-900 hover:brightness-95 transition-all"
              >
                <ChevronLeft size={18} />
                Log in
              </a>

              {/* Sign up button */}
              <button
                onClick={() => setShowSignupForm(true)}
                className="inline-flex items-center justify-center gap-2 px-6 py-3 text-sm font-semibold text-blue-700 bg-blue-50 border-2 border-blue-200 rounded-lg hover:border-blue-300 hover:bg-blue-100 transition-colors"
              >
                Sign up
                <ChevronRight size={18} />
              </button>
            </div>
          )}

          <div className="mt-6 pt-6 border-t border-blue-100">
            <p className="text-sm text-center text-blue-600">
              By continuing, you agree to our Terms of Service and Privacy Policy
            </p>
          </div>
        </div>

        {/* Footer */}
        <div className="mt-8 text-center">
          <p className="text-sm text-blue-600">
            © {new Date().getFullYear()} PinaColada.co. All rights reserved.
          </p>
        </div>
      </div>
    </div>
  );
};

export default LoginPage;
