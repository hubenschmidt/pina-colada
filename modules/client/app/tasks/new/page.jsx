"use client";

import { useRouter } from "next/navigation";
import { useContext, useEffect } from "react";
import TaskForm from "../../../components/TaskForm/TaskForm";
import { createTask } from "../../../api";
import { ProjectContext } from "../../../context/projectContext";
import { usePageLoading } from "../../../context/pageLoadingContext";

const NewTaskPage = () => {
  const router = useRouter();
  const { projectState } = useContext(ProjectContext);
  const selectedProject = projectState.selectedProject;
  const { dispatchPageLoading } = usePageLoading();

  useEffect(() => {
    dispatchPageLoading({ type: "SET_PAGE_LOADING", payload: false });
  }, [dispatchPageLoading]);

  const handleClose = () => {
    router.push("/tasks");
  };

  const handleAdd = async (data) => {
    const task = await createTask(data);
    router.push("/tasks");
    return task;
  };

  return <TaskForm onClose={handleClose} onAdd={handleAdd} selectedProject={selectedProject} />;
};

export default NewTaskPage;
