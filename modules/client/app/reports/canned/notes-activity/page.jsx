"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import {
  Stack,
  Center,
  Loader,
  Text,
  Card,
  SimpleGrid,
  Table,
  Anchor,
  Badge,
  Group,
} from "@mantine/core";
import { getNotesActivityReport } from "../../../../api";
import { useProjectContext } from "../../../../context/projectContext";

const NotesActivityPage = () => {
  const [report, setReport] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const { projectState } = useProjectContext();
  const { selectedProject } = projectState;

  useEffect(() => {
    const fetchReport = async () => {
      setLoading(true);
      try {
        const data = await getNotesActivityReport(selectedProject?.id);
        setReport(data);
      } catch (err) {
        setError(err instanceof Error ? err.message : "Failed to load report");
      } finally {
        setLoading(false);
      }
    };
    fetchReport();
  }, [selectedProject?.id]);

  if (loading) {
    return (
      <Center mih={400}>
        <Stack align="center" gap="md">
          <Loader size="xl" color="lime" />
          <Text c="dimmed">Loading report...</Text>
        </Stack>
      </Center>
    );
  }

  if (error) {
    return (
      <Stack gap="lg">
        <h1 className="text-xl font-bold text-zinc-900 dark:text-zinc-100">
          Notes Activity Report
        </h1>
        <Text c="red">{error}</Text>
      </Stack>
    );
  }

  if (!report) {
    return null;
  }

  const entityTypeEntries = Object.entries(report.by_entity_type);
  const entitiesWithNotesEntries = Object.entries(report.entities_with_notes);

  const formatDate = (dateStr) => {
    if (!dateStr) return "-";
    const date = new Date(dateStr);
    return date.toLocaleDateString("en-US", {
      month: "short",
      day: "numeric",
      year: "numeric",
      hour: "numeric",
      minute: "2-digit",
    });
  };

  const getEntityLink = (entityType, entityId) => {
    const routes = {
      Organization: `/accounts/organizations/${entityId}`,
      Individual: `/accounts/individuals/${entityId}`,
      Contact: `/contacts/${entityId}`,
      Lead: `/leads/${entityId}`,
    };
    return routes[entityType] || null;
  };

  return (
    <Stack gap="lg">
      <Group justify="space-between">
        <h1 className="text-xl font-bold text-zinc-900 dark:text-zinc-100">
          Notes Activity Report
        </h1>
        {selectedProject ? (
          <Badge variant="light" color="lime">
            {selectedProject.name}
          </Badge>
        ) : (
          <Badge variant="light" color="gray">
            Global
          </Badge>
        )}
      </Group>

      <SimpleGrid cols={{ base: 1, sm: 3 }} spacing="lg">
        <Card shadow="sm" padding="lg" radius="md" withBorder>
          <Text c="dimmed" size="sm">
            Total Notes
          </Text>
          <Text fw={700} size="xl">
            {report.total_notes}
          </Text>
        </Card>
        <Card shadow="sm" padding="lg" radius="md" withBorder>
          <Text c="dimmed" size="sm">
            Entity Types with Notes
          </Text>
          <Text fw={700} size="xl">
            {entityTypeEntries.length}
          </Text>
        </Card>
        <Card shadow="sm" padding="lg" radius="md" withBorder>
          <Text c="dimmed" size="sm">
            Entities with Notes
          </Text>
          <Text fw={700} size="xl">
            {Object.values(report.entities_with_notes).reduce((a, b) => a + b, 0)}
          </Text>
        </Card>
      </SimpleGrid>

      <SimpleGrid cols={{ base: 1, md: 2 }} spacing="lg">
        <Card shadow="sm" padding="lg" radius="md" withBorder>
          <Text fw={600} mb="md">
            Notes by Entity Type
          </Text>
          {entityTypeEntries.length === 0 ? (
            <Text c="dimmed" size="sm">
              No data available
            </Text>
          ) : (
            <Table>
              <Table.Thead>
                <Table.Tr>
                  <Table.Th>Entity Type</Table.Th>
                  <Table.Th>Note Count</Table.Th>
                </Table.Tr>
              </Table.Thead>
              <Table.Tbody>
                {entityTypeEntries.map(([type, count]) => (
                  <Table.Tr key={type}>
                    <Table.Td>{type}</Table.Td>
                    <Table.Td>{count}</Table.Td>
                  </Table.Tr>
                ))}
              </Table.Tbody>
            </Table>
          )}
        </Card>

        <Card shadow="sm" padding="lg" radius="md" withBorder>
          <Text fw={600} mb="md">
            Entities with Notes
          </Text>
          {entitiesWithNotesEntries.length === 0 ? (
            <Text c="dimmed" size="sm">
              No data available
            </Text>
          ) : (
            <Table>
              <Table.Thead>
                <Table.Tr>
                  <Table.Th>Entity Type</Table.Th>
                  <Table.Th>Unique Entities</Table.Th>
                </Table.Tr>
              </Table.Thead>
              <Table.Tbody>
                {entitiesWithNotesEntries.map(([type, count]) => (
                  <Table.Tr key={type}>
                    <Table.Td>{type}</Table.Td>
                    <Table.Td>{count}</Table.Td>
                  </Table.Tr>
                ))}
              </Table.Tbody>
            </Table>
          )}
        </Card>
      </SimpleGrid>

      <Card shadow="sm" padding="lg" radius="md" withBorder>
        <Text fw={600} mb="md">
          Recent Notes
        </Text>
        {report.recent_notes.length === 0 ? (
          <Text c="dimmed" size="sm">
            No recent notes
          </Text>
        ) : (
          <Table>
            <Table.Thead>
              <Table.Tr>
                <Table.Th>ID</Table.Th>
                <Table.Th>Entity</Table.Th>
                <Table.Th>Content</Table.Th>
                <Table.Th>Created</Table.Th>
              </Table.Tr>
            </Table.Thead>
            <Table.Tbody>
              {report.recent_notes.map((note) => {
                const entityLink = getEntityLink(note.entity_type, note.entity_id);
                const displayName = note.entity_name || `${note.entity_type} #${note.entity_id}`;
                return (
                  <Table.Tr key={note.id}>
                    <Table.Td>{note.id}</Table.Td>
                    <Table.Td>
                      {entityLink ? (
                        <Anchor component={Link} href={entityLink} size="sm" c="blue">
                          {displayName}
                        </Anchor>
                      ) : (
                        <Text size="sm">{displayName}</Text>
                      )}
                      <Text size="xs" c="dimmed">
                        {note.entity_type}
                      </Text>
                    </Table.Td>
                    <Table.Td style={{ maxWidth: 400 }}>
                      <Text size="sm" lineClamp={2}>
                        {note.content}
                      </Text>
                    </Table.Td>
                    <Table.Td>{formatDate(note.created_at)}</Table.Td>
                  </Table.Tr>
                );
              })}
            </Table.Tbody>
          </Table>
        )}
      </Card>
    </Stack>
  );
};

export default NotesActivityPage;
