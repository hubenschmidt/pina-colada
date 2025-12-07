"use client";

import { useState, useEffect } from "react";
import { useUserContext } from "../../context/userContext";
import {
  Stack,
  Paper,
  Text,
  Loader,
  Center,
  SegmentedControl,
  Group,
  Table,
  Tabs,
  Badge,
} from "@mantine/core";
import { BarChart2, User, Building, Code, Cpu } from "lucide-react";
import { usePageLoading } from "../../context/pageLoadingContext";
import {
  getUserUsage,
  getTenantUsage,
  getDeveloperAnalytics,
  checkDeveloperAccess,
} from "../../api";

const PERIODS = [
  { value: "daily", label: "Day" },
  { value: "weekly", label: "Week" },
  { value: "monthly", label: "Month" },
  { value: "quarterly", label: "Quarter" },
  { value: "annual", label: "Year" },
];

const formatTokens = (count) => {
  if (count >= 1_000_000) return `${(count / 1_000_000).toFixed(2)}M`;
  if (count >= 1_000) return `${(count / 1_000).toFixed(1)}K`;
  return count.toString();
};

const UsageCard = ({ title, icon: Icon, data, loading }) => (
  <Paper withBorder p="md" radius="md">
    <Group gap="xs" mb="xs">
      <Icon size={20} />
      <Text fw={500}>{title}</Text>
    </Group>
    {loading ? (
      <Loader size="sm" />
    ) : (
      <Stack gap="xs">
        <Group justify="space-between">
          <Text size="sm" c="dimmed">
            Total tokens
          </Text>
          <Text fw={600}>{formatTokens(data?.total_tokens || 0)}</Text>
        </Group>
        <Group justify="space-between">
          <Text size="sm" c="dimmed">
            Input / Output
          </Text>
          <Text size="sm">
            {formatTokens(data?.input_tokens || 0)} / {formatTokens(data?.output_tokens || 0)}
          </Text>
        </Group>
      </Stack>
    )}
  </Paper>
);

const AnalyticsTable = ({ data, labelKey, labelTitle }) => {
  if (!data?.length) {
    return (
      <Text c="dimmed" ta="center" py="md">
        No data available for this period
      </Text>
    );
  }

  return (
    <Table striped highlightOnHover>
      <Table.Thead>
        <Table.Tr>
          <Table.Th>{labelTitle}</Table.Th>
          <Table.Th ta="right">Total</Table.Th>
          <Table.Th ta="right">Input</Table.Th>
          <Table.Th ta="right">Output</Table.Th>
        </Table.Tr>
      </Table.Thead>
      <Table.Tbody>
        {data.map((row, i) => (
          <Table.Tr key={i}>
            <Table.Td>{row[labelKey] || "Unknown"}</Table.Td>
            <Table.Td ta="right">{formatTokens(row.total_tokens)}</Table.Td>
            <Table.Td ta="right">{formatTokens(row.input_tokens)}</Table.Td>
            <Table.Td ta="right">{formatTokens(row.output_tokens)}</Table.Td>
          </Table.Tr>
        ))}
      </Table.Tbody>
    </Table>
  );
};

const UsagePage = () => {
  const { userState } = useUserContext();
  const [period, setPeriod] = useState("monthly");
  const [userUsage, setUserUsage] = useState(null);
  const [tenantUsage, setTenantUsage] = useState(null);
  const [hasDeveloperAccess, setHasDeveloperAccess] = useState(false);
  const [nodeAnalytics, setNodeAnalytics] = useState([]);
  const [modelAnalytics, setModelAnalytics] = useState([]);
  const [loading, setLoading] = useState(true);
  const [analyticsTab, setAnalyticsTab] = useState("node");
  const { dispatchPageLoading } = usePageLoading();

  useEffect(() => {
    const fetchData = async () => {
      if (!userState.isAuthed) {
        dispatchPageLoading({ type: "SET_PAGE_LOADING", payload: false });
        return;
      }

      setLoading(true);
      try {
        const [user, tenant, devAccess] = await Promise.all([
          getUserUsage(period),
          getTenantUsage(period),
          checkDeveloperAccess(),
        ]);
        setUserUsage(user);
        setTenantUsage(tenant);
        setHasDeveloperAccess(devAccess.has_developer_access);

        if (devAccess.has_developer_access) {
          const [nodes, models] = await Promise.all([
            getDeveloperAnalytics(period, "node"),
            getDeveloperAnalytics(period, "model"),
          ]);
          setNodeAnalytics(nodes.data || []);
          setModelAnalytics(models.data || []);
        }
      } catch (error) {
        console.error("Failed to fetch usage data:", error);
      } finally {
        setLoading(false);
        dispatchPageLoading({ type: "SET_PAGE_LOADING", payload: false });
      }
    };

    fetchData();
  }, [userState.isAuthed, period, dispatchPageLoading]);

  if (!userState.isAuthed) {
    return (
      <Center mih={400}>
        <Text>Please log in to view usage analytics.</Text>
      </Center>
    );
  }

  if (loading && !userUsage) {
    return (
      <Center mih={400}>
        <Stack align="center" gap="md">
          <Loader size="xl" color="lime" />
          <Text c="dimmed">Loading usage data...</Text>
        </Stack>
      </Center>
    );
  }

  return (
    <Stack gap="lg">
      <Group justify="space-between">
        <h1 className="text-xl font-bold text-zinc-900 dark:text-zinc-100">Usage</h1>
        <SegmentedControl value={period} onChange={setPeriod} data={PERIODS} size="sm" />
      </Group>

      <Group grow>
        <UsageCard title="Your Usage" icon={User} data={userUsage} loading={loading} />
        <UsageCard
          title="Organization Usage"
          icon={Building}
          data={tenantUsage}
          loading={loading}
        />
      </Group>

      {hasDeveloperAccess && (
        <Paper withBorder p="md" radius="md">
          <Stack gap="md">
            <Group gap="xs">
              <BarChart2 size={20} />
              <Text fw={600}>Agent Analytics</Text>
              <Badge color="grape" size="sm">
                Developer
              </Badge>
            </Group>

            <Tabs value={analyticsTab} onChange={setAnalyticsTab}>
              <Tabs.List>
                <Tabs.Tab value="node" leftSection={<Code size={14} />}>
                  By Node
                </Tabs.Tab>
                <Tabs.Tab value="model" leftSection={<Cpu size={14} />}>
                  By Model
                </Tabs.Tab>
              </Tabs.List>

              <Tabs.Panel value="node" pt="md">
                <AnalyticsTable data={nodeAnalytics} labelKey="node_name" labelTitle="Node" />
              </Tabs.Panel>

              <Tabs.Panel value="model" pt="md">
                <AnalyticsTable data={modelAnalytics} labelKey="model_name" labelTitle="Model" />
              </Tabs.Panel>
            </Tabs>
          </Stack>
        </Paper>
      )}
    </Stack>
  );
};

export default UsagePage;
