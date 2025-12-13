"use client";

import { useUserContext } from "../../context/userContext";
import { useState, useEffect } from "react";
import { SET_THEME } from "../../reducers/userReducer";
import { Container, Stack, Title, Radio, Paper, Text, Loader, Center, Select } from "@mantine/core";
import {
  getUserPreferences,
  updateUserPreferences,
  getTenantPreferences,
  updateTenantPreferences,
  getTimezones,
} from "../../api";
import { usePageLoading } from "../../context/pageLoadingContext";

const SettingsPage = () => {
  const { userState, dispatchUser } = useUserContext();
  const [userTheme, setUserTheme] = useState(null);
  const [userTimezone, setUserTimezone] = useState("America/New_York");
  const [timezoneOptions, setTimezoneOptions] = useState([]);
  const [tenantTheme, setTenantTheme] = useState("light");
  const [loading, setLoading] = useState(true);
  const { dispatchPageLoading } = usePageLoading();

  useEffect(() => {
    dispatchPageLoading({ type: "SET_PAGE_LOADING", payload: false });
  }, [dispatchPageLoading]);

  useEffect(() => {
    const fetchPrefs = async () => {
      if (!userState.isAuthed) return;

      try {
        const [userPrefs, tzOptions] = await Promise.all([getUserPreferences(), getTimezones()]);
        setUserTheme(userPrefs.theme);
        setUserTimezone(userPrefs.timezone);
        setTimezoneOptions(tzOptions);

        if (userState.canEditTenantTheme) {
          const tenantPrefs = await getTenantPreferences();
          setTenantTheme(tenantPrefs.theme);
        }
      } catch (error) {
        console.error("Failed to fetch preferences:", error);
      }

      setLoading(false);
    };

    fetchPrefs();
  }, [userState.isAuthed, userState.canEditTenantTheme]);

  const updateUserTheme = (newTheme) => {
    if (!userState.isAuthed) return;

    updateUserPreferences({ theme: newTheme })
      .then((data) => {
        setUserTheme(newTheme);
        dispatchUser({
          type: SET_THEME,
          payload: {
            theme: data.effective_theme,
            canEditTenant: userState.canEditTenantTheme,
          },
        });
      })
      .catch((error) => {
        console.error("Failed to update user theme:", error);
      });
  };

  const updateTimezone = (newTimezone) => {
    if (!userState.isAuthed) return;

    updateUserPreferences({ timezone: newTimezone })
      .then(() => {
        setUserTimezone(newTimezone);
      })
      .catch((error) => {
        console.error("Failed to update timezone:", error);
      });
  };

  const updateTenantTheme = (newTheme) => {
    if (!userState.isAuthed || !userState.canEditTenantTheme) return;

    updateTenantPreferences(newTheme)
      .then(() => {
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
      })
      .catch((error) => {
        console.error("Failed to update tenant theme:", error);
      });
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
      <Center mih={400}>
        <Stack align="center" gap="md">
          <Loader size="xl" color="lime" />
          <Text c="dimmed">Loading settings...</Text>
        </Stack>
      </Center>
    );
  }

  return (
    <Container size="sm" py="xl">
      <Stack gap="xl">
        <Title order={1}>Settings</Title>

        <Stack gap="md">
          <Title order={2} size="h3">
            Personal Theme
          </Title>
          <Radio.Group
            value={userTheme === null ? "inherit" : userTheme}
            onChange={(value) => updateUserTheme(value === "inherit" ? null : value)}>
            <Stack gap="xs">
              <Radio value="inherit" label="Inherit from organization" />
              <Radio value="light" label="Light" />
              <Radio value="dark" label="Dark" />
            </Stack>
          </Radio.Group>
        </Stack>

        <Stack gap="md">
          <Title order={2} size="h3">
            Timezone
          </Title>
          <Select
            value={userTimezone}
            onChange={(value) => value && updateTimezone(value)}
            data={timezoneOptions}
            searchable
            maw={400}
          />
        </Stack>

        {userState.canEditTenantTheme && (
          <Paper withBorder p="md">
            <Stack gap="md">
              <div>
                <Title order={2} size="h3" mb="xs">
                  Organization Theme (Admin)
                </Title>
                <Text size="sm" c="dimmed">
                  This affects all users who inherit the organization theme.
                </Text>
              </div>
              <Radio.Group value={tenantTheme} onChange={(value) => updateTenantTheme(value)}>
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
};

export default SettingsPage;
