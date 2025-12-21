"use client";

import React, { useState, useEffect, useCallback } from "react";
import { Paper, Stack, Group, Table, Text, Checkbox, Loader, Alert, Badge } from "@mantine/core";
import { getPermissions, getAllRolePermissions, updateRolePermissions } from "../../api";
import { useRoles } from "./RolesContext";

const PermissionMatrix = () => {
  const { roles, loading: rolesLoading, error: rolesError } = useRoles();
  const [permissions, setPermissions] = useState([]);
  const [rolePermissions, setRolePermissions] = useState({});
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [updating, setUpdating] = useState(null);

  const loadData = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const [permsData, allRolePerms] = await Promise.all([
        getPermissions(),
        getAllRolePermissions(),
      ]);
      setPermissions(permsData);

      const permMap = {};
      for (const roleId of Object.keys(allRolePerms)) {
        permMap[roleId] = new Set(allRolePerms[roleId] || []);
      }
      setRolePermissions(permMap);
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

  const hasPermission = (roleId, permissionId) => {
    return rolePermissions[roleId]?.has(permissionId) || false;
  };

  const togglePermission = async (role, permission) => {
    if (role.is_system) return;

    const key = `${role.id}-${permission.id}`;
    setUpdating(key);

    try {
      const currentPerms = rolePermissions[role.id] || new Set();
      const newPerms = new Set(currentPerms);

      if (newPerms.has(permission.id)) {
        newPerms.delete(permission.id);
      } else {
        newPerms.add(permission.id);
      }

      await updateRolePermissions(role.id, Array.from(newPerms));

      setRolePermissions((prev) => ({
        ...prev,
        [role.id]: newPerms,
      }));
    } catch (e) {
      setError(e.message);
    } finally {
      setUpdating(null);
    }
  };

  // Group permissions by resource
  const groupedPermissions = permissions.reduce((acc, perm) => {
    if (!acc[perm.resource]) {
      acc[perm.resource] = [];
    }
    acc[perm.resource].push(perm);
    return acc;
  }, {});

  if (isLoading) {
    return (
      <Paper withBorder p="lg">
        <Group justify="center">
          <Loader size="sm" color="lime" />
          <Text>Loading permissions...</Text>
        </Group>
      </Paper>
    );
  }

  return (
    <Paper withBorder p="lg">
      <Stack gap="md">
        <Text fw={500}>Permission Matrix</Text>
        <Text size="sm" c="dimmed">
          Check/uncheck to assign permissions to roles. System roles cannot be modified.
        </Text>

        {displayError && (
          <Alert color="red" onClose={() => setError(null)} withCloseButton>
            {displayError}
          </Alert>
        )}

        <div style={{ overflowX: "auto" }}>
          <Table withColumnBorders>
            <Table.Thead>
              <Table.Tr>
                <Table.Th>Permission</Table.Th>
                {roles.map((role) => (
                  <Table.Th key={role.id} ta="center">
                    <Stack gap={2} align="center">
                      <Text size="sm">{role.name}</Text>
                      {role.is_system && (
                        <Badge size="xs" variant="light" color="blue">
                          System
                        </Badge>
                      )}
                    </Stack>
                  </Table.Th>
                ))}
              </Table.Tr>
            </Table.Thead>
            <Table.Tbody>
              {Object.entries(groupedPermissions).map(([resource, perms]) => (
                <React.Fragment key={resource}>
                  <Table.Tr style={{ backgroundColor: "var(--mantine-color-gray-0)" }}>
                    <Table.Td colSpan={roles.length + 1}>
                      <Text size="sm" fw={600} tt="capitalize">
                        {resource}
                      </Text>
                    </Table.Td>
                  </Table.Tr>
                  {perms.map((perm) => (
                    <Table.Tr key={perm.id}>
                      <Table.Td>
                        <Text size="sm">{perm.action}</Text>
                      </Table.Td>
                      {roles.map((role) => {
                        const key = `${role.id}-${perm.id}`;
                        return (
                          <Table.Td key={role.id} ta="center">
                            <Checkbox
                              checked={hasPermission(role.id, perm.id)}
                              onChange={() => togglePermission(role, perm)}
                              disabled={role.is_system || updating === key}
                            />
                          </Table.Td>
                        );
                      })}
                    </Table.Tr>
                  ))}
                </React.Fragment>
              ))}
            </Table.Tbody>
          </Table>
        </div>
      </Stack>
    </Paper>
  );
};

export default PermissionMatrix;
