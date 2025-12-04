"use client";

import { useEffect, useState } from "react";
import { Stack, Center, Loader, Text, Card, SimpleGrid, Table, Badge, Group } from "@mantine/core";
import { getLeadPipelineReport } from "../../../../api";
import { useProjectContext } from "../../../../context/projectContext";

const LeadPipelinePage = () => {
  const [report, setReport] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const { projectState } = useProjectContext();
  const { selectedProject } = projectState;

  useEffect(() => {
    const fetchReport = async () => {
      setLoading(true);
      try {
        const data = await getLeadPipelineReport(undefined, undefined, selectedProject?.id);
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
        <h1 className="text-xl font-bold text-zinc-900 dark:text-zinc-100">Lead Pipeline Report</h1>
        <Text c="red">{error}</Text>
      </Stack>
    );
  }

  if (!report) {
    return null;
  }

  const typeEntries = Object.entries(report.by_type);
  const sourceEntries = Object.entries(report.by_source);

  return (
    <Stack gap="lg">
      <Group justify="space-between">
        <h1 className="text-xl font-bold text-zinc-900 dark:text-zinc-100">Lead Pipeline Report</h1>
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
            Total Leads
          </Text>
          <Text fw={700} size="xl">
            {report.total_leads}
          </Text>
        </Card>
        <Card shadow="sm" padding="lg" radius="md" withBorder>
          <Text c="dimmed" size="sm">
            Lead Types
          </Text>
          <Text fw={700} size="xl">
            {typeEntries.length}
          </Text>
        </Card>
        <Card shadow="sm" padding="lg" radius="md" withBorder>
          <Text c="dimmed" size="sm">
            Lead Sources
          </Text>
          <Text fw={700} size="xl">
            {sourceEntries.length}
          </Text>
        </Card>
      </SimpleGrid>

      <SimpleGrid cols={{ base: 1, md: 2 }} spacing="lg">
        <Card shadow="sm" padding="lg" radius="md" withBorder>
          <Text fw={600} mb="md">
            Leads by Type
          </Text>
          {typeEntries.length === 0 ? (
            <Text c="dimmed" size="sm">
              No data available
            </Text>
          ) : (
            <Table>
              <Table.Thead>
                <Table.Tr>
                  <Table.Th>Type</Table.Th>
                  <Table.Th>Count</Table.Th>
                </Table.Tr>
              </Table.Thead>
              <Table.Tbody>
                {typeEntries.map(([type, count]) => (
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
            Leads by Source
          </Text>
          {sourceEntries.length === 0 ? (
            <Text c="dimmed" size="sm">
              No data available
            </Text>
          ) : (
            <Table>
              <Table.Thead>
                <Table.Tr>
                  <Table.Th>Source</Table.Th>
                  <Table.Th>Count</Table.Th>
                </Table.Tr>
              </Table.Thead>
              <Table.Tbody>
                {sourceEntries.map(([source, count]) => (
                  <Table.Tr key={source}>
                    <Table.Td>{source}</Table.Td>
                    <Table.Td>{count}</Table.Td>
                  </Table.Tr>
                ))}
              </Table.Tbody>
            </Table>
          )}
        </Card>
      </SimpleGrid>
    </Stack>
  );
};

export default LeadPipelinePage;
