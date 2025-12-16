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
import { BarChart2, User, Building, Code, Cpu, DollarSign } from "lucide-react";
import { usePageLoading } from "../../context/pageLoadingContext";
import { getUserUsage, getTenantUsage, getDeveloperAnalytics, getProviderCosts } from "../../api";
import DeveloperFeature from "../../components/DeveloperFeature/DeveloperFeature";

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

const formatCost = (amount) => {
  if (amount === null || amount === undefined) return "—";
  return `$${parseFloat(amount).toFixed(2)}`;
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

const formatAvg = (total, count) => {
  if (!count) return "—";
  return formatTokens(Math.round(total / count));
};

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
          <Table.Th ta="right">Requests</Table.Th>
          <Table.Th ta="right">Convos</Table.Th>
          <Table.Th ta="right">Total</Table.Th>
          <Table.Th ta="right">Avg/Req</Table.Th>
          <Table.Th ta="right">Input</Table.Th>
          <Table.Th ta="right">Output</Table.Th>
        </Table.Tr>
      </Table.Thead>
      <Table.Tbody>
        {data.map((row, i) => (
          <Table.Tr key={i}>
            <Table.Td>{row[labelKey] || "Unknown"}</Table.Td>
            <Table.Td ta="right">{row.request_count?.toLocaleString() || 0}</Table.Td>
            <Table.Td ta="right">{row.conversation_count?.toLocaleString() || 0}</Table.Td>
            <Table.Td ta="right">{formatTokens(row.total_tokens)}</Table.Td>
            <Table.Td ta="right">{formatAvg(row.total_tokens, row.request_count)}</Table.Td>
            <Table.Td ta="right">{formatTokens(row.input_tokens)}</Table.Td>
            <Table.Td ta="right">{formatTokens(row.output_tokens)}</Table.Td>
          </Table.Tr>
        ))}
      </Table.Tbody>
    </Table>
  );
};

const ProviderCostsCard = ({ costs, loading }) => (
  <Paper withBorder p="md" radius="md">
    <Group gap="xs" mb="xs">
      <DollarSign size={20} />
      <Text fw={500}>Provider Spend</Text>
      <Badge color="grape" size="sm">
        Developer
      </Badge>
    </Group>
    {loading ? (
      <Loader size="sm" />
    ) : costs?.error ? (
      <Text c="dimmed" size="sm">
        {costs.error}
      </Text>
    ) : (
      <Stack gap="xs">
        <Group justify="space-between">
          <Text size="sm" c="dimmed">
            OpenAI
          </Text>
          <Text fw={600}>{formatCost(costs?.openai?.spend)}</Text>
        </Group>
        <Group justify="space-between">
          <Text size="sm" c="dimmed">
            Anthropic (includes Claude Code raw API cost)
          </Text>
          <Text fw={600}>{formatCost(costs?.anthropic?.spend)}</Text>
        </Group>
        <Group
          justify="space-between"
          pt="xs"
          style={{ borderTop: "1px solid var(--mantine-color-default-border)" }}>
          <Text size="sm" fw={500}>
            Total
          </Text>
          <Text fw={700}>{formatCost(costs?.total_spend)}</Text>
        </Group>
      </Stack>
    )}
  </Paper>
);

const UsagePage = () => {
  const { userState } = useUserContext();
  const [period, setPeriod] = useState("monthly");
  const [userUsage, setUserUsage] = useState(null);
  const [tenantUsage, setTenantUsage] = useState(null);
  const [nodeAnalytics, setNodeAnalytics] = useState([]);
  const [modelAnalytics, setModelAnalytics] = useState([]);
  const [providerCosts, setProviderCosts] = useState(null);
  const [loadingUsage, setLoadingUsage] = useState(true);
  const [loadingCosts, setLoadingCosts] = useState(true);
  const [analyticsTab, setAnalyticsTab] = useState("node");
  const { dispatchPageLoading } = usePageLoading();

  // Fetch usage data (fast)
  useEffect(() => {
    const fetchUsageData = async () => {
      if (!userState.isAuthed) {
        dispatchPageLoading({ type: "SET_PAGE_LOADING", payload: false });
        return;
      }

      setLoadingUsage(true);
      try {
        const [user, tenant, nodes, models] = await Promise.all([
          getUserUsage(period),
          getTenantUsage(period),
          getDeveloperAnalytics(period, "node"),
          getDeveloperAnalytics(period, "model"),
        ]);
        setUserUsage(user);
        setTenantUsage(tenant);
        setNodeAnalytics(nodes.data || []);
        setModelAnalytics(models.data || []);
      } catch (error) {
        console.error("Failed to fetch usage data:", error);
      } finally {
        setLoadingUsage(false);
        dispatchPageLoading({ type: "SET_PAGE_LOADING", payload: false });
      }
    };

    fetchUsageData();
  }, [userState.isAuthed, period, dispatchPageLoading]);

  // Fetch provider costs separately (slow external API)
  useEffect(() => {
    const fetchCosts = async () => {
      if (!userState.isAuthed) return;

      setLoadingCosts(true);
      try {
        const costs = await getProviderCosts(period);
        setProviderCosts(costs);
      } catch (error) {
        console.error("Failed to fetch provider costs:", error);
      } finally {
        setLoadingCosts(false);
      }
    };

    fetchCosts();
  }, [userState.isAuthed, period]);

  if (!userState.isAuthed) {
    return (
      <Center mih={400}>
        <Text>Please log in to view usage analytics.</Text>
      </Center>
    );
  }

  if (loadingUsage && !userUsage) {
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
        <UsageCard title="Your Usage" icon={User} data={userUsage} loading={loadingUsage} />
        <UsageCard
          title="Organization Usage"
          icon={Building}
          data={tenantUsage}
          loading={loadingUsage}
        />
        <DeveloperFeature>
          <ProviderCostsCard costs={providerCosts} loading={loadingCosts} />
        </DeveloperFeature>
      </Group>

      <DeveloperFeature>
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
      </DeveloperFeature>
    </Stack>
  );
};

export default UsagePage;
