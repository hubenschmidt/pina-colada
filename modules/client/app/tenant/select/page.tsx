'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { useUser } from '@auth0/nextjs-auth0/client';
import { Container, Title, Loader, TextInput, Button, Paper, Text } from '@mantine/core';
import { fetchBearerToken } from '../../../lib/fetch-bearer-token';
import { useUserContext } from '../../../context/userContext';
import { SET_USER, SET_BEARER_TOKEN, SET_TENANT_NAME } from '../../../reducers/userReducer';
import axios from 'axios';
import { env } from 'next-runtime-env';

const TenantSelectPage = () => {
  const { user, isLoading: userLoading } = useUser();
  const router = useRouter();
  const { userState, dispatchUser } = useUserContext();
  const [loading, setLoading] = useState(false);
  const [tenantName, setTenantName] = useState('');
  const [error, setError] = useState('');

  useEffect(() => {
    if (!userLoading && !user) {
      router.push('/auth/login');
      return;
    }

    if (!user) return;

    const initializeAuth = async () => {
      try {
        // Store Auth0 user and fetch bearer token
        if (!userState.user) {
          dispatchUser({
            type: SET_USER,
            payload: user,
          });
        }

        if (!userState.bearerToken) {
          const bearerTokenData = await fetchBearerToken();
          dispatchUser({
            type: SET_BEARER_TOKEN,
            payload: bearerTokenData.headers.Authorization.replace('Bearer ', ''),
          });
        }
      } catch (error) {
        console.error('Error initializing auth:', error);
      }
    };

    initializeAuth();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [user, userLoading]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');

    if (!tenantName.trim()) {
      setError('Please enter an organization name');
      return;
    }

    setLoading(true);
    try {
      const authHeaders = await fetchBearerToken();
      const response = await axios.post(
        `${env('NEXT_PUBLIC_API_URL')}/api/auth/tenant/create`,
        { name: tenantName },
        authHeaders
      );

      const newTenant = response.data.tenant;

      // Store tenant in context
      dispatchUser({
        type: SET_TENANT_NAME,
        payload: newTenant.name,
      });

      // Switch to the new tenant
      await axios.post(
        `${env('NEXT_PUBLIC_API_URL')}/api/auth/tenant/switch`,
        { tenant_id: newTenant.id },
        authHeaders
      );

      // Store tenant in cookie
      document.cookie = `tenant_id=${newTenant.id}; path=/`;

      router.push('/');
    } catch (error: any) {
      console.error('Error creating tenant:', error);
      setError(error.response?.data?.message || 'Failed to create organization');
      setLoading(false);
    }
  };

  if (userLoading) {
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
      <Paper shadow="sm" p="xl" radius="md" withBorder>
        <Title order={2} mb="md" size="h3">
          Create Your Organization
        </Title>
        <Text size="sm" c="dimmed" mb="lg">
          Please enter a name for your organization to get started.
        </Text>
        <form onSubmit={handleSubmit}>
          <TextInput
            label="Organization Name"
            placeholder="Enter organization name"
            value={tenantName}
            onChange={(e) => setTenantName(e.currentTarget.value)}
            error={error}
            disabled={loading}
            mb="md"
            required
          />
          <Button
            type="submit"
            fullWidth
            loading={loading}
            disabled={!tenantName.trim()}
          >
            Create Organization
          </Button>
        </form>
      </Paper>
    </Container>
  );
};

export default TenantSelectPage;
