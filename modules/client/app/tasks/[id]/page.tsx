"use client";

import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import {
  Stack,
  Group,
  Text,
  Badge,
  Button,
  Card,
  Anchor,
  Center,
  Loader,
  TextInput,
  Textarea,
  Select,
  NumberInput,
} from "@mantine/core";
import { ArrowLeft } from "lucide-react";
import {
  getTask,
  updateTask,
  deleteTask,
  getTaskStatuses,
  getTaskPriorities,
  Task,
  TaskStatus,
  TaskComplexity,
} from "../../../api";
import { usePageLoading } from "../../../context/pageLoadingContext";

const COMPLEXITY_OPTIONS = [
  { value: "1", label: "1" },
  { value: "2", label: "2" },
  { value: "3", label: "3" },
  { value: "5", label: "5" },
  { value: "8", label: "8" },
  { value: "13", label: "13" },
  { value: "21", label: "21" },
];

const TaskDetailPage = () => {
  const params = useParams();
  const router = useRouter();
  const { dispatchPageLoading } = usePageLoading();
  const id = params.id as string;

  const [task, setTask] = useState<Task | null>(null);
  const [statuses, setStatuses] = useState<TaskStatus[]>([]);
  const [priorities, setPriorities] = useState<TaskStatus[]>([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Form state
  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [statusId, setStatusId] = useState<string | null>(null);
  const [priorityId, setPriorityId] = useState<string | null>(null);
  const [startDate, setStartDate] = useState("");
  const [dueDate, setDueDate] = useState("");
  const [estimatedHours, setEstimatedHours] = useState<number | string>("");
  const [actualHours, setActualHours] = useState<number | string>("");
  const [complexity, setComplexity] = useState<string | null>(null);

  useEffect(() => {
    const fetchData = async () => {
      try {
        const [taskData, statusData, priorityData] = await Promise.all([
          getTask(Number(id)),
          getTaskStatuses(),
          getTaskPriorities(),
        ]);
        setTask(taskData);
        setStatuses(statusData);
        setPriorities(priorityData);

        // Initialize form state
        setTitle(taskData.title);
        setDescription(taskData.description || "");
        setStatusId(taskData.status?.id?.toString() || null);
        setPriorityId(taskData.priority?.id?.toString() || null);
        setStartDate(taskData.start_date || "");
        setDueDate(taskData.due_date || "");
        setEstimatedHours(taskData.estimated_hours ?? "");
        setActualHours(taskData.actual_hours ?? "");
        setComplexity(taskData.complexity?.toString() || null);
      } catch (err) {
        setError(err instanceof Error ? err.message : "Failed to load task");
      } finally {
        setLoading(false);
        dispatchPageLoading({ type: "SET_PAGE_LOADING", payload: false });
      }
    };

    fetchData();
  }, [id, dispatchPageLoading]);

  const handleBack = () => {
    router.push("/tasks");
  };

  const handleSave = async () => {
    setSaving(true);
    setError(null);
    try {
      const updated = await updateTask(Number(id), {
        title,
        description: description || null,
        current_status_id: statusId ? Number(statusId) : null,
        priority_id: priorityId ? Number(priorityId) : null,
        start_date: startDate || null,
        due_date: dueDate || null,
        estimated_hours: estimatedHours !== "" ? Number(estimatedHours) : null,
        actual_hours: actualHours !== "" ? Number(actualHours) : null,
        complexity: complexity ? (Number(complexity) as TaskComplexity) : null,
      });
      setTask(updated);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to save task");
    } finally {
      setSaving(false);
    }
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

  if (error && !task) {
    return (
      <Stack p="md" gap="md">
        <Button
          variant="subtle"
          leftSection={<ArrowLeft size={16} />}
          onClick={handleBack}
        >
          Back to Tasks
        </Button>
        <Text c="red">{error}</Text>
      </Stack>
    );
  }

  if (!task) return null;

  const statusOptions = statuses.map((s) => ({
    value: s.id.toString(),
    label: s.name,
  }));

  const priorityOptions = priorities.map((p) => ({
    value: p.id.toString(),
    label: p.name,
  }));

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
        <Group>
          <Button color="lime" onClick={handleSave} loading={saving}>
            Save
          </Button>
          <Button color="red" variant="outline" onClick={handleDelete}>
            Delete
          </Button>
        </Group>
      </Group>

      {error && (
        <Text c="red" size="sm">
          {error}
        </Text>
      )}

      <Card withBorder p="lg">
        <Stack gap="md">
          <TextInput
            label="Title"
            value={title}
            onChange={(e) => setTitle(e.currentTarget.value)}
            required
          />

          <Textarea
            label="Description"
            value={description}
            onChange={(e) => setDescription(e.currentTarget.value)}
            rows={3}
          />

          <Group grow>
            <Select
              label="Status"
              data={statusOptions}
              value={statusId}
              onChange={setStatusId}
              clearable
            />

            <Select
              label="Priority"
              data={priorityOptions}
              value={priorityId}
              onChange={setPriorityId}
              clearable
            />

            <Select
              label="Complexity"
              data={COMPLEXITY_OPTIONS}
              value={complexity}
              onChange={setComplexity}
              clearable
            />
          </Group>

          <Group grow>
            <TextInput
              label="Start Date"
              type="date"
              value={startDate}
              onChange={(e) => setStartDate(e.currentTarget.value)}
            />

            <TextInput
              label="Due Date"
              type="date"
              value={dueDate}
              onChange={(e) => setDueDate(e.currentTarget.value)}
            />
          </Group>

          <Group grow>
            <NumberInput
              label="Estimated Hours"
              value={estimatedHours}
              onChange={setEstimatedHours}
              min={0}
              decimalScale={2}
            />

            <NumberInput
              label="Actual Hours"
              value={actualHours}
              onChange={setActualHours}
              min={0}
              decimalScale={2}
            />
          </Group>

          {task.entity.type && task.entity.display_name && (
            <Stack gap={4}>
              <Text size="xs" c="dimmed">
                Linked Entity
              </Text>
              <Group gap="xs">
                <Badge variant="light">{task.entity.type}</Badge>
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
              <Text size="xs" c="dimmed">
                Created
              </Text>
              <Text size="sm">
                {task.created_at ? task.created_at.slice(0, 10) : "—"}
              </Text>
            </Stack>

            {task.completed_at && (
              <Stack gap={4}>
                <Text size="xs" c="dimmed">
                  Completed
                </Text>
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
