"use client";

import { useState } from "react";
import { Button, TextInput, Card, Title, Text } from "@mantine/core";

const TenantSelector = ({ tenants, onSelectTenant, onCreateTenant }) => {
  const [isCreating, setIsCreating] = useState(false);
  const [newTenantName, setNewTenantName] = useState("");

  const handleCreateTenant = () => {
    if (!newTenantName.trim()) return;

    onCreateTenant(newTenantName.trim());
    setNewTenantName("");
    setIsCreating(false);
  };

  if (isCreating) {
    return (
      <Card shadow="sm" padding="lg" radius="md">
        <Title order={3} mb="md">
          Create New Organization
        </Title>
        <TextInput
          label="Organization Name"
          placeholder="Enter organization name"
          value={newTenantName}
          onChange={(e) => setNewTenantName(e.currentTarget.value)}
          mb="md"
        />

        <div className="flex gap-2">
          <Button onClick={handleCreateTenant} disabled={!newTenantName.trim()}>
            Create
          </Button>
          <Button variant="outline" onClick={() => setIsCreating(false)}>
            Cancel
          </Button>
        </div>
      </Card>
    );
  }

  return (
    <div className="space-y-4">
      {tenants.length > 0 && (
        <Card shadow="sm" padding="lg" radius="md">
          <Title order={3} mb="md">
            Select Organization
          </Title>
          <div className="space-y-2">
            {tenants.map((tenant) => (
              <Card
                key={tenant.id}
                shadow="xs"
                padding="md"
                className="cursor-pointer hover:bg-gray-50"
                onClick={() => onSelectTenant(tenant.id)}
              >
                <Text fw={500}>{tenant.name}</Text>
                <Text size="sm" c="dimmed">
                  Role: {tenant.role}
                </Text>
              </Card>
            ))}
          </div>
        </Card>
      )}

      <Button fullWidth onClick={() => setIsCreating(true)}>
        Create New Organization
      </Button>
    </div>
  );
};

export default TenantSelector;
