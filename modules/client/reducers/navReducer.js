export const SET_AGENT_OPEN = "SET_AGENT_OPEN";
export const SET_SIDEBAR_COLLAPSED = "SET_SIDEBAR_COLLAPSED";
export const SET_SIDEBAR_SECTION_EXPANDED = "SET_SIDEBAR_SECTION_EXPANDED";

export default (initialState) => {
  return (state, action) => {
    switch (action.type) {
      case SET_AGENT_OPEN:
        return setAgentOpen(state, action.payload);
      case SET_SIDEBAR_COLLAPSED:
        return setSidebarCollapsed(state, action.payload);
      case SET_SIDEBAR_SECTION_EXPANDED:
        return setSidebarSectionExpanded(state, action.payload);
      default:
        return state;
    }
  };
};

const setAgentOpen = (state, payload) => {
  return {
    ...state,
    agentOpen: payload,
  };
};

const setSidebarCollapsed = (state, payload) => {
  return {
    ...state,
    sidebarCollapsed: payload,
  };
};

const setSidebarSectionExpanded = (state, payload) => {
  return {
    ...state,
    sidebarSections: {
      ...state.sidebarSections,
      [payload.section]: payload.expanded,
    },
  };
};
