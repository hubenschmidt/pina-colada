"use client";

import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import Link from "next/link";
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
import CommentsSection from "../../../components/CommentsSection/CommentsSection";
import Timestamps from "../../../components/Timestamps/Timestamps";
import { usePageLoading } from "../../../context/pageLoadingContext";
import { DeleteConfirmBanner } from "../../../components/DeleteConfirmBanner";

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
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);
  const [deleting, setDeleting] = useState(false);

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

  const handleDeleteClick = () => {
    setShowDeleteConfirm(true);
  };

  const handleDeleteConfirm = async () => {
    setDeleting(true);
    try {
      await deleteTask(Number(id));
      router.push("/tasks");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to delete task");
      setDeleting(false);
      setShowDeleteConfirm(false);
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

  const getEntityColor = (entityType: string): string => {
    const colorMap: Record<string, string> = {
      Organization: "blue",
      Individual: "green",
      Project: "violet",
      Contact: "cyan",
      Lead: "orange",
    };
    return colorMap[entityType] || "gray";
  };

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
          <Button
            color="red"
            variant="outline"
            onClick={handleDeleteClick}
            disabled={showDeleteConfirm}
            loading={deleting}
          >
            Delete
          </Button>
        </Group>
      </Group>

      {showDeleteConfirm && (
        <DeleteConfirmBanner
          itemName={task.title}
          onConfirm={handleDeleteConfirm}
          onCancel={() => setShowDeleteConfirm(false)}
          loading={deleting}
        />
      )}

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
                Linked To
              </Text>
              <Group gap="xs">
                {task.entity.url ? (
                  <Anchor component={Link} href={task.entity.url} size="sm">
                    <Badge
                      size="sm"
                      variant="light"
                      color={getEntityColor(task.entity.type)}
                      style={{ cursor: "pointer" }}
                    >
                      {task.entity.display_name}
                    </Badge>
                  </Anchor>
                ) : (
                  <Badge
                    size="sm"
                    variant="light"
                    color={getEntityColor(task.entity.type)}
                  >
                    {task.entity.display_name}
                  </Badge>
                )}
              </Group>
            </Stack>
          )}

          <Timestamps createdAt={task.created_at} updatedAt={task.updated_at} />
        </Stack>
      </Card>

      <Card withBorder p="lg">
        <CommentsSection entityType="Task" entityId={Number(id)} />
      </Card>
    </Stack>
  );
};

export default TaskDetailPage;
