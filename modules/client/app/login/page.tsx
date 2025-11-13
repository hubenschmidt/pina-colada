'use client';

import { useEffect } from 'react';
import { useRouter } from 'next/navigation';

const LoginPage = () => {
  const router = useRouter();

  useEffect(() => {
    router.push('/api/auth/login');
  }, [router]);

  return (
    <div className="flex items-center justify-center min-h-screen">
      <div className="text-center">
        <h1 className="text-2xl font-bold mb-4">Redirecting to login...</h1>
      </div>
    </div>
  );
};

export default LoginPage;
