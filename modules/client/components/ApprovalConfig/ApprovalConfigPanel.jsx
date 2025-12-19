"use client";

import { useState, useEffect, useCallback } from "react";
import { Paper, Title, Stack, Group, Switch, Text, Loader, Alert } from "@mantine/core";
import { getApprovalConfig, updateApprovalConfig } from "../../api";

const entityLabels = {
  contact: "Contacts",
  account: "Accounts",
  organization: "Organizations",
  individual: "Individuals",
  job: "Job Leads",
  note: "Notes",
  task: "Tasks",
};

const ApprovalConfigPanel = () => {
  const [config, setConfig] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [updating, setUpdating] = useState(null);

  const loadConfig = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await getApprovalConfig();
      setConfig(data);
    } catch (e) {
      setError(e.message);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    loadConfig();
  }, [loadConfig]);

  const handleToggle = async (entityType, currentValue) => {
    setUpdating(entityType);
    try {
      await updateApprovalConfig(entityType, !currentValue);
      setConfig((prev) =>
        prev.map((c) =>
          c.entity_type === entityType ? { ...c, requires_approval: !currentValue } : c
        )
      );
    } catch (e) {
      setError(e.message);
    } finally {
      setUpdating(null);
    }
  };

  if (loading) {
    return (
      <Paper withBorder p="lg" mt="md">
        <Group justify="center">
          <Loader size="sm" />
          <Text>Loading configuration...</Text>
        </Group>
      </Paper>
    );
  }

  return (
    <Paper withBorder p="lg" mt="md">
      <Stack gap="md">
        <Title order={4}>Entity Approval Settings</Title>
        <Text size="sm" c="dimmed">
          Control which entity types require approval before agent changes take effect.
        </Text>

        {error && (
          <Alert color="red" onClose={() => setError(null)} withCloseButton>
            {error}
          </Alert>
        )}

        <Stack gap="xs">
          {config.map(({ entity_type, requires_approval }) => (
            <Group key={entity_type} justify="space-between" py="xs">
              <Text>{entityLabels[entity_type] || entity_type}</Text>
              <Switch
                checked={requires_approval}
                onChange={() => handleToggle(entity_type, requires_approval)}
                disabled={updating === entity_type}
                label={requires_approval ? "Requires approval" : "Auto-execute"}
              />
            </Group>
          ))}
        </Stack>
      </Stack>
    </Paper>
  );
};

export default ApprovalConfigPanel;
