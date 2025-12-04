"use client";

import { useEffect, useState } from "react";
import { Stack, Center, Loader, Text, Card, SimpleGrid, Table } from "@mantine/core";
import { getAccountOverviewReport } from "../../../../api";

const AccountOverviewPage = () => {
  const [report, setReport] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    const fetchReport = async () => {
      try {
        const data = await getAccountOverviewReport();
        setReport(data);
      } catch (err) {
        setError(err instanceof Error ? err.message : "Failed to load report");
      } finally {
        setLoading(false);
      }
    };
    fetchReport();
  }, []);

  if (loading) {
    return (
      <Center mih={400}>
        <Stack align="center" gap="md">
          <Loader size="xl" color="lime" />
          <Text c="dimmed">Loading report...</Text>
        </Stack>
      </Center>);

  }

  if (error) {
    return (
      <Stack gap="lg">
        <h1 className="text-xl font-bold text-zinc-900 dark:text-zinc-100">
          Account Overview Report
        </h1>
        <Text c="red">{error}</Text>
      </Stack>);

  }

  if (!report) {
    return null;
  }

  const countryEntries = Object.entries(report.organizations_by_country).sort((a, b) => b[1] - a[1]);
  const typeEntries = Object.entries(report.organizations_by_type).sort((a, b) => b[1] - a[1]);

  return (
    <Stack gap="lg">
      <h1 className="text-xl font-bold text-zinc-900 dark:text-zinc-100">
        Account Overview Report
      </h1>

      <SimpleGrid cols={{ base: 1, sm: 2 }} spacing="lg">
        <Card shadow="sm" padding="lg" radius="md" withBorder>
          <Text c="dimmed" size="sm">Total Organizations</Text>
          <Text fw={700} size="xl">{report.total_organizations}</Text>
        </Card>
        <Card shadow="sm" padding="lg" radius="md" withBorder>
          <Text c="dimmed" size="sm">Total Individuals</Text>
          <Text fw={700} size="xl">{report.total_individuals}</Text>
        </Card>
      </SimpleGrid>

      <SimpleGrid cols={{ base: 1, md: 2 }} spacing="lg">
        <Card shadow="sm" padding="lg" radius="md" withBorder>
          <Text fw={600} mb="md">Organizations by Country</Text>
          {countryEntries.length === 0 ?
          <Text c="dimmed" size="sm">No data available</Text> :

          <Table>
              <Table.Thead>
                <Table.Tr>
                  <Table.Th>Country</Table.Th>
                  <Table.Th>Count</Table.Th>
                </Table.Tr>
              </Table.Thead>
              <Table.Tbody>
                {countryEntries.map(([country, count]) =>
              <Table.Tr key={country}>
                    <Table.Td>{country}</Table.Td>
                    <Table.Td>{count}</Table.Td>
                  </Table.Tr>
              )}
              </Table.Tbody>
            </Table>
          }
        </Card>

        <Card shadow="sm" padding="lg" radius="md" withBorder>
          <Text fw={600} mb="md">Organizations by Type</Text>
          {typeEntries.length === 0 ?
          <Text c="dimmed" size="sm">No data available</Text> :

          <Table>
              <Table.Thead>
                <Table.Tr>
                  <Table.Th>Type</Table.Th>
                  <Table.Th>Count</Table.Th>
                </Table.Tr>
              </Table.Thead>
              <Table.Tbody>
                {typeEntries.map(([type, count]) =>
              <Table.Tr key={type}>
                    <Table.Td>{type}</Table.Td>
                    <Table.Td>{count}</Table.Td>
                  </Table.Tr>
              )}
              </Table.Tbody>
            </Table>
          }
        </Card>
      </SimpleGrid>
    </Stack>);

};

export default AccountOverviewPage;