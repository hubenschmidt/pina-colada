"use client";

import { createContext, useReducer, useContext, useEffect } from "react";
import projectReducer, { SET_PROJECTS, SET_SELECTED_PROJECT } from "../reducers/projectReducer";
import { getProjects, setSelectedProject as setSelectedProjectApi } from "../api";
import { useUserContext } from "./userContext";

const initialState = {
  selectedProject: null,
  projects: [],
};

export const ProjectContext = createContext({
  projectState: initialState,
  dispatchProject: () => {},
  selectProject: () => {},
  refreshProjects: async () => {},
});

export const useProjectContext = () => useContext(ProjectContext);

export const ProjectProvider = ({ children }) => {
  const { userState } = useUserContext();
  const reducer = projectReducer(initialState);
  const [projectState, dispatchProject] = useReducer(reducer, initialState);

  const refreshProjects = async () => {
    if (!userState.isAuthed) return;
    const projects = await getProjects();
    dispatchProject({ type: SET_PROJECTS, payload: projects });
  };

  const selectProject = async (project) => {
    dispatchProject({ type: SET_SELECTED_PROJECT, payload: project });
    try {
      await setSelectedProjectApi(project?.id ?? null);
    } catch (error) {
      console.error("Failed to persist selected project:", error);
    }
  };

  useEffect(() => {
    if (!userState.isAuthed) return;

    const loadProjects = async () => {
      const projects = await getProjects();
      dispatchProject({ type: SET_PROJECTS, payload: projects });

      // Use selected_project_id from user state (from /auth/me response)
      const savedProjectId = userState.selectedProjectId;
      if (savedProjectId) {
        const savedProject = projects.find((p) => p.id === savedProjectId);
        if (savedProject) {
          dispatchProject({
            type: SET_SELECTED_PROJECT,
            payload: savedProject,
          });
        }
      }
    };

    loadProjects();
  }, [userState.isAuthed, userState.selectedProjectId]);

  return (
    <ProjectContext.Provider
      value={{ projectState, dispatchProject, selectProject, refreshProjects }}>
      {children}
    </ProjectContext.Provider>
  );
};
