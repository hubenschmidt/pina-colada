"use client";

import { useEffect, useState } from "react";
import {
  Stack,
  Center,
  Loader,
  Text,
  Card,
  SimpleGrid,
  Table,
  Badge,
  Group,
  Select,
} from "@mantine/core";
import { getUserAuditReport, getTenantUsers } from "../../../../api";
import { usePageLoading } from "../../../../context/pageLoadingContext";

const UserAuditPage = () => {
  const [report, setReport] = useState(null);
  const [users, setUsers] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [userId, setUserId] = useState(null);
  const [searchValue, setSearchValue] = useState("");
  const { dispatchPageLoading } = usePageLoading();

  useEffect(() => {
    dispatchPageLoading({ type: "SET_PAGE_LOADING", payload: false });
  }, [dispatchPageLoading]);

  useEffect(() => {
    const fetchUsers = async () => {
      try {
        const data = await getTenantUsers();
        setUsers(data);
        setLoading(false);
      } catch (err) {
        console.error("Failed to load users:", err);
        setLoading(false);
      }
    };
    fetchUsers();
  }, []);

  useEffect(() => {
    if (!userId) {
      setReport(null);
      return;
    }
    const fetchReport = async () => {
      setLoading(true);
      try {
        const data = await getUserAuditReport(parseInt(userId, 10));
        setReport(data);
      } catch (err) {
        setError(err instanceof Error ? err.message : "Failed to load report");
      } finally {
        setLoading(false);
      }
    };
    fetchReport();
  }, [userId]);

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
        <h1 className="text-xl font-bold text-zinc-900 dark:text-zinc-100">User Audit Report</h1>
        <Text c="red">{error}</Text>
      </Stack>
    );
  }

  if (!userId) {
    return (
      <Stack gap="lg">
        <Group justify="space-between">
          <h1 className="text-xl font-bold text-zinc-900 dark:text-zinc-100">User Audit Report</h1>
          <Select
            label="User"
            placeholder="Search users..."
            data={users.map((u) => ({ value: u.id.toString(), label: u.name }))}
            value={userId}
            onChange={(val) => setUserId(val || null)}
            searchValue={searchValue}
            onSearchChange={setSearchValue}
            onDropdownOpen={() => setSearchValue("")}
            size="sm"
            w={250}
            searchable
            clearable
            nothingFoundMessage="No users found"
          />
        </Group>
        <Center mih={200}>
          <Text c="dimmed">Select a user to view their audit report</Text>
        </Center>
      </Stack>
    );
  }

  if (!report) {
    return null;
  }

  return (
    <Stack gap="lg">
      <Group justify="space-between">
        <h1 className="text-xl font-bold text-zinc-900 dark:text-zinc-100">User Audit Report</h1>
        <Select
          label="User"
          placeholder="Search users..."
          data={users.map((u) => ({ value: u.id.toString(), label: u.name }))}
          value={userId}
          onChange={(val) => setUserId(val || null)}
          searchValue={searchValue}
          onSearchChange={setSearchValue}
          onDropdownOpen={() => setSearchValue("")}
          size="sm"
          w={250}
          searchable
          clearable
          nothingFoundMessage="No users found"
        />
      </Group>

      <SimpleGrid cols={{ base: 1, sm: 3 }} spacing="lg">
        <Card shadow="sm" padding="lg" radius="md" withBorder>
          <Text c="dimmed" size="sm">
            Total Records Created
          </Text>
          <Text fw={700} size="xl">
            {report.total_created}
          </Text>
        </Card>
        <Card shadow="sm" padding="lg" radius="md" withBorder>
          <Text c="dimmed" size="sm">
            Total Records Updated
          </Text>
          <Text fw={700} size="xl">
            {report.total_updated}
          </Text>
        </Card>
        <Card shadow="sm" padding="lg" radius="md" withBorder>
          <Text c="dimmed" size="sm">
            Tables with Activity
          </Text>
          <Text fw={700} size="xl">
            {report.by_table.length}
          </Text>
        </Card>
      </SimpleGrid>

      <Card shadow="sm" padding="lg" radius="md" withBorder>
        <Text fw={600} mb="md">
          Activity by Table
        </Text>
        {report.by_table.length === 0 ? (
          <Text c="dimmed" size="sm">
            No audit data available
          </Text>
        ) : (
          <Table>
            <Table.Thead>
              <Table.Tr>
                <Table.Th>Table</Table.Th>
                <Table.Th>Created</Table.Th>
                <Table.Th>Updated</Table.Th>
              </Table.Tr>
            </Table.Thead>
            <Table.Tbody>
              {report.by_table.map((row) => (
                <Table.Tr key={row.table}>
                  <Table.Td>{row.table}</Table.Td>
                  <Table.Td>{row.created_count}</Table.Td>
                  <Table.Td>{row.updated_count}</Table.Td>
                </Table.Tr>
              ))}
            </Table.Tbody>
          </Table>
        )}
      </Card>

      <Card shadow="sm" padding="lg" radius="md" withBorder>
        <Text fw={600} mb="md">
          Recent Activity
        </Text>
        {report.recent_activity.length === 0 ? (
          <Text c="dimmed" size="sm">
            No recent activity
          </Text>
        ) : (
          <Table>
            <Table.Thead>
              <Table.Tr>
                <Table.Th>Table</Table.Th>
                <Table.Th>Record</Table.Th>
                <Table.Th>Updated At</Table.Th>
              </Table.Tr>
            </Table.Thead>
            <Table.Tbody>
              {report.recent_activity.map((activity, idx) => (
                <Table.Tr key={`${activity.table}-${activity.id}-${idx}`}>
                  <Table.Td>
                    <Badge variant="light" color="gray" size="sm">
                      {activity.table}
                    </Badge>
                  </Table.Td>
                  <Table.Td>
                    <Text size="sm">{activity.display_name}</Text>
                    <Text size="xs" c="dimmed">
                      ID: {activity.id}
                    </Text>
                  </Table.Td>
                  <Table.Td>{formatDate(activity.updated_at)}</Table.Td>
                </Table.Tr>
              ))}
            </Table.Tbody>
          </Table>
        )}
      </Card>
    </Stack>
  );
};

export default UserAuditPage;
