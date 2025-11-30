"use client";

import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import {
  Stack,
  Group,
  Title,
  Text,
  Badge,
  Button,
  Card,
  Anchor,
  Center,
  Loader,
} from "@mantine/core";
import { ArrowLeft } from "lucide-react";
import { getTask, deleteTask, Task } from "../../../api";
import { usePageLoading } from "../../../context/pageLoadingContext";

const TaskDetailPage = () => {
  const params = useParams();
  const router = useRouter();
  const { dispatchPageLoading } = usePageLoading();
  const id = params.id as string;

  const [task, setTask] = useState<Task | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchTask = async () => {
      try {
        const data = await getTask(Number(id));
        setTask(data);
      } catch (err) {
        setError(err instanceof Error ? err.message : "Failed to load task");
      } finally {
        setLoading(false);
        dispatchPageLoading({ type: "SET_PAGE_LOADING", payload: false });
      }
    };

    fetchTask();
  }, [id, dispatchPageLoading]);

  const handleBack = () => {
    router.push("/tasks");
  };

  const handleDelete = async () => {
    if (!confirm("Are you sure you want to delete this task?")) return;
    try {
      await deleteTask(Number(id));
      router.push("/tasks");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to delete task");
    }
  };

  if (loading) {
    return (
      <Center mih={400}>
        <Stack align="center" gap="md">
          <Loader size="xl" color="lime" />
          <Text c="dimmed">Loading task...</Text>
        </Stack>
      </Center>
    );
  }

  if (error || !task) {
    return (
      <Stack p="md" gap="md">
        <Button
          variant="subtle"
          leftSection={<ArrowLeft size={16} />}
          onClick={handleBack}
        >
          Back to Tasks
        </Button>
        <Text c="red">{error || "Task not found"}</Text>
      </Stack>
    );
  }

  return (
    <Stack gap="lg">
      <Group justify="space-between">
        <Button
          variant="subtle"
          leftSection={<ArrowLeft size={16} />}
          onClick={handleBack}
        >
          Back to Tasks
        </Button>
        <Button color="red" variant="outline" onClick={handleDelete}>
          Delete
        </Button>
      </Group>

      <Card withBorder p="lg">
        <Stack gap="md">
          <Title order={2}>{task.title}</Title>

          {task.description && (
            <Text c="dimmed">{task.description}</Text>
          )}

          <Group gap="xl">
            {task.status && (
              <Stack gap={4}>
                <Text size="xs" c="dimmed">Status</Text>
                <Badge variant="light">{task.status.name}</Badge>
              </Stack>
            )}

            {task.priority && (
              <Stack gap={4}>
                <Text size="xs" c="dimmed">Priority</Text>
                <Badge variant="outline">{task.priority.name}</Badge>
              </Stack>
            )}

            {task.due_date && (
              <Stack gap={4}>
                <Text size="xs" c="dimmed">Due Date</Text>
                <Text size="sm">{task.due_date}</Text>
              </Stack>
            )}
          </Group>

          {task.entity.type && task.entity.display_name && (
            <Stack gap={4}>
              <Text size="xs" c="dimmed">Linked Entity</Text>
              <Group gap="xs">
                <Text size="sm" c="dimmed">{task.entity.type}:</Text>
                {task.entity.url ? (
                  <Anchor href={task.entity.url} size="sm">
                    {task.entity.display_name}
                  </Anchor>
                ) : (
                  <Text size="sm">{task.entity.display_name}</Text>
                )}
              </Group>
            </Stack>
          )}

          <Group gap="xl">
            <Stack gap={4}>
              <Text size="xs" c="dimmed">Created</Text>
              <Text size="sm">{task.created_at ? task.created_at.slice(0, 10) : "—"}</Text>
            </Stack>

            {task.completed_at && (
              <Stack gap={4}>
                <Text size="xs" c="dimmed">Completed</Text>
                <Text size="sm">{task.completed_at.slice(0, 10)}</Text>
              </Stack>
            )}
          </Group>
        </Stack>
      </Card>
    </Stack>
  );
};

export default TaskDetailPage;
