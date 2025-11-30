"use client";

import { createContext, useReducer, useContext, useEffect } from "react";
import projectReducer, {
  ProjectState,
  Project,
  SET_PROJECTS,
  SET_SELECTED_PROJECT,
} from "../reducers/projectReducer";
import { getProjects } from "../api";
import { useUserContext } from "./userContext";

const initialState: ProjectState = {
  selectedProject: null,
  projects: [],
};

export const ProjectContext = createContext<{
  projectState: ProjectState;
  dispatchProject: React.Dispatch<{ type: string; payload?: unknown }>;
  selectProject: (project: Project | null) => void;
  refreshProjects: () => Promise<void>;
}>({
  projectState: initialState,
  dispatchProject: () => {},
  selectProject: () => {},
  refreshProjects: async () => {},
});

export const useProjectContext = () => useContext(ProjectContext);

export const ProjectProvider = ({
  children,
}: {
  children: React.ReactNode;
}) => {
  const { userState } = useUserContext();
  const reducer = projectReducer(initialState);
  const [projectState, dispatchProject] = useReducer(reducer, initialState);

  const refreshProjects = async () => {
    if (!userState.isAuthed) return;
    const projects = await getProjects();
    dispatchProject({ type: SET_PROJECTS, payload: projects });
  };

  const selectProject = (project: Project | null) => {
    dispatchProject({ type: SET_SELECTED_PROJECT, payload: project });
    const storageAction = project
      ? () => localStorage.setItem("selectedProjectId", String(project.id))
      : () => localStorage.removeItem("selectedProjectId");
    storageAction();
  };

  useEffect(() => {
    if (!userState.isAuthed) return;

    const loadProjects = async () => {
      const projects = await getProjects();
      dispatchProject({ type: SET_PROJECTS, payload: projects });

      const savedProjectId = localStorage.getItem("selectedProjectId");
      if (savedProjectId) {
        const savedProject = projects.find(
          (p: Project) => p.id === Number(savedProjectId)
        );
        if (savedProject) {
          dispatchProject({ type: SET_SELECTED_PROJECT, payload: savedProject });
        }
      }
    };

    loadProjects();
  }, [userState.isAuthed]);

  return (
    <ProjectContext.Provider
      value={{ projectState, dispatchProject, selectProject, refreshProjects }}
    >
      {children}
    </ProjectContext.Provider>
  );
};
