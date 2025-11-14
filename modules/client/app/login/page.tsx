'use client';

import { useUser } from '@auth0/nextjs-auth0/client';
import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { Flex, Loader, Box, Button, Text, TextInput } from '@mantine/core';
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
      <Flex style={{ height: '100vh' }} justify="center" align="center">
        <Loader size="xl" color="rgb(59, 130, 246)" />
      </Flex>
    );
  }

  // Already authenticated
  if (user) {
    return (
      <Flex style={{ height: '100vh' }} justify="center" align="center">
        <Text>Redirecting...</Text>
      </Flex>
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
      <Flex direction="column" align="center" gap="md" className="max-w-md w-full">
        <Box className="text-center mb-4">
          <Text
            className="text-4xl font-bold bg-gradient-to-r from-blue-600 to-blue-800 bg-clip-text text-transparent mb-2"
          >
            PinaColada.co
          </Text>
          <Text className="text-lg text-blue-700">
            Software and AI Solutions
          </Text>
        </Box>

        <Box className="w-full bg-white rounded-2xl shadow-xl p-8 border border-blue-100">
          <div className="text-center mb-6">
            <h2 className="text-2xl font-semibold text-blue-900 mb-2">
              Welcome Back
            </h2>
            <p className="text-blue-600">Sign in to continue to your dashboard</p>
          </div>

          {showSignupForm ? (
            <Flex direction="column" gap="md">
              <TextInput
                label="Organization Name"
                placeholder="Enter your organization name"
                value={tenantName}
                onChange={(e) => setTenantName(e.target.value)}
                required
                size="lg"
              />
              <Flex gap="md">
                <Button
                  onClick={() => {
                    setShowSignupForm(false);
                    setTenantName('');
                  }}
                  variant="light"
                  size="lg"
                  style={{ flex: 1 }}
                >
                  Cancel
                </Button>
                <Button
                  onClick={handleSignup}
                  size="lg"
                  disabled={!tenantName.trim() || isSubmitting}
                  className="bg-gradient-to-r from-blue-600 to-blue-700 hover:from-blue-700 hover:to-blue-800 text-white border-0"
                  style={{ flex: 1 }}
                  loading={isSubmitting}
                >
                  Continue
                </Button>
              </Flex>
            </Flex>
          ) : (
            <Flex direction="column" gap="md">
              {/* Login button */}
              <Button
                component="a"
                href="/auth/login"
                size="lg"
                className="bg-gradient-to-r from-blue-600 to-blue-700 hover:from-blue-700 hover:to-blue-800 text-white border-0"
                leftSection={<ChevronLeft size={18} />}
              >
                Log in
              </Button>

              {/* Sign up button */}
              <Button
                onClick={() => setShowSignupForm(true)}
                size="lg"
                variant="light"
                className="text-blue-700 border-2 border-blue-200 hover:border-blue-300 hover:bg-blue-50"
                rightSection={<ChevronRight size={18} />}
              >
                Sign up
              </Button>
            </Flex>
          )}

          <div className="mt-6 pt-6 border-t border-blue-100">
            <p className="text-sm text-center text-blue-600">
              By continuing, you agree to our Terms of Service and Privacy Policy
            </p>
          </div>
        </Box>

        {/* Footer */}
        <Box className="mt-8 text-center">
          <Text className="text-sm text-blue-600">
            © {new Date().getFullYear()} PinaColada.co. All rights reserved.
          </Text>
        </Box>
      </Flex>
    </div>
  );
};

export default LoginPage;
