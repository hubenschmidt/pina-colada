"use client";

import { useState, useEffect } from "react";
import { useRouter, useParams } from "next/navigation";
import { useUserContext } from "../../../../context/userContext";
import { Stack, Paper, Text, Loader, Center, Group, Button, Badge } from "@mantine/core";
import { ArrowLeft, Clock, DollarSign, Zap } from "lucide-react";
import { usePageLoading } from "../../../../context/pageLoadingContext";
import { getMetricSession } from "../../../../api";
import MetricsCharts from "../../../../components/Metrics/MetricsCharts";

const formatDate = (dateStr) => {
  if (!dateStr) return "—";
  const date = new Date(dateStr);
  return date.toLocaleDateString("en-US", {
    month: "short",
    day: "numeric",
    year: "numeric",
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

const SessionDetailPage = () => {
  const { userState } = useUserContext();
  const router = useRouter();
  const params = useParams();
  const sessionId = params.id;
  const [session, setSession] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const { dispatchPageLoading } = usePageLoading();

  useEffect(() => {
    const fetchSession = async () => {
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
        const data = await getMetricSession(sessionId);
        setSession(data);
      } catch (err) {
        setError(err.message);
      } finally {
        setLoading(false);
        dispatchPageLoading({ type: "SET_PAGE_LOADING", payload: false });
      }
    };

    fetchSession();
  }, [userState.isAuthed, userState.roles, sessionId, dispatchPageLoading, router]);

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

  if (loading) {
    return (
      <Center mih={400}>
        <Stack align="center" gap="md">
          <Loader size="xl" color="lime" />
          <Text c="dimmed">Loading session details...</Text>
        </Stack>
      </Center>
    );
  }

  if (error) {
    return (
      <Center mih={400}>
        <Stack align="center" gap="md">
          <Text c="red">{error}</Text>
          <Button variant="subtle" onClick={() => router.push("/metrics")}>
            Back to Sessions
          </Button>
        </Stack>
      </Center>
    );
  }

  if (!session) {
    return (
      <Center mih={400}>
        <Stack align="center" gap="md">
          <Text c="dimmed">Session not found</Text>
          <Button variant="subtle" onClick={() => router.push("/metrics")}>
            Back to Sessions
          </Button>
        </Stack>
      </Center>
    );
  }

  return (
    <Stack gap="lg">
      <Group justify="space-between">
        <Group gap="sm">
          <Button
            variant="subtle"
            leftSection={<ArrowLeft size={16} />}
            onClick={() => router.push("/metrics")}>
            Back
          </Button>
          <h1 className="text-xl font-bold text-zinc-900 dark:text-zinc-100">Session Details</h1>
          {!session.session?.ended_at && (
            <Badge color="red" size="sm" variant="dot">
              Recording
            </Badge>
          )}
        </Group>
      </Group>

      <Paper withBorder p="md" radius="md">
        <Group grow>
          <Stack gap="xs">
            <Group gap="xs">
              <Clock size={16} />
              <Text size="sm" fw={500}>Started</Text>
            </Group>
            <Text size="sm" c="dimmed">{formatDate(session.session?.started_at)}</Text>
          </Stack>
          <Stack gap="xs">
            <Group gap="xs">
              <Clock size={16} />
              <Text size="sm" fw={500}>Duration</Text>
            </Group>
            <Text size="sm" c="dimmed">{formatDuration(session.session?.started_at, session.session?.ended_at)}</Text>
          </Stack>
          <Stack gap="xs">
            <Group gap="xs">
              <Zap size={16} />
              <Text size="sm" fw={500}>Turns</Text>
            </Group>
            <Text size="sm" c="dimmed">{session.metrics?.length || 0}</Text>
          </Stack>
          <Stack gap="xs">
            <Group gap="xs">
              <Zap size={16} />
              <Text size="sm" fw={500}>Total Tokens</Text>
            </Group>
            <Text size="sm" c="dimmed">{formatTokens(session.session?.total_tokens)}</Text>
          </Stack>
          <Stack gap="xs">
            <Group gap="xs">
              <DollarSign size={16} />
              <Text size="sm" fw={500}>Total Cost</Text>
            </Group>
            <Text size="sm" c="dimmed">{formatCost(session.session?.total_cost)}</Text>
          </Stack>
        </Group>
      </Paper>

      <MetricsCharts session={session} />
    </Stack>
  );
};

export default SessionDetailPage;
