"use client";

import { useEffect, useState } from "react";
import { Stack, Center, Loader, Text, Card, SimpleGrid, Table, Progress } from "@mantine/core";
import { getContactCoverageReport } from "../../../../api";

const ContactCoveragePage = () => {
  const [report, setReport] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    const fetchReport = async () => {
      try {
        const data = await getContactCoverageReport();
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
      </Center>
    );
  }

  if (error) {
    return (
      <Stack gap="lg">
        <h1 className="text-xl font-bold text-zinc-900 dark:text-zinc-100">
          Contact Coverage Report
        </h1>
        <Text c="red">{error}</Text>
      </Stack>
    );
  }

  if (!report) {
    return null;
  }

  const coveragePercent =
    report.total_organizations > 0
      ? ((report.total_organizations - report.organizations_with_zero_contacts) /
          report.total_organizations) *
        100
      : 0;

  const decisionMakerPercent = report.decision_maker_ratio * 100;

  return (
    <Stack gap="lg">
      <h1 className="text-xl font-bold text-zinc-900 dark:text-zinc-100">
        Contact Coverage Report
      </h1>

      <SimpleGrid cols={{ base: 1, sm: 2, lg: 4 }} spacing="lg">
        <Card shadow="sm" padding="lg" radius="md" withBorder>
          <Text c="dimmed" size="sm">
            Total Organizations
          </Text>
          <Text fw={700} size="xl">
            {report.total_organizations}
          </Text>
        </Card>
        <Card shadow="sm" padding="lg" radius="md" withBorder>
          <Text c="dimmed" size="sm">
            Total Contacts
          </Text>
          <Text fw={700} size="xl">
            {report.total_contacts}
          </Text>
        </Card>
        <Card shadow="sm" padding="lg" radius="md" withBorder>
          <Text c="dimmed" size="sm">
            Avg Contacts/Org
          </Text>
          <Text fw={700} size="xl">
            {report.average_contacts_per_org}
          </Text>
        </Card>
        <Card shadow="sm" padding="lg" radius="md" withBorder>
          <Text c="dimmed" size="sm">
            Decision Makers
          </Text>
          <Text fw={700} size="xl">
            {report.decision_maker_count}
          </Text>
        </Card>
      </SimpleGrid>

      <SimpleGrid cols={{ base: 1, md: 2 }} spacing="lg">
        <Card shadow="sm" padding="lg" radius="md" withBorder>
          <Text fw={600} mb="md">
            Coverage Rate
          </Text>
          <Text c="dimmed" size="sm" mb="xs">
            {report.total_organizations - report.organizations_with_zero_contacts} of{" "}
            {report.total_organizations} organizations have contacts
          </Text>
          <Progress value={coveragePercent} color="lime" size="lg" />
          <Text size="sm" mt="xs">
            {coveragePercent.toFixed(1)}% coverage
          </Text>
        </Card>

        <Card shadow="sm" padding="lg" radius="md" withBorder>
          <Text fw={600} mb="md">
            Decision Maker Ratio
          </Text>
          <Text c="dimmed" size="sm" mb="xs">
            {report.decision_maker_count} of {report.total_contacts} contacts are decision makers
          </Text>
          <Progress value={decisionMakerPercent} color="blue" size="lg" />
          <Text size="sm" mt="xs">
            {decisionMakerPercent.toFixed(1)}% decision makers
          </Text>
        </Card>
      </SimpleGrid>

      <Card shadow="sm" padding="lg" radius="md" withBorder>
        <Text fw={600} mb="md">
          Top Organizations by Contact Count
        </Text>
        {report.coverage_by_org.length === 0 ? (
          <Text c="dimmed" size="sm">
            No data available
          </Text>
        ) : (
          <Table>
            <Table.Thead>
              <Table.Tr>
                <Table.Th>Organization</Table.Th>
                <Table.Th>Contact Count</Table.Th>
              </Table.Tr>
            </Table.Thead>
            <Table.Tbody>
              {report.coverage_by_org.map((org) => (
                <Table.Tr key={org.organization_id}>
                  <Table.Td>{org.organization_name}</Table.Td>
                  <Table.Td>{org.contact_count}</Table.Td>
                </Table.Tr>
              ))}
            </Table.Tbody>
          </Table>
        )}
      </Card>
    </Stack>
  );
};

export default ContactCoveragePage;
