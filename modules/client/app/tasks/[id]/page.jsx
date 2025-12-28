"use client";

import { useEffect, useState, useContext } from "react";
import { useParams, useRouter } from "next/navigation";
import { Center, Stack, Loader, Text } from "@mantine/core";
import { usePageLoading } from "../../../context/pageLoadingContext";
import { ProjectContext } from "../../../context/projectContext";
import TaskForm from "../../../components/TaskForm/TaskForm";
import { getTask, updateTask, deleteTask } from "../../../api";

const TaskDetailPage = () => {
  const params = useParams();
  const router = useRouter();
  const { dispatchPageLoading } = usePageLoading();
  const { projectState } = useContext(ProjectContext);
  const selectedProject = projectState.selectedProject;
  const id = params.id;

  const [task, setTask] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    const fetchTask = async () => {
      const taskId = Number(id);
      if (!id || isNaN(taskId)) {
        setError("Invalid task ID");
        setLoading(false);
        dispatchPageLoading({ type: "SET_PAGE_LOADING", payload: false });
        return;
      }

      try {
        const data = await getTask(taskId);
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

  const handleClose = () => {
    router.push("/tasks");
  };

  const handleUpdate = async (taskId, updates) => {
    await updateTask(taskId, updates);
    router.push("/tasks");
  };

  const handleDelete = async (taskId) => {
    await deleteTask(taskId);
    router.push("/tasks");
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
      <div className="p-6">
        <p className="text-red-600 dark:text-red-400">{error || "Task not found"}</p>
      </div>
    );
  }

  return (
    <TaskForm
      onClose={handleClose}
      task={task}
      onUpdate={handleUpdate}
      onDelete={handleDelete}
      selectedProject={selectedProject}
    />
  );
};

export default TaskDetailPage;
