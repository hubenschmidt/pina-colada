export const SET_SELECTED_PROJECT = "SET_SELECTED_PROJECT";
export const SET_PROJECTS = "SET_PROJECTS";

export default (initialState) => {
  return (state, action) => {
    switch (action.type) {
      case SET_SELECTED_PROJECT:
        return setSelectedProject(state, action.payload);
      case SET_PROJECTS:
        return setProjects(state, action.payload);
      default:
        return state;
    }
  };
};

const setSelectedProject = (state, payload) => {
  return {
    ...state,
    selectedProject: payload,
  };
};

const setProjects = (state, payload) => {
  return {
    ...state,
    projects: payload,
  };
};
