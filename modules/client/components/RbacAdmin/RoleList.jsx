"use client";

import { useState, useEffect, useCallback } from "react";
import {
  Paper,
  Stack,
  Group,
  Button,
  Table,
  Text,
  Badge,
  Loader,
  Alert,
  ActionIcon,
  Modal,
  TextInput,
  Textarea,
} from "@mantine/core";
import { Plus, Pencil, Trash2 } from "lucide-react";
import { getRoles, createRole, updateRole, deleteRole } from "../../api";

const RoleList = () => {
  const [roles, setRoles] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [editingRole, setEditingRole] = useState(null);
  const [isCreating, setIsCreating] = useState(false);
  const [formData, setFormData] = useState({ name: "", description: "" });
  const [saving, setSaving] = useState(false);

  const loadRoles = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await getRoles();
      setRoles(data);
    } catch (e) {
      setError(e.message);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    loadRoles();
  }, [loadRoles]);

  const openCreate = () => {
    setFormData({ name: "", description: "" });
    setIsCreating(true);
    setEditingRole(null);
  };

  const openEdit = (role) => {
    setFormData({ name: role.name, description: role.description || "" });
    setEditingRole(role);
    setIsCreating(false);
  };

  const closeModal = () => {
    setEditingRole(null);
    setIsCreating(false);
    setFormData({ name: "", description: "" });
  };

  const handleSave = async () => {
    if (!formData.name.trim()) return;
    setSaving(true);
    try {
      const payload = {
        name: formData.name.trim(),
        description: formData.description.trim() || null,
      };

      if (isCreating) {
        await createRole(payload);
      } else if (editingRole) {
        await updateRole(editingRole.id, payload);
      }

      closeModal();
      loadRoles();
    } catch (e) {
      setError(e.message);
    } finally {
      setSaving(false);
    }
  };

  const handleDelete = async (role) => {
    if (!confirm(`Delete role "${role.name}"? This cannot be undone.`)) return;
    try {
      await deleteRole(role.id);
      loadRoles();
    } catch (e) {
      setError(e.message);
    }
  };

  if (loading) {
    return (
      <Paper withBorder p="lg">
        <Group justify="center">
          <Loader size="sm" />
          <Text>Loading roles...</Text>
        </Group>
      </Paper>
    );
  }

  return (
    <Paper withBorder p="lg">
      <Stack gap="md">
        <Group justify="space-between">
          <Text fw={500}>Roles</Text>
          <Button size="xs" leftSection={<Plus size={14} />} onClick={openCreate}>
            New Role
          </Button>
        </Group>

        {error && (
          <Alert color="red" onClose={() => setError(null)} withCloseButton>
            {error}
          </Alert>
        )}

        <Table highlightOnHover>
          <Table.Thead>
            <Table.Tr>
              <Table.Th>Name</Table.Th>
              <Table.Th>Description</Table.Th>
              <Table.Th>Type</Table.Th>
              <Table.Th style={{ width: 100 }}>Actions</Table.Th>
            </Table.Tr>
          </Table.Thead>
          <Table.Tbody>
            {roles.map((role) => (
              <Table.Tr key={role.id}>
                <Table.Td>{role.name}</Table.Td>
                <Table.Td>
                  <Text size="sm" c="dimmed" lineClamp={1}>
                    {role.description || "-"}
                  </Text>
                </Table.Td>
                <Table.Td>
                  <Badge variant="light" color={role.is_system ? "blue" : "gray"}>
                    {role.is_system ? "System" : "Custom"}
                  </Badge>
                </Table.Td>
                <Table.Td>
                  {!role.is_system && (
                    <Group gap="xs">
                      <ActionIcon variant="subtle" size="sm" onClick={() => openEdit(role)}>
                        <Pencil size={14} />
                      </ActionIcon>
                      <ActionIcon
                        variant="subtle"
                        size="sm"
                        color="red"
                        onClick={() => handleDelete(role)}
                      >
                        <Trash2 size={14} />
                      </ActionIcon>
                    </Group>
                  )}
                </Table.Td>
              </Table.Tr>
            ))}
          </Table.Tbody>
        </Table>
      </Stack>

      <Modal
        opened={isCreating || !!editingRole}
        onClose={closeModal}
        title={isCreating ? "Create Role" : "Edit Role"}
      >
        <Stack gap="md">
          <TextInput
            label="Name"
            placeholder="Role name"
            value={formData.name}
            onChange={(e) => setFormData((prev) => ({ ...prev, name: e.target.value }))}
            required
          />
          <Textarea
            label="Description"
            placeholder="Optional description"
            value={formData.description}
            onChange={(e) => setFormData((prev) => ({ ...prev, description: e.target.value }))}
            minRows={2}
          />
          <Group justify="flex-end" gap="xs">
            <Button variant="default" onClick={closeModal}>
              Cancel
            </Button>
            <Button onClick={handleSave} loading={saving} disabled={!formData.name.trim()}>
              {isCreating ? "Create" : "Save"}
            </Button>
          </Group>
        </Stack>
      </Modal>
    </Paper>
  );
};

export default RoleList;
