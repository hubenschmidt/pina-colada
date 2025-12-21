"use client";

import { useState, useEffect, useCallback } from "react";
import { Paper, Stack, Group, Table, Text, Select, Loader, Alert } from "@mantine/core";
import { getUserRoles, updateUserRole } from "../../api";
import { useRoles } from "./RolesContext";

const UserRoleAssignment = () => {
  const { roles, loading: rolesLoading, error: rolesError } = useRoles();
  const [users, setUsers] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [updating, setUpdating] = useState(null);

  const loadData = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const usersData = await getUserRoles();
      setUsers(usersData);
    } catch (e) {
      setError(e.message);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    loadData();
  }, [loadData]);

  const isLoading = loading || rolesLoading;
  const displayError = error || rolesError;

  const handleRoleChange = async (userId, roleId) => {
    if (!roleId) return;
    setUpdating(userId);
    try {
      await updateUserRole(userId, parseInt(roleId, 10));
      setUsers((prev) =>
        prev.map((u) => {
          if (u.user_id !== userId) return u;
          const role = roles.find((r) => r.id === parseInt(roleId, 10));
          return { ...u, role_id: parseInt(roleId, 10), role_name: role?.name || null };
        })
      );
    } catch (e) {
      setError(e.message);
    } finally {
      setUpdating(null);
    }
  };

  const roleOptions = roles.map((r) => ({
    value: String(r.id),
    label: r.name + (r.is_system ? " (System)" : ""),
  }));

  if (isLoading) {
    return (
      <Paper withBorder p="lg">
        <Group justify="center">
          <Loader size="sm" color="lime" />
          <Text>Loading users...</Text>
        </Group>
      </Paper>
    );
  }

  return (
    <Paper withBorder p="lg">
      <Stack gap="md">
        <Text fw={500}>User Role Assignments</Text>
        <Text size="sm" c="dimmed">
          Assign roles to users in your organization.
        </Text>

        {displayError && (
          <Alert color="red" onClose={() => setError(null)} withCloseButton>
            {displayError}
          </Alert>
        )}

        <Table highlightOnHover>
          <Table.Thead>
            <Table.Tr>
              <Table.Th>First Name</Table.Th>
              <Table.Th>Last Name</Table.Th>
              <Table.Th>Email</Table.Th>
              <Table.Th>Role</Table.Th>
            </Table.Tr>
          </Table.Thead>
          <Table.Tbody>
            {users.map((user, idx) => (
              <Table.Tr key={`${user.user_id}-${idx}`}>
                <Table.Td>{user.first_name || "-"}</Table.Td>
                <Table.Td>{user.last_name || "-"}</Table.Td>
                <Table.Td>{user.email}</Table.Td>
                <Table.Td>
                  <Select
                    data={roleOptions}
                    value={user.role_id ? String(user.role_id) : null}
                    onChange={(value) => handleRoleChange(user.user_id, value)}
                    placeholder="Select role"
                    disabled={updating === user.user_id}
                    size="xs"
                    style={{ width: 200 }}
                  />
                </Table.Td>
              </Table.Tr>
            ))}
          </Table.Tbody>
        </Table>

        {users.length === 0 && (
          <Text c="dimmed" ta="center" py="md">
            No users found
          </Text>
        )}
      </Stack>
    </Paper>
  );
};

export default UserRoleAssignment;
