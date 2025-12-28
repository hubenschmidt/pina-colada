"use client";

import { useEffect, useState } from "react";
import { useRouter, useParams } from "next/navigation";
import { Stack, Center, Loader, Text } from "@mantine/core";
import ProjectForm from "../../../components/ProjectForm/ProjectForm";
import { getProject, updateProject, deleteProject } from "../../../api";
import { useProjectContext } from "../../../context/projectContext";

const ProjectDetailPage = () => {
  const router = useRouter();
  const params = useParams();
  const projectId = Number(params.id);
  const { refreshProjects, projectState, selectProject } = useProjectContext();

  const [project, setProject] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    const fetchProject = async () => {
      try {
        const data = await getProject(projectId);
        setProject(data);
      } catch (err) {
        setError(err instanceof Error ? err.message : "Failed to load project");
      } finally {
        setLoading(false);
      }
    };

    fetchProject();
  }, [projectId]);

  const handleClose = () => {
    router.push("/projects");
  };

  const handleUpdate = async (data) => {
    const updated = await updateProject(projectId, data);
    await refreshProjects();
    if (projectState.selectedProject?.id === projectId) {
      selectProject(updated);
    }
    router.push("/projects");
    return updated;
  };

  const handleDelete = async () => {
    await deleteProject(projectId);
    await refreshProjects();
    if (projectState.selectedProject?.id === projectId) {
      selectProject(null);
    }
    router.push("/projects");
  };

  if (loading) {
    return (
      <Center mih={400}>
        <Stack align="center" gap="md">
          <Loader size="xl" color="lime" />
          <Text c="dimmed">Loading project...</Text>
        </Stack>
      </Center>
    );
  }

  if (error) {
    return (
      <Stack gap="lg">
        <h1 className="text-xl font-bold text-zinc-900 dark:text-zinc-100">Project</h1>
        <Text c="red">{error}</Text>
      </Stack>
    );
  }

  return (
    <ProjectForm
      project={project}
      onClose={handleClose}
      onUpdate={handleUpdate}
      onDelete={handleDelete}
    />
  );
};

export default ProjectDetailPage;
