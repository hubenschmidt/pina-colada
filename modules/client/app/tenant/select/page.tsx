'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { useUser } from '@auth0/nextjs-auth0/client';
import { Container, Title, Loader } from '@mantine/core';
import TenantSelector from '../../../components/auth/TenantSelector';
import { fetchBearerToken } from '../../../lib/fetch-bearer-token';
import axios from 'axios';
import { env } from 'next-runtime-env';

const TenantSelectPage = () => {
  const { user, isLoading: userLoading } = useUser();
  const router = useRouter();
  const [tenants, setTenants] = useState([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (!userLoading && !user) {
      router.push('/auth/login');
      return;
    }

    if (!user) return;

    const loadTenants = async () => {
      try {
        const authHeaders = await fetchBearerToken();
        const response = await axios.get(
          `${env('NEXT_PUBLIC_API_URL')}/api/auth/me`,
          authHeaders
        );

        const userTenants = response.data.tenants || [];
        setTenants(userTenants);

        // Auto-select if only one tenant
        if (userTenants.length === 1) {
          handleSelectTenant(userTenants[0].id);
        }
      } catch (error) {
        console.error('Error loading tenants:', error);
      } finally {
        setLoading(false);
      }
    };

    loadTenants();
  }, [user, userLoading, router]);

  const handleSelectTenant = async (tenantId: number) => {
    try {
      const authHeaders = await fetchBearerToken();
      await axios.post(
        `${env('NEXT_PUBLIC_API_URL')}/api/auth/tenant/switch`,
        { tenant_id: tenantId },
        authHeaders
      );

      // Store tenant in cookie/localStorage
      document.cookie = `tenant_id=${tenantId}; path=/`;

      router.push('/');
    } catch (error) {
      console.error('Error switching tenant:', error);
    }
  };

  const handleCreateTenant = async (name: string) => {
    try {
      const authHeaders = await fetchBearerToken();
      const response = await axios.post(
        `${env('NEXT_PUBLIC_API_URL')}/api/auth/tenant/create`,
        { name },
        authHeaders
      );

      const newTenant = response.data.tenant;
      setTenants([...tenants, newTenant]);
      handleSelectTenant(newTenant.id);
    } catch (error) {
      console.error('Error creating tenant:', error);
    }
  };

  if (userLoading || loading) {
    return (
      <Container size="sm" className="flex items-center justify-center min-h-screen">
        <Loader size="lg" />
      </Container>
    );
  }

  return (
    <Container size="sm" className="py-12">
      <Title order={1} mb="xl" className="text-center">
        Welcome to PinaColada.co
      </Title>
      <TenantSelector
        tenants={tenants}
        onSelectTenant={handleSelectTenant}
        onCreateTenant={handleCreateTenant}
      />
    </Container>
  );
};

export default TenantSelectPage;
