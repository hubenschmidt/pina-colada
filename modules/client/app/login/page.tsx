'use client';

import { useUser } from '@auth0/nextjs-auth0/client';
import { useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { Flex, Loader, Box, Button, Text } from '@mantine/core';
import { ChevronRight, ChevronLeft } from 'lucide-react';
import Image from 'next/image';

const LoginPage = () => {
  const { user, isLoading } = useUser();
  const router = useRouter();

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

          <Flex direction="column" gap="md">
            {/* Login button */}
            <Button
              component="a"
              href="/api/auth/login"
              size="lg"
              className="bg-gradient-to-r from-blue-600 to-blue-700 hover:from-blue-700 hover:to-blue-800 text-white border-0"
              leftSection={<ChevronLeft size={18} />}
            >
              Log in
            </Button>

            {/* Sign up button */}
            <Button
              component="a"
              href="/api/auth/login?screen_hint=signup"
              size="lg"
              variant="light"
              className="text-blue-700 border-2 border-blue-200 hover:border-blue-300 hover:bg-blue-50"
              rightSection={<ChevronRight size={18} />}
            >
              Sign up
            </Button>
          </Flex>

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
