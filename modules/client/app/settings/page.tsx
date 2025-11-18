"use client";

import { useUserContext } from "../../context/userContext";
import { useState, useEffect } from "react";
import { SET_THEME } from "../../reducers/userReducer";
import { Container, Stack, Title, Radio, Paper, Text, Loader, Center } from "@mantine/core";

export default function SettingsPage() {
  const { userState, dispatchUser } = useUserContext();
  const [userTheme, setUserTheme] = useState<string | null>(null);
  const [tenantTheme, setTenantTheme] = useState<string>("light");
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchPrefs = async () => {
      if (!userState.isAuthed) return;

      try {
        const userPrefs = await fetch("/api/preferences/user");
        if (userPrefs.ok) {
          const data = await userPrefs.json();
          setUserTheme(data.theme);
        }

        if (userState.canEditTenantTheme) {
          const tenantPrefs = await fetch("/api/preferences/tenant");
          if (tenantPrefs.ok) {
            const data = await tenantPrefs.json();
            setTenantTheme(data.theme);
          }
        }
      } catch (error) {
        console.error("Failed to fetch preferences:", error);
      }

      setLoading(false);
    };

    fetchPrefs();
  }, [userState.isAuthed, userState.canEditTenantTheme]);

  const updateUserTheme = async (newTheme: "light" | "dark" | null) => {
    if (!userState.isAuthed) return;

    try {
      const res = await fetch("/api/preferences/user", {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ theme: newTheme }),
      });

      if (res.ok) {
        const data = await res.json();
        setUserTheme(newTheme);
        dispatchUser({
          type: SET_THEME,
          payload: {
            theme: data.effective_theme,
            canEditTenant: userState.canEditTenantTheme,
          },
        });
      }
    } catch (error) {
      console.error("Failed to update user theme:", error);
    }
  };

  const updateTenantTheme = async (newTheme: "light" | "dark") => {
    if (!userState.isAuthed || !userState.canEditTenantTheme) return;

    try {
      const res = await fetch("/api/preferences/tenant", {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ theme: newTheme }),
      });

      if (res.ok) {
        setTenantTheme(newTheme);

        if (userTheme === null) {
          dispatchUser({
            type: SET_THEME,
            payload: {
              theme: newTheme,
              canEditTenant: userState.canEditTenantTheme,
            },
          });
        }
      }
    } catch (error) {
      console.error("Failed to update tenant theme:", error);
    }
  };

  if (!userState.isAuthed) {
    return (
      <Container size="sm" py="xl">
        <Text>Please log in to access settings.</Text>
      </Container>
    );
  }

  if (loading) {
    return (
      <Center h="50vh">
        <Loader size="lg" />
      </Center>
    );
  }

  return (
    <Container size="sm" py="xl">
      <Stack gap="xl">
        <Title order={1}>Settings</Title>

        <Stack gap="md">
          <Title order={2} size="h3">Personal Theme</Title>
          <Radio.Group value={userTheme || ""} onChange={(value) => updateUserTheme(value as "light" | "dark" | null)}>
            <Stack gap="xs">
              <Radio value="light" label="Light" />
              <Radio value="dark" label="Dark" />
              <Radio value="" label="Inherit from organization" />
            </Stack>
          </Radio.Group>
        </Stack>

        {userState.canEditTenantTheme && (
          <Paper withBorder p="md">
            <Stack gap="md">
              <div>
                <Title order={2} size="h3" mb="xs">Organization Theme (Admin)</Title>
                <Text size="sm" c="dimmed">
                  This affects all users who inherit the organization theme.
                </Text>
              </div>
              <Radio.Group value={tenantTheme} onChange={(value) => updateTenantTheme(value as "light" | "dark")}>
                <Stack gap="xs">
                  <Radio value="light" label="Light" />
                  <Radio value="dark" label="Dark" />
                </Stack>
              </Radio.Group>
            </Stack>
          </Paper>
        )}
      </Stack>
    </Container>
  );
}
