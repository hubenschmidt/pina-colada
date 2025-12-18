"use client";

import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import { useUserContext } from "../../context/userContext";
import { Stack, Paper, Text, Loader, Center, Group, Badge, Button, Checkbox } from "@mantine/core";
import { BarChart2, Clock, Zap, GitCompare } from "lucide-react";
import { usePageLoading } from "../../context/pageLoadingContext";
import { getMetricSessions, compareMetricSessions } from "../../api";
import { DataTable } from "../../components/DataTable/DataTable";
import MetricsCharts from "../../components/Metrics/MetricsCharts";

const formatDate = (dateStr) => {
  if (!dateStr) return "—";
  const date = new Date(dateStr);
  return date.toLocaleDateString("en-US", {
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
};

const formatDuration = (startedAt, endedAt) => {
  if (!startedAt) return "—";
  const start = new Date(startedAt);
  const end = endedAt ? new Date(endedAt) : new Date();
  const durationMs = end - start;
  const minutes = Math.floor(durationMs / 60000);
  const seconds = Math.floor((durationMs % 60000) / 1000);
  if (minutes > 0) return `${minutes}m ${seconds}s`;
  return `${seconds}s`;
};

const formatCost = (amount) => {
  if (amount === null || amount === undefined) return "—";
  return `$${parseFloat(amount).toFixed(4)}`;
};

const formatTokens = (count) => {
  if (count === null || count === undefined) return "—";
  if (count >= 1_000_000) return `${(count / 1_000_000).toFixed(2)}M`;
  if (count >= 1_000) return `${(count / 1_000).toFixed(1)}K`;
  return count.toString();
};

const MetricsPage = () => {
  const { userState } = useUserContext();
  const router = useRouter();
  const [sessions, setSessions] = useState([]);
  const [selectedIds, setSelectedIds] = useState([]);
  const [loading, setLoading] = useState(true);
  const [comparisonData, setComparisonData] = useState(null);
  const [loadingComparison, setLoadingComparison] = useState(false);
  const { dispatchPageLoading } = usePageLoading();

  useEffect(() => {
    const fetchSessions = async () => {
      if (!userState.isAuthed) {
        dispatchPageLoading({ type: "SET_PAGE_LOADING", payload: false });
        return;
      }

      if (!userState.roles?.includes("developer")) {
        router.push("/");
        return;
      }

      setLoading(true);
      try {
        const data = await getMetricSessions(50);
        setSessions(data || []);
      } catch (error) {
        console.error("Failed to fetch sessions:", error);
      } finally {
        setLoading(false);
        dispatchPageLoading({ type: "SET_PAGE_LOADING", payload: false });
      }
    };

    fetchSessions();
  }, [userState.isAuthed, userState.roles, dispatchPageLoading, router]);

  const handleToggleSelect = (id) => {
    setSelectedIds((prev) => (prev.includes(id) ? prev.filter((i) => i !== id) : [...prev, id]));
  };

  const handleViewSession = (id) => {
    router.push(`/metrics/session/${id}`);
  };

  const handleCompare = async () => {
    if (selectedIds.length < 2) return;

    setLoadingComparison(true);
    try {
      const data = await compareMetricSessions(selectedIds);
      setComparisonData(data);
    } catch (error) {
      console.error("Failed to compare sessions:", error);
    } finally {
      setLoadingComparison(false);
    }
  };

  const handleBackToList = () => {
    setComparisonData(null);
  };

  if (!userState.isAuthed) {
    return (
      <Center mih={400}>
        <Text>Please log in to view metrics.</Text>
      </Center>
    );
  }

  if (!userState.roles?.includes("developer")) {
    return (
      <Center mih={400}>
        <Text>This feature is only available to developers.</Text>
      </Center>
    );
  }

  if (loading && sessions.length === 0) {
    return (
      <Center mih={400}>
        <Stack align="center" gap="md">
          <Loader size="xl" color="lime" />
          <Text c="dimmed">Loading recording sessions...</Text>
        </Stack>
      </Center>
    );
  }

  if (comparisonData) {
    return (
      <Stack gap="lg">
        <Group justify="space-between">
          <h1 className="text-xl font-bold text-zinc-900 dark:text-zinc-100">Session Comparison</h1>
          <Button variant="subtle" onClick={handleBackToList}>
            Back to Sessions
          </Button>
        </Group>
        {loadingComparison ? (
          <Center mih={300}>
            <Loader size="lg" />
          </Center>
        ) : (
          <MetricsCharts comparison={comparisonData} />
        )}
      </Stack>
    );
  }

  return (
    <Stack gap="lg">
      <Group justify="space-between">
        <Group gap="xs">
          <BarChart2 size={24} />
          <h1 className="text-xl font-bold text-zinc-900 dark:text-zinc-100">Agent Metrics</h1>
          <Badge color="grape" size="sm">
            Developer
          </Badge>
        </Group>
        <Button
          leftSection={<GitCompare size={16} />}
          disabled={selectedIds.length < 2}
          onClick={handleCompare}>
          Compare ({selectedIds.length})
        </Button>
      </Group>

      <DataTable
        data={{ items: sessions, currentPage: 1, totalPages: 1 }}
        columns={[
          {
            header: "",
            width: 40,
            render: (row) => (
              <Checkbox
                checked={selectedIds.includes(row.id)}
                onChange={(e) => {
                  e.stopPropagation();
                  handleToggleSelect(row.id);
                }}
                onClick={(e) => e.stopPropagation()}
              />
            ),
          },
          {
            header: "Started",
            render: (row) => (
              <Group gap="xs">
                <Text size="sm" fw={500}>
                  {formatDate(row.started_at)}
                </Text>
                {!row.ended_at && (
                  <Badge color="red" size="xs" variant="dot">
                    Recording
                  </Badge>
                )}
              </Group>
            ),
          },
          {
            header: "Turns",
            accessor: (row) => row.metric_count || 0,
            tdProps: { align: "center" },
            thProps: { align: "center" },
          },
          {
            header: "Duration",
            accessor: (row) => formatDuration(row.started_at, row.ended_at),
            tdProps: { align: "right" },
            thProps: { align: "right" },
          },
          {
            header: "Tokens",
            accessor: (row) => formatTokens(row.total_tokens),
            tdProps: { align: "right" },
            thProps: { align: "right" },
          },
          {
            header: "Cost",
            accessor: (row) => formatCost(row.total_cost),
            tdProps: { align: "right" },
            thProps: { align: "right" },
          },
        ]}
        rowKey={(row) => row.id}
        onRowClick={(row) => handleViewSession(row.id)}
        emptyText="No recording sessions yet. Click the record button on the chat page to start capturing metrics."
      />

      <Paper withBorder p="md" radius="md">
        <Stack gap="xs">
          <Group gap="xs">
            <Clock size={16} />
            <Text size="sm" fw={500}>
              Quick Stats
            </Text>
          </Group>
          <Group grow>
            <Paper p="sm" withBorder radius="sm">
              <Text size="xs" c="dimmed">
                Total Sessions
              </Text>
              <Text fw={600}>{sessions.length}</Text>
            </Paper>
            <Paper p="sm" withBorder radius="sm">
              <Text size="xs" c="dimmed">
                Active Recording
              </Text>
              <Text fw={600}>{sessions.filter((s) => !s.ended_at).length > 0 ? "Yes" : "No"}</Text>
            </Paper>
            <Paper p="sm" withBorder radius="sm">
              <Text size="xs" c="dimmed">
                Total Metrics
              </Text>
              <Text fw={600}>{sessions.reduce((acc, s) => acc + (s.metric_count || 0), 0)}</Text>
            </Paper>
          </Group>
        </Stack>
      </Paper>
    </Stack>
  );
};

export default MetricsPage;
