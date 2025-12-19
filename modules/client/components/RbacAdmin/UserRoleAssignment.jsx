"use client";

import { useState, useEffect, useCallback } from "react";
import { Paper, Stack, Group, Table, Text, Select, Loader, Alert } from "@mantine/core";
import { getRoles, getUserRoles, updateUserRole } from "../../api";

const UserRoleAssignment = () => {
  const [users, setUsers] = useState([]);
  const [roles, setRoles] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [updating, setUpdating] = useState(null);

  const loadData = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const [usersData, rolesData] = await Promise.all([getUserRoles(), getRoles()]);
      setUsers(usersData);
      setRoles(rolesData);
    } catch (e) {
      setError(e.message);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    loadData();
  }, [loadData]);

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

  if (loading) {
    return (
      <Paper withBorder p="lg">
        <Group justify="center">
          <Loader size="sm" />
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

        {error && (
          <Alert color="red" onClose={() => setError(null)} withCloseButton>
            {error}
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
