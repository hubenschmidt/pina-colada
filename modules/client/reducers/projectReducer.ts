export const SET_SELECTED_PROJECT = "SET_SELECTED_PROJECT";
export const SET_PROJECTS = "SET_PROJECTS";

export interface Project {
  id: number;
  name: string;
  description: string | null;
  status: string | null;
  current_status_id: number | null;
  start_date: string | null;
  end_date: string | null;
  created_at: string | null;
  updated_at: string | null;
  deals_count?: number;
  leads_count?: number;
}

export interface ProjectState {
  selectedProject: Project | null;
  projects: Project[];
}

export default (initialState: ProjectState) => {
  return (state: ProjectState, action: { type: string; payload?: unknown }) => {
    switch (action.type) {
      case SET_SELECTED_PROJECT:
        return setSelectedProject(state, action.payload as Project | null);
      case SET_PROJECTS:
        return setProjects(state, action.payload as Project[]);
      default:
        return state;
    }
  };
};

const setSelectedProject = (
  state: ProjectState,
  payload: Project | null
): ProjectState => {
  return {
    ...state,
    selectedProject: payload,
  };
};

const setProjects = (
  state: ProjectState,
  payload: Project[]
): ProjectState => {
  return {
    ...state,
    projects: payload,
  };
};
