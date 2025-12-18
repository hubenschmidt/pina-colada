import { Stack, Paper, Text, Group, Table, Badge, SimpleGrid } from "@mantine/core";
import { Clock, Cpu, DollarSign, Zap, MessageSquare } from "lucide-react";
import styles from "./MetricsCharts.module.css";

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

const TokenBar = ({ inputTokens, outputTokens, maxTotal }) => {
  const total = inputTokens + outputTokens;
  const scale = maxTotal > 0 ? (total / maxTotal) * 100 : 0;
  const inputPct = total > 0 ? (inputTokens / total) * 100 : 50;

  return (
    <div className={styles.tokenBarContainer}>
      <div className={styles.tokenBar} style={{ width: `${Math.max(scale, 2)}%` }}>
        <div className={styles.inputBar} style={{ width: `${inputPct}%` }} title={`Input: ${formatTokens(inputTokens)}`} />
        <div className={styles.outputBar} style={{ width: `${100 - inputPct}%` }} title={`Output: ${formatTokens(outputTokens)}`} />
      </div>
    </div>
  );
};

const CostBar = ({ cost, maxCost }) => {
  const scale = maxCost > 0 ? (cost / maxCost) * 100 : 0;

  return (
    <div className={styles.costBarContainer}>
      <div className={styles.costBar} style={{ width: `${Math.max(scale, 2)}%` }} />
    </div>
  );
};

const DurationBar = ({ durationMs, maxDuration }) => {
  const scale = maxDuration > 0 ? (durationMs / maxDuration) * 100 : 0;

  return (
    <div className={styles.durationBarContainer}>
      <div className={styles.durationBar} style={{ width: `${Math.max(scale, 2)}%` }} />
    </div>
  );
};

const SessionMetricsTable = ({ metrics }) => {
  if (!metrics || metrics.length === 0) {
    return (
      <Text c="dimmed" ta="center" py="md">
        No metrics recorded in this session
      </Text>
    );
  }

  const maxTotal = Math.max(...metrics.map((m) => m.input_tokens + m.output_tokens));
  const maxCost = Math.max(...metrics.map((m) => m.estimated_cost_usd || 0));
  const maxDuration = Math.max(...metrics.map((m) => m.duration_ms || 0));

  return (
    <Table striped highlightOnHover>
      <Table.Thead>
        <Table.Tr>
          <Table.Th w={80}>Time</Table.Th>
          <Table.Th>Model</Table.Th>
          <Table.Th>Agent</Table.Th>
          <Table.Th w={200}>Tokens</Table.Th>
          <Table.Th ta="right">Total</Table.Th>
          <Table.Th w={120}>Duration</Table.Th>
          <Table.Th ta="right">Cost</Table.Th>
        </Table.Tr>
      </Table.Thead>
      <Table.Tbody>
        {metrics.map((metric) => (
          <Table.Tr key={metric.id}>
            <Table.Td>
              <Text size="xs" c="dimmed">
                {formatTime(metric.started_at)}
              </Text>
            </Table.Td>
            <Table.Td>
              <Text size="sm">{metric.model || "—"}</Text>
            </Table.Td>
            <Table.Td>
              <Text size="sm">{metric.node_name || "—"}</Text>
            </Table.Td>
            <Table.Td>
              <TokenBar inputTokens={metric.input_tokens} outputTokens={metric.output_tokens} maxTotal={maxTotal} />
            </Table.Td>
            <Table.Td ta="right">
              <Text size="sm">{formatTokens(metric.input_tokens + metric.output_tokens)}</Text>
            </Table.Td>
            <Table.Td>
              <Group gap="xs">
                <DurationBar durationMs={metric.duration_ms} maxDuration={maxDuration} />
                <Text size="xs" c="dimmed" w={50}>
                  {formatDuration(metric.duration_ms)}
                </Text>
              </Group>
            </Table.Td>
            <Table.Td ta="right">
              <Text size="sm">{formatCost(metric.estimated_cost_usd)}</Text>
            </Table.Td>
          </Table.Tr>
        ))}
      </Table.Tbody>
    </Table>
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

const MetricsCharts = ({ session, comparison }) => {
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

  const { session: sessionData, metrics } = session;
  const totalInputTokens = metrics.reduce((acc, m) => acc + (m.input_tokens || 0), 0);
  const totalOutputTokens = metrics.reduce((acc, m) => acc + (m.output_tokens || 0), 0);
  const avgLatency = metrics.length > 0 ? Math.round(metrics.reduce((acc, m) => acc + (m.duration_ms || 0), 0) / metrics.length) : 0;

  return (
    <Stack gap="lg">
      <SimpleGrid cols={{ base: 2, md: 4 }}>
        <StatCard icon={Zap} label="Total Turns" value={metrics.length} />
        <StatCard icon={Cpu} label="Total Tokens" value={formatTokens(sessionData.total_tokens)} subvalue={`In: ${formatTokens(totalInputTokens)} / Out: ${formatTokens(totalOutputTokens)}`} />
        <StatCard icon={Clock} label="Avg Latency" value={formatDuration(avgLatency)} />
        <StatCard icon={DollarSign} label="Total Cost" value={formatCost(sessionData.total_cost)} />
      </SimpleGrid>

      <Paper withBorder p="md" radius="md">
        <Group mb="md" gap="xs">
          <MessageSquare size={16} />
          <Text fw={600}>Turn-by-Turn Metrics</Text>
        </Group>
        <SessionMetricsTable metrics={metrics} />
      </Paper>

      <Paper withBorder p="md" radius="md">
        <Text fw={600} mb="md">
          Legend
        </Text>
        <Group gap="lg">
          <Group gap="xs">
            <div className={styles.legendInput} />
            <Text size="sm">Input Tokens</Text>
          </Group>
          <Group gap="xs">
            <div className={styles.legendOutput} />
            <Text size="sm">Output Tokens</Text>
          </Group>
        </Group>
      </Paper>
    </Stack>
  );
};

export default MetricsCharts;
