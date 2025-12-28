import { Stack, Paper, Text, Group, Table, Badge, SimpleGrid } from "@mantine/core";
import { Clock, Cpu, DollarSign, Zap, MessageSquare } from "lucide-react";
import { DataTable } from "../DataTable/DataTable";

const formatCost = (amount) => {
  if (amount === null || amount === undefined) return "$0.0000";
  return `$${parseFloat(amount).toFixed(4)}`;
};

const formatTokens = (count) => {
  if (count === null || count === undefined) return "0";
  if (count >= 1_000_000) return `${(count / 1_000_000).toFixed(2)}M`;
  if (count >= 1_000) return `${(count / 1_000).toFixed(1)}K`;
  return count.toString();
};

const formatDuration = (ms) => {
  if (!ms) return "0ms";
  if (ms >= 1000) return `${(ms / 1000).toFixed(1)}s`;
  return `${ms}ms`;
};

const formatTime = (dateStr) => {
  if (!dateStr) return "";
  const date = new Date(dateStr);
  return date.toLocaleTimeString("en-US", { hour: "2-digit", minute: "2-digit", second: "2-digit" });
};

const formatSettings = (configSnapshot, nodeName) => {
  if (!configSnapshot || !nodeName) return "—";
  const nodeConfig = configSnapshot[nodeName];
  if (!nodeConfig?.settings) return "—";
  const s = nodeConfig.settings;
  const parts = [];
  // Go struct uses PascalCase (no json tags)
  if (s.Temperature !== undefined && s.Temperature !== null) parts.push(`t=${s.Temperature}`);
  if (s.TopP !== undefined && s.TopP !== null) parts.push(`p=${s.TopP}`);
  if (s.MaxTokens !== undefined && s.MaxTokens !== null) parts.push(`max=${s.MaxTokens}`);
  if (s.FrequencyPenalty !== undefined && s.FrequencyPenalty !== null) parts.push(`fp=${s.FrequencyPenalty}`);
  if (s.PresencePenalty !== undefined && s.PresencePenalty !== null) parts.push(`pp=${s.PresencePenalty}`);
  return parts.length > 0 ? parts.join(" ") : "defaults";
};

const truncatePrompt = (text, maxLen = 100) => {
  if (!text) return "—";
  if (text.length <= maxLen) return text;
  return text.slice(0, maxLen) + "…";
};

const StatCard = ({ icon: Icon, label, value, subvalue }) => (
  <Paper withBorder p="md" radius="md">
    <Group gap="xs" mb="xs">
      <Icon size={16} />
      <Text size="sm" c="dimmed">
        {label}
      </Text>
    </Group>
    <Text fw={700} size="xl">
      {value}
    </Text>
    {subvalue && (
      <Text size="xs" c="dimmed">
        {subvalue}
      </Text>
    )}
  </Paper>
);


const SessionMetricsTable = ({ metrics, pagination, queryParams, onQueryChange }) => {
  const columns = [
    {
      header: "Time",
      width: 90,
      sortKey: "started_at",
      sortable: true,
      render: (row) => (
        <Text size="xs" c="dimmed">
          {formatTime(row.started_at)}
        </Text>
      ),
    },
    {
      header: "Model",
      accessor: "model",
      sortKey: "model",
      sortable: true,
      render: (row) => <Text size="sm">{row.model || "—"}</Text>,
    },
    {
      header: "Agent",
      accessor: "node_name",
      width: 100,
      sortKey: "node_name",
      sortable: true,
      render: (row) => <Text size="sm">{row.node_name || "—"}</Text>,
    },
    {
      header: "Settings",
      width: 140,
      render: (row) => (
        <Text size="xs" c="dimmed" style={{ fontFamily: "monospace" }}>
          {formatSettings(row.config_snapshot, row.node_name)}
        </Text>
      ),
    },
    {
      header: "Prompt",
      render: (row) => (
        <Text size="xs" c="dimmed" title={row.user_message || ""} style={{ maxWidth: 300 }}>
          {truncatePrompt(row.user_message, 100)}
        </Text>
      ),
    },
    {
      header: "In",
      width: 70,
      sortKey: "input_tokens",
      sortable: true,
      tdProps: { align: "right" },
      render: (row) => <Text size="sm">{formatTokens(row.input_tokens)}</Text>,
    },
    {
      header: "Out",
      width: 70,
      sortKey: "output_tokens",
      sortable: true,
      tdProps: { align: "right" },
      render: (row) => <Text size="sm">{formatTokens(row.output_tokens)}</Text>,
    },
    {
      header: "Duration",
      width: 80,
      sortKey: "duration_ms",
      sortable: true,
      tdProps: { align: "right" },
      render: (row) => (
        <Text size="xs" c="dimmed">
          {formatDuration(row.duration_ms)}
        </Text>
      ),
    },
    {
      header: "Cost",
      width: 80,
      sortKey: "estimated_cost_usd",
      sortable: true,
      tdProps: { align: "right" },
      render: (row) => <Text size="sm">{formatCost(row.estimated_cost_usd)}</Text>,
    },
  ];

  // Use server-side pagination data if available, else fall back to client-side
  const page = pagination?.current_page || queryParams?.page || 1;
  const pageSize = pagination?.page_size || queryParams?.pageSize || 10;
  const totalPages = pagination?.total_pages || Math.ceil((metrics?.length || 1) / pageSize);
  const sortBy = queryParams?.sortBy || "started_at";
  const sortDirection = queryParams?.sortDirection || "DESC";

  const data = {
    items: metrics || [],
    currentPage: page,
    totalPages,
    pageSize,
  };

  const handleSortChange = ({ sortBy: newSortBy, direction }) => {
    if (onQueryChange) {
      onQueryChange({ sortBy: newSortBy, sortDirection: direction, page: 1 });
    }
  };

  const handlePageChange = (newPage) => {
    if (onQueryChange) {
      onQueryChange({ page: newPage });
    }
  };

  const handlePageSizeChange = (newSize) => {
    if (onQueryChange) {
      onQueryChange({ pageSize: newSize, page: 1 });
    }
  };

  return (
    <DataTable
      data={data}
      columns={columns}
      rowKey={(row) => row.id}
      emptyText="No metrics recorded in this session"
      highlightOnHover
      withTableBorder={false}
      withColumnBorders={false}
      sortBy={sortBy}
      sortDirection={sortDirection}
      onSortChange={handleSortChange}
      pageValue={page}
      pageSizeValue={pageSize}
      onPageChange={handlePageChange}
      onPageSizeChange={handlePageSizeChange}
      pageSizeOptions={[5, 10, 20, 50]}
    />
  );
};

const ComparisonView = ({ sessions }) => {
  if (!sessions || sessions.length === 0) {
    return (
      <Text c="dimmed" ta="center" py="md">
        No sessions to compare
      </Text>
    );
  }

  return (
    <Stack gap="lg">
      <Paper withBorder p="md" radius="md">
        <Text fw={600} mb="md">
          Session Comparison Summary
        </Text>
        <Table striped highlightOnHover>
          <Table.Thead>
            <Table.Tr>
              <Table.Th>Session</Table.Th>
              <Table.Th ta="center">Turns</Table.Th>
              <Table.Th ta="right">Total Tokens</Table.Th>
              <Table.Th ta="right">Avg/Turn</Table.Th>
              <Table.Th ta="right">Total Cost</Table.Th>
              <Table.Th ta="right">Avg Latency</Table.Th>
            </Table.Tr>
          </Table.Thead>
          <Table.Tbody>
            {sessions.map((item) => {
              const avgTokens = item.metrics.length > 0 ? Math.round(item.session.total_tokens / item.metrics.length) : 0;
              const avgLatency =
                item.metrics.length > 0 ? Math.round(item.metrics.reduce((acc, m) => acc + (m.duration_ms || 0), 0) / item.metrics.length) : 0;

              return (
                <Table.Tr key={item.session.id}>
                  <Table.Td>
                    <Text size="sm" fw={500}>
                      {new Date(item.session.started_at).toLocaleDateString()}
                    </Text>
                    <Text size="xs" c="dimmed">
                      {new Date(item.session.started_at).toLocaleTimeString()}
                    </Text>
                  </Table.Td>
                  <Table.Td ta="center">{item.metrics.length}</Table.Td>
                  <Table.Td ta="right">{formatTokens(item.session.total_tokens)}</Table.Td>
                  <Table.Td ta="right">{formatTokens(avgTokens)}</Table.Td>
                  <Table.Td ta="right">{formatCost(item.session.total_cost)}</Table.Td>
                  <Table.Td ta="right">{formatDuration(avgLatency)}</Table.Td>
                </Table.Tr>
              );
            })}
          </Table.Tbody>
        </Table>
      </Paper>

      {sessions.map((item) => (
        <Paper key={item.session.id} withBorder p="md" radius="md">
          <Group mb="md" gap="sm">
            <Text fw={600}>Session {new Date(item.session.started_at).toLocaleDateString()}</Text>
            <Badge size="sm" color="gray">
              {item.metrics.length} turns
            </Badge>
          </Group>
          <SessionMetricsTable metrics={item.metrics} />
        </Paper>
      ))}
    </Stack>
  );
};

const MetricsCharts = ({ session, comparison, queryParams, onQueryChange }) => {
  if (comparison) {
    return <ComparisonView sessions={comparison.sessions} />;
  }

  if (!session) {
    return (
      <Text c="dimmed" ta="center" py="md">
        No session data available
      </Text>
    );
  }

  const { session: sessionData, metrics, total_items, current_page, page_size, total_pages } = session;

  const pagination = total_items ? { total_items, current_page, page_size, total_pages } : null;

  return (
    <Stack gap="lg">
      <SimpleGrid cols={{ base: 2, md: 4 }}>
        <StatCard icon={Zap} label="Total Turns" value={sessionData.metric_count || 0} />
        <StatCard icon={Cpu} label="Total Tokens" value={formatTokens(sessionData.total_tokens)} subvalue={`In: ${formatTokens(sessionData.total_input_tokens)} / Out: ${formatTokens(sessionData.total_output_tokens)}`} />
        <StatCard icon={Clock} label="Avg Latency" value={formatDuration(sessionData.avg_latency_ms)} />
        <StatCard icon={DollarSign} label="Total Cost" value={formatCost(sessionData.total_cost)} />
      </SimpleGrid>

      <Paper withBorder p="md" radius="md">
        <Group mb="md" gap="xs">
          <MessageSquare size={16} />
          <Text fw={600}>Turn-by-Turn Metrics</Text>
        </Group>
        <SessionMetricsTable metrics={metrics} pagination={pagination} queryParams={queryParams} onQueryChange={onQueryChange} />
      </Paper>
    </Stack>
  );
};

export default MetricsCharts;
