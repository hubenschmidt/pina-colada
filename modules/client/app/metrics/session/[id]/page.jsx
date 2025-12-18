"use client";

import { useState, useEffect, useCallback } from "react";
import { useRouter, useParams } from "next/navigation";
import { useUserContext } from "../../../../context/userContext";
import { Stack, Paper, Text, Loader, Center, Group, Button, Badge } from "@mantine/core";
import { ArrowLeft, Clock } from "lucide-react";
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

const SessionDetailPage = () => {
  const { userState } = useUserContext();
  const router = useRouter();
  const params = useParams();
  const sessionId = params.id;
  const [session, setSession] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const { dispatchPageLoading } = usePageLoading();

  // Pagination state
  const [queryParams, setQueryParams] = useState({
    page: 1,
    pageSize: 10,
    sortBy: "started_at",
    sortDirection: "DESC",
  });

  const fetchSession = useCallback(async (fetchParams) => {
    if (!userState.isAuthed || !userState.roles?.includes("developer")) return;

    setLoading(true);
    try {
      const data = await getMetricSession(sessionId, fetchParams);
      setSession(data);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
      dispatchPageLoading({ type: "SET_PAGE_LOADING", payload: false });
    }
  }, [sessionId, userState.isAuthed, userState.roles, dispatchPageLoading]);

  useEffect(() => {
    if (!userState.isAuthed) {
      dispatchPageLoading({ type: "SET_PAGE_LOADING", payload: false });
      return;
    }
    if (!userState.roles?.includes("developer")) {
      router.push("/");
      return;
    }
    fetchSession(queryParams);
  }, [userState.isAuthed, userState.roles, router, dispatchPageLoading, fetchSession, queryParams]);

  const handleQueryChange = useCallback((newParams) => {
    setQueryParams((prev) => ({ ...prev, ...newParams }));
  }, []);

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
        </Group>
      </Paper>

      <MetricsCharts session={session} queryParams={queryParams} onQueryChange={handleQueryChange} />
    </Stack>
  );
};

export default SessionDetailPage;
