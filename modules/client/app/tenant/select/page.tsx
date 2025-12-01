"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { useUser } from "@auth0/nextjs-auth0/client";
import {
  Container,
  Title,
  TextInput,
  Button,
  Paper,
  Text,
  Select,
} from "@mantine/core";
import { fetchBearerToken } from "../../../lib/fetch-bearer-token";
import { useUserContext } from "../../../context/userContext";
import { usePageLoading } from "../../../context/pageLoadingContext";
import Header from "../../../components/Header/Header";
import {
  SET_USER,
  SET_BEARER_TOKEN,
  SET_TENANT_NAME,
} from "../../../reducers/userReducer";
import { createTenant, getTimezones, updateUserPreferences, TimezoneOption } from "../../../api";

const TenantSelectPage = () => {
  const { user, isLoading: userLoading } = useUser();
  const router = useRouter();
  const { userState, dispatchUser } = useUserContext();
  const { dispatchPageLoading } = usePageLoading();
  const [loading, setLoading] = useState(false);
  const [tenantName, setTenantName] = useState("");
  const [plan, setPlan] = useState("free");
  const [timezone, setTimezone] = useState("America/New_York");
  const [timezoneOptions, setTimezoneOptions] = useState<TimezoneOption[]>([]);
  const [error, setError] = useState("");

  // Update page loading state
  useEffect(() => {
    dispatchPageLoading({ type: "SET_PAGE_LOADING", payload: userLoading });
  }, [userLoading, dispatchPageLoading]);

  useEffect(() => {
    if (!userLoading && !user) {
      router.push("/auth/login");
      return;
    }

    if (!user) return;

    const initializeAuth = () => {
      // Store Auth0 user and fetch bearer token
      if (!userState.user) {
        dispatchUser({
          type: SET_USER,
          payload: user,
        });
      }

      if (!userState.bearerToken) {
        fetchBearerToken()
          .then((bearerTokenData) => {
            dispatchUser({
              type: SET_BEARER_TOKEN,
              payload: bearerTokenData.headers.Authorization.replace(
                "Bearer ",
                ""
              ),
            });
          })
          .catch((error) => {
            console.error("Error initializing auth:", error);
          });
      }
    };

    initializeAuth();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [user, userLoading]);

  // Fetch timezone options
  useEffect(() => {
    getTimezones()
      .then((options) => setTimezoneOptions(options))
      .catch((err) => console.error("Failed to fetch timezones:", err));
  }, []);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    setError("");

    if (!tenantName.trim()) {
      setError("Please enter an organization name");
      return;
    }

    setLoading(true);
    createTenant(tenantName, plan)
      .then((newTenant) => {
        // Store tenant in context
        dispatchUser({
          type: SET_TENANT_NAME,
          payload: newTenant.name,
        });

        // Save user's timezone preference
        return updateUserPreferences({ timezone }).then(() => {
          router.push("/chat");
        });
      })
      .catch((error: any) => {
        console.error("Error creating tenant:", error);
        setError(
          error.response?.data?.message || "Failed to create organization"
        );
      })
      .finally(() => {
        setLoading(false);
      });
  };

  return (
    <>
      <Header />
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
            disabled={loading}
            mb="md"
            required
          />
          <Select
            label="Plan"
            value={plan}
            onChange={(value) => setPlan(value || "free")}
            data={[
              { value: "free", label: "Free" },
              // todo, implement these when we have a marktable product ~
              // { value: 'starter', label: 'Starter (Coming Soon)', disabled: true },
              // { value: 'professional', label: 'Professional (Coming Soon)', disabled: true },
              // { value: 'enterprise', label: 'Enterprise (Coming Soon)', disabled: true },
            ]}
            disabled={loading}
            mb="md"
          />
          <Select
            label="Timezone"
            value={timezone}
            onChange={(value) => setTimezone(value || "America/New_York")}
            data={timezoneOptions}
            disabled={loading}
            searchable
            mb="md"
          />
          {error && (
            <Text c="red" size="sm" mb="md">
              {error}
            </Text>
          )}
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
    </>
  );
};

export default TenantSelectPage;
