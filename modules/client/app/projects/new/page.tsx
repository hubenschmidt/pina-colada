"use client";

import { useRouter } from "next/navigation";
import ProjectForm from "../../../components/ProjectForm/ProjectForm";
import { createProject } from "../../../api";
import { useProjectContext } from "../../../context/projectContext";

const NewProjectPage = () => {
  const router = useRouter();
  const { refreshProjects } = useProjectContext();

  const handleClose = () => {
    router.push("/projects");
  };

  const handleAdd = async (data: Parameters<typeof createProject>[0]) => {
    const created = await createProject(data);
    await refreshProjects();
    router.push("/projects");
    return created;
  };

  return <ProjectForm onClose={handleClose} onAdd={handleAdd} />;
};

export default NewProjectPage;
