"use client";

import Link from "next/link";
import { useRouter, usePathname } from "next/navigation";
import { Container, Stack, Title, Text, Tabs } from "@mantine/core";
import { Users, Shield, Key, ArrowLeft } from "lucide-react";
import { useUserContext } from "../../../context/userContext";
import { useEffect } from "react";
import { RolesProvider } from "../../../components/RbacAdmin/RolesContext";

const AccessControlLayout = ({ children }) => {
  const router = useRouter();
  const pathname = usePathname();
  const { userState } = useUserContext();

  // Redirect if user doesn't have permission
  useEffect(() => {
    if (!userState.isLoading && userState.isAuthed && !userState.canEditTenantTheme) {
      router.push("/settings");
    }
  }, [userState.isLoading, userState.isAuthed, userState.canEditTenantTheme, router]);

  // Determine active tab from pathname
  const getActiveTab = () => {
    if (pathname.includes("/user-roles")) return "user-roles";
    if (pathname.includes("/permissions")) return "permissions";
    return "roles";
  };

  const handleTabClick = (tab) => {
    router.push(`/settings/access-control/${tab}`);
  };

  if (userState.isLoading || !userState.canEditTenantTheme) {
    return null;
  }

  return (
    <RolesProvider>
      <Container size="lg" py="xl">
        <Stack gap="xl">
          <div>
            <Link
              href="/settings"
              style={{ textDecoration: "none" }}
            >
              <Text
                size="sm"
                c="dimmed"
                mb="xs"
                style={{ cursor: "pointer", display: "inline-flex", alignItems: "center", gap: 4 }}
              >
                <ArrowLeft size={14} /> Back to Settings
              </Text>
            </Link>
            <Title order={1}>Settings</Title>
          </div>

          <Stack gap="md">
            <div>
              <Title order={4}>Access Control</Title>
              <Text size="sm" c="dimmed">
                Manage roles, permissions, and user assignments.
              </Text>
            </div>

            <Tabs value={getActiveTab()}>
              <Tabs.List>
                <Tabs.Tab
                  value="roles"
                  leftSection={<Shield size={16} />}
                  onClick={() => handleTabClick("roles")}
                >
                  Roles
                </Tabs.Tab>
                <Tabs.Tab
                  value="permissions"
                  leftSection={<Key size={16} />}
                  onClick={() => handleTabClick("permissions")}
                >
                  Permissions
                </Tabs.Tab>
                <Tabs.Tab
                  value="user-roles"
                  leftSection={<Users size={16} />}
                  onClick={() => handleTabClick("user-roles")}
                >
                  User Roles
                </Tabs.Tab>
              </Tabs.List>

              {children}
            </Tabs>
          </Stack>
        </Stack>
      </Container>
    </RolesProvider>
  );
};

export default AccessControlLayout;
