"use client";

import { Stack, Title, Text, Tabs } from "@mantine/core";
import { Users, Shield, Key } from "lucide-react";
import RoleList from "./RoleList";
import PermissionMatrix from "./PermissionMatrix";
import UserRoleAssignment from "./UserRoleAssignment";

const RbacAdmin = () => {
  return (
    <Stack gap="md">
      <div>
        <Title order={4}>Access Control</Title>
        <Text size="sm" c="dimmed">
          Manage roles, permissions, and user assignments.
        </Text>
      </div>

      <Tabs defaultValue="roles">
        <Tabs.List>
          <Tabs.Tab value="roles" leftSection={<Shield size={16} />}>
            Roles
          </Tabs.Tab>
          <Tabs.Tab value="permissions" leftSection={<Key size={16} />}>
            Permissions
          </Tabs.Tab>
          <Tabs.Tab value="users" leftSection={<Users size={16} />}>
            User Roles
          </Tabs.Tab>
        </Tabs.List>

        <Tabs.Panel value="roles" pt="md">
          <RoleList />
        </Tabs.Panel>

        <Tabs.Panel value="permissions" pt="md">
          <PermissionMatrix />
        </Tabs.Panel>

        <Tabs.Panel value="users" pt="md">
          <UserRoleAssignment />
        </Tabs.Panel>
      </Tabs>
    </Stack>
  );
};

export default RbacAdmin;
